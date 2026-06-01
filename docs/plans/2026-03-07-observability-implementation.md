# Observability Enhancement Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enriquecer traces no Jaeger, documentar queries Prometheus, configurar Alertmanager com regras de alerta e adicionar Grafana com dois dashboards provisionados automaticamente.

**Architecture:** Melhorias no tracing Go (resource attributes + span status), infraestrutura Docker nova (alertmanager, grafana), provisioning declarativo do Grafana via arquivos YAML/JSON e regras de alerta em arquivo separado referenciado pelo Prometheus.

**Tech Stack:** Go OTel SDK, otel-collector-contrib, Prometheus Alertmanager, Grafana 10+, Docker Compose.

---

## Task 1: Enriquecer TracerProvider com resource attributes

**Files:**
- Modify: `internal/infrastructure/observability/tracing.go`
- Modify: `internal/infrastructure/config/config.go`

**Context:** O `NewTracerProvider` atual só passa `service.name`. Precisamos adicionar `service.version`, `deployment.environment`, detecção de host/OS/processo e sampler explícito.

**Step 1: Adicionar campos de versão/ambiente ao Config**

Editar `internal/infrastructure/config/config.go` — adicionar dois campos no struct `Config`:

```go
ServiceVersion string
```

E no `Load()`:

```go
ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),
```

**Step 2: Reescrever `internal/infrastructure/observability/tracing.go`**

```go
package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewTracerProvider(ctx context.Context, endpoint, serviceName, serviceVersion, environment string) (*sdktrace.TracerProvider, error) {
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironment(environment),
		),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithOS(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}
```

**Step 3: Atualizar `cmd/api/main.go` — passar novos parâmetros**

Localizar a chamada de `observability.NewTracerProvider` e alterar de:
```go
tp, err := observability.NewTracerProvider(ctx, cfg.OTELEndpoint, cfg.OTELServiceName)
```
para:
```go
tp, err := observability.NewTracerProvider(ctx, cfg.OTELEndpoint, cfg.OTELServiceName, cfg.ServiceVersion, cfg.AppEnv)
```

**Step 4: Verificar compilação**

```bash
go build ./... && echo "OK"
```
Expected: `OK`

**Step 5: Commit**

```bash
git add internal/infrastructure/observability/tracing.go \
        internal/infrastructure/config/config.go \
        cmd/api/main.go
git commit -m "feat(tracing): enrich resource with version, environment, host and OS attributes"
```

---

## Task 2: Enriquecer spans nos handlers com atributos HTTP e status de erro

**Files:**
- Modify: `internal/adapters/primary/http/handler.go`

**Context:** Os spans criados nos handlers não marcam erros nem adicionam atributos HTTP. O Jaeger fica com spans "vazios" sem contexto útil.

**Step 1: Adicionar imports necessários no handler.go**

Adicionar aos imports:
```go
"go.opentelemetry.io/otel/attribute"
"go.opentelemetry.io/otel/codes"
semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
```

**Step 2: Criar helper para enriquecer spans**

Adicionar função privada no final de `handler.go`:

```go
func setSpanHTTPAttributes(span trace.Span, method, route string, statusCode int) {
	span.SetAttributes(
		semconv.HTTPMethod(method),
		semconv.HTTPRoute(route),
		semconv.HTTPStatusCode(statusCode),
	)
	if statusCode >= 500 {
		span.SetStatus(codes.Error, http.StatusText(statusCode))
	} else {
		span.SetStatus(codes.Ok, "")
	}
}
```

Import necessário: `"go.opentelemetry.io/otel/trace"` já está indiretamente disponível via otel.

**Step 3: Aplicar helper em cada handler**

Em `ListVinyls`, após `c.Next()` equivalente (antes de cada `c.JSON`):
```go
// no início do handler, após criar o span:
defer func() { setSpanHTTPAttributes(span, c.Request.Method, "/api/v1/vinyls", c.Writer.Status()) }()
```

