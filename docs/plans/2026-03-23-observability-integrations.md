# Observability for All Integrations — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add full OTel traces, Prometheus metrics, structured logs, Grafana dashboards and Prometheus alerts covering every integration (translate, favorites, cover/photo upload) in both the Go backend and the Next.js frontend.

**Architecture:** Browser → OTLP/HTTP (port 4318, CORS enabled) → OTel Collector → Jaeger + Prometheus. Next.js route handlers emit structured JSON to stdout → Promtail → Loki. Go backend gaps closed with spans + slog + custom Prometheus counters.

**Tech Stack:** Go `slog` + OTel Go SDK (already present), `@opentelemetry/sdk-trace-web` + `@opentelemetry/exporter-trace-otlp-http` (new), Prometheus promauto, Grafana JSON dashboards.

---

### Task 1: Enable CORS on OTel Collector HTTP receiver

Port 4318 is already exposed in docker-compose. Need CORS so the browser can POST traces to the collector.

**Files:**
- Modify: `otel-collector-config.yaml`

**Step 1: Add CORS config to the HTTP receiver**

Replace the `http` block under `receivers.otlp.protocols`:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
        cors:
          allowed_origins:
            - "http://localhost:3001"
            - "http://localhost:3000"
            - "http://*"
          allowed_headers:
            - "*"
```

**Step 2: Verify the collector restarts cleanly**

```bash
docker compose restart otel-collector
docker compose logs otel-collector --tail=20
```

Expected: no errors, `"Starting OTLP HTTP server"` log line.

**Step 3: Smoke test CORS preflight**

```bash
curl -sv -X OPTIONS http://localhost:4318/v1/traces \
  -H "Origin: http://localhost:3001" \
  -H "Access-Control-Request-Method: POST" 2>&1 | grep -E "< HTTP|access-control"
```

Expected: `HTTP/1.1 200` with `access-control-allow-origin` header.

**Step 4: Commit**

```bash
git add otel-collector-config.yaml
git commit -m "feat(otel): enable CORS on HTTP receiver for browser traces"
```

---

### Task 2: Add custom Prometheus metrics to the Go backend

**Files:**
- Modify: `internal/infrastructure/observability/metrics.go`

**Step 1: Add 3 new counters and their increment functions**

Append to `metrics.go` after the existing `httpRequestDuration` declaration:

```go
var (
	coverUploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vinyl_cover_uploads_total",
			Help: "Total cover image uploads",
		},
		[]string{"status", "ext"},
	)

	favoritesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vinyl_favorites_total",
			Help: "Total favorite toggle events",
		},
		[]string{"action"},
	)

	profilePhotoUploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vinyl_profile_photo_uploads_total",
			Help: "Total profile photo uploads",
		},
		[]string{"status"},
	)
)

func RecordCoverUpload(status, ext string) {
	coverUploadsTotal.WithLabelValues(status, ext).Inc()
}

func RecordFavoriteToggle(action string) {
	favoritesTotal.WithLabelValues(action).Inc()
}

func RecordProfilePhotoUpload(status string) {
	profilePhotoUploadsTotal.WithLabelValues(status).Inc()
}
```

**Step 2: Verify build**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
go build ./...
```

Expected: no errors.

**Step 3: Commit**

```bash
git add internal/infrastructure/observability/metrics.go
git commit -m "feat(metrics): add cover upload, favorites, and profile photo counters"
```

---

### Task 3: Add OTel spans and slog to missing Go handlers

**Files:**
- Modify: `internal/adapters/primary/http/handler.go`

**Step 1: Add span + metric increment to `ToggleFavorite`**

Replace the `ToggleFavorite` function signature and body opening:

```go
func (h *VinylHandler) ToggleFavorite(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.ToggleFavorite")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/vinyls/:id/favorite"),
		attribute.String("vinyl.id", id),
	)
```

After `favorited` is determined and before the final `c.JSON`, add:

