package http

import "github.com/nataliagranato/vinyl-catalog/internal/domain"

// CreateVinylRequest representa o body de criação de um disco de vinil
type CreateVinylRequest struct {
	Title  string `json:"title"  binding:"required" example:"Kind of Blue"`
	Artist string `json:"artist" binding:"required" example:"Miles Davis"`
	Year   int    `json:"year"   binding:"required,min=1860" example:"1959"`
	Genre       string `json:"genre"        example:"Jazz"`
	Label       string `json:"label"        example:"Columbia"`
	Description string `json:"description"  example:"Modal jazz masterpiece recorded in two sessions."`
}

// UpdateVinylRequest representa o body de atualização de um disco de vinil
type UpdateVinylRequest struct {
	Title       string `json:"title"        binding:"required" example:"Kind of Blue"`
	Artist      string `json:"artist"       binding:"required" example:"Miles Davis"`
	Year        int    `json:"year"         binding:"required,min=1860" example:"1959"`
	Genre       string `json:"genre"        example:"Jazz"`
	Label       string `json:"label"        example:"Columbia"`
	Description string `json:"description"  example:"Modal jazz masterpiece recorded in two sessions."`
}

// LoginRequest representa as credenciais de autenticação
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin"`
}

// TokenResponse retorna o JWT gerado após autenticação
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ErrorResponse representa uma resposta de erro da API
type ErrorResponse struct {
	Error string `json:"error" example:"vinyl not found"`
}

// VinylResponse representa um disco de vinil na resposta da API
type VinylResponse struct {
	ID          string `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	Title       string `json:"title"       example:"Kind of Blue"`
	Artist      string `json:"artist"      example:"Miles Davis"`
	Year        int    `json:"year"        example:"1959"`
	Genre       string `json:"genre"       example:"Jazz"`
	Label       string `json:"label"       example:"Columbia"`
	Description string `json:"description" example:"Modal jazz masterpiece."`
	CoverURL    string `json:"cover_url"   example:"/uploads/abc123.jpg"`
	CreatedAt   string `json:"created_at"  example:"2026-03-06T00:00:00Z"`
	UpdatedAt   string `json:"updated_at"  example:"2026-03-06T00:00:00Z"`
}

func toVinylResponse(v *domain.Vinyl) VinylResponse {
	return VinylResponse{
		ID:          v.ID,
		Title:       v.Title,
		Artist:      v.Artist,
		Year:        v.Year,
		Genre:       v.Genre,
		Label:       v.Label,
		Description: v.Description,
		CoverURL:    v.CoverURL,
		CreatedAt:   v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// TrackRequest represents the request body for creating or updating a track
type TrackRequest struct {
	Title    string `json:"title"    binding:"required" example:"So What"`
	Position int    `json:"position" example:"1"`
	Lyrics   string `json:"lyrics"   example:"So What is a D Dorian..."`
}

// TrackResponse represents a track in the API response
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

type FavoriteVinylWithTracks struct {
	VinylResponse
	Tracks []TrackResponse `json:"tracks"`
}

type ProfileResponse struct {
	Username         string                   `json:"username"`
	DisplayName      string                   `json:"display_name"`
	Bio              string                   `json:"bio"`
	PhotoURL         string                   `json:"photo_url"`
	Links            []string                 `json:"links"`
	PreferredGenres  []string                 `json:"preferred_genres"`
	FavoriteVinylIDs []string                 `json:"favorite_vinyl_ids"`
	FavoriteVinyls   []FavoriteVinylWithTracks `json:"favorite_vinyls"`
}

type UpdateProfileRequest struct {
	DisplayName     string   `json:"display_name"`
	Bio             string   `json:"bio"`
	Links           []string `json:"links"`
	PreferredGenres []string `json:"preferred_genres"`
}
