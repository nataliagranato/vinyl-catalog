package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	httpAdapter "github.com/nataliagranato/vinyl-catalog/internal/adapters/primary/http"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
)

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	jwtSvc := auth.NewJWTService("test-secret", 1)
	r := gin.New()
	protected := r.Group("/protected")
	protected.Use(httpAdapter.JWTMiddleware(jwtSvc))
	protected.GET("", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestJWTMiddleware_MissingHeader(t *testing.T) {
	r := newTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTMiddleware_TokenWithoutBearerPrefix(t *testing.T) {
	// Middleware accepts bare token (without "Bearer ") for Swagger UI compatibility
	jwtSvc := auth.NewJWTService("test-secret", 1)
	token, _ := jwtSvc.GenerateToken("admin")

	r := newTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", token) // sem "Bearer "
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with bare token, got %d", w.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	r := newTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", 1)
	token, _ := jwtSvc.GenerateToken("admin")

	r := newTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