```go
	action := "add"
	if favorited {
		action = "remove"
	}
	span.SetAttributes(attribute.String("favorite.action", action))
	observability.RecordFavoriteToggle(action)
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
```

**Step 2: Add span to `GetProfile`**

At the start of `GetProfile`:

```go
func (h *VinylHandler) GetProfile(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.GetProfile")
	defer span.End()
	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/profile"),
	)
```

Replace `c.Request.Context()` calls inside the function body with `ctx`.

On the error path add:
```go
	span.SetStatus(codes.Error, err.Error())
	logWithTrace(ctx, "failed to get profile", err)
```

On success:
```go
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
```

**Step 3: Add span to `UpdateProfile`**

At the start of `UpdateProfile`:

```go
func (h *VinylHandler) UpdateProfile(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.UpdateProfile")
	defer span.End()
	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/profile"),
	)
```

On error paths add `span.SetStatus(codes.Error, err.Error())` and `logWithTrace(ctx, "...", err)`. On success add `span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))`.

**Step 4: Add span + metrics to `UploadProfilePhoto`**

At the start:

```go
func (h *VinylHandler) UploadProfilePhoto(c *gin.Context) {
	ctx, span := otel.Tracer("vinyl-catalog").Start(c.Request.Context(), "Handler.UploadProfilePhoto")
	defer span.End()
	span.SetAttributes(
		semconv.HTTPMethod(c.Request.Method),
		semconv.HTTPRoute("/api/v1/profile/photo"),
	)
	_ = ctx
```

After `ext` is resolved, add:
```go
	span.SetAttributes(attribute.String("file.extension", ext))
```

On each error return, add before the `c.JSON`:
```go
	observability.RecordProfilePhotoUpload("error")
	span.SetStatus(codes.Error, "upload failed")
```

Before the final `c.JSON(http.StatusOK, ...)`:
```go
	observability.RecordProfilePhotoUpload("success")
	span.SetAttributes(semconv.HTTPStatusCode(http.StatusOK))
```

**Step 5: Add file size + ext attrs to existing `UploadCover`**

After `ext` is resolved in `UploadCover`, add:
```go
	span.SetAttributes(attribute.String("file.extension", ext))
```

After `io.Copy` succeeds, add:
```go
	span.SetAttributes(attribute.Int64("file.size_bytes", header.Size))
	observability.RecordCoverUpload("success", ext)
```

On each error return in `UploadCover`, add:
```go
	observability.RecordCoverUpload("error", ext)
```

**Step 6: Add the `attribute` import**

In the import block of `handler.go`, add:
```go
"go.opentelemetry.io/otel/attribute"
```

Also add the observability import:
```go
"github.com/nataliagranato/vinyl-catalog/internal/infrastructure/observability"
```

**Step 7: Verify build and run tests**

```bash
go build ./...
go test ./internal/adapters/primary/http/... -v
```

Expected: all existing tests pass, no compilation errors.

**Step 8: Commit**

```bash
git add internal/adapters/primary/http/handler.go
git commit -m "feat(traces): add OTel spans and metrics to favorites, profile, and upload handlers"
```

---

### Task 4: Install OTel packages in the Next.js frontend

**Files:**
- Modify: `frontend/package.json` (via npm install)

**Step 1: Install packages**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npm install \
  @opentelemetry/api \
  @opentelemetry/sdk-trace-web \
  @opentelemetry/sdk-trace-base \
  @opentelemetry/exporter-trace-otlp-http \
  @opentelemetry/resources \
  @opentelemetry/semantic-conventions
```

**Step 2: Verify TypeScript compiles**

```bash
npx tsc --noEmit
```

Expected: no errors.

**Step 3: Commit**

```bash
git add package.json package-lock.json
git commit -m "feat(frontend): install OpenTelemetry web SDK packages"
```

---

### Task 5: Create frontend OTel initializer

**Files:**
- Create: `frontend/lib/telemetry.ts`

**Step 1: Write the file**

```typescript
import { WebTracerProvider } from "@opentelemetry/sdk-trace-web";
import { SimpleSpanProcessor } from "@opentelemetry/sdk-trace-base";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import { Resource } from "@opentelemetry/resources";
import { SEMRESATTRS_SERVICE_NAME, SEMRESATTRS_SERVICE_VERSION } from "@opentelemetry/semantic-conventions";
import { trace, type Tracer } from "@opentelemetry/api";

