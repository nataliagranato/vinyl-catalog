# Cloudflare GitHub Integration Setup Guide

## Prerequisites

✅ Project is ready:
- `npm run build` configured
- `npm run deploy` configured
- `wrangler.toml` properly configured
- Custom domain configured

## Step-by-Step Setup

### Step 1: Access Cloudflare Dashboard

1. Go to: https://dash.cloudflare.com
2. Navigate to: **Workers & Pages**
3. Select: **vinyl-catalog-api**

### Step 2: Access Builds Settings

1. Click on **Settings** tab
2. Click on **Builds** section
3. Click on **Connect** button (under "Git integration")

### Step 3: Authorize GitHub

1. You'll be redirected to GitHub
2. Authorize the "Cloudflare Workers & Pages" app
3. Select the organization/repo: `vinyl-catalog`
4. Grant permissions (read/write for deployments)

### Step 4: Configure Build Settings

After connecting, you'll see build configuration options:

**Production Branch:**
- Set to: `main`

**Build Command:**
- Set to: `npm run build`
- This compiles TypeScript to JavaScript

**Deploy Command:**
- Set to: `npm run deploy`
- This deploys the Worker to Cloudflare

**Root Directory:**
- Set to: `/` (root of the project)

**Environment Variables:**
- These will be automatically pulled from your Cloudflare Worker environment

### Step 5: Configure Environment Variables (Optional)

If you need environment-specific variables:

1. In Builds settings, click on **Environment Variables**
2. Add variables for different environments (production, preview)
3. For example:
   - `APP_ENV`: `production`
   - `API_VERSION`: `v1`

### Step 6: Trigger Initial Build

**Option A: Automatic (Recommended)**
- The build will trigger automatically on the next push to `main`

**Option B: Manual**
- Click on **Trigger Build** in the Builds section
- This will create the first deployment

### Step 7: Verify Deployment

1. In the Builds section, you'll see build history
2. Click on the latest build to see:
   - Build logs
   - Deployment status
   - Version ID
   - Deployment time

### Step 8: Test the Deployed API

```bash
# Test health check
curl https://vinyl-catalog-api.nataliagranato.xyz/api/v1/health

# Test login
curl -X POST https://vinyl-catalog-api.nataliagranato.xyz/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TWTO/VsWHuV+IuKlYXsZTw=="}'
```

## How It Works

### On Push to Main Branch
1. GitHub triggers Cloudflare webhook
2. Cloudflare pulls the code
3. Runs `npm run build` (TypeScript compilation)
4. Runs `npm run deploy` (Worker deployment)
5. Deploys to production (vinyl-catalog-api.nataliagranato.xyz)

### On Pull Request
1. GitHub triggers Cloudflare webhook
2. Cloudflare pulls the code
3. Runs `npm run build`
4. Runs `wrangler versions upload` (preview deployment)
5. Creates preview URL for testing
6. You can test before merging

### After Merge
1. Preview version is promoted to production
2. Main branch triggers full deployment
3. Custom domain is updated

## Preview URLs

When you create a PR or push to a non-main branch:

1. Cloudflare automatically creates a preview deployment
2. Preview URL format: `https://*.workers.dev` or custom preview domain
3. You can test changes before merging
4. Preview URL shows in GitHub PR checks

## Monitoring

### In GitHub
- Check the **Checks** tab in your PR
- See build status directly in GitHub
- View build logs if there are failures

### In Cloudflare Dashboard
- Go to Workers & Pages → vinyl-catalog-api → Builds
- See full build history
- View detailed logs
- Download logs if needed

## Troubleshooting

### Build Fails
1. Check build logs in Cloudflare Dashboard
2. Verify `npm run build` works locally
3. Check TypeScript compilation errors
4. Verify wrangler.toml is correct

### Deploy Fails
1. Check that `wrangler.toml` has correct account_id
2. Verify custom domain is properly configured
3. Check environment variables are set
4. Review deploy logs for specific errors

### Preview URLs Not Working
1. Ensure `workers_dev` is enabled in wrangler.toml
2. Check that preview URLs are enabled in build settings
3. Verify branch is not set as production branch

### Custom Domain Not Updating
1. Verify custom domain is configured in wrangler.toml
2. Check DNS is pointing to Cloudflare
3. Ensure SSL certificate is issued (automatic with custom_domain=true)

## Best Practices

### Branch Strategy
- `main` - Production branch
- `feature/*` - Feature branches (creates preview deployments)
- `hotfix/*` - Hotfix branches (urgent fixes)

### Before Merging
1. Check PR checks pass
2. Test preview URL
3. Run tests locally: `npm test`
4. Run lint: `npm run lint`

### Rollback
If a deployment causes issues:
1. Go to Cloudflare Dashboard → Workers & Pages → vinyl-catalog-api
2. Click on **Deployments**
3. Find the previous working version
4. Click **Rollback** to restore that version

## File Structure Considerations

The build configuration expects:
- `package.json` in root directory
- `wrangler.toml` in root directory
- `src/index.ts` as main entry point
- TypeScript files in `src/` directory

## Advanced Configuration

### Custom Build Commands
If you need custom build steps:

```toml
# In wrangler.toml (if using wrangler.jsonc)
{
  "build_command": "npm run build && npm run test",
  "deploy_command": "npm run deploy"
}
```

### Multiple Environments
For staging/production:

1. Create separate Workers in Cloudflare
2. Use environment variables to differentiate
3. Configure different build commands per environment

### Notifications
Enable build notifications:
1. In Cloudflare Dashboard → Builds → Notifications
2. Configure Slack, email, or Discord notifications
3. Get alerts on build failures/success

## Security

### Secrets Management
- Use Cloudflare Secrets for sensitive data (JWT_SECRET, ADMIN_PASSWORD)
- Never commit secrets to GitHub
- Build configuration automatically uses Cloudflare secrets

### Access Control
- Ensure GitHub repository has proper access controls
- Limit who can trigger builds
- Use protected branches if needed

## Current Configuration Summary

**Project:** vinyl-catalog-api
**Repository:** vinyl-catalog (GitHub)
**Custom Domain:** vinyl-catalog-api.nataliagranato.xyz
**Build Command:** npm run build
**Deploy Command:** npm run deploy
**Production Branch:** main

## Next Steps

1. Open Cloudflare Dashboard at: https://dash.cloudflare.com
2. Navigate to Workers & Pages → vinyl-catalog-api
3. Go to Settings → Builds
4. Click "Connect" to start GitHub integration
5. Follow the on-screen prompts
6. Configure build settings as described above
7. Trigger initial build to verify setup

## Support

- Cloudflare Docs: https://developers.cloudflare.com/workers/ci-cd/builds/git-integration/github-integration/
- GitHub App Issues: https://github.com/cloudflare/github-actions/issues