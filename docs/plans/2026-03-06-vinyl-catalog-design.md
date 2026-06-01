# Vinyl Catalog — Design Document

**Date:** 2026-03-06
**Status:** Approved

---

## Overview

Sistema de catalogação de discos de vinil com API REST, autenticação JWT, banco de dados PostgreSQL e observabilidade completa. Implementado em Go seguindo Clean Architecture e Arquitetura Hexagonal (Ports & Adapters).

---

## Stack Tecnológica

| Componente | Tecnologia |
|---|---|
| Linguagem | Go 1.22+ |
| Framework HTTP | Gin |
| ORM | GORM |
| Banco de Dados | PostgreSQL |
| Autenticação | JWT (golang-jwt/jwt) |
| Documentação API | Swagger (swaggo/swag) |
| Métricas | Prometheus (prometheus/client_golang) |
| Traces | OpenTelemetry + Jaeger |
| Logs | slog (stdlib Go 1.21+) |
| Telemetry Pipeline | OpenTelemetry Collector |
| Gestão do Projeto | mise |
| Containerização | Docker + Docker Compose |

---

## Arquitetura

### Princípios

- **Hexagonal (Ports & Adapters):** o domínio define interfaces (ports); implementações externas são adapters
- **Clean Architecture:** dependências sempre apontam para dentro — adapters → application → domain
- **Domínio isolado:** o pacote `domain` não importa nenhuma biblioteca externa

### Estrutura de Diretórios

```
vinyl-catalog/
├── cmd/
│   └── api/
│       └── main.go                        # entrypoint: wiring de dependências
├── internal/
│   ├── domain/
│   │   ├── vinyl.go                       # entidade Vinyl
│   │   └── ports.go                       # interfaces VinylRepository e VinylService
│   ├── application/
│   │   └── vinyl_service.go               # casos de uso (implementa VinylService)
│   ├── adapters/
│   │   ├── primary/
│   │   │   └── http/
│   │   │       ├── handler.go             # handlers Gin
│   │   │       ├── router.go              # rotas + middlewares
│   │   │       └── dto.go                 # structs de request/response
│   │   └── secondary/
│   │       └── postgres/
│   │           └── vinyl_repo.go          # implementa VinylRepository via GORM
│   └── infrastructure/
│       ├── config/
│       │   └── config.go                  # leitura de variáveis de ambiente
│       ├── database/
│       │   └── postgres.go                # conexão e auto-migrate GORM
│       ├── auth/
│       │   └── jwt.go                     # geração e validação de tokens JWT
│       └── observability/
│           ├── metrics.go                 # Prometheus: registro de métricas
│           └── tracing.go                 # OpenTelemetry: setup de tracer
├── docs/                                  # gerado pelo swaggo
│   └── plans/
│       └── 2026-03-06-vinyl-catalog-design.md
├── Dockerfile                             # multi-stage build
├── docker-compose.yml                     # app, db, otel-collector, jaeger, prometheus
├── otel-collector-config.yaml             # configuração do OTel Collector
├── mise.toml                              # tasks de desenvolvimento
├── .env.example                           # variáveis de ambiente de referência
├── README.md
└── go.mod
```

---

## Modelo de Domínio

### Entidade `Vinyl`

| Campo | Tipo | Descrição |
|---|---|---|
| `id` | UUID | identificador único |
| `title` | string | título do disco |
| `artist` | string | nome do artista/banda |
| `year` | int | ano de lançamento |
| `genre` | string | gênero musical |
| `label` | string | gravadora |
| `created_at` | timestamp | data de criação |
| `updated_at` | timestamp | data de atualização |

### Interfaces (Ports)

```go
type VinylRepository interface {
    Create(ctx context.Context, v *Vinyl) error
    FindByID(ctx context.Context, id string) (*Vinyl, error)
    FindAll(ctx context.Context) ([]Vinyl, error)
    Update(ctx context.Context, v *Vinyl) error
    Delete(ctx context.Context, id string) error
}

type VinylService interface {
    CreateVinyl(ctx context.Context, v *Vinyl) error
    GetVinyl(ctx context.Context, id string) (*Vinyl, error)
    ListVinyls(ctx context.Context) ([]Vinyl, error)
    UpdateVinyl(ctx context.Context, v *Vinyl) error
    DeleteVinyl(ctx context.Context, id string) error
}
```

