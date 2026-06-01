package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/attribute"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/observability"
)

// VinylHandler holds handler dependencies
type VinylHandler struct {
	service      domain.VinylService
	trackService domain.TrackService
	profileRepo  domain.ProfileRepository
	jwtService   *auth.JWTService
	adminUser    string
	adminPass    string
}

// NewVinylHandler creates a handler without auth (for testing)
func NewVinylHandler(service domain.VinylService) *VinylHandler {
	return &VinylHandler{service: service}
}

// NewVinylHandlerWithAuth creates a handler with JWT auth
func NewVinylHandlerWithAuth(
	service domain.VinylService,
	trackSvc domain.TrackService,
	profileRepo domain.ProfileRepository,
	jwtSvc *auth.JWTService,
	user, pass string,
) *VinylHandler {
	return &VinylHandler{
		service:      service,
		trackService: trackSvc,
		profileRepo:  profileRepo,
		jwtService:   jwtSvc,
		adminUser:    user,
		adminPass:    pass,
	}
}

// Login godoc
// @Summary      Autenticar usuário
// @Description  Gera um token JWT para acesso à API
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Credenciais"
// @Success      200 {object} TokenResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /auth/login [post]
func (h *VinylHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if req.Username != h.adminUser || req.Password != h.adminPass {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
		return
	}
	token, err := h.jwtService.GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, TokenResponse{Token: token})
}

// ListVinyls godoc
// @Summary      Listar discos
// @Description  Retorna todos os discos de vinil cadastrados
// @Tags         vinyls
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array}  VinylResponse
// @Failure      500 {object} ErrorResponse
// @Router       /vinyls [get]
func (h *VinylHandler) ListVinyls(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.ListVinyls")
	defer span.End()

	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/vinyls"),
		semconv.HTTPURL(c.Request.URL.String()),
	)

	vinyls, err := h.service.ListVinyls(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusInternalServerError))
		logWithTrace(ctx, "failed to list vinyls", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
	resp := make([]VinylResponse, len(vinyls))
	for i := range vinyls {
		resp[i] = toVinylResponse(&vinyls[i])
	}
	c.JSON(http.StatusOK, resp)
}

// GetVinyl godoc
// @Summary      Buscar disco por ID
// @Description  Retorna um disco de vinil pelo seu ID
// @Tags         vinyls
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Vinyl ID"
// @Success      200 {object} VinylResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /vinyls/{id} [get]
func (h *VinylHandler) GetVinyl(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.GetVinyl")
	defer span.End()

	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/vinyls/:id"),
		semconv.HTTPURL(c.Request.URL.String()),
	)

	id := c.Param("id")
	vinyl, err := h.service.GetVinyl(ctx, id)
	if errors.Is(err, domain.ErrVinylNotFound) {
		span.SetStatus(codes.Error, "vinyl not found")
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusNotFound))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
		return
	}
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
	c.JSON(http.StatusOK, toVinylResponse(vinyl))
}

// CreateVinyl godoc
// @Summary      Criar disco
// @Description  Cadastra um novo disco de vinil
// @Tags         vinyls
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateVinylRequest true "Dados do disco"
// @Success      201 {object} VinylResponse
// @Failure      400 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /vinyls [post]
func (h *VinylHandler) CreateVinyl(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.CreateVinyl")
	defer span.End()

	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/vinyls"),
		semconv.HTTPURL(c.Request.URL.String()),
	)

	var req CreateVinylRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	now := time.Now()
	vinyl := &domain.Vinyl{
		ID:          uuid.NewString(),
		Title:       req.Title,
		Artist:      req.Artist,
		Year:        req.Year,
		Genre:       req.Genre,
		Label:       req.Label,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := vinyl.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.service.CreateVinyl(ctx, vinyl); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusCreated))
	c.JSON(http.StatusCreated, toVinylResponse(vinyl))
}

