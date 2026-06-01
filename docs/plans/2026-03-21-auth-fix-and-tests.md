# Auth Fix & Comprehensive Test Suite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix the auth error message to hint at the correct `Bearer <token>` format, then add unit tests, integration tests, and E2E tests covering auth middleware, track/profile handlers, track/profile services, repositories, and frontend API client.

**Architecture:** The API uses a `jwtMiddleware` in `router.go` that requires `Authorization: Bearer <token>`. Current error message is opaque. Tests are layered: unit (mocks only, no DB), integration (needs PostgreSQL, use `t.Skip` when unavailable), and E2E (Playwright against a running docker-compose stack).

**Tech Stack:** Go 1.25 (mise), `testing` stdlib, `net/http/httptest`, `github.com/gin-gonic/gin` TestMode, Jest (frontend), Playwright (E2E).

---

### Task 1: Fix auth error message in jwtMiddleware

**Files:**
- Modify: `internal/adapters/primary/http/router.go:80`

**Step 1: Write the failing test**

Create `internal/adapters/primary/http/router_test.go`:

```go
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

func TestJWTMiddleware_MissingBearerPrefix(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", 1)
	token, _ := jwtSvc.GenerateToken("admin")

	r := newTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", token) // missing "Bearer "
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if !contains(w.Body.String(), "Bearer") {
		t.Errorf("error message should hint at Bearer format, got: %s", w.Body.String())
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

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  test ./internal/adapters/primary/http/... -run TestJWTMiddleware -v 2>&1 | tail -20
```

Expected: FAIL — `JWTMiddleware` not exported / test for Bearer hint will fail.

**Step 3: Export `jwtMiddleware` and fix the error message**

In `internal/adapters/primary/http/router.go`, change the unexported `jwtMiddleware` to an exported `JWTMiddleware`, and update the error message on line 80:

```go
// JWTMiddleware validates Bearer tokens and sets "username" in context.
func JWTMiddleware(jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "authorization header required"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization format: use 'Bearer <token>'"})
			return
		}
		claims, err := jwtSvc.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
			return
		}
		c.Set("username", claims.Username)
		c.Next()
	}
}
```

Also update all callers in `router.go` from `jwtMiddleware(jwtSvc)` to `JWTMiddleware(jwtSvc)`.

**Step 4: Run tests to verify they pass**

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  test ./internal/adapters/primary/http/... -run TestJWTMiddleware -v 2>&1
```

Expected: 4/4 PASS.

**Step 5: Verify full package still builds**

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  build ./... 2>&1
```

Expected: no output (success).

**Step 6: Commit**

```bash
git add internal/adapters/primary/http/router.go internal/adapters/primary/http/router_test.go
git commit -m "fix(auth): export JWTMiddleware, improve error message to hint Bearer format"
```

---

### Task 2: Unit tests for Login handler

**Files:**
- Modify: `internal/adapters/primary/http/handler_test.go`

**Step 1: Add auth mock setup and Login tests**

The existing `setupRouter` uses `NewVinylHandler` (no auth). We need a second helper that uses `NewVinylHandlerWithAuth`.

Add to `handler_test.go` after the existing mocks:

```go
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
```

Also add `"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/auth"` to the imports.

**Step 2: Run tests**

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  test ./internal/adapters/primary/http/... -v 2>&1 | tail -30
```

Expected: All tests PASS (9+ tests).

**Step 3: Commit**

```bash
git add internal/adapters/primary/http/handler_test.go
git commit -m "test(handler): add Login and authenticated endpoint unit tests"
```

---

### Task 3: Unit tests for TrackService

**Files:**
- Create: `internal/application/track_service_test.go`

**Step 1: Write the tests**

```go
package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nataliagranato/vinyl-catalog/internal/application"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

type mockTrackRepo struct {
	tracks map[string]*domain.Track
	err    error
}

func newMockTrackRepo() *mockTrackRepo {
	return &mockTrackRepo{tracks: make(map[string]*domain.Track)}
}

func (m *mockTrackRepo) Create(_ context.Context, t *domain.Track) error {
	if m.err != nil {
		return m.err
	}
	m.tracks[t.ID] = t
	return nil
}

