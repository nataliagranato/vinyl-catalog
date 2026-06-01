# Testing

This document describes the test suite for the Vinyl Catalog project, covering Go backend tests, frontend unit tests, and end-to-end (E2E) tests.

## Overview

The project follows a layered testing strategy aligned with the Clean Architecture boundaries: domain and application logic are tested in pure isolation, HTTP handlers are tested with in-memory mocks, database repositories are tested as integration tests (skipped when no PostgreSQL instance is available), and the frontend is covered by unit tests and Playwright E2E tests.

## Test Strategy

The test pyramid for this project:

```
          /\
         /E2E\          Playwright — browser-level flows
        /------\
       / Frontend\      Jest/Vitest — API client, filters, utilities
      /  Unit     \
     /-------------\
    / Go HTTP Layer \   handler_test, router_test, auth_flow_test
   /-----------------\
  /  Go Application   \ vinyl_service_test, track_service_test (mock repo)
 /---------------------\
/    Go Domain           \ vinyl_test (pure stdlib, no mocks)
\------------------------/
```

Integration tests (postgres layer) run separately and require a live database. They skip gracefully in CI without PostgreSQL.

## Go Tests (Backend)

All Go tests live under `internal/`. Run them with:

```bash
go test ./... -cover
# or via mise
mise run test
```

### Unit Tests

#### Domain — `internal/domain/vinyl_test.go`

Tests for the `Vinyl` entity constructor and validation rules. Zero external dependencies — uses stdlib only.

- Valid vinyl creation
- Rejection of blank title, blank artist, invalid year (future, too old), blank genre, blank label
- Total: 9 tests, 100% coverage

#### Application — `internal/application/vinyl_service_test.go`

Tests for `VinylService` use cases using an in-memory mock repository. Covers the full CRUD surface.

- Create, GetByID, List, Update, Delete
- Total: 4 tests, 100% coverage

#### Application — `internal/application/track_service_test.go`

Tests for `TrackService` use cases using an in-memory mock repository.

- ListTracks, CreateTrack, UpdateTrack, DeleteTrack
- Total: 4 tests, 100% coverage

#### Infrastructure / Auth — `internal/infrastructure/auth/jwt_test.go`

Tests for the JWT service: token generation and validation.

- GenerateToken produces a parseable token
- ValidateToken accepts a valid token
- ValidateToken rejects an expired token
- ValidateToken rejects a tampered token
- ValidateToken rejects a token signed with a different secret
- Total: 5 tests, ~83% coverage

#### Infrastructure / Config — `internal/infrastructure/config/config_test.go`

Tests for environment-based config loading via `joho/godotenv`.

- Default values are applied when env vars are absent
- Explicit env vars override defaults
- Total: 2 tests, 100% coverage

### HTTP Handler Tests

#### `internal/adapters/primary/http/handler_test.go`

Tests for `VinylHandler`: login and vinyl listing endpoints using `httptest`.

- `POST /api/v1/auth/login` returns 200 with token on valid credentials
- `GET /api/v1/vinyls` returns 200 with auth token
- `GET /api/v1/vinyls` returns 401 without token
- Total: 5 tests, ~30% coverage (handler surface is broad; remaining paths covered by auth_flow_test)

#### `internal/adapters/primary/http/router_test.go`

Tests for JWT middleware behaviour mounted on the router.

- Request with missing `Authorization` header is rejected (401)
- Request with an invalid token string is rejected (401)
- Request with a valid `Bearer <token>` is accepted (200)
- Request with a bare token (no `Bearer ` prefix) is accepted — Swagger UI compatibility

#### `internal/adapters/primary/http/track_handler_test.go`

Tests for track-related endpoints.

- `GET /api/v1/vinyls/:id/tracks` lists tracks for an existing vinyl
- `POST /api/v1/vinyls/:id/tracks` creates a track
- `GET /api/v1/vinyls/:id/tracks` returns 404 for a non-existent vinyl
- `DELETE /api/v1/vinyls/:id/tracks/:track_id` removes a track

#### `internal/adapters/primary/http/profile_handler_test.go`

Tests for profile endpoints.

- `GET /api/v1/profile` is publicly accessible (no auth required), returns 200
- `PUT /api/v1/profile` requires authentication, returns 401 without token

#### `internal/adapters/primary/http/auth_flow_test.go`

