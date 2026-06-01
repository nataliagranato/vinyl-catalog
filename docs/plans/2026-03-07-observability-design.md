# Observability Enhancement — Design Document

**Date:** 2026-03-07
**Status:** Approved

---

## Overview

Melhoria completa da stack de observabilidade do vinyl-catalog: enriquecimento de traces no Jaeger, documentação de queries Prometheus, regras de alerta com Alertmanager e dashboards Grafana (operacional + negócio).

---

## Seção 1: Traces no Jaeger

### Problemas atuais
- Spans existem mas faltam atributos HTTP (status, rota, método)
- Erros não são marcados nos spans (`span.SetStatus`)
- Resource sem versão, ambiente, host info
- OTel Collector sem memory limiter

### Melhorias em `internal/infrastructure/observability/tracing.go`
- `resource.WithDetectors(...)` para host, OS, processo
- Atributos de recurso: `service.version`, `deployment.environment`
- `sdktrace.WithSampler(sdktrace.AlwaysSample())` explícito

### Melhorias em `internal/adapters/primary/http/handler.go`
- `span.SetStatus(codes.Error, msg)` em todos os caminhos de erro
- Atributos: `http.method`, `http.route`, `http.status_code`, `http.url`
- `trace_id` e `span_id` propagados nos logs slog via context

### Melhorias em `otel-collector-config.yaml`
- Processor `memory_limiter` para estabilidade
- Processor `attributes` para enriquecer spans

---

## Seção 2: Documentação Prometheus + Alertas

### Arquivo `docs/prometheus-queries.md`
Queries organizadas por categoria:
- Taxa de requisições (RPS total, por rota, por status)
- Latência (p50, p95, p99, média)
- Erros (4xx rate, 5xx rate, error ratio)
- Disponibilidade (uptime, SLO)
- Operações CRUD por endpoint

### Arquivo `prometheus/alerts.yml`
| Alerta | Condição | Severidade |
|---|---|---|
| HighErrorRate | taxa 5xx > 5% por 5min | critical |
| HighLatency | p95 > 500ms por 10min | warning |
| ServiceDown | sem métricas por 1min | critical |
| HighRequestRate | RPS > 100 por 5min | warning |
| NoRequestsReceived | RPS = 0 por 10min | warning |

### Arquivo `alertmanager/alertmanager.yml`
Receiver `null` (UI only, sem notificação externa)

### `prometheus.yml` atualizado
- `rule_files` apontando para `prometheus/alerts.yml`
- Bloco `alerting` com endereço do Alertmanager

---

## Seção 3: Grafana + Dashboards

### Serviço `grafana` no docker-compose
- Porta 3000
- Provisioning automático via arquivos de configuração

### Estrutura
```
grafana/
├── provisioning/
│   ├── datasources/datasources.yml   # Prometheus + Jaeger
│   └── dashboards/dashboards.yml     # aponta para /dashboards
└── dashboards/
    ├── operational.json
    └── business.json
```

### Dashboard Operacional
- Request Rate (RPS) — graph
- Error Rate (%) — stat
- Latência p50/p95/p99 — graph
- Breakdown por rota — table
- HTTP status distribution — pie
- Uptime — stat

### Dashboard de Negócio
- Total de discos criados — stat
- Operações CRUD por hora — bar chart
- Rota mais acessada — table
- Taxa de erros por operação — graph
- Link para Jaeger — annotation

---

## Infraestrutura Docker Atualizada

| Serviço | Porta | Descrição |
|---|---|---|
| app | 8080 | API Go |
| db | 5432 | PostgreSQL |
| otel-collector | 4317/4318 | OTel Collector |
| jaeger | 16686 | Traces UI |
| prometheus | 9090 | Métricas UI |
| alertmanager | 9093 | Alertas UI |
| grafana | 3000 | Dashboards UI |
