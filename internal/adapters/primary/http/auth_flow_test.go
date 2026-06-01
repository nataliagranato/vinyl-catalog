package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	httpAdapter "github.com/nataliagranato/vinyl-catalog/internal/adapters/primary/http"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
	"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"
)

// TestAuthFlow_LoginThenAccess exercita o fluxo completo: login → token → rota protegida.
func TestAuthFlow_LoginThenAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := auth.NewJWTService("integration-secret", 1)
	vinylSvc := &mockService{vinyls: []domain.Vinyl{
		{ID: "1", Title: "Kind of Blue", Artist: "Miles Davis", Year: 1959, Genre: "Jazz", Label: "Columbia"},
	}}
	h := httpAdapter.NewVinylHandlerWithAuth(vinylSvc, &mockTrackService{}, &mockProfileRepo{}, jwtSvc, "admin", "admin123")

	r := gin.New()
	r.POST("/api/v1/auth/login", h.Login)
	protected := r.Group("/api/v1/vinyls")
	protected.Use(httpAdapter.JWTMiddleware(jwtSvc))
	protected.GET("", h.ListVinyls)

	// Passo 1: Login com credenciais corretas → 200 + token
	loginBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%s", w.Code, w.Body.String())
	}
	var tokenResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &tokenResp)
	token := tokenResp["token"]
	if token == "" {
		t.Fatal("expected token in response")
	}

	// Passo 2: Acesso com token válido → 200
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d body=%s", w2.Code, w2.Body.String())
	}

	// Passo 3: Sem token → 401
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w3.Code)
	}

	// Passo 4: Token sem prefixo Bearer → 200 (aceito para compatibilidade com Swagger UI)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	req4.Header.Set("Authorization", token) // sem Bearer
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Errorf("expected 200 with bare token (Swagger UI compat), got %d", w4.Code)
	}

	// Passo 5: Senha errada → 401
	wrongBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(wrongBody))
	req5.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w5, req5)
	if w5.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong password, got %d", w5.Code)
	}
}
