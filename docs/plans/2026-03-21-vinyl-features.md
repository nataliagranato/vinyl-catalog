# Vinyl Features Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Adicionar descrição e capa ao vinil, lista de músicas com letras, e perfil público compartilhável com foto, bio, links e gêneros favoritos.

**Architecture:**
- Backend (Go): novos campos em `Vinyl` (description, cover_url), nova entidade `Track` (vinyl_id, title, position, lyrics), nova entidade `Profile` (singleton por instância). Uploads de imagem salvos em volume Docker `/app/uploads/`, servidos como arquivos estáticos em `/uploads/*`. Perfil público acessível sem autenticação em `GET /api/v1/profile`.
- Frontend (Next.js 16): VinylCard mostra thumbnail da capa; detail page ganha textarea de descrição, upload de capa, lista de faixas com letras; nova página pública `/profile` e `/profile/edit`.

**Tech Stack:** Go + Gin + GORM + PostgreSQL (backend), Next.js 16 + Tailwind v4 + Framer Motion (frontend), multipart/form-data upload, Docker volume para arquivos.

---

## Task 1: Adicionar `description` e `cover_url` ao domínio Vinyl

**Files:**
- Modify: `internal/domain/vinyl.go`
- Modify: `internal/domain/vinyl_test.go`

**Step 1: Atualizar `Vinyl` struct**

Em `internal/domain/vinyl.go`, adicionar dois campos após `Label`:
```go
Description string    `json:"description"`
CoverURL    string    `json:"cover_url"`
```

O construtor `NewVinyl` não precisa aceitar esses campos — são opcionais e editados separadamente.

**Step 2: Verificar que os testes existentes ainda passam**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
go test ./internal/domain/... -v
```
Expected: todos os testes existentes passam (os novos campos são opcionais, não quebram validação).

**Step 3: Commit**

```bash
git add internal/domain/vinyl.go
git commit -m "feat(domain): add description and cover_url to Vinyl entity"
```

---

## Task 2: Atualizar repositório PostgreSQL e DTOs HTTP para Vinyl

**Files:**
- Modify: `internal/adapters/secondary/postgres/vinyl_repo.go`
- Modify: `internal/adapters/primary/http/dto.go`
- Modify: `internal/adapters/primary/http/handler.go`

**Step 1: Adicionar campos ao `vinylModel` e às funções de mapeamento**

Em `internal/adapters/secondary/postgres/vinyl_repo.go`:

Na struct `vinylModel`, adicionar após `Label`:
```go
Description string
CoverURL    string `gorm:"column:cover_url"`
```

Na função `toModel`, adicionar:
```go
Description: v.Description,
CoverURL:    v.CoverURL,
```

Na função `toDomain`, adicionar:
```go
Description: m.Description,
CoverURL:    m.CoverURL,
```

AutoMigrate adicionará as colunas automaticamente na próxima inicialização.

**Step 2: Atualizar DTOs**

Em `internal/adapters/primary/http/dto.go`:

Adicionar `Description` e `CoverURL` a `CreateVinylRequest` e `UpdateVinylRequest`:
```go
Description string `json:"description" example:"Modal jazz masterpiece recorded in two sessions."`
```

Adicionar ao `VinylResponse`:
```go
Description string `json:"description" example:"Modal jazz masterpiece."`
CoverURL    string `json:"cover_url"   example:"/uploads/abc123.jpg"`
```

Atualizar `toVinylResponse` para incluir os novos campos:
```go
Description: v.Description,
CoverURL:    v.CoverURL,
```

**Step 3: Atualizar handler `CreateVinyl` e `UpdateVinyl`**

Em `internal/adapters/primary/http/handler.go`, no handler `CreateVinyl`, adicionar ao construir o `vinyl`:
```go
Description: req.Description,
```

No handler `UpdateVinyl`, adicionar ao construir o `vinyl`:
```go
Description: req.Description,
```

`CoverURL` é atualizado apenas pelo endpoint de upload (Task 3), não pelo PUT geral.

**Step 4: Verificar compilação**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog && go build ./...
```
Expected: sem erros.

**Step 5: Rodar testes existentes**

```bash
go test ./internal/adapters/... ./internal/application/... -v
```
Expected: todos passam.

**Step 6: Commit**

```bash
git add internal/adapters/
git commit -m "feat(api): add description and cover_url fields to vinyl"
```

---

## Task 3: Upload de imagem — endpoint e servidor de arquivos estáticos

**Files:**
- Modify: `internal/adapters/primary/http/handler.go`
- Modify: `internal/adapters/primary/http/router.go`
- Modify: `docker-compose.yml`

**Step 1: Criar diretório de uploads**

```bash
mkdir -p /Users/natalia.granato/Downloads/vinyl-catalog/uploads
echo "# uploads placeholder" > /Users/natalia.granato/Downloads/vinyl-catalog/uploads/.gitkeep
```

**Step 2: Adicionar handler de upload em `handler.go`**

Adicionar ao final de `internal/adapters/primary/http/handler.go`:

```go
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
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[strings.ToLower(ext)] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only jpg, png and webp are allowed"})
		return
	}

	filename := uuid.NewString() + ext
	dst := filepath.Join("uploads", filename)

	if err := os.MkdirAll("uploads", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not create upload dir"})
		return
	}

	out, err := os.Create(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not save file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not write file"})
		return
	}

	vinyl.CoverURL = "/uploads/" + filename
	vinyl.UpdatedAt = time.Now()
	if err := h.service.UpdateVinyl(ctx, vinyl); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not update vinyl"})
		return
	}

	c.JSON(http.StatusOK, toVinylResponse(vinyl))
}
```

Adicionar os imports necessários ao topo do arquivo (se não existirem):
```go
"io"
"os"
"path/filepath"
"strings"
```

**Step 3: Registrar rota e servidor estático em `router.go`**

Em `internal/adapters/primary/http/router.go`, dentro de `NewRouter`, após as rotas de vinyls:

```go
// Upload de capa (protegido)
vinyls.POST("/:id/cover", handler.UploadCover)

// Servir arquivos de upload (público)
r.Static("/uploads", "./uploads")
```

**Step 4: Montar volume no docker-compose**

Em `docker-compose.yml`, no serviço `app`, adicionar volume:
```yaml
    volumes:
      - uploads_data:/app/uploads
```

No final do arquivo, em `volumes:`, adicionar:
```yaml
  uploads_data:
```

**Step 5: Verificar compilação**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog && go build ./...
```
Expected: sem erros.

**Step 6: Commit**

```bash
git add internal/adapters/primary/http/ docker-compose.yml uploads/.gitkeep
git commit -m "feat(api): add cover image upload endpoint and static file server"
```

---

## Task 4: Entidade Track — domínio, repositório e serviço

**Files:**
- Create: `internal/domain/track.go`
- Create: `internal/application/track_service.go`
- Create: `internal/adapters/secondary/postgres/track_repo.go`
- Modify: `internal/domain/errors.go`

**Step 1: Adicionar erro de Track**

Em `internal/domain/errors.go`, adicionar:
```go
ErrTrackNotFound = errors.New("track not found")
```

**Step 2: Criar `internal/domain/track.go`**

```go
package domain