Aplicar o mesmo padrão em `GetVinyl`, `CreateVinyl`, `UpdateVinyl`, `DeleteVinyl`.

Para erros explícitos, marcar o span imediatamente:
```go
span.SetStatus(codes.Error, "vinyl not found")
span.SetAttributes(attribute.String("error.type", "not_found"))
```

**Step 4: Verificar compilação**

```bash
go get go.opentelemetry.io/otel/codes@latest
go build ./internal/adapters/primary/http/... && echo "OK"
```

**Step 5: Rodar testes**

```bash
go test ./internal/adapters/primary/http/... -v
```
Expected: 5/5 PASS

**Step 6: Commit**

```bash
git add internal/adapters/primary/http/handler.go go.mod go.sum
git commit -m "feat(tracing): add HTTP attributes and error status to handler spans"
```

---

## Task 3: Atualizar otel-collector-config.yaml com memory_limiter

**Files:**
- Modify: `otel-collector-config.yaml`

**Step 1: Reescrever `otel-collector-config.yaml`**

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: 256
    spike_limit_mib: 64

  batch:
    timeout: 1s
    send_batch_size: 1024

  resource:
    attributes:
      - key: collector.version
        value: "latest"
        action: insert

exporters:
  otlp/jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

  prometheus:
    endpoint: "0.0.0.0:8889"

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [otlp/jaeger]
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [prometheus]
```

**Step 2: Reiniciar otel-collector**

```bash
docker compose restart otel-collector
```
Expected: container reinicia sem erros

**Step 3: Commit**

```bash
git add otel-collector-config.yaml
git commit -m "feat(otel): add memory_limiter and resource processor to collector pipeline"
```

---

## Task 4: Documentação de queries Prometheus

**Files:**
- Create: `docs/prometheus-queries.md`

**Step 1: Criar `docs/prometheus-queries.md`**

```markdown
# Prometheus Queries — Vinyl Catalog

Documentação das queries disponíveis para consulta no Prometheus e Grafana.

**Métricas disponíveis:**
- `http_requests_total{method, path, status}` — counter de requisições HTTP
- `http_request_duration_seconds{method, path}` — histogram de latência
- `up{job}` — saúde do serviço

---

## Taxa de Requisições (RPS)

### RPS total (últimos 5 minutos)
```
rate(http_requests_total{job="vinyl-catalog"}[5m])
```

### RPS por rota
```
sum by (path) (rate(http_requests_total{job="vinyl-catalog"}[5m]))
```

### RPS por método HTTP
```
sum by (method) (rate(http_requests_total{job="vinyl-catalog"}[5m]))
```

### RPS apenas para erros 5xx
```
rate(http_requests_total{job="vinyl-catalog", status=~"5.."}[5m])
```

---

## Latência

### Latência média por rota
```
rate(http_request_duration_seconds_sum{job="vinyl-catalog"}[5m])
/
rate(http_request_duration_seconds_count{job="vinyl-catalog"}[5m])
```

### Percentil 50 (mediana) geral
```
histogram_quantile(0.50, sum by (le) (rate(http_request_duration_seconds_bucket{job="vinyl-catalog"}[5m])))
```

### Percentil 95
```
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{job="vinyl-catalog"}[5m])))
```

### Percentil 99
```
histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket{job="vinyl-catalog"}[5m])))
```

### Percentil 95 por rota
```
histogram_quantile(0.95,
  sum by (le, path) (
    rate(http_request_duration_seconds_bucket{job="vinyl-catalog"}[5m])
  )
)
```

---

## Erros

### Taxa de erros 4xx (últimos 5 minutos)
```
sum(rate(http_requests_total{job="vinyl-catalog", status=~"4.."}[5m]))
/
sum(rate(http_requests_total{job="vinyl-catalog"}[5m]))
```

