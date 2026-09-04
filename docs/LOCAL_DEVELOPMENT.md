# Local Development Guide

Guia completo para desenvolvimento local do Vinyl Catalog API sem Docker.

## Prerequisites

- Node.js >=22.14.0
- npm
- Wrangler CLI (instalado via npm)

## Quick Start

### 1. Instalar Dependências

```bash
npm install
```

### 2. Iniciar Backend (Cloudflare Workers)

```bash
npm run dev
```

Backend estará disponível em: `http://localhost:8787`

### 3. Iniciar Frontend (Next.js)

```bash
cd frontend
npm install
npm run dev
```

Frontend estará disponível em: `http://localhost:3000`

---

## Backend Development

### Wrangler Dev

O comando `npm run dev` inicia o Wrangler em modo de desenvolvimento:

```bash
npm run dev
```

**O que acontece:**
- TypeScript é compilado automaticamente
- Worker roda em: `http://localhost:8787`
- D1 local é criado automaticamente
- R2 local (limitado) ou usa bucket prod
- Logs são exibidos no terminal
- Hot reload ativo

### D1 Local

Wrangler cria automaticamente um banco D1 local:

```bash
# Criar banco local (automático com wrangler dev)
npx wrangler d1 execute DB --local --file=scripts/schema.sql
```

**Como verificar:**
```bash
# Listar tabelas
npx wrangler d1 execute DB --local --command="SELECT name FROM sqlite_master WHERE type='table'"

# Query manual
npx wrangler d1 execute DB --local --command="SELECT * FROM vinyls LIMIT 5"
```

### R2 Local

R2 local é limitado. Opções:

**Opção 1: Usar R2 de Produção**
```toml
# wrangler.toml - usar bindings existentes
[[r2_buckets]]
binding = "UPLOADS"
bucket_name = "vinyl-catalog-uploads"
```

**Opção 2: Mock para Desenvolvimento**
- Implementar mock em código
- Usar sistema de arquivos local

### Logs Locais

```bash
# Ver logs em tempo real
npx wrangler tail
```

---

## Frontend Development

### Iniciar Frontend

```bash
cd frontend
npm run dev
```

Frontend estará disponível em: `http://localhost:3000`

### Configurar API URL

No frontend, configurar API URL local:

```typescript
// frontend/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8787
```

---

## Scripts Disponíveis

### Backend

```bash
npm run dev        # Iniciar Wrangler dev
npm run build      # Compilar TypeScript
npm run deploy     # Deploy para Cloudflare
npm test           # Rodar testes
npm run lint       # Lint do código
npm run format     # Formatar código
```

### Scripts Utilitários

```bash
./scripts/setup-db.sh      # Configurar D1 local
./scripts/setup-r2.sh      # Configurar R2
./scripts/setup-secrets.sh # Configurar secrets
./scripts/migrate-data.sh  # Migrar dados
```

---

## Debugging

### Backend

**1. Adicionar logs no código:**
```typescript
// src/utils/logger.ts
logger.info('Debug message', { data: 'value' });
```

**2. Ver logs no terminal:**
```bash
npm run dev
# Logs aparecem automaticamente
```

**3. Inspect com Wrangler:**
```bash
npx wrangler tail
```

### Frontend

**1. Usar browser DevTools:**
- Console logs
- Network tab
- React DevTools

**2. Next.js debug:**
```bash
# Terminal frontend
NODE_OPTIONS='--inspect' npm run dev
```

---

## Testing

### Backend Tests

```bash
npm test
```

### API Testing Local

```bash
# Health check
curl http://localhost:8787/api/v1/health

# Login
curl -X POST http://localhost:8787/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'

# List vinyls (usando token)
curl http://localhost:8787/api/v1/vinyls \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Bruno Collection

Use a Bruno collection configurada com ambiente Local:
- Base URL: `http://localhost:8787`
- Environment: Local

---

## Troubleshooting

### Porta 8787 já em uso

```bash
# Mudar porta no wrangler.toml
[dev]
port = 8788  # Use outra porta
```

### D1 local não funciona

```bash
# Recriar banco local
npx wrangler d1 execute DB --local --command="DROP TABLE IF EXISTS vinyls"
npx wrangler d1 execute DB --local --file=scripts/schema.sql
```

### R2 upload falha localmente

- R2 local é limitado
- Use bucket de produção para testes
- Ou implemente mock para desenvolvimento

### TypeScript compilation errors

```bash
# Limpar cache do TypeScript
rm -rf .wrangler
npm run build
```

---

## Environment Variables

### Backend (Wrangler)

```bash
# Usar secrets via Wrangler
npx wrangler secret put JWT_SECRET
npx wrangler secret put ADMIN_PASSWORD
```

### Frontend

```bash
# frontend/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8787
```

---

## Produtos Alternativos

### Miniflare (Wrangler alternativo)

```bash
npm install -D miniflare
```

Miniflare é um runtime local para Cloudflare Workers, mas Wrangler dev é mais completo.

---

## Workflow de Desenvolvimento

### Fluxo Típico

1. **Iniciar backend**
   ```bash
   npm run dev
   ```

2. **Iniciar frontend**
   ```bash
   cd frontend && npm run dev
   ```

3. **Fazer mudanças**
   - Backend: hot reload automático
   - Frontend: hot reload automático

4. **Testar mudanças**
   - API em `http://localhost:8787`
   - Frontend em `http://localhost:3000`

5. **Commit e push**
   ```bash
   git add .
   git commit -m "feat: nova feature"
   git push origin develop
   ```

6. **Deploy automático**
   - GitHub Actions cria release
   - Cloudflare deploya automaticamente

---

## Performance

### Backend (Wrangler Dev)

- Startup: ~1-2 segundos
- Hot reload: <1 segundo
- D1 queries: <10ms
- HTTP requests: <5ms

### Frontend (Next.js Dev)

- Startup: ~5-10 segundos
- Hot reload: <1 segundo
- Page refresh: ~2-3 segundos

---

## Limitações Locais vs Produção

| Feature | Local | Produção |
|---------|-------|----------|
| D1 | SQLite local | D1 Cloudflare |
| R2 | Limitado/Prod | R2 Cloudflare |
| Observability | Terminal logs | Workers Logs + Traces |
| Edge Network | Não | 330+ locations |
| Cold starts | Não | Sim |
| Global deployment | Não | Sim |

---

## Best Practices

### 1. Usar Wrangler dev sempre
- Simula Workers perfeitamente
- Mais rápido que Docker
- Hot reload nativo

### 2. Manter secrets locais
- Usar `.secrets` local
- Não commitar secrets
- Usar Wrangler secrets para produção

### 3. Testar antes de deploy
- Rodar testes localmente
- Verificar endpoints com Bruno
- Testar integration manual

### 4. Usar branches de feature
- `git checkout -b feature/nova-feature`
- Testar localmente
- Merge para develop quando pronto

---

## Suporte

- Wrangler Docs: https://developers.cloudflare.com/workers/wrangler/
- Cloudflare Workers Docs: https://developers.cloudflare.com/workers/
- Project Issues: GitHub repository