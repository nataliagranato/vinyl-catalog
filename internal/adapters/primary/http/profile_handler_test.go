package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	httpAdapter "github.com/nataliagranato/vinyl-catalog/internal/adapters/primary/http"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
)

func setupProfileRouter() (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	jwtSvc := auth.NewJWTService("test-secret", 1)
	h := httpAdapter.NewVinylHandlerWithAuth(&mockService{}, &mockTrackService{}, &mockProfileRepo{}, jwtSvc, "admin", "secret")
	r := gin.New()
	r.GET("/api/v1/profile", h.GetProfile)
	protected := r.Group("/api/v1/profile")
	protected.Use(httpAdapter.JWTMiddleware(jwtSvc))
	protected.PUT("", h.UpdateProfile)
	token, _ := jwtSvc.GenerateToken("admin")
	return r, token
}

func TestHandler_GetProfile_Public(t *testing.T) {
	r, _ := setupProfileRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestHandler_UpdateProfile_RequiresAuth(t *testing.T) {
	r, _ := setupProfileRouter()
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"display_name": "Natalia"})
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandler_UpdateProfile_WithAuth(t *testing.T) {
	r, token := setupProfileRouter()
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]any{
		"display_name":     "Natalia",
		"bio":              "Colecionadora de jazz.",
		"links":            []string{"https://github.com/nataliagranato"},
		"preferred_genres": []string{"Jazz", "Soul"},
	})
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}
