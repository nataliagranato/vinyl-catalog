# Cloudflare Migration Guide

This guide explains how to migrate the Vinyl Catalog from the current Docker-based setup to Cloudflare's free tier services.

## Architecture Overview

### Current Stack (Docker-based)
- Backend: Go with Gin framework
- Database: PostgreSQL 16 (Docker)
- Storage: Local filesystem
- Observability: Prometheus, Jaeger, Loki, Grafana
- Frontend: Next.js 16

### New Stack (Cloudflare Free Tier)
- Backend: Cloudflare Workers (TypeScript)
- Database: Cloudflare D1 (SQLite)
- Storage: Cloudflare R2 (S3-compatible)
- Observability: Cloudflare Workers Logs + Traces + OpenTelemetry Export
- Security: JWT with HMAC-SHA256 (Web Crypto API)
- Custom Domain: vinyl-catalog-api.nataliagranato.xyz
- Frontend: Cloudflare Pages (Next.js)

## Migration Steps

### 1. Prerequisites
- Install Node.js and npm
- Install Cloudflare Wrangler CLI: `npm install -g wrangler`
- Create a Cloudflare account (free tier)
- Authenticate: `wrangler login`

### 2. Backend Migration (Go → TypeScript Workers)

#### Install dependencies
```bash
npm install
```

#### Configure Wrangler
Update `wrangler.toml` with your Cloudflare account details:
- Run `wrangler whoami` to get your account ID
- Update `account_id` in wrangler.toml

#### Setup Database
```bash
chmod +x setup-db.sh
./setup-db.sh
```
This will:
- Create a D1 database named `vinyl-catalog-db`
- Run the schema migration to create tables
- Output a database ID that you need to add to wrangler.toml

#### Setup R2 Storage
```bash
chmod +x setup-r2.sh
./setup-r2.sh
```
This creates an R2 bucket for file uploads.

#### Update Environment Variables
In `wrangler.toml`, update:
- `JWT_SECRET` - use a strong random string
- `ADMIN_PASSWORD` - set a secure password
- Add the `database_id` from the D1 setup

#### Local Development
```bash
npm run dev
```
This starts the Workers dev server on port 8787.

#### Deploy to Cloudflare
```bash
npm run deploy
```

### 3. Frontend Migration (Next.js)

#### Update Environment Variables
Edit `frontend/.env.production`:
```
NEXT_PUBLIC_API_URL=https://your-workers-url.workers.dev
```

#### Build for Cloudflare Pages
```bash
cd frontend
npm run build:cloudflare
```

#### Deploy to Cloudflare Pages
Option 1: Using Wrangler CLI
```bash
npm run deploy:cloudflare
```

Option 2: Using Cloudflare Dashboard
1. Go to Cloudflare Dashboard → Pages
2. Create new project → Connect to Git
3. Connect your repository
4. Configure build settings:
   - Build command: `npm run build`
   - Build output directory: `.next`
   - Environment variables: Add `NEXT_PUBLIC_API_URL`

### 4. Testing

#### Test Backend
```bash
# Login
curl -X POST https://vinyl-catalog-api.nataliagranato.xyz/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'

# Create vinyl (using token from login)
curl -X POST https://vinyl-catalog-api.nataliagranato.xyz/api/v1/vinyls \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Album","artist":"Test Artist","year":2024,"genre":"Rock","label":"Test Label"}'
```

#### Test Frontend
Navigate to your Cloudflare Pages URL and test the application.

## Key Differences from Original Stack

### Database
- **Before**: PostgreSQL with complex queries
- **After**: SQLite (D1) with simpler schema
- **Note**: D1 is optimized for read-heavy workloads

### Storage
- **Before**: Local filesystem with Docker volumes
- **After**: Cloudflare R2 (S3-compatible, global CDN)
- **Benefit**: Better performance and reliability

### Authentication
- **Before**: JWT with Go implementation
- **After**: JWT with TypeScript (compatible)
- **Note**: Same token format, seamless migration

### Observability
- **Before**: Full stack (Prometheus, Jaeger, Loki, Grafana)
- **After**: Cloudflare Analytics + Workers traces
- **Note**: Simplified but sufficient for most use cases

### Performance
- **Before**: Single region deployment
- **After**: Global edge deployment
- **Benefit**: Low latency worldwide

## Cost Comparison

### Current Stack (Docker-based)
- VPS: $5-20/month minimum
- Database: $5-15/month
- Storage: $5-10/month
- **Total**: $15-45/month

### Cloudflare Free Tier
- Workers: 100k requests/day free
- D1: 5GB storage, 25M reads/day free
- R2: 10GB storage, 1M Class A operations/month free
- Pages: Unlimited bandwidth free
- **Total**: $0/month (within free limits)

## Maintenance

### Database Backups
```bash
# Export D1 database
wrangler d1 export vinyl-catalog-db --output=backup.sql

# Import D1 database
wrangler d1 execute vinyl-catalog-db --file=backup.sql
```

### R2 Management
```bash
# List objects
wrangler r2 object list vinyl-catalog-uploads

# Delete object
wrangler r2 object delete vinyl-catalog-uploads filename.jpg
```

### Logs and Analytics
- Workers logs: `wrangler tail`
- Analytics: Cloudflare Dashboard → Workers → Analytics

## Troubleshooting

### Common Issues

1. **Database connection errors**
   - Ensure database_id is correct in wrangler.toml
   - Check that D1 database exists: `wrangler d1 list`

2. **R2 upload failures**
   - Verify R2 bucket exists: `wrangler r2 bucket list`
   - Check bucket name matches wrangler.toml

3. **JWT authentication issues**
   - Ensure JWT_SECRET is set consistently
   - Check token expiration

4. **Frontend API connection**
   - Verify NEXT_PUBLIC_API_URL is correct
   - Check CORS settings if needed

## Rollback Plan

If you need to rollback to the original Docker setup:

1. Stop Cloudflare deployments
2. Update DNS to point to original server
3. Start original docker-compose: `docker-compose up -d`
4. Restore PostgreSQL database from backup

## Migration Status

### ✅ Completed
- ✅ Backend migrated to Cloudflare Workers (TypeScript)
- ✅ Database migrated to Cloudflare D1
- ✅ Storage migrated to Cloudflare R2
- ✅ JWT authentication with HMAC-SHA256 (Web Crypto API)
- ✅ Structured logging with Workers Logs
- ✅ Automatic tracing with Workers Traces
- ✅ Health check and metrics endpoints
- ✅ Security audit completed (9/10 score)
- ✅ Wrangler updated to v4.129.0
- ✅ OpenTelemetry export configured

### ⏳ Pending
- ⏳ Frontend migration to Cloudflare Pages
- ⏳ OpenTelemetry destination setup (optional)

### Deployed Resources
- **Worker:** https://vinyl-catalog-api.nataliagranato.xyz
- **Database:** vinyl-catalog-db (D1)
- **Storage:** vinyl-catalog-uploads (R2)
- **Account ID:** 4839c9636a58fa9490bbe3d2e686ad98

## Support

- Cloudflare Workers docs: https://developers.cloudflare.com/workers/
- D1 docs: https://developers.cloudflare.com/d1/
- R2 docs: https://developers.cloudflare.com/r2/
- Pages docs: https://developers.cloudflare.com/pages/