// UpdateVinyl godoc
// @Summary      Atualizar disco
// @Description  Atualiza os dados de um disco de vinil existente
// @Tags         vinyls
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string             true "Vinyl ID"
// @Param        request body UpdateVinylRequest true "Dados atualizados"
// @Success      200 {object} VinylResponse
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /vinyls/{id} [put]
func (h *VinylHandler) UpdateVinyl(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.UpdateVinyl")
	defer span.End()

	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/vinyls/:id"),
		semconv.HTTPURL(c.Request.URL.String()),
	)

	id := c.Param("id")
	var req UpdateVinylRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	existing, err := h.service.GetVinyl(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrVinylNotFound) {
			span.SetStatus(codes.Error, "vinyl not found")
			span.SetAttributes(semconv.HTTPStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
			return
		}
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	vinyl := &domain.Vinyl{
		ID:          id,
		Title:       req.Title,
		Artist:      req.Artist,
		Year:        req.Year,
		Genre:       req.Genre,
		Label:       req.Label,
		Description: req.Description,
		CoverURL:    existing.CoverURL,
		UpdatedAt:   time.Now(),
	}
	if err := vinyl.Validate(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusBadRequest))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.service.UpdateVinyl(ctx, vinyl); err != nil {
		if errors.Is(err, domain.ErrVinylNotFound) {
			span.SetStatus(codes.Error, "vinyl not found")
			span.SetAttributes(semconv.HTTPStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
			return
		}
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
	c.JSON(http.StatusOK, toVinylResponse(vinyl))
}

// DeleteVinyl godoc
// @Summary      Remover disco
// @Description  Remove um disco de vinil pelo seu ID
// @Tags         vinyls
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Vinyl ID"
// @Success      204
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /vinyls/{id} [delete]
func (h *VinylHandler) DeleteVinyl(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.DeleteVinyl")
	defer span.End()

	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/vinyls/:id"),
		semconv.HTTPURL(c.Request.URL.String()),
	)

	id := c.Param("id")
	if err := h.service.DeleteVinyl(ctx, id); err != nil {
		if errors.Is(err, domain.ErrVinylNotFound) {
			span.SetStatus(codes.Error, "vinyl not found")
			span.SetAttributes(semconv.HTTPStatusCode(http.StatusNotFound))
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
			return
		}
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(semconv.HTTPStatusCode(http.StatusInternalServerError))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusNoContent))
	c.Status(http.StatusNoContent)
}

// UploadCover godoc
// @Summary      Upload capa do disco
// @Description  Faz upload de uma imagem e atualiza o cover_url do vinil
// @Tags         vinyls
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "Vinyl ID"
// @Param        file    formData  file    true  "Imagem da capa (jpg, png, webp)"
// @Success      200 {object} VinylResponse
// @Failure      400 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /vinyls/{id}/cover [post]
func (h *VinylHandler) UploadCover(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.UploadCover")
	defer span.End()

	id := c.Param("id")

	vinyl, err := h.service.GetVinyl(ctx, id)
	if errors.Is(err, domain.ErrVinylNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	span.SetAttributes(attribute.String("file.extension", ext))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[strings.ToLower(ext)] {
		observability.RecordCoverUpload("error", ext)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only jpg, png and webp are allowed"})
		return
	}

	filename := uuid.NewString() + ext
	dst := filepath.Join("uploads", filename)

	if err := os.MkdirAll("uploads", 0755); err != nil {
		observability.RecordCoverUpload("error", ext)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not create upload dir"})
		return
	}

	out, err := os.Create(dst)
	if err != nil {
		observability.RecordCoverUpload("error", ext)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not save file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		observability.RecordCoverUpload("error", ext)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not write file"})
		return
	}

	span.SetAttributes(attribute.Int64("file.size_bytes", header.Size))
	observability.RecordCoverUpload("success", ext)

	vinyl.CoverURL = "/uploads/" + filename
	vinyl.UpdatedAt = time.Now()
	if err := h.service.UpdateVinyl(ctx, vinyl); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not update vinyl"})
		return
	}

	c.JSON(http.StatusOK, toVinylResponse(vinyl))
}

// ListTracks godoc
// @Summary      Listar faixas do disco
// @Description  Retorna todas as faixas de um disco ordenadas por posição
// @Tags         tracks
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Vinyl ID"
// @Success      200  {array}   TrackResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /vinyls/{id}/tracks [get]
func (h *VinylHandler) ListTracks(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.ListTracks")
	defer span.End()
	vinylID := c.Param("id")
	tracks, err := h.trackService.ListTracks(ctx, vinylID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	resp := make([]TrackResponse, len(tracks))
	for i := range tracks {
		resp[i] = toTrackResponse(&tracks[i])
	}
	c.JSON(http.StatusOK, resp)
}

// CreateTrack godoc
// @Summary      Criar faixa
// @Description  Adiciona uma nova faixa ao disco
// @Tags         tracks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string        true  "Vinyl ID"
// @Param        request  body      TrackRequest  true  "Dados da faixa"
// @Success      201  {object}  TrackResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /vinyls/{id}/tracks [post]
func (h *VinylHandler) CreateTrack(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.CreateTrack")
	defer span.End()
	vinylID := c.Param("id")
	var req TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if _, err := h.service.GetVinyl(ctx, vinylID); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
		return
	}
	track := &domain.Track{
		ID:        uuid.NewString(),
		VinylID:   vinylID,
		Title:     req.Title,
		Position:  req.Position,
		Lyrics:    req.Lyrics,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := track.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.trackService.CreateTrack(ctx, track); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, toTrackResponse(track))
}