### Taxa de erros 5xx (últimos 5 minutos)
```
sum(rate(http_requests_total{job="vinyl-catalog", status=~"5.."}[5m]))
/
sum(rate(http_requests_total{job="vinyl-catalog"}[5m]))
```

### Error ratio total (4xx + 5xx)
```
sum(rate(http_requests_total{job="vinyl-catalog", status=~"[45].."}[5m]))
/
sum(rate(http_requests_total{job="vinyl-catalog"}[5m]))
```

### Rotas com mais erros
```
topk(5, sum by (path) (rate(http_requests_total{job="vinyl-catalog", status=~"[45].."}[5m])))
```

---

## Disponibilidade

### Serviço está no ar
```
up{job="vinyl-catalog"}
```
Retorna `1` (up) ou `0` (down).

### SLO: percentual de requisições bem-sucedidas (últimas 1h)
```
sum(rate(http_requests_total{job="vinyl-catalog", status=~"2.."}[1h]))
/
sum(rate(http_requests_total{job="vinyl-catalog"}[1h]))
* 100
```

### Uptime (desde o último restart)
```
time() - process_start_time_seconds{job="vinyl-catalog"}
```

---

## Operações CRUD

### Requisições de criação (POST /vinyls)
```
rate(http_requests_total{job="vinyl-catalog", method="POST", path="/api/v1/vinyls"}[5m])
```

### Requisições de listagem (GET /vinyls)
```
rate(http_requests_total{job="vinyl-catalog", method="GET", path="/api/v1/vinyls"}[5m])
```

### Requisições de busca por ID (GET /vinyls/:id)
```
rate(http_requests_total{job="vinyl-catalog", method="GET", path="/api/v1/vinyls/:id"}[5m])
```

### Requisições de atualização (PUT /vinyls/:id)
```
rate(http_requests_total{job="vinyl-catalog", method="PUT", path="/api/v1/vinyls/:id"}[5m])
```

### Requisições de remoção (DELETE /vinyls/:id)
```
rate(http_requests_total{job="vinyl-catalog", method="DELETE", path="/api/v1/vinyls/:id"}[5m])
```

### Total acumulado de discos criados (aproximado)
```
increase(http_requests_total{job="vinyl-catalog", method="POST", path="/api/v1/vinyls", status="201"}[30d])
```

### Ranking de rotas mais acessadas
```
topk(10, sum by (path, method) (increase(http_requests_total{job="vinyl-catalog"}[24h])))
```

---

## Exemplos de Alertas

### Detectar alta latência manualmente
```
histogram_quantile(0.95,
  sum by (le) (rate(http_request_duration_seconds_bucket[5m]))
) > 0.5
```

### Detectar aumento súbito de erros
```
(
  sum(rate(http_requests_total{status=~"5.."}[5m]))
  /
  sum(rate(http_requests_total[5m]))
) > 0.05
```
```

**Step 2: Commit**

```bash
git add docs/prometheus-queries.md
git commit -m "docs: add comprehensive Prometheus query reference for vinyl-catalog"
```

---

## Task 5: Criar regras de alerta Prometheus e configurar Alertmanager

**Files:**
- Create: `prometheus/alerts.yml`
- Create: `alertmanager/alertmanager.yml`
- Modify: `prometheus.yml`
- Modify: `docker-compose.yml`

**Step 1: Criar diretórios**

```bash
mkdir -p prometheus alertmanager
```

**Step 2: Criar `prometheus/alerts.yml`**

```yaml
groups:
  - name: vinyl-catalog
    interval: 30s
    rules:

      - alert: ServiceDown
        expr: up{job="vinyl-catalog"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Vinyl Catalog API está fora do ar"
          description: "O serviço vinyl-catalog não responde há mais de 1 minuto."

      - alert: HighErrorRate
        expr: |
          (
            sum(rate(http_requests_total{job="vinyl-catalog", status=~"5.."}[5m]))
            /
            sum(rate(http_requests_total{job="vinyl-catalog"}[5m]))
          ) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Alta taxa de erros 5xx"
          description: "Taxa de erros 5xx acima de 5% nos últimos 5 minutos (atual: {{ $value | humanizePercentage }})."

      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            sum by (le) (
              rate(http_request_duration_seconds_bucket{job="vinyl-catalog"}[5m])
            )
          ) > 0.5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Latência p95 elevada"
          description: "Latência no percentil 95 acima de 500ms por mais de 10 minutos (atual: {{ $value | humanizeDuration }})."

      - alert: HighRequestRate
        expr: |
          sum(rate(http_requests_total{job="vinyl-catalog"}[5m])) > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Alta taxa de requisições"
          description: "API recebendo mais de 100 req/s por 5 minutos (atual: {{ $value | humanize }} req/s)."

      - alert: NoRequestsReceived
        expr: |
          sum(rate(http_requests_total{job="vinyl-catalog"}[10m])) == 0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Nenhuma requisição recebida"
          description: "A API não recebeu nenhuma requisição nos últimos 10 minutos. Pode indicar problema de roteamento ou inatividade."
