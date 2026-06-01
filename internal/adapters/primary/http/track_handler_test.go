package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	httpAdapter "github.com/nataliagranato/vinyl-catalog/internal/adapters/primary/http"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
)

type mockTrackSvc struct {
	tracks []domain.Track
	err    error
}

func (m *mockTrackSvc) CreateTrack(_ context.Context, t *domain.Track) error { return m.err }
func (m *mockTrackSvc) ListTracks(_ context.Context, _ string) ([]domain.Track, error) {
	return m.tracks, m.err
}
func (m *mockTrackSvc) UpdateTrack(_ context.Context, _ *domain.Track) error { return m.err }
func (m *mockTrackSvc) DeleteTrack(_ context.Context, _ string) error        { return m.err }

func setupTrackRouter(vinylSvc domain.VinylService, trackSvc domain.TrackService) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	jwtSvc := auth.NewJWTService("test-secret", 1)
	h := httpAdapter.NewVinylHandlerWithAuth(vinylSvc, trackSvc, &mockProfileRepo{}, jwtSvc, "admin", "secret")
	r := gin.New()
	protected := r.Group("/api/v1/vinyls")
	protected.Use(httpAdapter.JWTMiddleware(jwtSvc))
	protected.GET("/:id/tracks", h.ListTracks)
	protected.POST("/:id/tracks", h.CreateTrack)
	protected.PUT("/:id/tracks/:track_id", h.UpdateTrack)
	protected.DELETE("/:id/tracks/:track_id", h.DeleteTrack)
	token, _ := jwtSvc.GenerateToken("admin")
	return r, token
}

func TestHandler_ListTracks(t *testing.T) {
	vinylSvc := &mockService{vinyls: []domain.Vinyl{{ID: "v1", Title: "A", Artist: "B", Year: 2000, Genre: "Rock", Label: "L"}}}
	trackSvc := &mockTrackSvc{tracks: []domain.Track{{ID: "t1", VinylID: "v1", Title: "Track 1", Position: 1}}}
	r, token := setupTrackRouter(vinylSvc, trackSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls/v1/tracks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateTrack(t *testing.T) {
	vinylSvc := &mockService{vinyls: []domain.Vinyl{{ID: "v1", Title: "A", Artist: "B", Year: 2000, Genre: "Rock", Label: "L"}}}
	trackSvc := &mockTrackSvc{}
	r, token := setupTrackRouter(vinylSvc, trackSvc)

	body, _ := json.Marshal(map[string]any{"title": "New Track", "position": 1})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vinyls/v1/tracks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateTrack_VinylNotFound(t *testing.T) {
	vinylSvc := &mockService{} // sem vinyls
	trackSvc := &mockTrackSvc{}
	r, token := setupTrackRouter(vinylSvc, trackSvc)

	body, _ := json.Marshal(map[string]any{"title": "Track", "position": 1})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vinyls/nonexistent/tracks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_DeleteTrack(t *testing.T) {
	vinylSvc := &mockService{}
	trackSvc := &mockTrackSvc{}
	r, token := setupTrackRouter(vinylSvc, trackSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/vinyls/v1/tracks/t1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}
