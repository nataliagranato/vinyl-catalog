# Vinyl Catalog - Cloudflare Deployment

🎵 Vinyl Catalog deployed on Cloudflare's free tier services with full observability.

## Quick Start

### Prerequisites
- Node.js 18+
- Cloudflare account (free)
- Wrangler CLI v4+

### Installation

1. **Install dependencies**
```bash
npm install
```

2. **Authenticate with Cloudflare**
```bash
npx wrangler login
```

3. **Configure environment**
Edit `wrangler.toml`:
- Update `account_id` (run `npx wrangler whoami` to get it)
- Set secure values for `JWT_SECRET` and `ADMIN_PASSWORD` (see `.secrets`)
- Add `database_id` after creating the database

4. **Setup database**
```bash
chmod +x setup-db.sh
./setup-db.sh
```

5. **Setup storage**
```bash
chmod +x setup-r2.sh
./setup-r2.sh
```

6. **Deploy**
```bash
npm run deploy
```

## Architecture

```
Frontend (Cloudflare Pages)
    ↓
Backend (Cloudflare Workers)
    ↓
Database (Cloudflare D1)
    ↓
Storage (Cloudflare R2)
```

## Development

### Local development
```bash
npm run dev
```

### Run tests
```bash
npm test
```

### Build
```bash
npm run build
```

## API Endpoints

### System Endpoints
- `GET /` - API documentation and status page
- `GET /api/v1/health` - Health check (verifies DB, R2, API status)
- `GET /api/v1/metrics` - API metrics (vinyls count, tracks count, storage stats)

### Authentication
- `POST /api/v1/auth/login` - Login and get JWT token
  - **Credentials:** Username: `admin`, Password: see `.secrets` file

### Vinyls (Requires Authentication)
- `GET /api/v1/vinyls` - List all vinyls
- `GET /api/v1/vinyls/:id` - Get specific vinyl
- `POST /api/v1/vinyls` - Create vinyl
- `PUT /api/v1/vinyls/:id` - Update vinyl
- `DELETE /api/v1/vinyls/:id` - Delete vinyl
- `POST /api/v1/vinyls/:id/favorite` - Toggle favorite

### Tracks (Requires Authentication)
- `GET /api/v1/vinyls/:id/tracks` - List tracks
- `POST /api/v1/vinyls/:id/tracks` - Create track
- `PUT /api/v1/vinyls/:id/tracks/:track_id` - Update track
- `DELETE /api/v1/vinyls/:id/tracks/:track_id` - Delete track

### Profile
- `GET /api/v1/profile` - Get public profile (no auth required)
- `PUT /api/v1/profile` - Update profile (requires auth)
- `POST /api/v1/profile/photo` - Upload profile photo (requires auth)

### Uploads (Requires Authentication)
- `POST /api/v1/vinyls/:id/cover` - Upload cover image
- `GET /uploads/:filename` - Serve uploaded files

## Security & Observability

### Security Features
- ✅ **JWT Authentication** with HMAC-SHA256 (Web Crypto API)
- ✅ **Secure credentials** (cryptographically generated)
- ✅ **Structured logging** with JSON format
- ✅ **Automatic tracing** for all requests
- ✅ **Health checks** with database connectivity
- ✅ **Metrics endpoints** for monitoring

### Observability
- **Workers Logs:** Structured JSON logs stored in Cloudflare
- **Workers Traces:** Automatic tracing with 10% sampling
- **Real-time logs:** View via `npx wrangler tail`
- **Dashboard access:** Cloudflare Dashboard → Workers & Pages → Observability
- **OpenTelemetry Export:** Configured (see `OPENTELEMETRY_SETUP.md`)

### View Logs
```bash
# Real-time logs
npx wrangler tail

# View in dashboard
# Visit: https://dash.cloudflare.com → Workers & Pages → vinyl-catalog-api → Observability
```

### Health Check
```bash
curl https://vinyl-catalog-api.nataliagranato.xyz/api/v1/health
```

### Metrics
```bash
curl https://vinyl-catalog-api.nataliagranato.xyz/api/v1/metrics
```

## Free Tier Limits

### Cloudflare Workers
- 100,000 requests/day
- 10ms CPU time per request
- 128MB memory

### Cloudflare D1
- 5GB storage
- 25M row reads/day
- 5M row writes/day

### Cloudflare R2
- 10GB storage
- 1M Class A operations/month
- 10M Class B operations/month
- Unlimited egress (free!)

### Cloudflare Pages
- Unlimited bandwidth
- 500 builds/month
- Global CDN

## Monitoring & Debugging

### Real-time Logs
```bash
npx wrangler tail
```

### Dashboard Observability
1. Visit https://dash.cloudflare.com
2. Navigate to Workers & Pages
3. Select vinyl-catalog-api
4. Click Observability
5. View Logs, Traces, and Metrics

### Database Queries
```bash
npx wrangler d1 execute vinyl-catalog-db --command="SELECT * FROM vinyls"
```

### Analytics
Visit Cloudflare Dashboard → Workers → Analytics

## Backup & Restore

### Backup Database
```bash
npx wrangler d1 export vinyl-catalog-db --output=backup.sql
```

### Restore Database
```bash
npx wrangler d1 execute vinyl-catalog-db --file=backup.sql
```

## Troubleshooting

### Authentication fails
- Check credentials in `.secrets` file
- Verify JWT_SECRET in wrangler.toml
- Ensure token is not expired (24h default)

### Database errors
- Ensure database_id is correct in wrangler.toml
- Check table schema with `npx wrangler d1 execute vinyl-catalog-db --command="SELECT * FROM sqlite_master"`
- Verify health check: `curl /api/v1/health`

### Upload failures
- Verify R2 bucket exists
- Check file size limits (10MB default)
- Ensure R2 binding is configured in wrangler.toml

### Observability issues
- Check [observability] section in wrangler.toml
- Verify Wrangler version: `npx wrangler --version` (should be v4+)
- Review logs in dashboard Cloudflare

## Migration from Docker

See [CLOUDFLARE_MIGRATION.md](./CLOUDFLARE_MIGRATION.md) for detailed migration guide.

## Security Documentation

See [SECURITY_AUDIT_12FACTOR.md](./SECURITY_AUDIT_12FACTOR.md) for:
- Complete security audit based on 12-Factor App methodology
- Vulnerability assessment and corrections
- Security score: 9/10
- Conformidade 12-Factor: 78%

## OpenTelemetry Export

See [OPENTELEMETRY_SETUP.md](./OPENTELEMETRY_SETUP.md) for:
- Guide to configure OpenTelemetry export
- Instructions for Grafana, Sentry, Honeycomb, Datadog
- Step-by-step destination setup

## Cost Comparison

| Service | Docker | Cloudflare Free |
|---------|--------|-----------------|
| Hosting | $15-45/mo | $0 |
| Database | $5-15/mo | $0 |
| Storage | $5-10/mo | $0 |
| Bandwidth | Variable | $0 |
| Observability | $10-30/mo | $0 |
| **Total** | **$35-100/mo** | **$0** |

## Current Status

- ✅ **Backend deployed:** https://vinyl-catalog-api.nataliagranato.xyz
- ✅ **Custom domain:** vinyl-catalog-api.nataliagranato.xyz
- ✅ **Security score:** 9/10
- ✅ **Observability:** Full (Logs + Traces + Metrics)
- ✅ **Wrangler version:** v4.129.0
- ⏳ **Frontend:** Pending migration

## Support

- Cloudflare Docs: https://developers.cloudflare.com/
- Project Issues: Check GitHub repository
- Dashboard: https://dash.cloudflare.com