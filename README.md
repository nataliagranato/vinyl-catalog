# Vinyl Catalog

Sistema de catalogação de discos de vinil com API REST em Go e frontend Next.js.

## Arquitetura

Clean Architecture + Hexagonal (Ports & Adapters):

```
adapters (HTTP/Gin, PostgreSQL/GORM)
    └── application (use cases)
            └── domain (entidades, interfaces — stdlib puro)
```

A regra de dependência aponta sempre para dentro. O domínio não importa nenhuma biblioteca externa.

## Stack

| Componente | Tecnologia |
|---|---|
| Linguagem | Go 1.22+ |
| Framework HTTP | Gin |
| ORM | GORM |
| Banco de Dados | PostgreSQL 16 |
| Autenticação | JWT (golang-jwt/jwt) |
| Documentação API | Swagger (swaggo/swag) |
| Métricas | Prometheus |
| Alertas | Alertmanager |
| Traces | OpenTelemetry + Jaeger |
| Dashboards | Grafana |
| Logs | Loki + Promtail + slog |
| Telemetry Pipeline | OpenTelemetry Collector |
| Error Tracking Frontend | Grafana Faro + Grafana Alloy |
| Frontend | Next.js 16 + Tailwind v4 + Framer Motion |
| Gestão de Tasks | mise |
| Containerização | Docker + Docker Compose |

## Início Rápido

### Pré-requisitos