```

**Step 3: Criar `alertmanager/alertmanager.yml`**

```yaml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'null'

receivers:
  - name: 'null'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname']
```

**Step 4: Atualizar `prometheus.yml`**

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

rule_files:
  - /etc/prometheus/alerts/*.yml

scrape_configs:
  - job_name: 'vinyl-catalog'
    static_configs:
      - targets: ['app:8080']
    metrics_path: /metrics

  - job_name: 'otel-collector'
    static_configs:
      - targets: ['otel-collector:8889']
```

**Step 5: Adicionar alertmanager ao `docker-compose.yml`**

Adicionar após o serviço `prometheus`:
```yaml
  alertmanager:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    restart: unless-stopped
```

E atualizar o serviço `prometheus` para montar a pasta de alertas e depender do alertmanager:
```yaml
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./prometheus:/etc/prometheus/alerts:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    depends_on:
      - alertmanager
    restart: unless-stopped
```

**Step 6: Reiniciar stack**

```bash
docker compose up -d alertmanager prometheus
```

**Step 7: Verificar alertas carregados**

```bash
curl -s http://localhost:9090/api/v1/rules | python3 -m json.tool | grep '"name"'
```
Expected: listar os 5 alertas configurados

**Step 8: Commit**

```bash
git add prometheus/ alertmanager/ prometheus.yml docker-compose.yml
git commit -m "feat(alerting): add Alertmanager and 5 alert rules for vinyl-catalog"
```

---

## Task 6: Adicionar Grafana com provisioning automático

**Files:**
- Create: `grafana/provisioning/datasources/datasources.yml`
- Create: `grafana/provisioning/dashboards/dashboards.yml`
- Create: `grafana/dashboards/operational.json`
- Create: `grafana/dashboards/business.json`
- Modify: `docker-compose.yml`

**Step 1: Criar estrutura de diretórios**

```bash
mkdir -p grafana/provisioning/datasources grafana/provisioning/dashboards grafana/dashboards
```

**Step 2: Criar `grafana/provisioning/datasources/datasources.yml`**

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false

  - name: Jaeger
    type: jaeger
    access: proxy
    url: http://jaeger:16686
    editable: false
```

**Step 3: Criar `grafana/provisioning/dashboards/dashboards.yml`**

```yaml
apiVersion: 1

providers:
  - name: vinyl-catalog
    orgId: 1
    folder: "Vinyl Catalog"
    type: file
    disableDeletion: false
    updateIntervalSeconds: 30
    options:
      path: /var/lib/grafana/dashboards