// UpdateTrack godoc
// @Summary      Atualizar faixa
// @Description  Atualiza título, posição e letra de uma faixa
// @Tags         tracks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id        path      string        true  "Vinyl ID"
// @Param        track_id  path      string        true  "Track ID"
// @Param        request   body      TrackRequest  true  "Dados da faixa"
// @Success      200  {object}  TrackResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /vinyls/{id}/tracks/{track_id} [put]
func (h *VinylHandler) UpdateTrack(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.UpdateTrack")
	defer span.End()
	trackID := c.Param("track_id")
	var req TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	track := &domain.Track{
		ID:        trackID,
		VinylID:   c.Param("id"),
		Title:     req.Title,
		Position:  req.Position,
		Lyrics:    req.Lyrics,
		UpdatedAt: time.Now(),
	}
	if err := h.trackService.UpdateTrack(ctx, track); err != nil {
		if errors.Is(err, domain.ErrTrackNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "track not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toTrackResponse(track))
}

// DeleteTrack godoc
// @Summary      Remover faixa
// @Description  Remove uma faixa do disco
// @Tags         tracks
// @Produce      json
// @Security     BearerAuth
// @Param        id        path  string  true  "Vinyl ID"
// @Param        track_id  path  string  true  "Track ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /vinyls/{id}/tracks/{track_id} [delete]
func (h *VinylHandler) DeleteTrack(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.DeleteTrack")
	defer span.End()
	trackID := c.Param("track_id")
	if err := h.trackService.DeleteTrack(ctx, trackID); err != nil {
		if errors.Is(err, domain.ErrTrackNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "track not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetProfile godoc
// @Summary      Perfil público
// @Description  Retorna o perfil público do catálogo (sem autenticação)
// @Tags         profile
// @Produce      json
// @Success      200  {object}  ProfileResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /profile [get]
func (h *VinylHandler) GetProfile(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.GetProfile")
	defer span.End()
	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/profile"),
	)

	profile, err := h.profileRepo.Get(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		logWithTrace(ctx, "failed to get profile", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	// Resolve favorite vinyl details with tracks (public, no auth needed)
	favoriteVinyls := make([]FavoriteVinylWithTracks, 0, len(profile.FavoriteVinylIDs))
	for _, id := range profile.FavoriteVinylIDs {
		v, err := h.service.GetVinyl(ctx, id)
		if err != nil {
			continue // vinyl may have been deleted
		}
		tracks, err := h.trackService.ListTracks(ctx, id)
		trackResponses := make([]TrackResponse, 0, len(tracks))
		if err == nil {
			for i := range tracks {
				trackResponses = append(trackResponses, toTrackResponse(&tracks[i]))
			}
		}
		favoriteVinyls = append(favoriteVinyls, FavoriteVinylWithTracks{
			VinylResponse: toVinylResponse(v),
			Tracks:        trackResponses,
		})
	}
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
	c.JSON(http.StatusOK, ProfileResponse{
		Username:         profile.Username,
		DisplayName:      profile.DisplayName,
		Bio:              profile.Bio,
		PhotoURL:         profile.PhotoURL,
		Links:            profile.Links,
		PreferredGenres:  profile.PreferredGenres,
		FavoriteVinylIDs: profile.FavoriteVinylIDs,
		FavoriteVinyls:   favoriteVinyls,
	})
}

// UpdateProfile godoc
// @Summary      Atualizar perfil
// @Description  Atualiza nome, bio, links e gêneros preferidos do perfil
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      UpdateProfileRequest  true  "Dados do perfil"
// @Success      200  {object}  ProfileResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /profile [put]
func (h *VinylHandler) UpdateProfile(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.UpdateProfile")
	defer span.End()
	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/profile"),
	)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.SetStatus(codes.Error, err.Error())
		logWithTrace(ctx, "failed to update profile", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	current, err := h.profileRepo.Get(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		logWithTrace(ctx, "failed to update profile", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	if current.ID == "" {
		current.ID = uuid.NewString()
		current.Username = h.adminUser
	}
	current.DisplayName = req.DisplayName
	current.Bio = req.Bio
	current.Links = req.Links
	current.PreferredGenres = req.PreferredGenres
	if err := h.profileRepo.Upsert(ctx, current); err != nil {
		span.SetStatus(codes.Error, err.Error())
		logWithTrace(ctx, "failed to update profile", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not save profile"})
		return
	}
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
	c.JSON(http.StatusOK, ProfileResponse{
		Username:         current.Username,
		DisplayName:      current.DisplayName,
		Bio:              current.Bio,
		PhotoURL:         current.PhotoURL,
		Links:            current.Links,
		PreferredGenres:  current.PreferredGenres,
		FavoriteVinylIDs: current.FavoriteVinylIDs,
		FavoriteVinyls:   []FavoriteVinylWithTracks{},
	})
}

// UploadProfilePhoto godoc
// @Summary      Upload foto do perfil
// @Description  Faz upload de uma imagem de perfil (jpg, png, webp)
// @Tags         profile
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Imagem do perfil"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /profile/photo [post]
func (h *VinylHandler) UploadProfilePhoto(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.UploadProfilePhoto")
	defer span.End()
	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/profile/photo"),
	)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		observability.RecordProfilePhotoUpload("error")
		span.SetStatus(codes.Error, "upload failed")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	span.SetAttributes(attribute.String("file.extension", ext))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[strings.ToLower(ext)] {
		observability.RecordProfilePhotoUpload("error")
		span.SetStatus(codes.Error, "upload failed")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only jpg, png and webp are allowed"})
		return
	}

	filename := "profile-" + uuid.NewString() + ext
	dst := filepath.Join("uploads", filename)
	if err := os.MkdirAll("uploads", 0755); err != nil {
		observability.RecordProfilePhotoUpload("error")
		span.SetStatus(codes.Error, "upload failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not create upload dir"})
		return
	}
	out, err := os.Create(dst)
	if err != nil {
		observability.RecordProfilePhotoUpload("error")
		span.SetStatus(codes.Error, "upload failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not save file"})
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		observability.RecordProfilePhotoUpload("error")
		span.SetStatus(codes.Error, "upload failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not write file"})
		return
	}

	current, _ := h.profileRepo.Get(ctx)
	if current.ID == "" {
		current.ID = uuid.NewString()
		current.Username = h.adminUser
	}
	current.PhotoURL = "/uploads/" + filename
	_ = h.profileRepo.Upsert(ctx, current)

	observability.RecordProfilePhotoUpload("success")
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"photo_url": current.PhotoURL})
}

// ToggleFavorite godoc
// @Summary      Favoritar/desfavoritar disco
// @Description  Adiciona ou remove um disco dos favoritos do perfil
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "ID do disco"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /vinyls/{id}/favorite [post]
func (h *VinylHandler) ToggleFavorite(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.ToggleFavorite")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/vinyls/:id/favorite"),
		attribute.String("vinyl.id", id),
	)
	// Ensure vinyl exists
	if _, err := h.service.GetVinyl(ctx, id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "vinyl not found"})
		return
	}
	profile, err := h.profileRepo.Get(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	if profile.ID == "" {
		profile.ID = uuid.NewString()
		profile.Username = h.adminUser
	}
	if profile.FavoriteVinylIDs == nil {
		profile.FavoriteVinylIDs = []string{}
	}

	favorited := false
	newFavs := make([]string, 0, len(profile.FavoriteVinylIDs))
	for _, fid := range profile.FavoriteVinylIDs {
		if fid == id {
			favorited = true // will be removed
		} else {
			newFavs = append(newFavs, fid)
		}
	}
	if !favorited {
		newFavs = append(newFavs, id)
	}
	profile.FavoriteVinylIDs = newFavs

	if err := h.profileRepo.Upsert(ctx, profile); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not save profile"})
		return
	}
	action := "add"
	if favorited {
		action = "remove"
	}
	span.SetAttributes(attribute.String("favorite.action", action))
	observability.RecordFavoriteToggle(action)
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
	c.JSON(http.StatusOK, gin.H{"favorited": !favorited, "favorite_vinyl_ids": newFavs})
}

// logWithTrace logs an error with trace_id and span_id extracted from ctx.
func logWithTrace(ctx context.Context, msg string, err error) {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	slog.ErrorContext(ctx, msg, "error", err, "trace_id", spanCtx.TraceID().String(), "span_id", spanCtx.SpanID().String())
}