func (m *mockTrackRepo) FindByVinylID(_ context.Context, vinylID string) ([]domain.Track, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Track
	for _, t := range m.tracks {
		if t.VinylID == vinylID {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (m *mockTrackRepo) FindByID(_ context.Context, id string) (*domain.Track, error) {
	if m.err != nil {
		return nil, m.err
	}
	t, ok := m.tracks[id]
	if !ok {
		return nil, domain.ErrTrackNotFound
	}
	return t, nil
}

func (m *mockTrackRepo) Update(_ context.Context, t *domain.Track) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.tracks[t.ID]; !ok {
		return domain.ErrTrackNotFound
	}
	m.tracks[t.ID] = t
	return nil
}

func (m *mockTrackRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.tracks[id]; !ok {
		return domain.ErrTrackNotFound
	}
	delete(m.tracks, id)
	return nil
}

func TestTrackService_CreateAndList(t *testing.T) {
	repo := newMockTrackRepo()
	svc := application.NewTrackService(repo)
	ctx := context.Background()

	track := &domain.Track{ID: "t1", VinylID: "v1", Title: "So What", Position: 1}
	if err := svc.CreateTrack(ctx, track); err != nil {
		t.Fatalf("CreateTrack() error = %v", err)
	}

	tracks, err := svc.ListTracks(ctx, "v1")
	if err != nil {
		t.Fatalf("ListTracks() error = %v", err)
	}
	if len(tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Title != "So What" {
		t.Errorf("expected title 'So What', got %s", tracks[0].Title)
	}
}

func TestTrackService_UpdateAndDelete(t *testing.T) {
	repo := newMockTrackRepo()
	svc := application.NewTrackService(repo)
	ctx := context.Background()

	track := &domain.Track{ID: "t1", VinylID: "v1", Title: "Original", Position: 1}
	_ = svc.CreateTrack(ctx, track)

	track.Title = "Updated"
	if err := svc.UpdateTrack(ctx, track); err != nil {
		t.Fatalf("UpdateTrack() error = %v", err)
	}

	if err := svc.DeleteTrack(ctx, "t1"); err != nil {
		t.Fatalf("DeleteTrack() error = %v", err)
	}
}

func TestTrackService_DeleteNotFound(t *testing.T) {
	repo := newMockTrackRepo()
	svc := application.NewTrackService(repo)

	err := svc.DeleteTrack(context.Background(), "nonexistent")
	if !errors.Is(err, domain.ErrTrackNotFound) {
		t.Errorf("expected ErrTrackNotFound, got %v", err)
	}
}

func TestTrackService_ListEmpty(t *testing.T) {
	repo := newMockTrackRepo()
	svc := application.NewTrackService(repo)

	tracks, err := svc.ListTracks(context.Background(), "no-such-vinyl")
	if err != nil {
		t.Fatalf("ListTracks() error = %v", err)
	}
	if len(tracks) != 0 {
		t.Errorf("expected empty list, got %d", len(tracks))
	}
}
```

**Step 2: Run tests**

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  test ./internal/application/... -v 2>&1
```

Expected: 8 tests PASS (4 existing + 4 new).

**Step 3: Commit**

```bash
git add internal/application/track_service_test.go
git commit -m "test(application): add TrackService unit tests"
```

---

### Task 4: Unit tests for Track and Profile handlers

**Files:**
- Create: `internal/adapters/primary/http/track_handler_test.go`
- Create: `internal/adapters/primary/http/profile_handler_test.go`

**Step 1: Create track handler tests**

```go
// track_handler_test.go
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
	vinylSvc := &mockService{} // empty — no vinyls
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
```

**Step 2: Create profile handler tests**

```go
// profile_handler_test.go
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
		"bio":             "Colecionadora de jazz.",
		"links":           []string{"https://github.com/nataliagranato"},
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
```

**Step 3: Run all handler tests**

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  test ./internal/adapters/primary/http/... -v 2>&1 | tail -40
```

Expected: 20+ tests PASS.

**Step 4: Commit**

```bash
git add internal/adapters/primary/http/track_handler_test.go \
        internal/adapters/primary/http/profile_handler_test.go
git commit -m "test(handler): add track and profile handler unit tests"
```

---

### Task 5: Integration tests for TrackRepository and ProfileRepository

**Files:**
- Create: `internal/adapters/secondary/postgres/track_repo_test.go`
- Create: `internal/adapters/secondary/postgres/profile_repo_test.go`

**Step 1: Create track repository integration test**

```go
// internal/adapters/secondary/postgres/track_repo_test.go
package postgres_test

import (
	"context"
	"testing"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nataliagranato/vinyl-catalog/internal/adapters/secondary/postgres"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=vinyl_catalog_test sslmode=disable"
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("PostgreSQL not available, skipping integration test: %v", err)
	}
	if err := postgres.AutoMigrateAll(db); err != nil {
		t.Fatalf("AutoMigrateAll error: %v", err)
	}
	return db
}

func TestTrackRepository_Implements(t *testing.T) {
	var _ domain.TrackRepository = (*postgres.TrackRepository)(nil)
	t.Log("TrackRepository correctly implements domain.TrackRepository")
}

func TestTrackRepository_CRUD(t *testing.T) {
	db := openTestDB(t)
	repo := postgres.NewTrackRepository(db)
	ctx := context.Background()

	// First create a vinyl (track has foreign key constraint)
	vinylRepo := postgres.NewVinylRepository(db)
	vinyl := &domain.Vinyl{ID: "test-vinyl-" + t.Name(), Title: "Test", Artist: "Artist", Year: 2000, Genre: "Rock", Label: "Label"}
	if err := vinylRepo.Create(ctx, vinyl); err != nil {
		t.Fatalf("create vinyl: %v", err)
	}
	t.Cleanup(func() { vinylRepo.Delete(ctx, vinyl.ID) })

	// Create track
	track := &domain.Track{ID: "track-" + t.Name(), VinylID: vinyl.ID, Title: "Track 1", Position: 1}
	if err := repo.Create(ctx, track); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() { repo.Delete(ctx, track.ID) })

	// FindByVinylID
	tracks, err := repo.FindByVinylID(ctx, vinyl.ID)
	if err != nil {
		t.Fatalf("FindByVinylID() error = %v", err)
	}
	if len(tracks) == 0 {
		t.Fatal("expected at least 1 track")
	}

	// FindByID
	got, err := repo.FindByID(ctx, track.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Title != "Track 1" {
		t.Errorf("expected 'Track 1', got %s", got.Title)
	}

	// Update
	track.Title = "Updated Track"
	if err := repo.Update(ctx, track); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Delete
	if err := repo.Delete(ctx, track.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify gone
	_, err = repo.FindByID(ctx, track.ID)
	if err != domain.ErrTrackNotFound {
		t.Errorf("expected ErrTrackNotFound after delete, got %v", err)
	}
}
```

**Step 2: Create profile repository integration test**

```go
// internal/adapters/secondary/postgres/profile_repo_test.go
package postgres_test

import (
	"context"
	"testing"

	"github.com/nataliagranato/vinyl-catalog/internal/adapters/secondary/postgres"
	"github.com/nataliagranato/vinyl-catalog/internal/domain"
)

func TestProfileRepository_Implements(t *testing.T) {
	var _ domain.ProfileRepository = (*postgres.ProfileRepository)(nil)
	t.Log("ProfileRepository correctly implements domain.ProfileRepository")
}

func TestProfileRepository_GetEmpty(t *testing.T) {
	db := openTestDB(t)
	repo := postgres.NewProfileRepository(db)

	// Should return empty profile (not error) when none exists
	profile, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
}

func TestProfileRepository_Upsert(t *testing.T) {
	db := openTestDB(t)
	repo := postgres.NewProfileRepository(db)
	ctx := context.Background()

	profile := &domain.Profile{
		Username:        "admin",
		DisplayName:     "Natalia",
		Bio:             "Colecionadora de jazz.",
		Links:           []string{"https://github.com/nataliagranato"},
		PreferredGenres: []string{"Jazz", "Soul"},
	}

	if err := repo.Upsert(ctx, profile); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get() after Upsert error = %v", err)
	}
	if got.DisplayName != "Natalia" {
		t.Errorf("expected DisplayName='Natalia', got %s", got.DisplayName)
	}
	if len(got.Links) == 0 {
		t.Error("expected links to be populated")
	}
}
```

**Step 3: Run integration tests**

These will skip automatically when PostgreSQL is unavailable:

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  test ./internal/adapters/secondary/postgres/... -v 2>&1
```

Expected: Interface tests PASS, CRUD tests either PASS (if DB available) or SKIP.

**Step 4: Commit**

```bash
git add internal/adapters/secondary/postgres/track_repo_test.go \
        internal/adapters/secondary/postgres/profile_repo_test.go
git commit -m "test(postgres): add TrackRepository and ProfileRepository integration tests"
```

---

### Task 6: Full auth flow integration test (HTTP-level)

**Files:**
- Create: `internal/adapters/primary/http/auth_flow_test.go`

This test spins up a full in-process Gin router with mock services and walks through the entire auth flow.

**Step 1: Write the test**

```go
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

// TestAuthFlow_LoginThenAccess exercises the complete login → token → protected route flow.
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
	protected.POST("", h.CreateVinyl)

	// Step 1: Login with correct credentials
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

	// Step 2: Access protected route WITH token
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d body=%s", w2.Code, w2.Body.String())
	}

	// Step 3: Access protected route WITHOUT token → 401
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w3.Code)
	}

	// Step 4: Access protected route with token WITHOUT Bearer prefix → 401
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/api/v1/vinyls", nil)
	req4.Header.Set("Authorization", token) // missing Bearer
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without Bearer prefix, got %d", w4.Code)
	}

	// Step 5: Login with WRONG password → 401
	wrongBody, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(wrongBody))
	req5.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w5, req5)
	if w5.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong password, got %d", w5.Code)
	}
}
```

Note: Add `"github.com/nataliagranato/vinyl-catalog/internal/domain"` import.

**Step 2: Run the test**

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  test ./internal/adapters/primary/http/... -run TestAuthFlow -v 2>&1
```

Expected: PASS.

**Step 3: Run full Go test suite**

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  test ./... 2>&1
```

Expected: all pass (integration DB tests skip when no DB available).

**Step 4: Commit**

```bash
git add internal/adapters/primary/http/auth_flow_test.go
git commit -m "test(http): add full auth flow integration test (login → token → access)"
```

---

### Task 7: Frontend unit tests — extend api.test.ts for tracks and profile

**Files:**
- Modify: `frontend/lib/api.test.ts`

**Step 1: Add track and profile API tests**

Append to `frontend/lib/api.test.ts`:

```typescript
import { buildTracksApi, buildProfileApi } from "./api";

describe("buildTracksApi", () => {
  it("lists tracks for a vinyl", async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => [{ id: "t1", vinyl_id: "v1", title: "So What", position: 1, lyrics: "" }],
    });
    const api = buildTracksApi("http://test", "tok", mockFetch as typeof fetch);
    await api.list("v1");
    expect(mockFetch).toHaveBeenCalledWith(
      "http://test/api/v1/vinyls/v1/tracks",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("creates a track", async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: "t1", vinyl_id: "v1", title: "Freddie Freeloader", position: 2, lyrics: "" }),
    });
    const api = buildTracksApi("http://test", "tok", mockFetch as typeof fetch);
    await api.create("v1", { title: "Freddie Freeloader", position: 2 });
    expect(mockFetch).toHaveBeenCalledWith(
      "http://test/api/v1/vinyls/v1/tracks",
      expect.objectContaining({ method: "POST" })
    );
  });

  it("deletes a track", async () => {
    const mockFetch = jest.fn().mockResolvedValue({ ok: true, json: async () => ({}) });
    const api = buildTracksApi("http://test", "tok", mockFetch as typeof fetch);
    await api.delete("v1", "t1");
    expect(mockFetch).toHaveBeenCalledWith(
      "http://test/api/v1/vinyls/v1/tracks/t1",
      expect.objectContaining({ method: "DELETE" })
    );
  });
});

describe("buildProfileApi", () => {
  it("gets public profile", async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ username: "admin", display_name: "Natalia", bio: "", links: [], preferred_genres: [] }),
    });
    const api = buildProfileApi("http://test", "tok", mockFetch as typeof fetch);
    await api.get();
    expect(mockFetch).toHaveBeenCalledWith(
      "http://test/api/v1/profile",
      expect.objectContaining({ method: "GET" })
    );
  });

  it("updates profile", async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ username: "admin", display_name: "Natalia", bio: "Jazz fan", links: [], preferred_genres: ["Jazz"] }),
    });
    const api = buildProfileApi("http://test", "tok", mockFetch as typeof fetch);
    await api.update({ display_name: "Natalia", bio: "Jazz fan", links: [], preferred_genres: ["Jazz"] });
    expect(mockFetch).toHaveBeenCalledWith(
      "http://test/api/v1/profile",
      expect.objectContaining({ method: "PUT" })
    );
  });
});
```

Also verify that `buildTracksApi` and `buildProfileApi` are exported from `frontend/lib/api.ts`. If they are not, export them.

**Step 2: Run frontend tests**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npm test -- --passWithNoTests 2>&1 | tail -20
```

Expected: All tests PASS.

**Step 3: Commit**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
git add frontend/lib/api.test.ts frontend/lib/api.ts
git commit -m "test(frontend): add unit tests for tracks and profile API client"
```

---

### Task 8: E2E tests with Playwright

**Files:**
- Create: `frontend/e2e/auth.spec.ts`
- Create: `frontend/e2e/vinyls.spec.ts`
- Modify: `frontend/package.json` (add playwright dev dependency and script)
- Create: `frontend/playwright.config.ts`

**Step 1: Install Playwright**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npm install --save-dev @playwright/test
npx playwright install chromium
```

**Step 2: Create `playwright.config.ts`**

```typescript
import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  use: {
    baseURL: "http://localhost:3001",
    headless: true,
  },
  webServer: {
    command: "npm run dev",
    url: "http://localhost:3001",
    reuseExistingServer: true,
    timeout: 30000,
  },
});
```

**Step 3: Add script to `package.json`**

In the `"scripts"` block, add:
```json
"test:e2e": "playwright test"
```

**Step 4: Create auth E2E test**

```typescript
// frontend/e2e/auth.spec.ts
import { test, expect } from "@playwright/test";