const SERVICE_NAME = "vinyl-catalog-frontend";

function initTelemetry(): void {
  if (typeof window === "undefined") return; // SSR guard

  const endpoint =
    (process.env.NEXT_PUBLIC_OTEL_ENDPOINT ?? "http://localhost:4318") +
    "/v1/traces";

  const exporter = new OTLPTraceExporter({ url: endpoint });

  const provider = new WebTracerProvider({
    resource: new Resource({
      [SEMRESATTRS_SERVICE_NAME]: SERVICE_NAME,
      [SEMRESATTRS_SERVICE_VERSION]: "1.0.0",
    }),
  });

  provider.addSpanProcessor(new SimpleSpanProcessor(exporter));
  provider.register();
}

// Initialize once per browser session
const g = globalThis as typeof globalThis & { __otelInit?: boolean };
if (!g.__otelInit) {
  g.__otelInit = true;
  initTelemetry();
}

export function getTracer(): Tracer {
  return trace.getTracer(SERVICE_NAME);
}
```

**Step 2: Verify TypeScript compiles**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npx tsc --noEmit
```

Expected: no errors.

**Step 3: Commit**

```bash
git add lib/telemetry.ts
git commit -m "feat(frontend): add OTel web tracer provider"
```

---

### Task 6: Instrument translateLyrics with a span

**Files:**
- Modify: `frontend/lib/translate.ts`

**Step 1: Update the file**

Add the import at the top:
```typescript
import { getTracer } from "./telemetry";
import { SpanStatusCode } from "@opentelemetry/api";
```

Replace the `translateLyrics` function body:

```typescript
export async function translateLyrics(
  text: string,
  sourceLang: string,
  targetLang: string
): Promise<string> {
  if (sourceLang === targetLang) return text;

  const chunks = splitChunks(text);
  const tracer = getTracer();

  return tracer.startActiveSpan("translate.lyrics", async (span) => {
    span.setAttributes({
      "translate.source_lang": sourceLang,
      "translate.target_lang": targetLang,
      "translate.chunk_count": chunks.length,
      "translate.char_count": text.length,
    });
    try {
      const results = await Promise.all(
        chunks.map(async (chunk) => {
          const url = `/api/translate?q=${encodeURIComponent(chunk)}&langpair=${sourceLang}|${targetLang}`;
          const res = await fetch(url);
          if (res.status === 429) throw new Error("Quota de tradução esgotada. Tente mais tarde.");
          if (!res.ok) throw new Error("Serviço de tradução indisponível.");
          const data = await res.json();
          if (data.error) throw new Error(data.error);
          return (data.translatedText as string) ?? chunk;
        })
      );
      span.setAttributes({ "translate.status": "success" });
      span.setStatus({ code: SpanStatusCode.OK });
      return results.join("\n");
    } catch (e) {
      const isQuota = e instanceof Error && e.message.includes("Quota");
      span.setAttributes({
        "translate.status": isQuota ? "quota_exceeded" : "error",
        "translate.error": e instanceof Error ? e.message : String(e),
      });
      span.setStatus({ code: SpanStatusCode.ERROR, message: String(e) });
      throw e;
    } finally {
      span.end();
    }
  });
}
```

**Step 2: Verify TypeScript compiles**

```bash
npx tsc --noEmit
```

Expected: no errors.

**Step 3: Commit**

```bash
git add lib/translate.ts
git commit -m "feat(frontend): instrument translateLyrics with OTel span"
```

---

### Task 7: Instrument favorite toggle and cover upload in the vinyl detail page

