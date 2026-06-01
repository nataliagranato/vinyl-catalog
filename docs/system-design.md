# System Design — Vinyl Catalog

## Overview

Vinyl Catalog is a full-stack application for cataloging vinyl records. It exposes a Go REST API backed by PostgreSQL and a Next.js frontend with a dark theme ("Dark Groove"). The system supports a single authenticated administrator who manages a catalog of vinyl records, their tracks, lyrics, and a public profile page.

The backend follows Clean Architecture + Hexagonal (Ports & Adapters). The frontend uses Next.js App Router with Server Components for data fetching and Client Components for interactive UI. The full stack is containerized with Docker Compose and ships with a complete observability stack (metrics, traces, logs, dashboards, and alerts).

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Browser / Client                           │
└────────────────────────────┬────────────────────────────────────────┘
                             │ HTTP (port 3001)
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Next.js Frontend (port 3001)                     │
│                                                                     │
│  Server Components          Route Handlers (/app/api/*)             │
│  (data fetching, SSR)       (proxy → backend, inject auth cookie)   │
│                                                                     │
│  Client Components          JWT stored in httpOnly cookie           │
│  (interactive UI, Framer    NEXT_PUBLIC_API_URL (browser uploads)   │
│   Motion, favorites,        API_URL (server-side, container-to-     │
│   lyrics translator)         container)                             │
└────────────────────────────┬────────────────────────────────────────┘
                             │ HTTP (port 8080)
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Go REST API (port 8080)                        │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  Adapters / Primary (HTTP)                                   │   │
│  │  Gin router — JWT middleware — VinylHandler, TrackHandler,   │   │
│  │  ProfileHandler — Swagger UI — Prometheus /metrics           │   │
│  └───────────────────────┬─────────────────────────────────────┘   │
│                          │                                          │
│  ┌───────────────────────▼─────────────────────────────────────┐   │
│  │  Application (Use Cases)                                     │   │
│  │  VinylService, TrackService, ProfileService                  │   │
│  └───────────────────────┬─────────────────────────────────────┘   │
│                          │                                          │
│  ┌───────────────────────▼─────────────────────────────────────┐   │
│  │  Domain (Entities + Interfaces)                              │   │
│  │  Vinyl, Track, Profile — VinylRepository, TrackRepository,  │   │
│  │  ProfileRepository (interfaces)                              │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                          │                                          │
│  ┌───────────────────────▼─────────────────────────────────────┐   │
│  │  Adapters / Secondary (PostgreSQL)                           │   │
│  │  GORM-backed implementations of domain repository interfaces │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  Infrastructure: auth (JWT), config (env), database, observability  │
└────────────────────────────┬────────────────────────────────────────┘
                             │
          ┌──────────────────┼──────────────────┐
          ▼                  ▼                  ▼
  ┌───────────────┐  ┌──────────────┐  ┌───────────────────────────┐
  │ PostgreSQL 16 │  │ uploads_data │  │  Observability Stack       │
  │ (port 5432)   │  │ (Docker vol) │  │                           │
  └───────────────┘  └──────────────┘  │  OTel Collector :4317     │
                                       │  Jaeger UI     :16686     │
                                       │  Prometheus    :9090      │
                                       │  Alertmanager  :9093      │
                                       │  Grafana       :3000      │
                                       │  Loki          :3100      │
                                       │  Promtail      (sidecar)  │
                                       └───────────────────────────┘
```

## Domain Model

The domain layer contains three entities and their repository interfaces. It has zero external imports — only Go's standard library.

### Vinyl

Represents a record in the catalog.

| Field | Type | Notes |
|---|---|---|
| id | string (UUID) | Supplied by handler layer at creation |
| title | string | Required |
| artist | string | Required |
| year | int | Required, 1900–current year |
| genre | string | Required |
| label | string | Required |
| description | string | Optional, free text |
| cover_url | string | Set after cover upload |
| created_at | time.Time | |
| updated_at | time.Time | |

### Track

Represents a track on a vinyl.

| Field | Type | Notes |
|---|---|---|
| id | string (UUID) | |
| vinyl_id | string (UUID) | Foreign key |
| title | string | Required |
| position | int | Track order on the record |
| lyrics | string | Optional, full lyrics text |
| created_at | time.Time | |
| updated_at | time.Time | |

### Profile

Represents the single administrator's public profile.

| Field | Type | Notes |
|---|---|---|
| id | string (UUID) | Single row |
| username | string | |
| display_name | string | |
| bio | string | |
| photo_url | string | Set after photo upload |
| links | []string | Stored as JSON array |
| preferred_genres | []string | Stored as JSON array |
| favorite_vinyl_ids | []string | Stored as comma-separated IDs |
| favorite_vinyls | []VinylWithTracks | Computed at read time, not persisted |

### Repository Interfaces (Ports)

```
VinylRepository   — Create, GetByID, List, Update, Delete
TrackRepository   — ListByVinylID, Create, Update, Delete
ProfileRepository — Get, Update
```

These interfaces are defined in the domain layer. The PostgreSQL GORM implementations live in the secondary adapters layer, satisfying the dependency inversion principle.

## API Design

### Base URL

```
http://localhost:8080/api/v1
```

### Authentication

JWT Bearer token. The middleware accepts both `Authorization: Bearer <token>` and `Authorization: <token>` (bare token, for Swagger UI compatibility).

### Endpoints

| Method | Path | Description | Auth |
|---|---|---|---|
| POST | /auth/login | Obtain JWT token | No |
| GET | /vinyls | List all vinyls | Yes |
| GET | /vinyls/:id | Get vinyl by ID | Yes |
| POST | /vinyls | Create vinyl | Yes |
| PUT | /vinyls/:id | Update vinyl | Yes |
| DELETE | /vinyls/:id | Delete vinyl | Yes |
| POST | /vinyls/:id/cover | Upload cover image | Yes |
| POST | /vinyls/:id/favorite | Toggle favorite (add/remove) | Yes |
| GET | /vinyls/:id/tracks | List tracks for a vinyl | Yes |
| POST | /vinyls/:id/tracks | Create track | Yes |
| PUT | /vinyls/:id/tracks/:track_id | Update track | Yes |
| DELETE | /vinyls/:id/tracks/:track_id | Delete track | Yes |
| GET | /profile | Public profile (with favorites + tracklists) | No |
| PUT | /profile | Update profile | Yes |
| POST | /profile/photo | Upload profile photo | Yes |
| GET | /health | Health check | No |
| GET | /metrics | Prometheus metrics | No |
| GET | /uploads/:filename | Static file server for uploads | No |
| GET | /swagger/index.html | Swagger UI | No |

### Error Responses

All error responses follow the format:

```json
{ "error": "human-readable message" }
```

HTTP status codes: 400 (validation), 401 (unauthenticated), 404 (not found), 500 (internal error).

## Data Flow Diagrams

### Login Flow

```
Client
  │  POST /api/v1/auth/login {"username","password"}
  ▼
Gin Router (no JWT middleware on this route)
  │
  ▼
VinylHandler.Login
  │  compare against ADMIN_USERNAME / ADMIN_PASSWORD env vars
  │
  ▼
auth.JWTService.GenerateToken(username)
  │
  ▼
200 OK {"token": "<jwt>"}
  │
Client stores token in httpOnly cookie (frontend)
or uses it directly in Authorization header (API clients)
```

### Create Vinyl Flow

```
Client
  │  POST /api/v1/vinyls  Authorization: Bearer <token>
  ▼
Gin Router
  │
  ▼
JWT Middleware
  │  ValidateToken → extract claims
  │
  ▼
VinylHandler.CreateVinyl
  │  parse JSON body → CreateVinylRequest DTO
  │  generate UUID (google/uuid)
  │  domain.NewVinyl(id, title, artist, year, genre, label)
  │
  ▼
VinylService.CreateVinyl(vinyl)
  │
  ▼
PostgresVinylRepository.Create(vinyl)
  │  GORM INSERT INTO vinyls
  │
  ▼
201 Created {vinyl JSON}
```

### GET /profile (Public — Favorites + Tracklist)

```
Client
  │  GET /api/v1/profile  (no auth required)
  ▼
Gin Router → public route, JWT middleware skipped
  │
  ▼
VinylHandler.GetProfile
  │
  ▼
ProfileRepository.Get()
  │  SELECT * FROM profiles LIMIT 1
  │
  ▼
for each id in profile.favorite_vinyl_ids:
  │  VinylService.GetVinyl(id)     → SELECT * FROM vinyls WHERE id=?
  │  TrackService.ListTracks(id)   → SELECT * FROM tracks WHERE vinyl_id=?
  │  assemble VinylWithTracks{Vinyl, Tracks}
  │
  ▼
200 OK { profile fields + favorite_vinyls: [VinylWithTracks...] }
```

### Toggle Favorite Flow

```
Client
  │  POST /api/v1/vinyls/:id/favorite  Authorization: Bearer <token>
  ▼
JWT Middleware → valid token
  │
  ▼
VinylHandler.ToggleFavorite
  │  ProfileRepository.Get() → load current profile
  │  if id in favorite_vinyl_ids → remove it
  │  else → append id
  │  ProfileRepository.Update(profile)
  │
  ▼
200 OK { favorited: true|false }
```

### Cover Upload Flow

```
Client
  │  POST /api/v1/vinyls/:id/cover  multipart/form-data  file=<image>
  ▼
JWT Middleware
  │
  ▼
VinylHandler.UploadCover
  │  read multipart file
  │  generate filename: cover-<uuid>.<ext>
  │  write to /uploads/ (Docker named volume: uploads_data)
  │  VinylService.Update(vinyl with cover_url="/uploads/<filename>")
  │
  ▼
200 OK { cover_url: "/uploads/cover-<uuid>.<ext>" }
```

## Frontend Architecture

### Stack

- Next.js 16 with App Router
- Tailwind CSS v4
- Framer Motion for animations
- TypeScript throughout
- OpenTelemetry Web SDK (browser traces → OTel Collector)

### Directory Structure

```
frontend/
  app/
    (auth)/login/          — login page (public)
    vinyls/                — catalog list + detail pages
    profile/               — public profile page
    api/                   — Route Handlers (server-side proxy)
      vinyls/
      tracks/
      profile/
      auth/
      translate/           — proxy para Google Translate (evita CORS e expõe logging)
  lib/
    api.ts                 — typed fetch wrappers (buildVinylsApi, etc.)
    filterVinyls.ts        — pure filter utility
    vinylColor.ts          — deterministic color palette for vinyl cards
    telemetry.ts           — OTel WebTracerProvider (browser → Collector :4318)
    translate.ts           — splitChunks + translateLyrics com span OTel
    logger.ts              — log() helper para JSON estruturado → stdout → Loki
  components/
    VinylCard              — card com botão de favorito no hover
    TrackList              — faixas colapsáveis com letras + tradutor
    ProfileFavorites       — favoritos com tracklist expansível
    LyricsTranslator       — seletor de idioma (PT/EN/ES/FR/DE/IT/JA), chama /api/translate
  e2e/                     — Playwright specs
```

### Request Routing

```
Browser fetch (NEXT_PUBLIC_API_URL)
  │  used for file uploads (multipart needs direct backend URL)
  ▼
Go backend :8080

Browser fetch to /app/api/* Route Handler
  │  server-side, runs in Node.js
  │  reads JWT from httpOnly cookie
  │  adds Authorization header
  ▼
Go backend :8080 (via API_URL, container-to-container)
  │
  ▼
JSON response forwarded back to browser
```

### Authentication Model

- Login: `POST /api/v1/auth/login` → JWT stored as httpOnly cookie named `token`
- Route Handlers read the cookie and inject `Authorization: Bearer <token>` before proxying
- Server Components fetch data via Route Handlers (cookie forwarding)
- Client Components use `fetch('/api/...')` to hit Route Handlers
- File uploads use `NEXT_PUBLIC_API_URL` directly (browser → backend), including cookie in request

### Key Frontend Features

- Vinyl list with real-time search and filter by genre and year
- Vinyl card hover reveals a heart button to toggle favorite status (instrumented with OTel span `favorite.toggle`)
- Vinyl detail page shows track list; click a track to expand its lyrics
- Cover upload with preview (instrumented with OTel span `cover.upload`)
- Lyrics translator: botão "Traduzir" com seletor de idioma (PT, EN, ES, FR, DE, IT, JA); chama o proxy interno `/api/translate` (Google Translate não-oficial), instrumentado com span `translate.lyrics`
- Public profile page (`/profile`) shows favorites gallery with expandable tracklist per vinyl
- Cover image and profile photo served from `/uploads/:filename` on the Go backend

## Infrastructure

### Docker Compose Services

| Service | Image | Port | Role |
|---|---|---|---|
| app | custom Go build | 8080 | REST API |
| frontend | custom Next.js build | 3001 | Web UI |
| db | postgres:16 | 5432 | Primary database |
| otel-collector | otel/opentelemetry-collector | 4317, 4318 | Telemetry pipeline |
| jaeger | jaegertracing/all-in-one | 16686 | Distributed tracing UI |
| prometheus | prom/prometheus | 9090 | Metrics scraping and storage |
| alertmanager | prom/alertmanager | 9093 | Alert routing |
| grafana | grafana/grafana | 3000 | Dashboards |
| loki | grafana/loki | 3100 | Log aggregation |
| promtail | grafana/promtail | — | Log collector (Docker logs → Loki) |

### Persistence

- PostgreSQL data: Docker named volume `postgres_data`
- Uploaded files (covers, profile photos): Docker named volume `uploads_data`, mounted at `/uploads` in the `app` container and served via `GET /uploads/:filename`

### Configuration

All configuration is environment-variable driven. See the `Variáveis de Ambiente` section in the README. The `app` container reads from a `.env` file at startup. The `infrastructure/config` package uses `joho/godotenv` to load it.

## Observability

### Metrics (Prometheus)

The Go application exposes a `/metrics` endpoint via `prometheus/client_golang`. Custom metrics recorded in `internal/infrastructure/observability`:

| Metric | Labels | Description |
|---|---|---|
| `http_requests_total` | method, path, status | Counter de todas as requisições HTTP |
| `http_request_duration_seconds` | method, path | Histograma de latência por rota |
| `vinyl_cover_uploads_total` | status, ext | Uploads de capa (success/error, .jpg/.png) |
| `vinyl_favorites_total` | action | Toggles de favorito (add/remove) |
| `vinyl_profile_photo_uploads_total` | status | Uploads de foto de perfil (success/error) |

Prometheus scrapes this endpoint every 15 seconds (default). Queries and recording rules are documented in `docs/prometheus-queries.md`.

### Traces (OpenTelemetry + Jaeger)

**Backend (Go):**
- Instrumented with `go.opentelemetry.io/otel` + `otelgin`
- Custom spans: `Handler.ToggleFavorite`, `Handler.GetProfile`, `Handler.UpdateProfile`, `Handler.UploadProfilePhoto`, `Handler.UploadCover`
- Exported via OTLP gRPC to OTel Collector on port 4317
- `trace_id` and `span_id` are injected into every structured log entry

**Frontend (Next.js/Browser):**
- `lib/telemetry.ts` initializes `WebTracerProvider` once per browser session
- Custom spans: `translate.lyrics`, `favorite.toggle`, `cover.upload`
- Exported via OTLP/HTTP to OTel Collector on port 4318 (CORS enabled)

Trace pipeline:

```
Browser (OTel Web SDK)
  │  OTLP/HTTP :4318 (CORS)
  ▼
OTel Collector
  │  exporters: jaeger
  ▼
Jaeger (UI :16686)
  service: vinyl-catalog-frontend

Go API (otelgin)
  │  OTLP gRPC :4317
  ▼
OTel Collector
  │  exporters: jaeger
  ▼
Jaeger (UI :16686)
  service: vinyl-catalog
```

### Logs (Loki + Promtail)

**Backend (Go):** `log/slog` with JSON output. `trace_id` and `span_id` included in every log entry for correlation.

**Frontend (Next.js):** `lib/logger.ts` emits structured JSON to stdout on every API route call:
```json
{"timestamp":"2026-03-24T00:00:30Z","level":"info","route":"/api/translate","langpair":"en|pt-BR","chars":85,"status":200,"duration_ms":174}
```

Promtail collects stdout from all Docker containers automatically.

Log pipeline:

```
Go API (slog JSON stdout)           Next.js Route Handlers (logger.ts JSON stdout)
  │                                   │
  └──────────────┬────────────────────┘
                 │ Docker container logs
                 ▼
              Promtail
                 │  push API
                 ▼
              Loki (:3100)
                 │
              Grafana (datasource: Loki)
```

### Alerts (Alertmanager)

Alerts are defined in `prometheus/alerts.yml`:

**Grupo `vinyl-catalog`:**

| Alert | Condition | Severity |
|---|---|---|
| HighErrorRate | 5xx rate > 5% for 5 min | critical |
| HighLatency | p95 latency > 500ms for 10 min | warning |
| ServiceDown | no metrics received for 1 min | critical |
| HighRequestRate | RPS > 100 for 5 min | warning |
| NoRequestsReceived | RPS = 0 for 10 min | warning |
| P99HighLatency | p99 latency elevated | warning |
| AuthErrorRate | high rate of 401 responses | warning |
| HighClientErrorRate | high rate of 4xx responses | warning |

**Grupo `vinyl-catalog-integrations`** (group_wait: 30s via Alertmanager):

| Alert | Condition | Severity |
|---|---|---|
| TranslationQuotaExceeded | quota_exceeded > 0 in 5 min | warning |
| CoverUploadErrorRate | upload error rate > 10% for 5 min | warning |
| FavoriteEndpointErrors | 5xx on favorite endpoint for 2 min | warning |
| ProfilePhotoUploadErrors | > 3 upload errors in 5 min | warning |

### Grafana Dashboards

Five dashboards are provisioned automatically at startup:

- **Operacional** — RPS, error rate, latency p50/p95/p99, route breakdown, HTTP status distribution, uptime
- **Negócio** — vinyls created, CRUD operations per hour, most-accessed route, error rate per operation, translations totals, favorites, cover uploads by extension
- **Frontend — Integrações** — translation requests by status, p95 translation latency, quota errors, favorite toggles, cover/photo upload errors, Loki frontend logs
- **Traces** — recent traces, error traces, latency per operation, span volume; integrated with Jaeger datasource
- **Logs** — log volume by level (ERROR/WARN/INFO), live app log panel with JSON parsing, full stack logs

## Key Design Decisions

### Domain has zero external imports

The `internal/domain` package imports nothing outside the Go standard library. This makes the domain entities and business rules trivially testable without any mocking framework and ensures the core logic never leaks infrastructure concerns.

### UUID generation at the handler layer

The `domain.NewVinyl` constructor accepts an `id` parameter supplied by the caller. UUIDs are generated in the HTTP handler using `google/uuid` before constructing the domain entity. This keeps the domain free from external dependencies while still ensuring globally unique identifiers.

### Single admin user

There is no user management system. The admin username and password are configured via `ADMIN_USERNAME` and `ADMIN_PASSWORD` environment variables. This keeps the authentication surface minimal for a single-owner catalog.

### Favorites stored as comma-separated IDs

`favorite_vinyl_ids` is persisted as a comma-separated string in the `profiles` table rather than a join table. This avoids schema complexity for a feature that never exceeds a few dozen entries. The full `VinylWithTracks` objects are assembled at read time by joining the vinyl and track queries in the handler.

### Tracks and favorites embedded in GET /profile

The public profile endpoint returns the complete profile including fully resolved `favorite_vinyls` (vinyl details + track list per vinyl). This means a single unauthenticated request is sufficient to render the entire public profile page, avoiding client-side waterfalls.

### JWT middleware accepts bare tokens

The JWT middleware checks for `Authorization: Bearer <token>` first, then falls back to treating the entire header value as the token. This is intentional: the Swagger UI sends a bare token by default, and requiring users to type `Bearer ` manually is a common source of confusion.

### Uploads in Docker named volume

Cover images and profile photos are saved to `/uploads` inside the `app` container and served as static files via `GET /uploads/:filename`. The directory is backed by a Docker named volume (`uploads_data`) so uploads persist across container restarts without an external object store.

### Google Translate proxy for lyrics translation

Lyrics translation uses an internal Next.js route handler (`/api/translate`) as a server-side proxy to the unofficial Google Translate API (`translate.googleapis.com/translate_a/single?client=gtx`). This approach:
- Avoids CORS issues (the fetch happens server-side in Node.js)
- Adds structured logging with `langpair`, `chars`, `duration_ms`
- Emits OTel spans from the browser (`translate.lyrics`) for end-to-end tracing
- Handles quota errors (429) and upstream failures with user-friendly messages

The translated text is displayed inline and never persisted. Long lyrics are split into ≤450-character chunks and translated in parallel.

### Next.js Route Handlers as auth proxy

The frontend's `/app/api/*` Route Handlers run server-side in Node.js. They read the httpOnly JWT cookie (inaccessible to client-side JavaScript) and inject it as an `Authorization` header before forwarding requests to the Go backend. This pattern keeps the token secure while letting Client Components make authenticated requests through the same `/api` prefix.

## Security

| Concern | Approach |
|---|---|
| Authentication | JWT HS256, configurable expiration (default 24h), secret via env var |
| Token storage | httpOnly cookie in the browser — inaccessible to JavaScript |
| Admin credentials | Env vars (`ADMIN_USERNAME`, `ADMIN_PASSWORD`) — not hardcoded |
| JWT secret | Env var (`JWT_SECRET`) — must be changed from default in production |
| Database | Credentials via env vars; SSL mode configurable (`DB_SSLMODE=require` in production) |
| File uploads | Stored in a named volume; filenames are UUID-based to avoid path traversal |
| Public routes | Profile GET and health check are explicitly excluded from JWT middleware |
| CORS | Configured at the Gin router level |
| Swagger UI | Available in all environments; restrict access via reverse proxy in production |