test.describe("Authentication", () => {
  test("shows login page when not authenticated", async ({ page }) => {
    await page.goto("/vinyls");
    await expect(page).toHaveURL(/\/login/);
  });

  test("redirects to /vinyls after successful login", async ({ page }) => {
    await page.goto("/login");
    await page.fill('input[name="username"]', "admin");
    await page.fill('input[name="password"]', "admin");
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/vinyls/);
  });

  test("shows error on wrong credentials", async ({ page }) => {
    await page.goto("/login");
    await page.fill('input[name="username"]', "admin");
    await page.fill('input[name="password"]', "wrong");
    await page.click('button[type="submit"]');
    await expect(page.locator("text=Invalid credentials")).toBeVisible({ timeout: 5000 });
  });

  test("profile page accessible without login", async ({ page }) => {
    await page.goto("/profile");
    await expect(page).not.toHaveURL(/\/login/);
    await expect(page).toHaveURL(/\/profile/);
  });
});
```

**Step 5: Create vinyl list E2E test**

```typescript
// frontend/e2e/vinyls.spec.ts
import { test, expect } from "@playwright/test";

test.use({ storageState: { cookies: [], origins: [] } });

async function loginAndGetCookie(page: any) {
  await page.goto("/login");
  await page.fill('input[name="username"]', "admin");
  await page.fill('input[name="password"]', "admin");
  await page.click('button[type="submit"]');
  await page.waitForURL(/\/vinyls/);
}