**Files:**
- Modify: `frontend/app/vinyls/[id]/page.tsx`

**Step 1: Add telemetry import at the top of the file**

```typescript
import { getTracer } from "@/lib/telemetry";
import { SpanStatusCode } from "@opentelemetry/api";
```

**Step 2: Wrap the cover upload `onChange` handler with a span**

Replace the `onChange` body inside the cover `<input type="file">`:

```typescript
onChange={async (e) => {
  const file = e.target.files?.[0];
  if (!file) return;
  const tracer = getTracer();
  await tracer.startActiveSpan("cover.upload", async (span) => {
    span.setAttributes({
      "upload.file_size_bytes": file.size,
      "upload.file_type": file.type,
      "vinyl.id": id,
    });
    try {
      const fd = new FormData();
      fd.append("file", file);
      const res = await fetch(`/api/vinyls/${id}/cover`, { method: "POST", body: fd });
      if (res.ok) {
        const data = await res.json();
        setVinyl((prev) => prev ? { ...prev, cover_url: data.cover_url } : prev);
        toast("Cover updated", "success");
        span.setAttributes({ "upload.status": "success" });
        span.setStatus({ code: SpanStatusCode.OK });
      } else {
        toast("Failed to upload cover", "error");
        span.setAttributes({ "upload.status": "error", "upload.http_status": res.status });
        span.setStatus({ code: SpanStatusCode.ERROR, message: "upload failed" });
      }
    } catch (err) {
      span.setStatus({ code: SpanStatusCode.ERROR, message: String(err) });
      toast("Failed to upload cover", "error");
    } finally {
      span.end();
    }
  });
}}
```

**Step 3: Wrap the favorite toggle in TrackItem (or wherever ToggleFavorite is called)**

In `TrackItem`, the favorite is triggered from `VinylCard`. In this page, the cover upload is the main client-side integration. The favorite toggle goes through `/api/vinyls/${id}/favorite` which is already a Next.js route handler — the span from `LyricsTranslator` covers client translate. The favorite is called via `VinylCard` component. Check `VinylCard.tsx` for the toggle call and add a span there if it exists. If the favorite fetch is in this page, wrap it similarly.

Check `frontend/components/VinylCard.tsx` for the favorite toggle fetch and wrap it:

```typescript
const tracer = getTracer();
await tracer.startActiveSpan("favorite.toggle", async (span) => {
  span.setAttributes({ "vinyl.id": id, "favorite.action": currentlyFavorited ? "remove" : "add" });
  try {
    const res = await fetch(`/api/vinyls/${id}/favorite`, { method: "POST" });
    // ... existing logic
    span.setStatus({ code: SpanStatusCode.OK });
  } catch (e) {
    span.setStatus({ code: SpanStatusCode.ERROR });
    throw e;
  } finally {
    span.end();
  }
});
```

**Step 4: Read VinylCard.tsx first, then apply the correct wrapping**

```bash
cat frontend/components/VinylCard.tsx
```

Apply the span wrapping to wherever `fetch('/api/vinyls/${id}/favorite')` is called.

**Step 5: Verify TypeScript compiles**

```bash
npx tsc --noEmit
```

**Step 6: Commit**

```bash
git add app/vinyls/[id]/page.tsx components/VinylCard.tsx
git commit -m "feat(frontend): instrument cover upload and favorite toggle with OTel spans"
```

---

### Task 8: Add structured JSON logging to Next.js API routes

**Files:**
- Create: `frontend/lib/logger.ts`
- Modify: `frontend/app/api/translate/route.ts`
- Modify: `frontend/app/api/vinyls/[id]/favorite/route.ts`
- Modify: `frontend/app/api/vinyls/[id]/cover/route.ts`
- Modify: `frontend/app/api/profile/photo/route.ts`

**Step 1: Create the log helper**

