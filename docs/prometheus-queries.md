# Prometheus Queries — Vinyl Catalog

Reference guide for querying metrics exposed by the vinyl-catalog service.

**Prometheus UI:** http://localhost:9090

---

## Taxa de Requisições (RPS)

**Total de requisições por segundo:**
```promql
rate(http_requests_total[5m])
```

**RPS por rota:**
```promql
sum by (handler) (rate(http_requests_total[5m]))
```

**RPS por status HTTP:**
```promql
sum by (status_code) (rate(http_requests_total[5m]))
```

**RPS por método HTTP:**
```promql
sum by (method) (rate(http_requests_total[5m]))
```

---

## Latência

**Latência média:**
```promql
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])
```

**Percentil 50 (mediana) por rota:**
```promql
histogram_quantile(0.50, sum by (le, handler) (rate(http_request_duration_seconds_bucket[5m])))
```

**Percentil 95 por rota:**
```promql
histogram_quantile(0.95, sum by (le, handler) (rate(http_request_duration_seconds_bucket[5m])))
```

**Percentil 99 por rota:**
```promql
histogram_quantile(0.99, sum by (le, handler) (rate(http_request_duration_seconds_bucket[5m])))
```

**Latência p95 global:**
```promql
histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))
```

---

## Erros

**Taxa de erros 5xx (últimos 5 min):**
```promql
sum(rate(http_requests_total{status_code=~"5.."}[5m]))
```

**Taxa de erros 4xx (últimos 5 min):**
```promql
sum(rate(http_requests_total{status_code=~"4.."}[5m]))
```

**Percentual de erros 5xx sobre total:**
```promql
sum(rate(http_requests_total{status_code=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100
```

**Erros por rota:**
```promql
sum by (handler) (rate(http_requests_total{status_code=~"5.."}[5m]))
```

---

## Disponibilidade

**Uptime do processo (segundos):**
```promql
process_uptime_seconds
```

**SLO de disponibilidade — proporção de requisições bem-sucedidas:**
```promql
1 - (sum(rate(http_requests_total{status_code=~"5.."}[30m])) / sum(rate(http_requests_total[30m])))
```

**Serviço ativo (1 = sim, 0 = não):**
```promql
up{job="vinyl-catalog"}
```

---

## Operações CRUD por Endpoint

**Criações (POST /vinyls) por segundo:**
```promql
rate(http_requests_total{handler="CreateVinyl", method="POST"}[5m])
```

**Leituras (GET /vinyls) por segundo:**
```promql
rate(http_requests_total{handler="ListVinyls", method="GET"}[5m])
```

**Leituras por ID (GET /vinyls/:id) por segundo:**
```promql
rate(http_requests_total{handler="GetVinyl", method="GET"}[5m])
```

**Atualizações (PUT /vinyls/:id) por segundo:**
```promql
rate(http_requests_total{handler="UpdateVinyl", method="PUT"}[5m])
```

**Deleções (DELETE /vinyls/:id) por segundo:**
```promql
rate(http_requests_total{handler="DeleteVinyl", method="DELETE"}[5m])
```

**Total de operações por tipo (últimas 24h):**
```promql
increase(http_requests_total[24h])
```

---

## Integrações (Favoritos, Uploads, Tradução)

### Favoritos

**Total de toggles de favorito (add/remove) nas últimas 24h:**
```promql
increase(vinyl_favorites_total[24h])
```

**Apenas adições:**
```promql
increase(vinyl_favorites_total{action="add"}[24h])
```

**Apenas remoções:**
```promql
increase(vinyl_favorites_total{action="remove"}[24h])
```

**Taxa de toggles por segundo:**
```promql
rate(vinyl_favorites_total[5m])
```

---

### Uploads de Capa

**Total de uploads de capa (últimas 24h) por extensão:**
```promql
increase(vinyl_cover_uploads_total[24h])
```

**Apenas bem-sucedidos:**
```promql
increase(vinyl_cover_uploads_total{status="success"}[24h])
```

**Apenas com erro:**
```promql
increase(vinyl_cover_uploads_total{status="error"}[24h])
```

**Taxa de erro de upload (%):**
```promql
rate(vinyl_cover_uploads_total{status="error"}[5m]) /
rate(vinyl_cover_uploads_total[5m]) * 100
```

---

### Uploads de Foto de Perfil

**Total nas últimas 24h:**
```promql
increase(vinyl_profile_photo_uploads_total[24h])
```

**Taxa de sucesso vs erro:**
```promql
sum by (status) (increase(vinyl_profile_photo_uploads_total[24h]))
```

---

### Tradução de Letras (Frontend)

**Total de requisições de tradução por status:**
```promql
sum by (status) (increase(frontend_translate_requests_total[24h]))
```

**Apenas traduções com sucesso:**
```promql
increase(frontend_translate_requests_total{status="success"}[24h])
```

**Erros de quota nas últimas 24h:**
```promql
increase(frontend_translate_requests_total{status="quota_exceeded"}[24h])
```

**Taxa de erro de tradução (%):**
```promql
rate(frontend_translate_requests_total{status=~"error|quota_exceeded"}[5m]) /
rate(frontend_translate_requests_total[5m]) * 100
```

---

## OTel Collector (porta 8889)

**Spans recebidos por segundo:**
```promql
rate(otelcol_receiver_accepted_spans_total[5m])
```

**Spans exportados para Jaeger:**
```promql
rate(otelcol_exporter_sent_spans_total[5m])
```

**Uso de memória do collector (bytes):**
```promql
otelcol_process_memory_rss
```