import (
	"context"
	"errors"
	"time"
)

type Track struct {
	ID        string    `json:"id"`
	VinylID   string    `json:"vinyl_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`
	Lyrics    string    `json:"lyrics"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *Track) Validate() error {
	if t.Title == "" {
		return errors.New("track title is required")
	}
	if t.VinylID == "" {
		return errors.New("track vinyl_id is required")
	}
	return nil
}

type TrackRepository interface {
	Create(ctx context.Context, t *Track) error
	FindByVinylID(ctx context.Context, vinylID string) ([]Track, error)
	FindByID(ctx context.Context, id string) (*Track, error)
	Update(ctx context.Context, t *Track) error
	Delete(ctx context.Context, id string) error
}

type TrackService interface {
	CreateTrack(ctx context.Context, t *Track) error
	ListTracks(ctx context.Context, vinylID string) ([]Track, error)
	UpdateTrack(ctx context.Context, t *Track) error
	DeleteTrack(ctx context.Context, id string) error
}
```

**Step 3: Criar `internal/adapters/secondary/postgres/track_repo.go`**

```go
package postgres

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"gorm.io/gorm"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type trackModel struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	VinylID   string    `gorm:"not null;index"`
	Title     string    `gorm:"not null"`
	Position  int
	Lyrics    string    `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (trackModel) TableName() string { return "tracks" }

func toTrackDomain(m *trackModel) *domain.Track {
	return &domain.Track{
		ID:        m.ID,
		VinylID:   m.VinylID,
		Title:     m.Title,
		Position:  m.Position,
		Lyrics:    m.Lyrics,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type TrackRepository struct {
	db *gorm.DB
}

func NewTrackRepository(db *gorm.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(&vinylModel{}, &trackModel{}, &profileModel{})
}

func (r *TrackRepository) Create(ctx context.Context, t *domain.Track) error {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.Create")
	defer span.End()
	m := &trackModel{
		ID: t.ID, VinylID: t.VinylID, Title: t.Title,
		Position: t.Position, Lyrics: t.Lyrics,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *TrackRepository) FindByVinylID(ctx context.Context, vinylID string) ([]domain.Track, error) {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.FindByVinylID")
	defer span.End()
	var models []trackModel
	if err := r.db.WithContext(ctx).Where("vinyl_id = ?", vinylID).Order("position asc").Find(&models).Error; err != nil {
		return nil, err
	}
	tracks := make([]domain.Track, len(models))
	for i, m := range models {
		tracks[i] = *toTrackDomain(&m)
	}
	return tracks, nil
}

func (r *TrackRepository) FindByID(ctx context.Context, id string) (*domain.Track, error) {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.FindByID")
	defer span.End()
	var m trackModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTrackNotFound
	}
	if err != nil {
		return nil, err
	}
	return toTrackDomain(&m), nil
}

func (r *TrackRepository) Update(ctx context.Context, t *domain.Track) error {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.Update")
	defer span.End()
	result := r.db.WithContext(ctx).Model(&trackModel{}).Where("id = ?", t.ID).Updates(map[string]any{
		"title": t.Title, "position": t.Position, "lyrics": t.Lyrics, "updated_at": t.UpdatedAt,
	})
	if result.RowsAffected == 0 {
		return domain.ErrTrackNotFound
	}
	return result.Error
}

func (r *TrackRepository) Delete(ctx context.Context, id string) error {
	_, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackRepository.Delete")
	defer span.End()
	result := r.db.WithContext(ctx).Delete(&trackModel{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return domain.ErrTrackNotFound
	}
	return result.Error
}
```

**Step 4: Criar `internal/application/track_service.go`**

```go
package application

import (
	"context"

	"go.opentelemetry.io/otel"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type trackService struct {
	repo domain.TrackRepository
}

func NewTrackService(repo domain.TrackRepository) domain.TrackService {
	return &trackService{repo: repo}
}

func (s *trackService) CreateTrack(ctx context.Context, t *domain.Track) error {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackService.CreateTrack")
	defer span.End()
	return s.repo.Create(ctx, t)
}

func (s *trackService) ListTracks(ctx context.Context, vinylID string) ([]domain.Track, error) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackService.ListTracks")
	defer span.End()
	return s.repo.FindByVinylID(ctx, vinylID)
}

func (s *trackService) UpdateTrack(ctx context.Context, t *domain.Track) error {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackService.UpdateTrack")
	defer span.End()
	return s.repo.Update(ctx, t)
}

func (s *trackService) DeleteTrack(ctx context.Context, id string) error {
	ctx, span := otel.Tracer("vinyl-catalog").Start(ctx, "TrackService.DeleteTrack")
	defer span.End()
	return s.repo.Delete(ctx, id)
}
```

**Step 5: Verificar compilação**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog && go build ./...
```

**Step 6: Commit**

```bash
git add internal/domain/track.go internal/domain/errors.go \
        internal/application/track_service.go \
        internal/adapters/secondary/postgres/track_repo.go
git commit -m "feat(domain): add Track entity, repository and service"
```

---

## Task 5: HTTP handlers para Tracks

**Files:**
- Modify: `internal/adapters/primary/http/handler.go`
- Modify: `internal/adapters/primary/http/dto.go`
- Modify: `internal/adapters/primary/http/router.go`

**Step 1: Adicionar DTOs de Track em `dto.go`**

```go
type TrackRequest struct {
	Title    string `json:"title"    binding:"required" example:"So What"`
	Position int    `json:"position" example:"1"`
	Lyrics   string `json:"lyrics"   example:"So What is a D Dorian..."`
}

type TrackResponse struct {
	ID        string `json:"id"`
	VinylID   string `json:"vinyl_id"`
	Title     string `json:"title"`
	Position  int    `json:"position"`
	Lyrics    string `json:"lyrics"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toTrackResponse(t *domain.Track) TrackResponse {
	return TrackResponse{
		ID:        t.ID,
		VinylID:   t.VinylID,
		Title:     t.Title,
		Position:  t.Position,
		Lyrics:    t.Lyrics,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
```

**Step 2: Adicionar `trackService` ao `VinylHandler` e novos handlers**

Em `handler.go`, modificar `VinylHandler` para incluir trackService:
```go
type VinylHandler struct {
	service      domain.VinylService
	trackService domain.TrackService
	jwtService   *auth.JWTService
	adminUser    string
	adminPass    string
}
```

Atualizar `NewVinylHandlerWithAuth`:
```go
func NewVinylHandlerWithAuth(service domain.VinylService, trackSvc domain.TrackService, jwtSvc *auth.JWTService, user, pass string) *VinylHandler {
	return &VinylHandler{service: service, trackService: trackSvc, jwtService: jwtSvc, adminUser: user, adminPass: pass}
}
```

Adicionar handlers de track ao final do arquivo:

```go
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

func (h *VinylHandler) CreateTrack(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.CreateTrack")
	defer span.End()
	vinylID := c.Param("id")
	var req TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
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
```

**Step 3: Registrar rotas em `router.go`**

Dentro do bloco `vinyls`, após as rotas existentes:
```go
vinyls.GET("/:id/tracks", handler.ListTracks)
vinyls.POST("/:id/tracks", handler.CreateTrack)
vinyls.PUT("/:id/tracks/:track_id", handler.UpdateTrack)
vinyls.DELETE("/:id/tracks/:track_id", handler.DeleteTrack)
```

**Step 4: Verificar compilação**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog && go build ./...
```

**Step 5: Commit**

```bash
git add internal/adapters/primary/http/
git commit -m "feat(api): add track CRUD endpoints (list, create, update, delete)"
```

---

## Task 6: Entidade Profile — domínio, repositório e endpoint

**Files:**
- Create: `internal/domain/profile.go`
- Create: `internal/adapters/secondary/postgres/profile_repo.go`
- Modify: `internal/adapters/primary/http/handler.go`
- Modify: `internal/adapters/primary/http/dto.go`
- Modify: `internal/adapters/primary/http/router.go`

**Step 1: Criar `internal/domain/profile.go`**

```go
package domain

import "context"

type Profile struct {
	ID               string   `json:"id"`
	Username         string   `json:"username"`
	DisplayName      string   `json:"display_name"`
	Bio              string   `json:"bio"`
	PhotoURL         string   `json:"photo_url"`
	Links            []string `json:"links"`
	PreferredGenres  []string `json:"preferred_genres"`
}

type ProfileRepository interface {
	Get(ctx context.Context) (*Profile, error)
	Upsert(ctx context.Context, p *Profile) error
}
```

**Step 2: Criar `internal/adapters/secondary/postgres/profile_repo.go`**

```go
package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type profileModel struct {
	ID              string `gorm:"primaryKey;type:uuid"`
	Username        string `gorm:"uniqueIndex;not null"`
	DisplayName     string
	Bio             string `gorm:"type:text"`
	PhotoURL        string `gorm:"column:photo_url"`
	Links           string `gorm:"type:text"` // comma-separated
	PreferredGenres string `gorm:"type:text"` // comma-separated
}

func (profileModel) TableName() string { return "profiles" }

func splitCSV(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func joinCSV(ss []string) string {
	return strings.Join(ss, ",")
}

func toProfileDomain(m *profileModel) *domain.Profile {
	return &domain.Profile{
		ID:              m.ID,
		Username:        m.Username,
		DisplayName:     m.DisplayName,
		Bio:             m.Bio,
		PhotoURL:        m.PhotoURL,
		Links:           splitCSV(m.Links),
		PreferredGenres: splitCSV(m.PreferredGenres),
	}
}

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Get(ctx context.Context) (*domain.Profile, error) {
	var m profileModel
	err := r.db.WithContext(ctx).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Return empty profile if none exists yet
		return &domain.Profile{Username: "admin", Links: []string{}, PreferredGenres: []string{}}, nil
	}
	if err != nil {
		return nil, err
	}
	return toProfileDomain(&m), nil
}

func (r *ProfileRepository) Upsert(ctx context.Context, p *domain.Profile) error {
	m := &profileModel{
		ID:              p.ID,
		Username:        p.Username,
		DisplayName:     p.DisplayName,
		Bio:             p.Bio,
		PhotoURL:        p.PhotoURL,
		Links:           joinCSV(p.Links),
		PreferredGenres: joinCSV(p.PreferredGenres),
	}
	return r.db.WithContext(ctx).Save(m).Error
}
```

**Step 3: Atualizar `AutoMigrateAll`**

A função `AutoMigrateAll` já foi definida na Task 4 com `&profileModel{}`. Verificar que está assim:
```go
func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(&vinylModel{}, &trackModel{}, &profileModel{})
}
```

**Step 4: Adicionar DTOs de Profile em `dto.go`**

```go
type ProfileResponse struct {
	Username        string   `json:"username"`
	DisplayName     string   `json:"display_name"`
	Bio             string   `json:"bio"`
	PhotoURL        string   `json:"photo_url"`
	Links           []string `json:"links"`
	PreferredGenres []string `json:"preferred_genres"`
}

type UpdateProfileRequest struct {
	DisplayName     string   `json:"display_name"`
	Bio             string   `json:"bio"`
	Links           []string `json:"links"`
	PreferredGenres []string `json:"preferred_genres"`
}
```

**Step 5: Adicionar `ProfileHandler` em `handler.go`**

Adicionar campo `profileRepo` ao `VinylHandler`:
```go
profileRepo domain.ProfileRepository
```

Atualizar `NewVinylHandlerWithAuth`:
```go
func NewVinylHandlerWithAuth(
	service domain.VinylService,
	trackSvc domain.TrackService,
	profileRepo domain.ProfileRepository,
	jwtSvc *auth.JWTService,
	user, pass string,
) *VinylHandler {
	return &VinylHandler{
		service:     service,
		trackService: trackSvc,
		profileRepo: profileRepo,
		jwtService:  jwtSvc,
		adminUser:   user,
		adminPass:   pass,
	}
}
```

Adicionar handlers:
```go
func (h *VinylHandler) GetProfile(c *gin.Context) {
	profile, err := h.profileRepo.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, ProfileResponse{
		Username:        profile.Username,
		DisplayName:     profile.DisplayName,
		Bio:             profile.Bio,
		PhotoURL:        profile.PhotoURL,
		Links:           profile.Links,
		PreferredGenres: profile.PreferredGenres,
	})
}

func (h *VinylHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	current, err := h.profileRepo.Get(c.Request.Context())
	if err != nil {
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
	if err := h.profileRepo.Upsert(c.Request.Context(), current); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not save profile"})
		return
	}
	c.JSON(http.StatusOK, ProfileResponse{
		Username:        current.Username,
		DisplayName:     current.DisplayName,
		Bio:             current.Bio,
		PhotoURL:        current.PhotoURL,
		Links:           current.Links,
		PreferredGenres: current.PreferredGenres,
	})
}

func (h *VinylHandler) UploadProfilePhoto(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[strings.ToLower(ext)] {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "only jpg, png and webp are allowed"})
		return
	}

	filename := "profile-" + uuid.NewString() + ext
	dst := filepath.Join("uploads", filename)
	if err := os.MkdirAll("uploads", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not create upload dir"})
		return
	}
	out, err := os.Create(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not save file"})
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not write file"})
		return
	}

	current, _ := h.profileRepo.Get(c.Request.Context())
	if current.ID == "" {
		current.ID = uuid.NewString()
		current.Username = h.adminUser
	}
	current.PhotoURL = "/uploads/" + filename
	_ = h.profileRepo.Upsert(c.Request.Context(), current)

	c.JSON(http.StatusOK, gin.H{"photo_url": current.PhotoURL})
}
```

**Step 6: Registrar rotas em `router.go`**

```go
// Profile — GET público, PUT e foto protegidos
r.GET("/api/v1/profile", handler.GetProfile)

profileRoutes := r.Group("/api/v1/profile")
profileRoutes.Use(jwtMiddleware(jwtSvc))
profileRoutes.PUT("", handler.UpdateProfile)
profileRoutes.POST("/photo", handler.UploadProfilePhoto)
```

**Step 7: Atualizar `main.go`**

Em `cmd/api/main.go`, substituir `AutoMigrate` por `AutoMigrateAll` e adicionar wiring do TrackRepo, TrackService, ProfileRepo:

```go
// Migrations
if err := postgres.AutoMigrateAll(db); err != nil {
    slog.Error("failed to run migrations", "error", err)
    os.Exit(1)
}

// Wiring
repo := postgres.NewVinylRepository(db)
trackRepo := postgres.NewTrackRepository(db)
profileRepo := postgres.NewProfileRepository(db)
svc := application.NewVinylService(repo)
trackSvc := application.NewTrackService(trackRepo)
jwtSvc := auth.NewJWTServiceFromConfig(cfg.JWTSecret, cfg.JWTExpirationHours)
handler := httpAdapter.NewVinylHandlerWithAuth(svc, trackSvc, profileRepo, jwtSvc, cfg.AdminUsername, cfg.AdminPassword)
router := httpAdapter.NewRouter(handler, jwtSvc)
```

**Step 8: Compilar tudo**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog && go build ./...
```
Expected: sem erros.

**Step 9: Rodar todos os testes Go**

```bash
go test ./... -v 2>&1 | grep -E "(PASS|FAIL|ok|---)"
```
Expected: todos passam.

**Step 10: Commit**

```bash
git add internal/domain/profile.go \
        internal/adapters/secondary/postgres/profile_repo.go \
        internal/adapters/primary/http/ \
        cmd/api/main.go
git commit -m "feat(api): add Profile entity, photo upload and public profile endpoint"
```

---

## Task 7: Frontend — atualizar lib/api.ts e proxy routes

**Files:**
- Modify: `frontend/lib/api.ts`
- Create: `frontend/app/api/vinyls/[id]/cover/route.ts`
- Create: `frontend/app/api/vinyls/[id]/tracks/route.ts`
- Create: `frontend/app/api/vinyls/[id]/tracks/[track_id]/route.ts`
- Create: `frontend/app/api/profile/route.ts`
- Create: `frontend/app/api/profile/photo/route.ts`

**Step 1: Atualizar `frontend/lib/api.ts`**

Adicionar os novos tipos ao arquivo existente:

```ts
// Atualizar VinylResponse — adicionar campos:
export type VinylResponse = {
  id: string;
  title: string;
  artist: string;
  year: number;
  genre: string;
  label: string;
  description: string;   // novo
  cover_url: string;     // novo
  created_at: string;
  updated_at: string;
};

// Atualizar CreateVinylInput — adicionar:
export type CreateVinylInput = {
  title: string;
  artist: string;
  year: number;
  genre: string;
  label: string;
  description: string;   // novo
};

// Novos tipos:
export type TrackResponse = {
  id: string;
  vinyl_id: string;
  title: string;
  position: number;
  lyrics: string;
  created_at: string;
  updated_at: string;
};

export type TrackInput = {
  title: string;
  position: number;
  lyrics: string;
};

export type ProfileResponse = {
  username: string;
  display_name: string;
  bio: string;
  photo_url: string;
  links: string[];
  preferred_genres: string[];
};

export type UpdateProfileInput = {
  display_name: string;
  bio: string;
  links: string[];
  preferred_genres: string[];
};
```

**Step 2: Criar proxy route para upload de capa**

Criar `frontend/app/api/vinyls/[id]/cover/route.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";

  const formData = await req.formData();

  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/cover`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  });

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
```

**Step 3: Criar proxy route para tracks**

Criar `frontend/app/api/vinyls/[id]/tracks/route.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

async function getToken() {
  const cookieStore = await cookies();
  return cookieStore.get("token")?.value ?? "";
}

export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const token = await getToken();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/tracks`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const token = await getToken();
  const body = await req.json();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/tracks`, {
    method: "POST",
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
```

**Step 4: Criar proxy route para track individual**

Criar `frontend/app/api/vinyls/[id]/tracks/[track_id]/route.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

async function getToken() {
  const cookieStore = await cookies();
  return cookieStore.get("token")?.value ?? "";
}

export async function PUT(
  req: NextRequest,
  { params }: { params: Promise<{ id: string; track_id: string }> }
) {
  const { id, track_id } = await params;
  const token = await getToken();
  const body = await req.json();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/tracks/${track_id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function DELETE(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string; track_id: string }> }
) {
  const { id, track_id } = await params;
  const token = await getToken();
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/tracks/${track_id}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) return NextResponse.json({ error: "Delete failed" }, { status: res.status });
  return NextResponse.json({ ok: true });
}
```

**Step 5: Criar proxy route para perfil**

Criar `frontend/app/api/profile/route.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function GET() {
  const res = await fetch(`${process.env.API_URL}/api/v1/profile`);
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}

export async function PUT(req: NextRequest) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";
  const body = await req.json();
  const res = await fetch(`${process.env.API_URL}/api/v1/profile`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
```

**Step 6: Criar proxy route para foto do perfil**

Criar `frontend/app/api/profile/photo/route.ts`:
```ts
import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function POST(req: NextRequest) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";
  const formData = await req.formData();
  const res = await fetch(`${process.env.API_URL}/api/v1/profile/photo`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  });
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
```

**Step 7: Verificar TypeScript**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend && npx tsc --noEmit
```
Expected: sem erros.

**Step 8: Commit**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
git add frontend/lib/api.ts frontend/app/api/
git commit -m "feat(frontend): add API types and proxy routes for tracks, profile and uploads"
```

---

## Task 8: Frontend — VinylForm com descrição e upload de capa

**Files:**
- Modify: `frontend/components/VinylForm.tsx`
- Modify: `frontend/app/vinyls/new/page.tsx`

**Step 1: Atualizar `VinylForm.tsx`**

Adicionar campo `description` ao schema e ao form:

No `vinylSchema`, adicionar:
```ts
description: z.string().optional().default(""),
```

No componente, adicionar após o campo `label`:
```tsx
<div className="flex flex-col gap-1.5">
  <label htmlFor="description" className="text-sm font-medium text-muted">Description</label>
  <textarea
    id="description"
    rows={3}
    placeholder="What makes this album special…"
    {...register("description")}
    className="bg-surface border border-border rounded px-4 py-2.5 text-foreground placeholder:text-muted focus:outline-none focus:border-accent transition-colors resize-none"
  />
  {errors.description && <p className="text-xs text-red-400">{errors.description.message}</p>}
</div>
```

Adicionar prop `vinylId` opcional e campo de upload de capa (mostrado apenas quando `vinylId` está presente):

```tsx
type Props = {
  defaultValues?: Partial<VinylFormData>;
  onSubmit: (data: VinylFormData) => Promise<void>;
  submitLabel?: string;
  vinylId?: string;  // se fornecido, mostra upload de capa
  onCoverUploaded?: (url: string) => void;
};
```

Adicionar ao componente (após o botão submit, quando `vinylId` existe):
```tsx
{vinylId && (
  <div className="pt-4 border-t border-border">
    <p className="text-sm font-medium text-muted mb-2">Album Cover</p>
    <input
      type="file"
      accept="image/jpeg,image/png,image/webp"
      onChange={async (e) => {
        const file = e.target.files?.[0];
        if (!file) return;
        const fd = new FormData();
        fd.append("file", file);
        const res = await fetch(`/api/vinyls/${vinylId}/cover`, { method: "POST", body: fd });
        if (res.ok) {
          const data = await res.json();
          onCoverUploaded?.(data.cover_url);
        }
      }}
      className="text-sm text-muted file:mr-3 file:py-1.5 file:px-3 file:rounded file:border-0 file:text-xs file:font-medium file:bg-accent file:text-background hover:file:bg-accent/90 cursor-pointer"
    />
  </div>
)}
```

**Step 2: Verificar TypeScript**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend && npx tsc --noEmit
```

**Step 3: Commit**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
git add frontend/components/VinylForm.tsx
git commit -m "feat(frontend): add description field and cover upload to VinylForm"
```

---

## Task 9: Frontend — VinylCard com thumbnail da capa

**Files:**
- Modify: `frontend/components/VinylCard.tsx`

**Step 1: Atualizar `VinylCard.tsx`**

Substituir o bloco de capa (a `div` com `backgroundColor: bgColor`) para mostrar a imagem quando `cover_url` existe:

```tsx
{/* Cover */}
<div
  className="aspect-square rounded-lg flex flex-col items-center justify-center p-4 mb-3 relative overflow-hidden"
  style={{ backgroundColor: bgColor }}
>
  {vinyl.cover_url ? (
    <img
      src={vinyl.cover_url}
      alt={`${vinyl.title} cover`}
      className="absolute inset-0 w-full h-full object-cover"
    />
  ) : (
    <>
      {/* Vinyl groove rings */}
      <div className="absolute inset-0 opacity-20">
        {[20, 35, 50, 65].map((size) => (
          <div
            key={size}
            className="absolute rounded-full border border-black/30"
            style={{
              width: `${size}%`,
              height: `${size}%`,
              top: `${(100 - size) / 2}%`,
              left: `${(100 - size) / 2}%`,
            }}
          />
        ))}
      </div>
      <div className="w-4 h-4 rounded-full bg-background/60 z-10" />
    </>
  )}
  <div className="absolute inset-0 rounded-lg ring-0 group-hover:ring-2 group-hover:ring-accent/50 transition-all duration-200" />
</div>
```

**Step 2: Verificar TypeScript**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend && npx tsc --noEmit
```

**Step 3: Commit**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
git add frontend/components/VinylCard.tsx
git commit -m "feat(frontend): show album cover thumbnail in VinylCard"
```

---

## Task 10: Frontend — página de detalhe com descrição, upload de capa e lista de faixas

**Files:**
- Modify: `frontend/app/vinyls/[id]/page.tsx`

**Step 1: Substituir `frontend/app/vinyls/[id]/page.tsx` completamente**

```tsx
"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { motion, AnimatePresence } from "framer-motion";
import { ArrowLeft, Trash2, Plus, ChevronDown, ChevronUp, Music } from "lucide-react";
import { VinylResponse, TrackResponse } from "@/lib/api";
import { VinylForm, VinylFormData } from "@/components/VinylForm";
import { artistToHsl } from "@/lib/vinylColor";
import { useToast } from "@/components/ToastProvider";
import { Button } from "@/components/ui/Button";

// ─── Track item ────────────────────────────────────────────────────────────────

function TrackItem({
  track,
  vinylId,
  onUpdated,
  onDeleted,
}: {
  track: TrackResponse;
  vinylId: string;
  onUpdated: (t: TrackResponse) => void;
  onDeleted: (id: string) => void;
}) {
  const [showLyrics, setShowLyrics] = useState(false);
  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState(track.title);
  const [lyrics, setLyrics] = useState(track.lyrics);
  const { toast } = useToast();

  const save = async () => {
    const res = await fetch(`/api/vinyls/${vinylId}/tracks/${track.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title, position: track.position, lyrics }),
    });
    if (res.ok) {
      const updated = await res.json();
      onUpdated(updated);
      setEditing(false);
      toast("Track updated", "success");
    } else {
      toast("Failed to update track", "error");
    }
  };

  const del = async () => {
    if (!confirm(`Delete "${track.title}"?`)) return;
    const res = await fetch(`/api/vinyls/${vinylId}/tracks/${track.id}`, { method: "DELETE" });
    if (res.ok) {
      onDeleted(track.id);
      toast("Track deleted", "success");
    } else {
      toast("Failed to delete track", "error");
    }
  };

  return (
    <div className="border border-border rounded-lg overflow-hidden">
      <div className="flex items-center gap-3 px-4 py-3 bg-surface">
        <span className="text-muted text-xs w-5 text-right">{track.position}</span>
        {editing ? (
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="flex-1 bg-background border border-border rounded px-2 py-1 text-sm text-foreground focus:outline-none focus:border-accent"
          />
        ) : (
          <span className="flex-1 text-sm font-medium text-foreground">{track.title}</span>
        )}
        <div className="flex items-center gap-1">
          {editing ? (
            <>
              <button onClick={save} className="text-xs text-accent hover:text-accent/80 px-2 py-1">Save</button>
              <button onClick={() => setEditing(false)} className="text-xs text-muted hover:text-foreground px-2 py-1">Cancel</button>
            </>
          ) : (
            <button onClick={() => setEditing(true)} className="text-xs text-muted hover:text-accent px-2 py-1">Edit</button>
          )}
          {track.lyrics && !editing && (
            <button onClick={() => setShowLyrics(!showLyrics)} className="text-muted hover:text-accent p-1">
              {showLyrics ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
            </button>
          )}
          <button onClick={del} className="text-muted hover:text-red-400 p-1"><Trash2 size={13} /></button>
        </div>
      </div>

      <AnimatePresence>
        {(showLyrics || editing) && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden"
          >
            <div className="px-4 py-3 border-t border-border bg-background/50">
              {editing ? (
                <textarea
                  value={lyrics}
                  onChange={(e) => setLyrics(e.target.value)}
                  rows={6}
                  placeholder="Paste lyrics here…"
                  className="w-full bg-surface border border-border rounded px-3 py-2 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent resize-none"
                />
              ) : (
                <pre className="text-sm text-muted whitespace-pre-wrap font-sans leading-relaxed">{track.lyrics}</pre>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

// ─── Add track form ─────────────────────────────────────────────────────────────

function AddTrackForm({ vinylId, onAdded }: { vinylId: string; onAdded: (t: TrackResponse) => void }) {
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [position, setPosition] = useState(1);
  const [lyrics, setLyrics] = useState("");
  const { toast } = useToast();

  const submit = async () => {
    if (!title.trim()) return;
    const res = await fetch(`/api/vinyls/${vinylId}/tracks`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title, position, lyrics }),
    });
    if (res.ok) {
      const t = await res.json();
      onAdded(t);
      setTitle(""); setLyrics(""); setOpen(false);
      toast("Track added", "success");
    } else {
      toast("Failed to add track", "error");
    }
  };

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        className="flex items-center gap-2 text-sm text-muted hover:text-accent transition-colors border border-dashed border-border rounded-lg px-4 py-3 w-full"
      >
        <Plus size={14} /> Add track
      </button>
    );
  }

  return (
    <div className="border border-accent/30 rounded-lg p-4 flex flex-col gap-3">
      <div className="flex gap-3">
        <input
          type="number"
          value={position}
          onChange={(e) => setPosition(Number(e.target.value))}
          className="w-16 bg-surface border border-border rounded px-2 py-1.5 text-sm text-foreground focus:outline-none focus:border-accent"
          placeholder="#"
          min={1}
        />
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="flex-1 bg-surface border border-border rounded px-3 py-1.5 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent"
          placeholder="Track title"
        />
      </div>
      <textarea
        value={lyrics}
        onChange={(e) => setLyrics(e.target.value)}
        rows={4}
        placeholder="Lyrics (optional)"
        className="bg-surface border border-border rounded px-3 py-2 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent resize-none"
      />
      <div className="flex gap-2">
        <button onClick={submit} className="bg-accent text-background px-4 py-1.5 rounded text-sm font-medium hover:bg-accent/90">Add</button>
        <button onClick={() => setOpen(false)} className="text-sm text-muted hover:text-foreground px-3">Cancel</button>
      </div>
    </div>
  );
}

// ─── Main page ──────────────────────────────────────────────────────────────────

export default function VinylDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { toast } = useToast();
  const [vinyl, setVinyl] = useState<VinylResponse | null>(null);
  const [tracks, setTracks] = useState<TrackResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    Promise.all([
      fetch(`/api/vinyls/${id}`).then((r) => r.json()),
      fetch(`/api/vinyls/${id}/tracks`).then((r) => r.json()),
    ])
      .then(([v, t]) => { setVinyl(v); setTracks(Array.isArray(t) ? t : []); })
      .catch(() => toast("Failed to load vinyl", "error"))
      .finally(() => setLoading(false));
  }, [id]);

  const handleUpdate = async (data: VinylFormData) => {
    try {
      const res = await fetch(`/api/vinyls/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error("Update failed");
      const updated = await res.json();
      setVinyl(updated);
      setEditing(false);
      toast("Vinyl updated", "success");
    } catch {
      toast("Failed to update vinyl", "error");
    }
  };

  const handleDelete = async () => {
    if (!confirm("Delete this vinyl permanently?")) return;
    setDeleting(true);
    const res = await fetch(`/api/vinyls/${id}`, { method: "DELETE" });
    if (res.ok) {
      toast("Vinyl deleted", "success");
      router.push("/vinyls");
    } else {
      toast("Failed to delete", "error");
      setDeleting(false);
    }
  };

  if (loading) return (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
    </div>
  );

  if (!vinyl) return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center gap-4">
      <p className="font-serif text-2xl">Record not found</p>
      <Link href="/vinyls" className="text-accent text-sm">← Back to collection</Link>
    </div>
  );

  return (
    <div className="min-h-screen bg-background px-4 py-8">
      <div className="max-w-5xl mx-auto">
        <Link href="/vinyls" className="inline-flex items-center gap-2 text-muted hover:text-accent transition-colors text-sm mb-8">
          <ArrowLeft size={14} /> Back to collection
        </Link>

        {/* Top: cover + info */}
        <div className="grid md:grid-cols-2 gap-12 items-start mb-12">
          {/* Cover */}
          <motion.div initial={{ opacity: 0, x: -20 }} animate={{ opacity: 1, x: 0 }}>
            <div
              className="aspect-square rounded-2xl relative overflow-hidden flex items-center justify-center"
              style={{ backgroundColor: artistToHsl(vinyl.artist) }}
            >
              {vinyl.cover_url ? (
                <img src={vinyl.cover_url} alt={`${vinyl.title} cover`} className="absolute inset-0 w-full h-full object-cover" />
              ) : (
                <>
                  {[20, 35, 50, 65, 78].map((size) => (
                    <div key={size} className="absolute rounded-full border border-black/20"
                      style={{ width: `${size}%`, height: `${size}%`, top: `${(100-size)/2}%`, left: `${(100-size)/2}%` }} />
                  ))}
                  <div className="w-6 h-6 rounded-full bg-background/50 z-10" />
                </>
              )}
            </div>
            {/* Upload cover button */}
            {!editing && (
              <label className="mt-3 flex items-center justify-center gap-2 text-xs text-muted hover:text-accent cursor-pointer transition-colors">
                <input type="file" accept="image/jpeg,image/png,image/webp" className="hidden"
                  onChange={async (e) => {
                    const file = e.target.files?.[0];
                    if (!file) return;
                    const fd = new FormData();
                    fd.append("file", file);
                    const res = await fetch(`/api/vinyls/${id}/cover`, { method: "POST", body: fd });
                    if (res.ok) {
                      const data = await res.json();
                      setVinyl((prev) => prev ? { ...prev, cover_url: data.cover_url } : prev);
                      toast("Cover updated", "success");
                    } else {
                      toast("Failed to upload cover", "error");
                    }
                  }}
                />
                Change cover
              </label>
            )}
          </motion.div>

          {/* Info / edit */}
          <motion.div initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }}>
            {editing ? (
              <>
                <h2 className="font-serif text-2xl font-bold mb-6">Edit Record</h2>
                <VinylForm
                  defaultValues={{ title: vinyl.title, artist: vinyl.artist, year: vinyl.year, genre: vinyl.genre, label: vinyl.label, description: vinyl.description }}
                  onSubmit={handleUpdate}
                  submitLabel="Save changes"
                />
                <button onClick={() => setEditing(false)} className="mt-4 text-sm text-muted hover:text-foreground transition-colors w-full text-center">Cancel</button>
              </>
            ) : (
              <>
                <h1 className="font-serif text-4xl font-bold leading-tight">{vinyl.title}</h1>
                <p className="text-accent text-xl mt-1">{vinyl.artist}</p>

                <dl className="mt-6 grid grid-cols-2 gap-4">
                  {[
                    ["Year", vinyl.year],
                    ["Genre", vinyl.genre || "—"],
                    ["Label", vinyl.label || "—"],
                    ["Added", new Date(vinyl.created_at).toLocaleDateString()],
                  ].map(([k, v]) => (
                    <div key={String(k)}>
                      <dt className="text-xs text-muted uppercase tracking-wider">{k}</dt>
                      <dd className="text-foreground mt-0.5">{v}</dd>
                    </div>
                  ))}
                </dl>

                {vinyl.description && (
                  <p className="mt-6 text-sm text-muted leading-relaxed">{vinyl.description}</p>
                )}

                <div className="flex gap-3 mt-8">
                  <Button onClick={() => setEditing(true)} variant="ghost">Edit</Button>
                  <Button onClick={handleDelete} variant="danger" loading={deleting}>
                    <Trash2 size={14} className="mr-1.5" /> Delete
                  </Button>
                </div>
              </>
            )}
          </motion.div>
        </div>

        {/* Track list */}
        <section>
          <div className="flex items-center gap-2 mb-4">
            <Music size={16} className="text-accent" />
            <h2 className="font-serif text-xl font-bold">Tracklist</h2>
            <span className="text-muted text-sm ml-1">({tracks.length})</span>
          </div>

          <div className="flex flex-col gap-2">
            {tracks.map((t) => (
              <TrackItem
                key={t.id}
                track={t}
                vinylId={id}
                onUpdated={(updated) => setTracks((prev) => prev.map((x) => x.id === updated.id ? updated : x))}
                onDeleted={(tid) => setTracks((prev) => prev.filter((x) => x.id !== tid))}
              />
            ))}
            <AddTrackForm vinylId={id} onAdded={(t) => setTracks((prev) => [...prev, t].sort((a,b) => a.position - b.position))} />
          </div>
        </section>
      </div>
    </div>
  );
}
```

**Step 2: Verificar TypeScript**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend && npx tsc --noEmit
```
Expected: sem erros.

**Step 3: Commit**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
git add frontend/app/vinyls/[id]/
git commit -m "feat(frontend): add tracklist, lyrics, description and cover upload to detail page"
```

---

## Task 11: Frontend — perfil público e página de edição

**Files:**
- Create: `frontend/app/profile/page.tsx`
- Create: `frontend/app/profile/edit/page.tsx`
- Modify: `frontend/app/vinyls/VinylListClient.tsx` (adicionar link para perfil no header)
- Modify: `frontend/proxy.ts` (adicionar /profile como rota pública)

**Step 1: Tornar `/profile` público no middleware**

Em `frontend/proxy.ts`, atualizar `PUBLIC_PATHS`:
```ts
const PUBLIC_PATHS = ["/login", "/api/auth/login", "/profile"];
```

**Step 2: Criar `frontend/app/profile/page.tsx`**

```tsx
import { ProfileResponse, VinylResponse } from "@/lib/api";
import { ProfilePublicClient } from "./ProfilePublicClient";

async function fetchProfile(): Promise<ProfileResponse> {
  const res = await fetch(`${process.env.API_URL}/api/v1/profile`, { cache: "no-store" });
  if (!res.ok) return { username: "admin", display_name: "", bio: "", photo_url: "", links: [], preferred_genres: [] };
  return res.json();
}

async function fetchVinyls(token: string): Promise<VinylResponse[]> {
  if (!token) return [];
  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function ProfilePage() {
  const profile = await fetchProfile();
  // Perfil público não mostra todos os vinyls — só conta e gêneros
  return <ProfilePublicClient profile={profile} />;
}
```

**Step 3: Criar `frontend/app/profile/ProfilePublicClient.tsx`**

```tsx
"use client";

import { motion } from "framer-motion";
import Link from "next/link";
import { ExternalLink, Music2 } from "lucide-react";
import { ProfileResponse } from "@/lib/api";

export function ProfilePublicClient({ profile }: { profile: ProfileResponse }) {
  const displayName = profile.display_name || profile.username;

  return (
    <div className="min-h-screen bg-background px-4 py-12">
      <div className="max-w-2xl mx-auto">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }}>

          {/* Avatar + name */}
          <div className="flex items-center gap-6 mb-8">
            <div className="w-20 h-20 rounded-full overflow-hidden bg-surface border border-border flex-shrink-0 flex items-center justify-center">
              {profile.photo_url ? (
                <img src={profile.photo_url} alt={displayName} className="w-full h-full object-cover" />
              ) : (
                <span className="text-3xl text-muted">◉</span>
              )}
            </div>
            <div>
              <h1 className="font-serif text-3xl font-bold text-foreground">{displayName}</h1>
              {profile.bio && <p className="text-muted text-sm mt-1 leading-relaxed">{profile.bio}</p>}
            </div>
          </div>

          {/* Preferred genres */}
          {profile.preferred_genres.length > 0 && (
            <div className="mb-6">
              <div className="flex items-center gap-2 mb-3">
                <Music2 size={14} className="text-accent" />
                <h2 className="text-xs text-muted uppercase tracking-wider">Preferred Genres</h2>
              </div>
              <div className="flex flex-wrap gap-2">
                {profile.preferred_genres.map((g) => (
                  <span key={g} className="px-3 py-1 rounded-full bg-surface border border-border text-sm text-foreground">
                    {g}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Links */}
          {profile.links.length > 0 && (
            <div className="mb-8">
              <h2 className="text-xs text-muted uppercase tracking-wider mb-3">Links</h2>
              <div className="flex flex-col gap-2">
                {profile.links.map((link) => (
                  <a
                    key={link}
                    href={link.startsWith("http") ? link : `https://${link}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-2 text-sm text-accent hover:text-accent/80 transition-colors"
                  >
                    <ExternalLink size={13} />
                    {link}
                  </a>
                ))}
              </div>
            </div>
          )}

          <div className="pt-6 border-t border-border text-center">
            <Link href="/vinyls" className="text-sm text-muted hover:text-accent transition-colors">
              View collection →
            </Link>
          </div>
        </motion.div>
      </div>
    </div>
  );
}
```

**Step 4: Criar `frontend/app/profile/edit/page.tsx`**

```tsx
"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { ArrowLeft, Plus, X } from "lucide-react";
import { ProfileResponse } from "@/lib/api";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { useToast } from "@/components/ToastProvider";