```typescript
// frontend/lib/logger.ts
type LogLevel = "info" | "warn" | "error";

export function log(level: LogLevel, route: string, data: Record<string, unknown>): void {
  console.log(
    JSON.stringify({
      timestamp: new Date().toISOString(),
      level,
      route,
      ...data,
    })
  );
}
```

**Step 2: Update translate route**

Wrap the handler body to time the request and log:

```typescript
import { log } from "@/lib/logger";

export async function GET(req: NextRequest) {
  const start = Date.now();
  const { searchParams } = req.nextUrl;
  const q = searchParams.get("q");
  const langpair = searchParams.get("langpair");

  if (!q || !langpair) {
    log("warn", "/api/translate", { error: "missing params", langpair });
    return NextResponse.json({ error: "Missing q or langpair" }, { status: 400 });
  }

  const [sl, tl] = langpair.split("|");
  const url = new URL("https://translate.googleapis.com/translate_a/single");
  url.searchParams.set("client", "gtx");
  url.searchParams.set("sl", sl);
  url.searchParams.set("tl", tl);
  url.searchParams.set("dt", "t");
  url.searchParams.set("q", q);

  const res = await fetch(url.toString(), { headers: { "User-Agent": "Mozilla/5.0" } });

  if (res.status === 429) {
    log("warn", "/api/translate", { langpair, status: 429, error: "quota_exceeded", duration_ms: Date.now() - start });
    return NextResponse.json({ error: "Quota de tradução esgotada. Tente mais tarde." }, { status: 429 });
  }
  if (!res.ok) {
    log("error", "/api/translate", { langpair, status: res.status, error: "upstream_error", duration_ms: Date.now() - start });
    return NextResponse.json({ error: "Serviço de tradução indisponível." }, { status: 502 });
  }

  const data: [[string, string][]] = await res.json();
  const translated = data[0].map((seg) => seg[0]).join("");
  log("info", "/api/translate", { langpair, chars: q.length, status: 200, duration_ms: Date.now() - start });
  return NextResponse.json({ translatedText: translated });
}
```

**Step 3: Update favorite route**

```typescript
import { log } from "@/lib/logger";

export async function POST(
  _req: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const start = Date.now();
  const { id } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value ?? "";

  const res = await fetch(`${process.env.API_URL}/api/v1/vinyls/${id}/favorite`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
  });

  const data = await res.json();
  log(res.ok ? "info" : "error", "/api/vinyls/[id]/favorite", {
    vinyl_id: id,
    status: res.status,
    favorited: data?.favorited,
    duration_ms: Date.now() - start,
  });
  return NextResponse.json(data, { status: res.status });
}
```

**Step 4: Update cover route — read the file first**

```bash
cat frontend/app/api/vinyls/[id]/cover/route.ts
```

Add `log` import and timing at start/end of the handler, similar pattern to the favorite route above.

**Step 5: Update profile photo route — read the file first**

```bash
cat frontend/app/api/profile/photo/route.ts
```

Add `log` import and timing at start/end.

**Step 6: Verify TypeScript compiles**

```bash
npx tsc --noEmit
```

**Step 7: Commit**

```bash
git add lib/logger.ts app/api/translate/route.ts app/api/vinyls/[id]/favorite/route.ts \
  app/api/vinyls/[id]/cover/route.ts app/api/profile/photo/route.ts
git commit -m "feat(frontend): add structured JSON logging to API routes"
```

---

### Task 9: Add integration alerts to Prometheus

**Files:**
- Modify: `prometheus/alerts.yml`

**Step 1: Append new alert group**

Add a new group at the end of `prometheus/alerts.yml`:

```yaml
  - name: vinyl-catalog-integrations
    rules:
      - alert: TranslationQuotaExceeded
        expr: |
          increase(calls_total{span_name="translate.lyrics",
            "translate.status"="quota_exceeded"}[5m]) > 0
        for: 0m
        labels:
          severity: warning
          category: integration
        annotations:
          summary: "Translation quota exceeded"
          description: "Google Translate quota has been hit. Users cannot translate lyrics."

      - alert: CoverUploadErrorRate
        expr: |
          rate(vinyl_cover_uploads_total{status="error"}[5m]) /
          (rate(vinyl_cover_uploads_total[5m]) + 0.001) > 0.10
        for: 5m
        labels:
          severity: warning
          category: integration
        annotations:
          summary: "High cover upload error rate"
          description: "Cover upload error rate is {{ humanizePercentage $value }} (threshold: 10%)."

      - alert: FavoriteEndpointErrors
        expr: |
          rate(http_requests_total{path="/api/v1/vinyls/:id/favorite",status=~"5.."}[5m]) > 0
        for: 2m
        labels:
          severity: warning
          category: integration
        annotations:
          summary: "Favorite endpoint returning 5xx errors"
          description: "The favorite toggle endpoint has been failing for 2 minutes."

      - alert: ProfilePhotoUploadErrors
        expr: |
          rate(vinyl_profile_photo_uploads_total{status="error"}[5m]) /
          (rate(vinyl_profile_photo_uploads_total[5m]) + 0.001) > 0.10
        for: 5m
        labels:
          severity: warning
          category: integration
        annotations:
          summary: "High profile photo upload error rate"
          description: "Profile photo upload error rate is {{ humanizePercentage $value }} (threshold: 10%)."
```

**Step 2: Reload Prometheus config**

```bash
curl -X POST http://localhost:9090/-/reload
```

Expected: `200 OK`.

**Step 3: Verify alerts appear**

```bash
curl -s http://localhost:9090/api/v1/rules | python3 -m json.tool | grep "TranslationQuota\|CoverUpload\|FavoriteEndpoint\|ProfilePhoto"
```

Expected: all 4 alert names appear.

**Step 4: Commit**

```bash
git add prometheus/alerts.yml
git commit -m "feat(alerts): add integration-level alerts for translate, upload, and favorites"
```

---

### Task 10: Add integration route to Alertmanager

**Files:**
- Modify: `alertmanager/alertmanager.yml`

**Step 1: Add integration route**

Add a new route inside the `routes:` list, before the `severity = "critical"` route:

```yaml
    - matchers:
        - category = "integration"
      receiver: 'null'
      group_wait: 30s
      group_interval: 2m
      repeat_interval: 1h
```

**Step 2: Reload Alertmanager**

```bash
curl -X POST http://localhost:9093/-/reload
```

Expected: `200 OK`.

**Step 3: Commit**

```bash
git add alertmanager/alertmanager.yml
git commit -m "feat(alertmanager): add integration alert route with shorter group_wait"
```

---

### Task 11: Create the Grafana frontend dashboard

**Files:**
- Create: `grafana/dashboards/frontend.json`

**Step 1: Write the dashboard JSON**

