# Vinyl Catalog API

🎵 Sistema de catalogação de discos de vinil migrado para Cloudflare Workers com custo zero.

## 📊 Status Atual

- ✅ **Backend:** Deployed na Cloudflare Workers
- ✅ **Database:** Cloudflare D1 configurado
- ✅ **Storage:** Cloudflare R2 configurado
- ✅ **Security:** 9/10 (corrigidas vulnerabilidades críticas)
- ✅ **Observability:** Full (Logs + Traces + Metrics)
- ⏳ **Frontend:** Pendente migração para Cloudflare Pages

## 🚀 Deployed Resources

- **API:** https://vinyl-catalog-api.nataliagranato.xyz
- **Health Check:** /api/v1/health
- **Metrics:** /api/v1/metrics
- **Documentation:** https://vinyl-catalog-api.nataliagranato.xyz/

## 📁 Documentação

- **[README_CLOUDFLARE.md](./README_CLOUDFLARE.md)** - Guia completo de deployment Cloudflare
- **[CLOUDFLARE_MIGRATION.md](./CLOUDFLARE_MIGRATION.md)** - Guia de migração do Docker
- **[SECURITY_AUDIT_12FACTOR.md](./SECURITY_AUDIT_12FACTOR.md)** - Auditoria de segurança 12-Factor
- **[OPENTELEMETRY_SETUP.md](./OPENTELEMETRY_SETUP.md)** - Guia de configuração OpenTelemetry

## 🔐 Credenciais

Ver arquivo `.secrets` (não versionado) para credenciais de login atuais.

## 🛠️ Quick Start

```bash
# Instalar dependências
npm install

# Deploy para Cloudflare
npm run deploy

# Ver logs em tempo real
npx wrangler tail

# Build
npm run build
```

## 📋 API Endpoints

### System
- `GET /` - Documentação da API
- `GET /api/v1/health` - Health check
- `GET /api/v1/metrics` - Métricas

### Authentication
- `POST /api/v1/auth/login` - Login JWT

### Vinyls (Requires Auth)
- `GET /api/v1/vinyls` - Listar discos
- `POST /api/v1/vinyls` - Criar disco
- `PUT /api/v1/vinyls/:id` - Atualizar disco
- `DELETE /api/v1/vinyls/:id` - Remover disco

### Profile
- `GET /api/v1/profile` - Perfil público
- `PUT /api/v1/profile` - Atualizar perfil (auth)

### Uploads (Requires Auth)
- `POST /api/v1/vinyls/:id/cover` - Upload capa
- `GET /uploads/:filename` - Servir arquivos

## 🔍 Observabilidade

### Dashboard Cloudflare
1. Acesse: https://dash.cloudflare.com
2. Workers & Pages → vinyl-catalog-api → Observability
3. Visualize Logs, Traces e Metrics

### CLI
```bash
npx wrangler tail
```

## 💰 Custo

**Total: $0/mês** (Cloudflare Free Tier)

- Workers: 100k requests/day (grátis)
- D1: 5GB storage (grátis)
- R2: 10GB storage (grátis)
- Observability: Logs + Traces (grátis)

## 📊 Métricas de Segurança

- **Score:** 9/10
- **Conformidade 12-Factor:** 78%
- **Vulnerabilidades Críticas:** 0
- **JWT:** HMAC-SHA256 (Web Crypto API)
- **Credenciais:** Cryptographically secure

## 🎯 Stack Tecnológico

- **Backend:** TypeScript (Cloudflare Workers)
- **Database:** Cloudflare D1 (SQLite)
- **Storage:** Cloudflare R2
- **Observability:** Workers Logs + Traces + OpenTelemetry
- **Auth:** JWT com Web Crypto API
- **Router:** itty-router

## 📖 Desenvolvimento

```bash
# Desenvolvimento local
npm run dev

# Testes
npm test

# Lint
npm run lint

# Format
npm run format
```

## 🔧 Configuração

Ver `wrangler.toml` para configuração completa:
- Account ID
- Database bindings
- R2 bindings
- Environment variables
- Observability settings

## 🚨 Troubleshooting

### Login falha
- Verificar `.secrets` para credenciais corretas
- Checar JWT_SECRET no wrangler.toml
- Verificar se token não expirou (24h)

### Observabilidade não funciona
- Verificar Wrangler v4+: `npx wrangler --version`
- Checar seção [observability] no wrangler.toml
- Acessar dashboard Cloudflare

### Database errors
- Verificar database_id no wrangler.toml
- Checar health check: `curl /api/v1/health`
- Re-run schema: `./setup-db.sh`

## 📞 Suporte

- Cloudflare Docs: https://developers.cloudflare.com/
- Dashboard: https://dash.cloudflare.com
- Project Issues: GitHub repository

---

**Deployed with [Devin](https://devin.ai)**