```

**Step 4: Criar `grafana/dashboards/operational.json`**

Dashboard operacional com 6 painéis:
- Request Rate (RPS) — timeseries
- Error Rate (%) — stat
- Latência p50/p95/p99 — timeseries
- Requisições por rota — table
- HTTP status distribution — piechart
- Uptime — stat

```json
{
  "__inputs": [],
  "__requires": [],
  "annotations": { "list": [] },
  "description": "Métricas operacionais da API Vinyl Catalog",
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 1,
  "id": null,
  "links": [],
  "panels": [
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "reqps" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
      "id": 1,
      "options": { "legend": { "calcs": ["mean", "max"], "displayMode": "list" }, "tooltip": { "mode": "multi" } },
      "targets": [
        {
          "expr": "sum by (path) (rate(http_requests_total{job=\"vinyl-catalog\"}[5m]))",
          "legendFormat": "{{path}}",
          "refId": "A"
        }
      ],
      "title": "Request Rate por Rota (RPS)",
      "type": "timeseries"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": {
          "color": { "mode": "thresholds" },
          "thresholds": { "mode": "absolute", "steps": [
            { "color": "green", "value": null },
            { "color": "yellow", "value": 0.01 },
            { "color": "red", "value": 0.05 }
          ]},
          "unit": "percentunit",
          "mappings": []
        },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 6, "x": 12, "y": 0 },
      "id": 2,
      "options": { "colorMode": "background", "graphMode": "area", "orientation": "auto", "reduceOptions": { "calcs": ["lastNotNull"] } },
      "targets": [
        {
          "expr": "sum(rate(http_requests_total{job=\"vinyl-catalog\",status=~\"5..\"}[5m])) / sum(rate(http_requests_total{job=\"vinyl-catalog\"}[5m]))",
          "legendFormat": "Error Rate",
          "refId": "A"
        }
      ],
      "title": "Error Rate (5xx)",
      "type": "stat"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": {
          "color": { "mode": "thresholds" },
          "thresholds": { "mode": "absolute", "steps": [
            { "color": "green", "value": null },
            { "color": "yellow", "value": 0.8 },
            { "color": "red", "value": 0 }
          ]},
          "unit": "short",
          "mappings": [{ "options": { "0": { "text": "DOWN" }, "1": { "text": "UP" } }, "type": "value" }]
        },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 6, "x": 18, "y": 0 },
      "id": 3,
      "options": { "colorMode": "background", "graphMode": "none", "orientation": "auto", "reduceOptions": { "calcs": ["lastNotNull"] } },
      "targets": [
        {
          "expr": "up{job=\"vinyl-catalog\"}",
          "legendFormat": "Status",
          "refId": "A"
        }
      ],
      "title": "Serviço Status",
      "type": "stat"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "s" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 8 },
      "id": 4,
      "options": { "legend": { "calcs": ["mean", "max"], "displayMode": "list" }, "tooltip": { "mode": "multi" } },
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum by (le) (rate(http_request_duration_seconds_bucket{job=\"vinyl-catalog\"}[5m])))",
          "legendFormat": "p50",
          "refId": "A"
        },
        {
          "expr": "histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket{job=\"vinyl-catalog\"}[5m])))",
          "legendFormat": "p95",
          "refId": "B"
        },
        {
          "expr": "histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket{job=\"vinyl-catalog\"}[5m])))",
          "legendFormat": "p99",
          "refId": "C"
        }
      ],
      "title": "Latência p50 / p95 / p99",
      "type": "timeseries"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": { "custom": { "align": "auto" }, "unit": "reqps" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 12, "x": 12, "y": 8 },
      "id": 5,
      "options": { "footer": { "show": false }, "showHeader": true, "sortBy": [{ "desc": true, "displayName": "Value" }] },
      "targets": [
        {
          "expr": "sort_desc(sum by (path, method) (rate(http_requests_total{job=\"vinyl-catalog\"}[5m])))",
          "format": "table",
          "instant": true,
          "legendFormat": "",
          "refId": "A"
        }
      ],
      "title": "Requisições por Rota e Método",
      "transformations": [{ "id": "organize", "options": { "excludeByName": { "Time": true, "__name__": true, "job": true }, "renameByName": { "Value": "RPS", "method": "Método", "path": "Rota" } } }],
      "type": "table"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": { "defaults": {}, "overrides": [] },
      "gridPos": { "h": 8, "w": 8, "x": 0, "y": 16 },
      "id": 6,
      "options": { "legend": { "displayMode": "table", "placement": "right", "values": ["percent"] }, "pieType": "pie" },
      "targets": [
        {
          "expr": "sum by (status) (increase(http_requests_total{job=\"vinyl-catalog\"}[1h]))",
          "legendFormat": "HTTP {{status}}",
          "refId": "A"
        }
      ],
      "title": "Distribuição de Status HTTP (última 1h)",
      "type": "piechart"
    }
  ],
  "refresh": "30s",
  "schemaVersion": 38,
  "tags": ["vinyl-catalog", "operational"],
  "templating": {
    "list": [
      {
        "hide": 2,
        "name": "DS_PROMETHEUS",
        "type": "datasource",
        "query": "prometheus"
      }
    ]
  },
  "time": { "from": "now-1h", "to": "now" },
  "timepicker": {},
  "timezone": "browser",
  "title": "Vinyl Catalog — Operacional",
  "uid": "vinyl-operational",
  "version": 1
}
```

**Step 5: Criar `grafana/dashboards/business.json`**

```json
{
  "__inputs": [],
  "__requires": [],
  "annotations": { "list": [] },
  "description": "Métricas de negócio da aplicação Vinyl Catalog",
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 1,
  "id": null,
  "links": [
    {
      "icon": "external link",
      "tags": [],
      "targetBlank": true,
      "title": "Jaeger Traces",
      "tooltip": "Ver traces no Jaeger",
      "type": "link",
      "url": "http://localhost:16686/search?service=vinyl-catalog"
    }
  ],
  "panels": [
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": {
          "color": { "mode": "thresholds" },
          "thresholds": { "mode": "absolute", "steps": [{ "color": "blue", "value": null }] },
          "unit": "short",
          "mappings": []
        },
        "overrides": []
      },
      "gridPos": { "h": 6, "w": 6, "x": 0, "y": 0 },
      "id": 1,
      "options": { "colorMode": "background", "graphMode": "area", "orientation": "auto", "reduceOptions": { "calcs": ["sum"] } },
      "targets": [
        {
          "expr": "increase(http_requests_total{job=\"vinyl-catalog\", method=\"POST\", path=\"/api/v1/vinyls\", status=\"201\"}[30d])",
          "legendFormat": "Discos criados",
          "refId": "A"
        }
      ],
      "title": "Discos Criados (30 dias)",
      "type": "stat"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "thresholds" }, "unit": "short",
          "thresholds": { "mode": "absolute", "steps": [{ "color": "green", "value": null }] }
        },
        "overrides": []
      },
      "gridPos": { "h": 6, "w": 6, "x": 6, "y": 0 },
      "id": 2,
      "options": { "colorMode": "background", "graphMode": "area", "reduceOptions": { "calcs": ["sum"] } },
      "targets": [
        {
          "expr": "increase(http_requests_total{job=\"vinyl-catalog\", method=\"GET\", path=\"/api/v1/vinyls\"}[24h])",
          "legendFormat": "Listagens",
          "refId": "A"
        }
      ],
      "title": "Listagens de Discos (24h)",
      "type": "stat"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "thresholds" }, "unit": "short",
          "thresholds": { "mode": "absolute", "steps": [
            { "color": "green", "value": null },
            { "color": "yellow", "value": 5 },
            { "color": "red", "value": 20 }
          ]}
        },
        "overrides": []
      },
      "gridPos": { "h": 6, "w": 6, "x": 12, "y": 0 },
      "id": 3,
      "options": { "colorMode": "background", "graphMode": "area", "reduceOptions": { "calcs": ["sum"] } },
      "targets": [
        {
          "expr": "increase(http_requests_total{job=\"vinyl-catalog\", status=~\"[45]..\"}[24h])",
          "legendFormat": "Erros",
          "refId": "A"
        }
      ],
      "title": "Total de Erros (24h)",
      "type": "stat"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "thresholds" }, "unit": "short",
          "thresholds": { "mode": "absolute", "steps": [{ "color": "purple", "value": null }] }
        },
        "overrides": []
      },
      "gridPos": { "h": 6, "w": 6, "x": 18, "y": 0 },
      "id": 4,
      "options": { "colorMode": "background", "graphMode": "area", "reduceOptions": { "calcs": ["sum"] } },
      "targets": [
        {
          "expr": "increase(http_requests_total{job=\"vinyl-catalog\", method=\"DELETE\", path=\"/api/v1/vinyls/:id\", status=\"204\"}[30d])",
          "legendFormat": "Removidos",
          "refId": "A"
        }
      ],
      "title": "Discos Removidos (30 dias)",
      "type": "stat"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "short" },
        "overrides": []
      },
      "gridPos": { "h": 9, "w": 12, "x": 0, "y": 6 },
      "id": 5,
      "options": { "barRadius": 0.1, "barWidth": 0.7, "groupWidth": 0.7, "legend": { "displayMode": "list", "placement": "bottom" }, "orientation": "auto", "tooltip": { "mode": "multi" }, "xTickLabelRotation": 0 },
      "targets": [
        {
          "expr": "sum by (method) (increase(http_requests_total{job=\"vinyl-catalog\", path=~\"/api/v1/vinyls.*\"}[1h]))",
          "legendFormat": "{{method}}",
          "refId": "A"
        }
      ],
      "title": "Operações CRUD por Hora",
      "type": "barchart"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": { "custom": { "align": "auto" }, "unit": "short" },
        "overrides": []
      },
      "gridPos": { "h": 9, "w": 12, "x": 12, "y": 6 },
      "id": 6,
      "options": { "footer": { "show": false }, "showHeader": true, "sortBy": [{ "desc": true, "displayName": "Acessos" }] },
      "targets": [
        {
          "expr": "sort_desc(sum by (path, method) (increase(http_requests_total{job=\"vinyl-catalog\"}[24h])))",
          "format": "table",
          "instant": true,
          "legendFormat": "",
          "refId": "A"
        }
      ],
      "title": "Rotas Mais Acessadas (24h)",
      "transformations": [{ "id": "organize", "options": { "excludeByName": { "Time": true, "__name__": true, "job": true }, "renameByName": { "Value": "Acessos", "method": "Método", "path": "Rota" } } }],
      "type": "table"
    },
    {
      "datasource": { "type": "prometheus", "uid": "${DS_PROMETHEUS}" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "short" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 24, "x": 0, "y": 15 },
      "id": 7,
      "options": { "legend": { "calcs": ["sum"], "displayMode": "list" }, "tooltip": { "mode": "multi" } },
      "targets": [
        {
          "expr": "sum by (path) (increase(http_requests_total{job=\"vinyl-catalog\", status=~\"[45]..\"}[5m]))",
          "legendFormat": "Erros — {{path}}",
          "refId": "A"
        }
      ],
      "title": "Erros por Endpoint",
      "type": "timeseries"
    }
  ],
  "refresh": "30s",
  "schemaVersion": 38,
  "tags": ["vinyl-catalog", "business"],
  "templating": {
    "list": [
      {
        "hide": 2,
        "name": "DS_PROMETHEUS",
        "type": "datasource",
        "query": "prometheus"
      }
    ]
  },
  "time": { "from": "now-24h", "to": "now" },
  "timepicker": {},
  "timezone": "browser",
  "title": "Vinyl Catalog — Negócio",
  "uid": "vinyl-business",
  "version": 1
}
```

**Step 6: Adicionar Grafana ao `docker-compose.yml`**

```yaml
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_USER: admin
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_USERS_ALLOW_SIGN_UP: "false"
      GF_FEATURE_TOGGLES_ENABLE: traceqlEditor
    volumes:
      - ./grafana/provisioning:/etc/grafana/provisioning:ro
      - ./grafana/dashboards:/var/lib/grafana/dashboards:ro
      - grafana_data:/var/lib/grafana
    depends_on:
      - prometheus
      - jaeger
    restart: unless-stopped