```json
{
  "id": null,
  "uid": "frontend-integrations",
  "title": "Vinyl Catalog — Frontend",
  "tags": ["frontend", "integrations"],
  "timezone": "browser",
  "schemaVersion": 38,
  "version": 1,
  "refresh": "30s",
  "time": { "from": "now-3h", "to": "now" },
  "panels": [
    {
      "id": 1,
      "title": "Traduções por Status (1h)",
      "type": "timeseries",
      "gridPos": { "x": 0, "y": 0, "w": 12, "h": 8 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "targets": [
        {
          "expr": "sum by (translate_status) (increase(calls_total{span_name=\"translate.lyrics\"}[1h]))",
          "legendFormat": "{{translate_status}}"
        }
      ]
    },
    {
      "id": 2,
      "title": "P95 Latência de Tradução",
      "type": "stat",
      "gridPos": { "x": 12, "y": 0, "w": 6, "h": 8 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "options": { "reduceOptions": { "calcs": ["lastNotNull"] } },
      "fieldConfig": {
        "defaults": {
          "unit": "ms",
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 1000 },
              { "color": "red", "value": 3000 }
            ]
          }
        }
      },
      "targets": [
        {
          "expr": "histogram_quantile(0.95, sum by (le) (rate(duration_milliseconds_bucket{span_name=\"translate.lyrics\"}[5m])))",
          "legendFormat": "p95"
        }
      ]
    },
    {
      "id": 3,
      "title": "Erros de Quota (24h)",
      "type": "stat",
      "gridPos": { "x": 18, "y": 0, "w": 6, "h": 8 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "options": { "reduceOptions": { "calcs": ["sum"] } },
      "fieldConfig": {
        "defaults": {
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "red", "value": 1 }
            ]
          }
        }
      },
      "targets": [
        {
          "expr": "increase(calls_total{span_name=\"translate.lyrics\",\"translate.status\"=\"quota_exceeded\"}[24h])",
          "legendFormat": "quota exceeded"
        }
      ]
    },
    {
      "id": 4,
      "title": "Favoritos — Adicionar vs Remover (24h)",
      "type": "stat",
      "gridPos": { "x": 0, "y": 8, "w": 8, "h": 6 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "options": { "reduceOptions": { "calcs": ["sum"] } },
      "targets": [
        { "expr": "increase(vinyl_favorites_total{action=\"add\"}[24h])", "legendFormat": "Adicionados" },
        { "expr": "increase(vinyl_favorites_total{action=\"remove\"}[24h])", "legendFormat": "Removidos" }
      ]
    },
    {
      "id": 5,
      "title": "Uploads de Capa por Status (24h)",
      "type": "stat",
      "gridPos": { "x": 8, "y": 8, "w": 8, "h": 6 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "options": { "reduceOptions": { "calcs": ["sum"] } },
      "fieldConfig": {
        "defaults": {
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "red", "value": 1 }
            ]
          }
        }
      },
      "targets": [
        { "expr": "increase(vinyl_cover_uploads_total{status=\"success\"}[24h])", "legendFormat": "Sucesso" },
        { "expr": "increase(vinyl_cover_uploads_total{status=\"error\"}[24h])", "legendFormat": "Erro" }
      ]
    },
    {
      "id": 6,
      "title": "Uploads de Foto de Perfil (24h)",
      "type": "stat",
      "gridPos": { "x": 16, "y": 8, "w": 8, "h": 6 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "options": { "reduceOptions": { "calcs": ["sum"] } },
      "targets": [
        { "expr": "increase(vinyl_profile_photo_uploads_total{status=\"success\"}[24h])", "legendFormat": "Sucesso" },
        { "expr": "increase(vinyl_profile_photo_uploads_total{status=\"error\"}[24h])", "legendFormat": "Erro" }
      ]
    },
    {
      "id": 7,
      "title": "Logs do Frontend (Loki)",
      "type": "logs",
      "gridPos": { "x": 0, "y": 14, "w": 24, "h": 10 },
      "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
      "options": { "dedupStrategy": "none", "showLabels": true, "showTime": true, "sortOrder": "Descending" },
      "targets": [
        {
          "expr": "{container=\"vinyl-catalog-frontend-1\"} | json | line_format \"[{{.level}}] {{.route}} {{.status}} {{.duration_ms}}ms {{.error}}\"",
          "legendFormat": ""
        }
      ]
    },
    {
      "id": 8,
      "title": "Taxa de Erro de Tradução (%)",
      "type": "timeseries",
      "gridPos": { "x": 0, "y": 24, "w": 12, "h": 8 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": { "defaults": { "unit": "percentunit" } },
      "targets": [
        {
          "expr": "rate(calls_total{span_name=\"translate.lyrics\",status_code=\"ERROR\"}[5m]) / rate(calls_total{span_name=\"translate.lyrics\"}[5m])",
          "legendFormat": "error rate"
        }
      ]
    },
    {
      "id": 9,
      "title": "Uploads de Capa por Extensão (30d)",
      "type": "piechart",
      "gridPos": { "x": 12, "y": 24, "w": 12, "h": 8 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "targets": [
        {
          "expr": "sum by (ext) (increase(vinyl_cover_uploads_total{status=\"success\"}[30d]))",
          "legendFormat": "{{ext}}"
        }
      ]
    }
  ],
  "templating": {
    "list": [
      {
        "name": "DS_PROMETHEUS",
        "type": "datasource",
        "query": "prometheus",
        "current": { "text": "Prometheus", "value": "prometheus" }
      },
      {
        "name": "DS_LOKI",
        "type": "datasource",
        "query": "loki",
        "current": { "text": "Loki", "value": "loki" }
      }
    ]
  }
}
```