export default function ProfileEditPage() {
  const router = useRouter();
  const { toast } = useToast();
  const [profile, setProfile] = useState<ProfileResponse | null>(null);
  const [displayName, setDisplayName] = useState("");
  const [bio, setBio] = useState("");
  const [links, setLinks] = useState<string[]>([]);
  const [genres, setGenres] = useState<string[]>([]);
  const [newLink, setNewLink] = useState("");
  const [newGenre, setNewGenre] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetch("/api/profile").then((r) => r.json()).then((p: ProfileResponse) => {
      setProfile(p);
      setDisplayName(p.display_name || "");
      setBio(p.bio || "");
      setLinks(p.links || []);
      setGenres(p.preferred_genres || []);
    });
  }, []);

  const save = async () => {
    setSaving(true);
    try {
      const res = await fetch("/api/profile", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ display_name: displayName, bio, links, preferred_genres: genres }),
      });
      if (!res.ok) throw new Error("Save failed");
      toast("Profile saved", "success");
      router.push("/profile");
    } catch {
      toast("Failed to save profile", "error");
    } finally {
      setSaving(false);
    }
  };

  const uploadPhoto = async (file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    const res = await fetch("/api/profile/photo", { method: "POST", body: fd });
    if (res.ok) {
      const data = await res.json();
      setProfile((prev) => prev ? { ...prev, photo_url: data.photo_url } : prev);
      toast("Photo updated", "success");
    } else {
      toast("Failed to upload photo", "error");
    }
  };

  if (!profile) return (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
    </div>
  );

  return (
    <div className="min-h-screen bg-background px-4 py-8">
      <div className="max-w-lg mx-auto">
        <Link href="/vinyls" className="inline-flex items-center gap-2 text-muted hover:text-accent text-sm mb-8 transition-colors">
          <ArrowLeft size={14} /> Back
        </Link>

        <motion.div initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }}>
          <h1 className="font-serif text-3xl font-bold mb-8">Edit Profile</h1>

          <div className="bg-surface border border-border rounded-xl p-8 flex flex-col gap-6">

            {/* Photo */}
            <div className="flex items-center gap-4">
              <div className="w-16 h-16 rounded-full overflow-hidden bg-background border border-border flex items-center justify-center flex-shrink-0">
                {profile.photo_url ? (
                  <img src={profile.photo_url} alt="" className="w-full h-full object-cover" />
                ) : (
                  <span className="text-2xl text-muted">◉</span>
                )}
              </div>
              <label className="text-sm text-muted hover:text-accent cursor-pointer transition-colors">
                <input type="file" accept="image/jpeg,image/png,image/webp" className="hidden"
                  onChange={(e) => { const f = e.target.files?.[0]; if (f) uploadPhoto(f); }} />
                Change photo
              </label>
            </div>

            {/* Display name */}
            <Input
              id="display_name"
              label="Display Name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="Your name"
            />

            {/* Bio */}
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-muted">Bio</label>
              <textarea
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                rows={3}
                placeholder="Tell the world about your music taste…"
                className="bg-background border border-border rounded px-4 py-2.5 text-foreground placeholder:text-muted focus:outline-none focus:border-accent transition-colors resize-none text-sm"
              />
            </div>

            {/* Preferred genres */}
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium text-muted">Preferred Genres</label>
              <div className="flex flex-wrap gap-2 mb-2">
                {genres.map((g) => (
                  <span key={g} className="flex items-center gap-1 px-2.5 py-1 rounded-full bg-background border border-border text-sm text-foreground">
                    {g}
                    <button onClick={() => setGenres(genres.filter((x) => x !== g))} className="text-muted hover:text-red-400 ml-0.5"><X size={11} /></button>
                  </span>
                ))}
              </div>
              <div className="flex gap-2">
                <input
                  value={newGenre}
                  onChange={(e) => setNewGenre(e.target.value)}
                  onKeyDown={(e) => { if (e.key === "Enter" && newGenre.trim()) { setGenres([...genres, newGenre.trim()]); setNewGenre(""); } }}
                  placeholder="Add genre (Enter to add)"
                  className="flex-1 bg-background border border-border rounded px-3 py-1.5 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent transition-colors"
                />
                <button onClick={() => { if (newGenre.trim()) { setGenres([...genres, newGenre.trim()]); setNewGenre(""); } }}
                  className="p-1.5 border border-border rounded hover:border-accent text-muted hover:text-accent transition-colors">
                  <Plus size={16} />
                </button>
              </div>
            </div>

            {/* Links */}
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium text-muted">Links</label>
              <div className="flex flex-col gap-1.5 mb-2">
                {links.map((l) => (
                  <div key={l} className="flex items-center gap-2">
                    <span className="flex-1 text-sm text-accent truncate">{l}</span>
                    <button onClick={() => setLinks(links.filter((x) => x !== l))} className="text-muted hover:text-red-400"><X size={12} /></button>
                  </div>
                ))}
              </div>
              <div className="flex gap-2">
                <input
                  value={newLink}
                  onChange={(e) => setNewLink(e.target.value)}
                  onKeyDown={(e) => { if (e.key === "Enter" && newLink.trim()) { setLinks([...links, newLink.trim()]); setNewLink(""); } }}
                  placeholder="https://… (Enter to add)"
                  className="flex-1 bg-background border border-border rounded px-3 py-1.5 text-sm text-foreground placeholder:text-muted focus:outline-none focus:border-accent transition-colors"
                />
                <button onClick={() => { if (newLink.trim()) { setLinks([...links, newLink.trim()]); setNewLink(""); } }}
                  className="p-1.5 border border-border rounded hover:border-accent text-muted hover:text-accent transition-colors">
                  <Plus size={16} />
                </button>
              </div>
            </div>

            <Button onClick={save} loading={saving} className="w-full mt-2">Save Profile</Button>
          </div>

          <div className="mt-4 text-center">
            <Link href="/profile" className="text-sm text-muted hover:text-accent transition-colors">
              View public profile →
            </Link>
          </div>
        </motion.div>
      </div>
    </div>
  );
}
```

**Step 5: Adicionar link de perfil no header da lista**

Em `frontend/app/vinyls/VinylListClient.tsx`, no header, adicionar um link para editar perfil ao lado do logout:

```tsx
import { User } from "lucide-react";