```

Adicionar `grafana_data:` na seção `volumes:` do docker-compose.

**Step 7: Subir Grafana**

```bash
docker compose up -d grafana
```

**Step 8: Verificar dashboards**

```bash
curl -s -u admin:admin http://localhost:3000/api/dashboards/uid/vinyl-operational | python3 -m json.tool | grep '"title"'
curl -s -u admin:admin http://localhost:3000/api/dashboards/uid/vinyl-business | python3 -m json.tool | grep '"title"'
```
Expected: `"title": "Vinyl Catalog — Operacional"` e `"title": "Vinyl Catalog — Negócio"`

**Step 9: Commit**

```bash
git add grafana/ docker-compose.yml
git commit -m "feat(grafana): add Grafana with auto-provisioned operational and business dashboards"
```

---

## Task 7: Atualizar README com novos serviços

**Files:**
- Modify: `README.md`

**Step 1: Adicionar seção de novos serviços**

Na tabela de serviços do README, adicionar:

| Serviço | URL |
|---|---|
| Alertmanager | http://localhost:9093 |
| Grafana | http://localhost:3000 (admin/admin) |

**Step 2: Adicionar seção de dashboards Grafana**

```markdown
## Grafana Dashboards

Acesse http://localhost:3000 com usuário `admin` e senha `admin`.

