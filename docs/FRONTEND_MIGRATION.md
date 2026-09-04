# Frontend Migration to Cloudflare Pages

Guia completo para migrar o frontend Next.js para Cloudflare Pages.

## Overview

**Antes:**
- Next.js frontend local
- Observabilidade com Grafana Faro e OpenTelemetry
- Deploy via Docker

**Depois:**
- Next.js estático em Cloudflare Pages
- Observabilidade via Cloudflare Workers Traces
- Deploy automático via GitHub Actions
- Custom domain: vinyl-catalog.nataliagranato.xyz

---

## Changes Made

### 1. Removed Dependencies

**Removed Grafana Faro and OpenTelemetry:**
```json
// Removed from package.json
"@grafana/faro-web-sdk": "^2.3.1",
"@grafana/faro-web-tracing": "^2.3.1",
"@opentelemetry/api": "^1.9.0",
"@opentelemetry/exporter-trace-otlp-http": "^0.213.0",
"@opentelemetry/resources": "^2.6.0",
"@opentelemetry/sdk-trace-base": "^2.6.0",
"@opentelemetry/sdk-trace-web": "^2.6.0",
"@opentelemetry/semantic-conventions": "^1.40.0",
```

**Reason:** Cloudflare Pages não suporta essas dependências para sites estáticos. Observabilidade será via Cloudflare Workers Traces.

### 2. Updated Next.js Config

**next.config.cloudflare.ts:**
```typescript
import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  output: 'export',
  images: {
    unoptimized: true,
  },
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'https://vinyl-catalog-api.nataliagranato.xyz',
  },
};

export default nextConfig;
```

**Changes:**
- `output: 'export'` - Build estático para Cloudflare Pages
- `images.unoptimized: true` - Cloudflare Pages não otimiza imagens
- API URL padrão atualizado para custom domain

### 3. Created Wrangler Config

**frontend/wrangler.toml:**
```toml
name = "vinyl-catalog-frontend"
compatibility_date = "2024-01-01"
compatibility_flags = ["nodejs_compat"]

# Account configuration
account_id = "4839c9636a58fa9490bbe3d2e686ad98"

# Pages configuration
pages_build_output_dir = ".next"

# Environment variables
[vars]
NEXT_PUBLIC_API_URL = "https://vinyl-catalog-api.nataliagranato.xyz"
```

### 4. Updated Deploy Script

**package.json:**
```json
"deploy:cloudflare": "npx wrangler pages deploy"
```

---

## Deployment Steps

### 1. Build for Cloudflare Pages

```bash
cd frontend
npm install
npm run build:cloudflare
```

### 2. Deploy via Wrangler

```bash
cd frontend
npm run deploy:cloudflare
```

**First deployment will:**
- Create Cloudflare Pages project
- Upload build output
- Provide deployment URL

### 3. Configure Custom Domain

**Option A: Via Dashboard**
1. Acesse: https://dash.cloudflare.com
2. Pages → vinyl-catalog-frontend
3. Custom domains → Add domain
4. Adicionar: vinyl-catalog.nataliagranato.xyz
5. Configurar DNS

**Option B: Via Wrangler (configurar depois via dashboard)**
O custom domain para Pages é configurado diferente de Workers, via dashboard.

---

## GitHub Actions Integration

### Configure Automatic Deployment

**1. Conectar GitHub Repository**
- Dashboard Cloudflare → Pages → vinyl-catalog-frontend
- Settings → Builds → Connect GitHub
- Selecionar repositório

**2. Configurar Build Settings**
- **Build command:** `npm run build:cloudflare`
- **Build output directory:** `.next`
- **Root directory:** `frontend`

**3. Trigger Deployment**
- Push para branch principal
- Cloudflare Pages fará build e deploy automático

---

## Environment Variables

### Production (Cloudflare Pages)

Configurar via Dashboard Cloudflare:
```
NEXT_PUBLIC_API_URL = https://vinyl-catalog-api.nataliagranato.xyz
```

### Local Development

```bash
# frontend/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8787
```

---

## Limitations

### Static Export Limitations

**Next.js Static Export:**
- ❌ No server-side rendering (SSR)
- ❌ No API routes
- ❌ No getServerSideProps
- ❌ No incremental static regeneration (ISR)
- ✅ Static pages only
- ✅ Client-side navigation
- ✅ Image optimization desabilitado

**Workarounds:**
- Usar client-side fetching para dados
- Usar SWR ou React Query para cache
- Usar API Workers para server-side logic

### Image Optimization

Cloudflare Pages não otimiza imagens automaticamente:

**Soluções:**
1. Otimizar imagens antes do upload
2. Usar serviço externo (Cloudflare Images - pago)
3. Usar imagens CDN já otimizadas

---

## Observability

### Cloudflare Pages Analytics

**Dashboard:**
- Pages → vinyl-catalog-frontend → Analytics
- View count, bandwidth, requests
- Page views e visitors

### Workers Traces (Frontend)

O frontend pode enviar traces para Workers Traces:
- Integrar com backend Workers
- Usar fetch para API backend
- Traces correlacionados entre frontend e backend

---

## Testing

### Local Testing

```bash
cd frontend
npm run dev
```

Frontend roda em: `http://localhost:3000`

### Production Testing

Após deploy:
- Acessar URL fornecida pelo Cloudflare
- Testar todas as páginas
- Verificar integração com API backend

---

## Rollback

### Via Dashboard
1. Pages → vinyl-catalog-frontend
2. Deployments
3. Selecionar deployment anterior
4. Rollback

### Via Wrangler
```bash
# Listar deployments
npx wrangler pages deployment list

# Rollback
npx wrangler pages deployment rollback <deployment-id>
```

---

## Troubleshooting

### Build Fails

**Issues comuns:**
- TypeScript errors
- Missing dependencies
- Image optimization conflicts

**Soluções:**
```bash
# Limpar cache
rm -rf .next

# Rebuild
npm run build:cloudflare
```

### Deployment Fails

**Issues comuns:**
- Account ID incorreto
- Build output directory errado
- Build command falha

**Soluções:**
- Verificar wrangler.toml
- Checar logs no dashboard
- Testar build localmente

### API Calls Fail

**Issues comuns:**
- CORS errors
- API URL incorreta
- Token expirado

**Soluções:**
- Verificar NEXT_PUBLIC_API_URL
- Configurar CORS no backend Workers
- Verificar se backend está operacional

---

## Next Steps

### 1. Configure Custom Domain
- Adicionar vinyl-catalog.nataliagranato.xyz via dashboard
- Configurar DNS records

### 2. Setup GitHub Actions
- Conectar repositório GitHub
- Configurar build settings
- Habilitar automatic deployments

### 3. Update DNS
- Configurar CNAME ou A records
- Verificar SSL certificate

### 4. Monitor Analytics
- Configurar Web Analytics (opcional)
- Verificar Page Analytics no dashboard

---

## Costs

**Cloudflare Pages Free Tier:**
- 500 builds per month
- Unlimited bandwidth
- Unlimited requests
- SSL certificates included

**Total: $0/mês**

---

## Support

- Cloudflare Pages Docs: https://developers.cloudflare.com/pages/
- Next.js Static Export: https://nextjs.org/docs/app/building-your-application/deploying/static-exports
- Project Issues: GitHub repository