// No JSX, antes do botão de logout:
<Link href="/profile/edit" className="p-2 text-muted hover:text-foreground transition-colors" title="Edit profile">
  <User size={16} />
</Link>
```

**Step 6: Verificar TypeScript**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend && npx tsc --noEmit
```
Expected: sem erros.

**Step 7: Build de produção**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend && npm run build
```
Expected: build completa sem erros.

**Step 8: Commit**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
git add frontend/app/profile/ frontend/app/vinyls/VinylListClient.tsx frontend/proxy.ts
git commit -m "feat(frontend): add public profile page and profile edit with photo, bio, links and genres"
```

---

## Task 12: Verificação final

**Step 1: Rodar todos os testes Go**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
go test ./... -v 2>&1 | grep -E "(PASS|FAIL|ok|---)"
```
Expected: todos passam.

**Step 2: Rodar testes frontend**

```bash
cd frontend && npm test
```
Expected: 12 testes passam.

**Step 3: TypeScript e lint**

```bash
npx tsc --noEmit && npm run lint
```
Expected: sem erros.

**Step 4: Build de produção**

```bash
npm run build
```
Expected: build completa.

**Step 5: Subir com docker-compose**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
docker-compose up --build
```

Verificar:
- `http://localhost:8080/api/v1/profile` retorna JSON do perfil
- `http://localhost:3001` redireciona para login
- Após login: detalhe do vinil tem tracklist e upload de capa
- `http://localhost:3001/profile` exibe perfil público (sem login)
- `http://localhost:3001/profile/edit` permite editar (requer login)

**Step 6: Commit final**

```bash
git add -A
git commit -m "feat: complete vinyl features — tracks, lyrics, covers and public profile"
```