Dois dashboards disponíveis na pasta **Vinyl Catalog**:

- **Operacional** — RPS, latência p50/p95/p99, error rate, distribuição de status HTTP
- **Negócio** — operações CRUD, discos criados/removidos, rotas mais acessadas, erros por endpoint

Os datasources Prometheus e Jaeger são provisionados automaticamente.
```

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: update README with Alertmanager and Grafana service URLs"
```

---

## Task 8: Verificação final da stack completa

**Step 1: Reiniciar toda a stack**

```bash
docker compose down && docker compose up -d --build
```

**Step 2: Verificar todos os serviços**

```bash
docker compose ps
```
Expected: 7 serviços Up (app, db, otel-collector, jaeger, prometheus, alertmanager, grafana)

**Step 3: Verificar alertas carregados no Prometheus**

```bash
curl -s http://localhost:9090/api/v1/rules | python3 -m json.tool | grep '"name"'
```
Expected: ServiceDown, HighErrorRate, HighLatency, HighRequestRate, NoRequestsReceived

**Step 4: Verificar Alertmanager**

```bash
curl -s http://localhost:9093/api/v1/status | python3 -m json.tool | grep '"uptime"'
```

**Step 5: Verificar datasources Grafana**

```bash
curl -s -u admin:admin http://localhost:3000/api/datasources | python3 -m json.tool | grep '"name"'
```
Expected: "Prometheus", "Jaeger"

**Step 6: Smoke test de trace — verificar no Jaeger**

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

curl -s -X POST http://localhost:8080/api/v1/vinyls \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Nevermind","artist":"Nirvana","year":1991,"genre":"Grunge","label":"DGC"}'

echo "Verificar trace em http://localhost:16686 > Service: vinyl-catalog"
```

**Step 7: Commit final**

```bash
git add -A
git commit -m "chore: finalize observability stack - Jaeger traces, alerts, Grafana dashboards"
```