**Step 2: Commit**

```bash
git add grafana/dashboards/frontend.json
git commit -m "feat(grafana): add frontend integrations dashboard"
```

---

### Task 12: Update business.json dashboard with integration panels

**Files:**
- Modify: `grafana/dashboards/business.json`

**Step 1: Read the current highest panel ID in business.json**

```bash
grep '"id":' grafana/dashboards/business.json | grep -oE '[0-9]+' | sort -n | tail -1
```

**Step 2: Append 3 panels to the `panels` array**

Using IDs 30, 31, 32 (above current max), add after the last panel entry in the `panels` array:

```json
    {
      "id": 30,
      "title": "Traduções Totais (24h)",
      "type": "stat",
      "gridPos": { "x": 0, "y": 40, "w": 6, "h": 4 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "options": { "reduceOptions": { "calcs": ["sum"] } },
      "targets": [
        { "expr": "increase(calls_total{span_name=\"translate.lyrics\"}[24h])", "legendFormat": "Traduções" }
      ]
    },
    {
      "id": 31,
      "title": "Favoritos Adicionados (24h)",
      "type": "stat",
      "gridPos": { "x": 6, "y": 40, "w": 6, "h": 4 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "options": { "reduceOptions": { "calcs": ["sum"] } },
      "targets": [
        { "expr": "increase(vinyl_favorites_total{action=\"add\"}[24h])", "legendFormat": "Favoritos" }
      ]
    },
    {
      "id": 32,
      "title": "Uploads de Capa (30d)",
      "type": "stat",
      "gridPos": { "x": 12, "y": 40, "w": 6, "h": 4 },
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "options": { "reduceOptions": { "calcs": ["sum"] } },
      "targets": [
        { "expr": "increase(vinyl_cover_uploads_total{status=\"success\"}[30d])", "legendFormat": "Uploads" }
      ]
    }
```

**Step 3: Commit**

```bash
git add grafana/dashboards/business.json
git commit -m "feat(grafana): add translation, favorites, and upload panels to business dashboard"
```

---

### Task 13: Rebuild containers and smoke test

**Step 1: Rebuild backend and frontend**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
docker compose up -d --build app frontend
```

**Step 2: Verify backend compiles and serves metrics**

```bash
curl -s http://localhost:8080/metrics | grep "vinyl_cover_uploads\|vinyl_favorites\|vinyl_profile_photo"
```

Expected: 3 metric families appear (even with 0 values).

**Step 3: Trigger a translation to generate a trace**

Open the browser at `http://localhost:3001`, navigate to a vinyl with lyrics, click "Traduzir". Then check Jaeger:

```bash
curl -s "http://localhost:16686/api/traces?service=vinyl-catalog-frontend&limit=5" | python3 -m json.tool | grep "operationName"
```

Expected: `translate.lyrics` span appears.

**Step 4: Verify frontend logs in Loki**

```bash
curl -s -G "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={container="vinyl-catalog-frontend-1"} | json' \
  --data-urlencode 'limit=5' | python3 -m json.tool | grep "route"
```

Expected: `"/api/translate"` appears in log lines.

**Step 5: Verify Grafana frontend dashboard loads**

Open `http://localhost:3000` → Dashboards → "Vinyl Catalog — Frontend". All panels should render without errors (data may be sparse initially).

**Step 6: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix: post-integration observability adjustments"
```
