# Observability for All Integrations

**Date:** 2026-03-23
**Status:** Approved
**Approach:** OpenTelemetry full-stack (Option A)

## Context

The vinyl-catalog stack already runs OTel Collector, Jaeger, Prometheus, Loki, and Grafana. The backend has partial instrumentation (vinyls CRUD has spans; favorites, profile, and uploads are missing spans and custom metrics). The frontend (Next.js) has zero telemetry. The translation integration (`/api/translate`) calls Google Translate with no visibility into latency, errors, or quota events.

## Goals

- Close all gaps in backend span coverage
- Add custom Prometheus metrics per integration (favorites, uploads, translation)
- Instrument frontend browser actions via OTel Web SDK → OTel Collector
- Emit structured JSON logs from Next.js route handlers → Promtail → Loki
- New Grafana dashboard for the frontend
- Update business dashboard with integration panels
- New Prometheus alerts for integration-level failures
- Alertmanager routing for integration alerts

## Architecture

```
Browser (OTel Web SDK)
  └─► OTLP/HTTP → OTel Collector :4318 (CORS enabled)
                      ├─► Jaeger   (traces)
                      ├─► Prometheus (metrics via prometheusremotewrite)
                      └─► Loki     (logs via loki exporter)

Next.js route handlers (JSON logs → stdout)
  └─► Promtail → Loki

Go backend (gaps closed)
  └─► OTLP/gRPC → OTel Collector :4317
```

### Infrastructure changes
- `otel-collector-config.yaml`: add `otlp/http` receiver on port 4318 with CORS for browser origin; add `loki` exporter for frontend logs
- `docker-compose.yml`: expose port 4318 externally

## Backend Go — Changes

### New spans (handler.go)
| Handler | Span name | Key attributes |
|---|---|---|
| ToggleFavorite | `Handler.ToggleFavorite` | `vinyl.id`, `favorite.action` (add/remove) |
| GetProfile | `Handler.GetProfile` | HTTP standard |
| UpdateProfile | `Handler.UpdateProfile` | HTTP standard |
| UploadProfilePhoto | `Handler.UploadProfilePhoto` | `file.extension`, `file.size_bytes` |
| UploadCover (existing) | add attrs | `file.extension`, `file.size_bytes` |

### New Prometheus metrics (metrics.go)
```
vinyl_cover_uploads_total{status, ext}
vinyl_favorites_total{action}
vinyl_profile_photo_uploads_total{status}
```

### Structured logs
All handlers without `logWithTrace` get `slog` calls with `trace_id` + `span_id`.

## Frontend Next.js — Changes

### Route handler logs (server-side)
Each API route emits structured JSON to stdout on every request/error:
```json
{"level":"info","route":"/api/translate","langpair":"en|pt-BR","duration_ms":320,"status":200}
{"level":"error","route":"/api/translate","error":"quota exceeded","status":429}
```
Promtail collects container stdout — no config change needed.

### Browser instrumentation (lib/telemetry.ts)
- Install: `@opentelemetry/sdk-web`, `@opentelemetry/exporter-trace-otlp-http`, `@opentelemetry/sdk-metrics`
- Tracer exports to `NEXT_PUBLIC_OTEL_ENDPOINT/v1/traces` (default: `http://localhost:4318`)
- Instrument:
  - `translateLyrics()` — span attrs: `langpair`, `chunk_count`, `error`
  - Favorite toggle — span attrs: `vinyl_id`, `action`
  - Cover/photo upload — span attrs: `file_size_bytes`, `file_type`

### Frontend metrics (exported via OTel Metrics → Collector)
```
frontend_translate_requests_total{status}       # success | error | quota_exceeded
frontend_favorite_toggles_total{action}         # add | remove
frontend_cover_uploads_total{status}            # success | error
frontend_profile_photo_uploads_total{status}
```

## Grafana

### New dashboard: frontend.json
Panels:
1. Translate requests by status — timeseries
2. P95 translation latency — stat
3. Translation quota errors (24h) — stat
4. Favorite toggles add/remove (24h) — stat + timeseries
5. Cover upload errors (24h) — stat
6. Frontend logs (Loki) — logs panel

### Updates to business.json
Add panels: Traduções totais, Favoritos totais, Uploads por extensão

## Prometheus — New Alerts

```yaml
- alert: TranslationQuotaExceeded
  expr: increase(frontend_translate_requests_total{status="quota_exceeded"}[5m]) > 0
  severity: warning

- alert: CoverUploadErrorRate
  expr: rate(vinyl_cover_uploads_total{status="error"}[5m]) /
        rate(vinyl_cover_uploads_total[5m]) > 0.10
  severity: warning

- alert: FavoriteEndpointErrors
  expr: rate(http_requests_total{path="/api/v1/vinyls/:id/favorite",status=~"5.."}[2m]) > 0
  severity: warning

- alert: TranslationServiceDown
  expr: rate(http_requests_total{path="/api/v1/vinyls/:id/tracks",status=~"5.."}[5m]) > 0.5
  severity: critical
```

## Alertmanager — New Route

Add an `integration` severity route that groups integration alerts separately from infrastructure alerts, with a shorter group_wait (30s vs 5m).

## Files to Create / Modify

| File | Action |
|---|---|
| `otel-collector-config.yaml` | add otlp/http receiver + loki exporter |
| `docker-compose.yml` | expose 4318 |
| `internal/infrastructure/observability/metrics.go` | add 3 new metrics |
| `internal/adapters/primary/http/handler.go` | add spans + slog to 5 handlers |
| `frontend/lib/telemetry.ts` | create OTel web initializer |
| `frontend/lib/translate.ts` | wrap in span |
| `frontend/app/api/translate/route.ts` | add structured logging |
| `frontend/app/api/vinyls/[id]/favorite/route.ts` | add structured logging |
| `frontend/app/api/vinyls/[id]/cover/route.ts` | add structured logging |
| `frontend/app/api/profile/photo/route.ts` | add structured logging |
| `frontend/app/vinyls/[id]/page.tsx` | instrument favorite + cover upload with spans |
| `prometheus/alerts.yml` | add 4 new alerts |
| `alertmanager/alertmanager.yml` | add integration route |
| `grafana/dashboards/frontend.json` | create new dashboard |
| `grafana/dashboards/business.json` | add integration panels |
