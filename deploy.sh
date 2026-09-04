#!/bin/bash

# Complete deployment script for Vinyl Catalog on Cloudflare

set -e

echo "🚀 Starting Vinyl Catalog deployment to Cloudflare..."

# Check if wrangler is installed
if ! command -v wrangler &> /dev/null; then
    echo "❌ Wrangler CLI not found. Installing..."
    npm install -g wrangler
fi

# Check if user is authenticated
echo "🔐 Checking Cloudflare authentication..."
if ! wrangler whoami &> /dev/null; then
    echo "Please authenticate with Cloudflare:"
    wrangler login
fi

# Install dependencies
echo "📦 Installing dependencies..."
npm install

# Setup database if it doesn't exist
echo "🗄️  Setting up D1 database..."
if ! wrangler d1 list | grep -q "vinyl-catalog-db"; then
    echo "Creating new D1 database..."
    wrangler d1 create vinyl-catalog-db
    echo "⚠️  Please update the database_id in wrangler.toml with the output above"
    echo "Then run this script again."
    exit 1
fi

# Setup R2 bucket if it doesn't exist
echo "📁 Setting up R2 storage..."
if ! wrangler r2 bucket list | grep -q "vinyl-catalog-uploads"; then
    echo "Creating new R2 bucket..."
    wrangler r2 bucket create vinyl-catalog-uploads
fi

# Build TypeScript
echo "🔨 Building TypeScript..."
npm run build

# Deploy backend Workers
echo "🌐 Deploying backend Workers..."
npm run deploy

# Deploy frontend
echo "🎨 Deploying frontend to Cloudflare Pages..."
cd frontend
npm install
npm run build:cloudflare
npm run deploy:cloudflare
cd ..

echo "✅ Deployment complete!"
echo ""
echo "📝 Next steps:"
echo "1. Update your DNS to point to the Cloudflare Pages URL"
echo "2. Update frontend/.env.production with the Workers URL"
echo "3. Test the application"
echo ""
echo "🔗 Useful commands:"
echo "- View logs: wrangler tail"
echo "- Database management: wrangler d1 execute vinyl-catalog-db --command='SELECT * FROM vinyls'"
echo "- R2 management: wrangler r2 object list vinyl-catalog-uploads"