End-to-end auth flow test through the full Gin router stack (no HTTP server, uses `httptest`).

1. Login with valid credentials → 200 + token
2. Use token to access protected endpoint → 200
3. Access protected endpoint without token → 401
4. Access with a bare token (no `Bearer ` prefix) → 200
5. Login with wrong password → 401

### Integration Tests (PostgreSQL)

These tests connect to a real PostgreSQL instance. They call `t.Skip()` automatically when no database is available, so they never fail in a local or CI environment without a DB.

| File | What it tests |
|---|---|
| `internal/adapters/secondary/postgres/vinyl_repo_test.go` | Full CRUD lifecycle for `VinylRepository` |
| `internal/adapters/secondary/postgres/track_repo_test.go` | Full CRUD lifecycle for `TrackRepository` |
| `internal/adapters/secondary/postgres/profile_repo_test.go` | Get and Update for `ProfileRepository` |

To run integration tests locally, bring up the database first:

```bash
docker compose up -d db
go test ./internal/adapters/secondary/postgres/... -cover
```

## Frontend Tests (TypeScript)

Frontend tests live under `frontend/`. Run them with:

```bash
cd frontend && npm test
```

### Unit Tests

#### `frontend/lib/api.test.ts`

Tests for the API client factory functions that build typed fetch wrappers for each domain resource.

- `buildVinylsApi` — list, get, create, update, delete, uploadCover, toggleFavorite
- `buildTracksApi` — list, create, update, delete
- `buildProfileApi` — get, update, uploadPhoto

Each function is tested with a mocked `fetch`, asserting the correct HTTP method, URL, and headers.

#### `frontend/lib/filterVinyls.test.ts`

Tests for the pure `filterVinyls` utility function used by the vinyl list page.

- Filter by search query (title, artist)
- Filter by genre
- Filter by year
- Combined filters
- Empty result set when no match

#### `frontend/lib/vinylColor.test.ts`

Tests for the deterministic color function that maps a vinyl ID to a consistent Tailwind color class.

- Same ID always produces the same color (determinism)
- Different IDs produce valid color values from the palette
- Edge cases: empty string, very long string

## E2E Tests (Playwright)

E2E tests live under `frontend/e2e/`. They require the full application stack to be running.

```bash
# Start the full stack
docker compose up -d

# Run E2E tests
cd frontend && npx playwright test
```

### `frontend/e2e/auth.spec.ts`

Browser-level auth flows:

- Unauthenticated user is redirected to `/login`
- Successful login with valid credentials navigates to the main page
- Invalid credentials display an error message
- Public profile page (`/profile`) is accessible without login

### `frontend/e2e/vinyls.spec.ts`

Vinyl list page flows (authenticated):

- Vinyl list page loads and displays the catalog
- "Add vinyl" button is visible to authenticated users

## How to Run

### All Go tests

```bash
# Run all tests with coverage output
go test ./... -cover

# Via mise
mise run test

# Generate HTML coverage report
mise run coverage
```

### Specific packages

```bash
go test ./internal/domain/... -v
go test ./internal/application/... -v
go test ./internal/adapters/primary/http/... -v
go test ./internal/infrastructure/... -v
```

### Frontend unit tests

```bash
cd frontend && npm test
```

### E2E tests

```bash
# Requires the full stack (docker compose up)
cd frontend && npx playwright test

# Run a specific spec
cd frontend && npx playwright test e2e/auth.spec.ts

# Open the Playwright report
cd frontend && npx playwright show-report
```

## Coverage

| Package | Tests | Coverage |
|---|---|---|
| `internal/domain` | 9 | 100% |
| `internal/application` | 8 | 100% |
| `internal/infrastructure/auth` | 5 | ~83% |
| `internal/infrastructure/config` | 2 | 100% |
| `internal/adapters/primary/http` | 16 | ~30% (handler surface; core flows fully exercised) |
| `internal/adapters/secondary/postgres` | 3 (skip without DB) | — |
| `frontend/lib` | ~15 | — |
| `frontend/e2e` | 6 | — |

The low percentage on the HTTP adapter package is expected: handler functions have many branches (validation, error paths, content-type handling) that are tested qualitatively through the integration and auth-flow tests rather than line-by-line unit coverage. The domain and application layers — where business logic lives — are at 100%.