test.describe("Vinyl list (authenticated)", () => {
  test.beforeEach(async ({ page }) => {
    await loginAndGetCookie(page);
  });

  test("shows vinyl list page after login", async ({ page }) => {
    await expect(page.locator("h1, h2")).toBeVisible();
  });

  test("can navigate to add vinyl form", async ({ page }) => {
    // Look for a button/link to add a new vinyl
    const addButton = page.locator("text=Add, text=New, text=Create, button").first();
    await expect(addButton).toBeVisible({ timeout: 5000 });
  });
});
```

**Step 6: Run E2E tests (requires running docker-compose)**

Make sure `docker-compose up` is running first:

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npm run test:e2e 2>&1 | tail -30
```

Expected: Tests PASS (or note any selectors that need adjusting to match actual DOM).

**Step 7: Commit**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
git add frontend/playwright.config.ts frontend/e2e/ frontend/package.json
git commit -m "test(e2e): add Playwright E2E tests for auth flow and vinyl list"
```

---

### Task 9: Run full test suite and verify

**Step 1: Run all Go tests**

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  test ./... -v 2>&1 | grep -E "^(ok|FAIL|---)" | head -40
```

Expected: All `ok`, integration DB tests show `(cached)` or `SKIP`.

**Step 2: Run frontend unit tests**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npm test 2>&1 | tail -20
```

Expected: All PASS.

**Step 3: Build check**

```bash
GOROOT=/Users/natalia.granato/.local/share/mise/installs/go/1.25.7 \
  /Users/natalia.granato/.local/share/mise/installs/go/1.25.7/bin/go \
  build ./... 2>&1
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend && npm run build 2>&1 | tail -5
```

Expected: No errors.

**Step 4: Final commit if any fixes needed**

```bash
git add -A
git commit -m "test: all unit, integration and E2E tests passing"
```
