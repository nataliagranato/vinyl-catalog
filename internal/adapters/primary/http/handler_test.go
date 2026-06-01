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

type mockService struct {
	vinyls []domain.Vinyl
	err    error
}

func (m *mockService) CreateVinyl(_ context.Context, _ *domain.Vinyl) error { return m.err }
func (m *mockService) GetVinyl(_ context.Context, id string) (*domain.Vinyl, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.vinyls {
		if m.vinyls[i].ID == id {
			return &m.vinyls[i], nil
		}
	}
	return nil, domain.ErrVinylNotFound
}
func (m *mockService) ListVinyls(_ context.Context) ([]domain.Vinyl, error) {
	return m.vinyls, m.err
}
func (m *mockService) UpdateVinyl(_ context.Context, _ *domain.Vinyl) error { return m.err }
func (m *mockService) DeleteVinyl(_ context.Context, _ string) error        { return m.err }

func setupRouter(svc domain.VinylService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httpAdapter.NewVinylHandler(svc)
	v1 := r.Group("/api/v1/vinyls")
	v1.GET("", h.ListVinyls)
	v1.GET("/:id", h.GetVinyl)
	v1.POST("", h.CreateVinyl)
	v1.PUT("/:id", h.UpdateVinyl)
	v1.DELETE("/:id", h.DeleteVinyl)
	return r
}

// mockTrackService satisfies domain.TrackService for testing
type mockTrackService struct{}

func (m *mockTrackService) CreateTrack(_ context.Context, _ *domain.Track) error { return nil }
func (m *mockTrackService) ListTracks(_ context.Context, _ string) ([]domain.Track, error) {
	return nil, nil
}
func (m *mockTrackService) UpdateTrack(_ context.Context, _ *domain.Track) error { return nil }
func (m *mockTrackService) DeleteTrack(_ context.Context, _ string) error        { return nil }

// mockProfileRepo satisfies domain.ProfileRepository for testing
type mockProfileRepo struct{}

func (m *mockProfileRepo) Get(_ context.Context) (*domain.Profile, error) {
	return &domain.Profile{}, nil
}
func (m *mockProfileRepo) Upsert(_ context.Context, _ *domain.Profile) error { return nil }

func setupAuthRouter(svc domain.VinylService) (*gin.Engine, *auth.JWTService) {
	gin.SetMode(gin.TestMode)
	jwtSvc := auth.NewJWTService("test-secret", 1)
	h := httpAdapter.NewVinylHandlerWithAuth(svc, &mockTrackService{}, &mockProfileRepo{}, jwtSvc, "admin", "secret")
	r := gin.New()
	r.POST("/api/v1/auth/login", h.Login)
	protected := r.Group("/api/v1/vinyls")
	protected.Use(httpAdapter.JWTMiddleware(jwtSvc))
	protected.GET("", h.ListVinyls)
	return r, jwtSvc
}

func TestHandler_Login_Success(t *testing.T) {
	svc := &mockService{}
	r, _ := setupAuthRouter(svc)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "secret"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	svc := &mockService{}
	r, _ := setupAuthRouter(svc)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandler_ListVinyls_WithValidToken(t *testing.T) {
	svc := &mockService{vinyls: []domain.Vinyl{
		{ID: "1", Title: "Blue", Artist: "Miles", Year: 1959, Genre: "Jazz", Label: "Columbia"},
	}}
	r, jwtSvc := setupAuthRouter(svc)

	token, _ := jwtSvc.GenerateToken("admin")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandler_ListVinyls_WithoutToken(t *testing.T) {
	svc := &mockService{}
	r, _ := setupAuthRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandler_ListVinyls(t *testing.T) {
	svc := &mockService{vinyls: []domain.Vinyl{
		{ID: "1", Title: "Album A", Artist: "Artist", Year: 2000, Genre: "Rock", Label: "Label"},
	}}
	r := setupRouter(svc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandler_GetVinyl_NotFound(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls/not-found", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandler_CreateVinyl(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc)
	body, _ := json.Marshal(map[string]any{
		"title": "Kind of Blue", "artist": "Miles Davis", "year": 1959,
		"genre": "Jazz", "label": "Columbia",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vinyls", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestHandler_CreateVinyl_InvalidBody(t *testing.T) {
	svc := &mockService{}
	r := setupRouter(svc)
	body := bytes.NewBufferString(`{"title": ""}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/vinyls", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandler_DeleteVinyl_NotFound(t *testing.T) {
	svc := &mockService{err: domain.ErrVinylNotFound}
	r := setupRouter(svc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/vinyls/unknown", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