- [mise](https://mise.jdx.dev/)
- Docker e Docker Compose

### Subir tudo com Docker Compose

```bash
cp .env.example .env
# Edite .env com valores seguros para produção

docker-compose up --build
```

Serviços disponíveis:

| Serviço | URL | Credenciais |
|---|---|---|
| Frontend | http://localhost:3001 | admin / admin |
| API | http://localhost:8080 | — |
| Swagger UI | http://localhost:8080/swagger/index.html | — |
| Métricas | http://localhost:8080/metrics | — |
| Prometheus | http://localhost:9090 | — |
| Alertmanager | http://localhost:9093 | — |
| Jaeger UI | http://localhost:16686 | — |
| Loki | http://localhost:3100 | — |
| Grafana | http://localhost:3000 | admin / admin |
| Alloy (Faro receiver) | http://localhost:12347 | — |

### Desenvolvimento local

```bash
cp .env.example .env
# Suba apenas a infraestrutura (db, otel, jaeger, prometheus, loki)
docker compose up -d db otel-collector jaeger prometheus loki promtail

# Rode a aplicação localmente
mise run dev
```

## Frontend

O frontend é uma aplicação Next.js 16 com tema dark ("Dark Groove") acessível em http://localhost:3001.

**Funcionalidades:**
- Login com autenticação JWT via cookie httpOnly
- Listagem de discos com busca e filtros por gênero e ano
- Criação, edição e remoção de discos
- Upload de capa do disco com preview em miniatura
- Descrição por disco
- Lista de faixas com letras (collapsible por faixa)
- Perfil público compartilhável em `/profile` (sem login)
- Edição de perfil com foto, bio, links e gêneros preferidos
- Discos favoritos: botão de coração no hover de cada card para favoritar/desfavoritar
- Perfil público mostra discos favoritos com tracklist expansível (clique para ver faixas)
- Letras das músicas expansíveis: clique na faixa para ver a letra
- Tradutor de letras: botão "Traduzir" com seletor de idioma (PT, EN, ES, FR, DE, IT, JA) via proxy interno `/api/translate` → Google Translate
- Imagens de capa e foto de perfil corretamente servidas pelo backend
- Telemetria no browser: traces OTel (tradução, favoritos, upload) enviados ao Jaeger via OTel Collector

## Autenticação

A API usa JWT. Faça login para obter um token:

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

A resposta é um objeto JSON:

```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

> ⚠️ **Atenção:** use apenas o valor do campo `"token"`, não o objeto JSON completo.

Use o token nas requisições protegidas:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Fluxo completo em uma linha** (captura o token automaticamente):

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/vinyls
```

### No Swagger UI

1. Acesse http://localhost:8080/swagger/index.html
2. Faça `POST /auth/login` e copie **apenas o valor** do campo `token` da resposta
3. Clique em **Authorize** (cadeado)
4. No campo, digite: `Bearer ` seguido do token (ex: `Bearer eyJhbGc...`)
5. Clique em **Authorize**

## Endpoints

### Autenticação

| Método | Rota | Descrição | Auth |
|---|---|---|---|
| POST | /api/v1/auth/login | Gera token JWT | Não |

### Discos

| Método | Rota | Descrição | Auth |
|---|---|---|---|
| GET | /api/v1/vinyls | Lista todos os discos | Sim |
| GET | /api/v1/vinyls/:id | Busca disco por ID | Sim |
| POST | /api/v1/vinyls | Cria novo disco | Sim |
| PUT | /api/v1/vinyls/:id | Atualiza disco | Sim |
| DELETE | /api/v1/vinyls/:id | Remove disco | Sim |
| POST | /api/v1/vinyls/:id/cover | Upload de capa (multipart/form-data) | Sim |
| POST | /api/v1/vinyls/:id/favorite | Toggle favorito (adiciona/remove) | Sim |

### Faixas

| Método | Rota | Descrição | Auth |
|---|---|---|---|
| GET | /api/v1/vinyls/:id/tracks | Lista faixas do disco | Sim |
| POST | /api/v1/vinyls/:id/tracks | Cria faixa | Sim |
| PUT | /api/v1/vinyls/:id/tracks/:track_id | Atualiza faixa | Sim |
| DELETE | /api/v1/vinyls/:id/tracks/:track_id | Remove faixa | Sim |

### Perfil

| Método | Rota | Descrição | Auth |
|---|---|---|---|
| GET | /api/v1/profile | Perfil público | Não |
| PUT | /api/v1/profile | Atualiza perfil | Sim |
| POST | /api/v1/profile/photo | Upload de foto do perfil | Sim |

### Utilitários

| Método | Rota | Descrição | Auth |
|---|---|---|---|
| GET | /api/v1/health | Health check | Não |
| GET | /metrics | Métricas Prometheus | Não |
| GET | /swagger/index.html | Documentação Swagger | Não |
| GET | /uploads/:filename | Arquivos enviados (capas, fotos) | Não |

## Exemplos de Uso

```bash
# Obter token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

# Criar disco com descrição
curl -X POST http://localhost:8080/api/v1/vinyls \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Kind of Blue","artist":"Miles Davis","year":1959,"genre":"Jazz","label":"Columbia","description":"Modal jazz masterpiece gravado em apenas duas sessões."}'

# Listar discos
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/vinyls

# Upload de capa
curl -X POST http://localhost:8080/api/v1/vinyls/<id>/cover \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@capa.jpg"

# Criar faixa com letra
curl -X POST http://localhost:8080/api/v1/vinyls/<id>/tracks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"So What","position":1,"lyrics":"So What é um tema em D Dórico..."}'

# Listar faixas
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/vinyls/<id>/tracks

# Atualizar faixa
curl -X PUT http://localhost:8080/api/v1/vinyls/<id>/tracks/<track_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"So What","position":1,"lyrics":"Letra atualizada..."}'

# Remover faixa
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/vinyls/<id>/tracks/<track_id>

# Ver perfil público (sem token)
curl http://localhost:8080/api/v1/profile

# Atualizar perfil
curl -X PUT http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"Natalia","bio":"Colecionadora de jazz e soul.","links":["https://github.com/nataliagranato"],"preferred_genres":["Jazz","Soul","Funk"]}'

# Upload de foto de perfil
curl -X POST http://localhost:8080/api/v1/profile/photo \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@foto.jpg"
```

## Modelos de Dados

### Vinyl

```json
{
  "id": "uuid",
  "title": "Kind of Blue",
  "artist": "Miles Davis",
  "year": 1959,
  "genre": "Jazz",
  "label": "Columbia",
  "description": "Modal jazz masterpiece.",
  "cover_url": "/uploads/abc123.jpg",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

### Track

```json
{
  "id": "uuid",
  "vinyl_id": "uuid",
  "title": "So What",
  "position": 1,
  "lyrics": "...",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

### Profile

```json
{
  "username": "admin",
  "display_name": "Natalia",
  "bio": "Colecionadora de jazz e soul.",
  "photo_url": "/uploads/profile-abc123.jpg",
  "links": ["https://github.com/nataliagranato"],
  "preferred_genres": ["Jazz", "Soul"],
  "favorite_vinyl_ids": ["uuid-1", "uuid-2"],
  "favorite_vinyls": [
    {
      "id": "uuid-1",
      "title": "Kind of Blue",
      "artist": "Miles Davis",
      "year": 1959,
      "genre": "Jazz",
      "label": "Columbia",
      "cover_url": "/uploads/abc123.jpg",
      "tracks": [
        { "id": "uuid", "vinyl_id": "uuid-1", "title": "So What", "position": 1, "lyrics": "..." }
      ]
    }
  ]
}
```

## Tasks mise

```bash
mise run dev          # Roda a aplicação
mise run build        # Compila o binário em bin/vinyl-catalog
mise run test         # Executa todos os testes com cobertura
mise run coverage     # Gera relatório HTML de cobertura
mise run swag         # Regenera documentação Swagger
mise run lint         # Executa golangci-lint
mise run docker:up    # Sobe todos os serviços
mise run docker:down  # Derruba todos os serviços
mise run docker:logs  # Acompanha logs da aplicação
```

## Variáveis de Ambiente

| Variável | Padrão | Descrição |
|---|---|---|
| APP_PORT | 8080 | Porta da aplicação |
| APP_ENV | development | Ambiente (development/production) |
| DB_HOST | localhost | Host do PostgreSQL |
| DB_PORT | 5432 | Porta do PostgreSQL |
| DB_USER | postgres | Usuário do banco |
| DB_PASSWORD | postgres | Senha do banco |
| DB_NAME | vinyl_catalog | Nome do banco |
| DB_SSLMODE | disable | Modo SSL (use require em produção) |
| JWT_SECRET | change-me | Segredo JWT (use valor forte em produção) |
| JWT_EXPIRATION_HOURS | 24 | Duração do token em horas |
| ADMIN_USERNAME | admin | Usuário administrador |
| ADMIN_PASSWORD | admin | Senha administrador |
| OTEL_EXPORTER_OTLP_ENDPOINT | localhost:4317 | Endpoint do OTel Collector |
| OTEL_SERVICE_NAME | vinyl-catalog | Nome do serviço para traces |
| SERVICE_VERSION | 1.0.0 | Versão do serviço (atributo de trace) |

## Observabilidade

### Traces
Os traces são enviados via OTLP gRPC para o OTel Collector, que os encaminha ao Jaeger.
Acesse o Jaeger UI em http://localhost:16686 e filtre pelo serviço `vinyl-catalog`.

### Métricas
O Prometheus faz scrape do endpoint `/metrics` da aplicação.
Métricas disponíveis:
- `http_requests_total` — total de requisições por método, rota e status
- `http_request_duration_seconds` — latência por método e rota
- `vinyl_cover_uploads_total{status,ext}` — uploads de capa (success/error, .jpg/.png)
- `vinyl_favorites_total{action}` — toggles de favorito (add/remove)
- `vinyl_profile_photo_uploads_total{status}` — uploads de foto de perfil

Referência completa de queries: [`docs/prometheus-queries.md`](docs/prometheus-queries.md)

### Alertas
O Alertmanager gerencia alertas definidos em `prometheus/alerts.yml`:

**Infraestrutura (`vinyl-catalog`):**

| Alerta | Condição | Severidade |
|---|---|---|
| HighErrorRate | taxa 5xx > 5% por 5min | critical |
| HighLatency | p95 > 500ms por 10min | warning |
| ServiceDown | sem métricas por 1min | critical |
| HighRequestRate | RPS > 100 por 5min | warning |
| NoRequestsReceived | RPS = 0 por 10min | warning |
| P99HighLatency | p99 > 1s por 5min | warning |
| AuthErrorRate | taxa 401 > 15% por 5min | warning |
| HighClientErrorRate | taxa 4xx > 20% por 5min | warning |

**Integrações (`vinyl-catalog-integrations`, group_wait: 30s):**

| Alerta | Condição | Severidade |
|---|---|---|
| TranslationQuotaExceeded | quota esgotada em 5min | warning |
| CoverUploadErrorRate | taxa de erro upload > 10% por 5min | warning |
| FavoriteEndpointErrors | 5xx em `/favorite` por 2min | warning |
| ProfilePhotoUploadErrors | > 3 erros de upload em 5min | warning |

Acesse o Alertmanager em http://localhost:9093.

### Dashboards Grafana
Seis dashboards provisionados automaticamente em http://localhost:3000 (admin/admin):

- **Operacional** — RPS, taxa de erros, latência p50/p95/p99, breakdown por rota, distribuição de status HTTP, uptime
- **Negócio** — discos criados, operações CRUD por hora, rota mais acessada, traduções totais, favoritos, uploads por extensão
- **Frontend — Integrações** — requisições de tradução por status, latência p95, erros de quota, toggles de favorito, erros de upload, logs do frontend via Loki
- **Frontend — Faro Errors** — exceções JS (24h), console errors, LCP, CLS, live log de exceções e todos os logs do frontend (via Grafana Faro + Alloy)
- **Traces** — traces recentes, traces com erro, latência por operação, volume de spans; integrado com Jaeger via datasource nativo
- **Logs** — volume de logs por nível (ERROR/WARN/INFO), painel live do app com parse JSON, logs de todos os serviços do stack

### Logs
Logs estruturados em JSON via `log/slog` (backend) e `lib/logger.ts` (frontend). O Promtail coleta os logs de todos os containers Docker e os envia ao Loki. Os campos `trace_id` e `span_id` são extraídos como labels, permitindo correlacionar um log diretamente com o trace correspondente no Jaeger a partir do painel de logs do Grafana.

## Banco de Dados

Documentação completa para conexão e consultas SQL: [`docs/database.md`](docs/database.md)

Conexão rápida:

```bash
# Via Docker (sem instalar nada)
docker exec -it vinyl-catalog-db-1 psql -U postgres -d vinyl_catalog

# Via psql no host
psql -h localhost -p 5432 -U postgres -d vinyl_catalog
```