---

## API REST

### Endpoints

| Método | Rota | Descrição | Auth |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | gera token JWT | não |
| `GET` | `/api/v1/health` | health check | não |
| `GET` | `/metrics` | métricas Prometheus | não |
| `GET` | `/api/v1/vinyls` | lista todos os discos | sim |
| `GET` | `/api/v1/vinyls/:id` | busca disco por ID | sim |
| `POST` | `/api/v1/vinyls` | cria novo disco | sim |
| `PUT` | `/api/v1/vinyls/:id` | atualiza disco | sim |
| `DELETE` | `/api/v1/vinyls/:id` | remove disco | sim |

### Autenticação

- `POST /api/v1/auth/login` recebe `{"username": "...", "password": "..."}` e retorna `{"token": "..."}`
- Credenciais configuradas via variáveis de ambiente (`ADMIN_USERNAME`, `ADMIN_PASSWORD`)
- Todas as rotas `/api/v1/vinyls/*` exigem header `Authorization: Bearer <token>`
- Middleware JWT valida e extrai claims antes de chamar o handler

---

## Fluxo de uma Requisição

```
HTTP Request
    │
    ▼
[JWT Middleware]              ← valida token
    │
    ▼
[Handler]                     ← adapter primário (Gin): valida DTO, extrai params
    │
    ▼
[VinylService]                ← application: executa caso de uso
    │  usa interface VinylRepository
    ▼
[PostgresVinylRepo]           ← adapter secundário: query GORM
    │
    ▼
[PostgreSQL]
```

---

## Observabilidade

### Métricas — Prometheus
- Endpoint `GET /metrics` via `promhttp.Handler()`
- Métricas: latência por rota (histogram), total de requisições (counter), erros por status code
- Prometheus faz scrape direto na aplicação

### Traces — OpenTelemetry + Jaeger
- App instrumentada com `go.opentelemetry.io/otel`
- Traces propagados: handler → service → repository
- Exportados via OTLP gRPC para o OTel Collector
- OTel Collector encaminha para Jaeger

### Logs — slog
- Structured logging em JSON (produção) e texto (desenvolvimento)
- Campos padrão: `level`, `time`, `msg`, `trace_id`, `span_id`
- Correlação automática entre logs e traces via context propagation

### Pipeline de Telemetria

```
App (Go)
  │ OTLP gRPC :4317
  ▼
otel-collector
  ├──► Jaeger    (traces)  :16686
  └──► Prometheus (métricas via scrape) :9090
```

---

## Infraestrutura Docker

### Serviços (`docker-compose.yml`)

| Serviço | Porta | Descrição |
|---|---|---|
| `app` | 8080 | API Go |
| `db` | 5432 | PostgreSQL 16 |
| `otel-collector` | 4317, 4318 | OpenTelemetry Collector |
| `jaeger` | 16686 | UI de traces |
| `prometheus` | 9090 | UI de métricas |

### Dockerfile
- Multi-stage build: estágio `builder` (golang:alpine) + estágio `runner` (alpine)
- Imagem final mínima, sem toolchain Go

---

## Gestão com mise

| Task | Comando | Descrição |
|---|---|---|
| `mise run dev` | air / go run | sobe servidor com hot reload |
| `mise run build` | go build | compila binário |
| `mise run test` | go test ./... | executa todos os testes |
| `mise run swag` | swag init | gera documentação Swagger |
| `mise run docker:up` | docker compose up | sobe todos os serviços |
| `mise run docker:down` | docker compose down | derruba todos os serviços |
| `mise run lint` | golangci-lint run | executa linter |

---

## Variáveis de Ambiente

```env
# App
APP_PORT=8080
APP_ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=vinyl_catalog

# Auth
JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
OTEL_SERVICE_NAME=vinyl-catalog
```

---

## Documentação Swagger

- Anotações `swaggo` nos handlers
- Gerado em `docs/` via `swag init`
- Acessível em `GET /swagger/index.html`
