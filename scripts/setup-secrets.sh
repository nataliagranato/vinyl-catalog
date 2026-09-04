#!/bin/bash

# Setup script for Cloudflare Secrets
# This script will set up secure secrets for the Vinyl Catalog API

echo "🔐 Setting up Cloudflare Secrets for Vinyl Catalog API..."

# Check if secrets already exist
echo "Checking existing secrets..."
npx wrangler secret list

echo ""
echo "Please enter a secure JWT_SECRET (press Enter to generate one):"
read -s JWT_SECRET_INPUT

if [ -z "$JWT_SECRET_INPUT" ]; then
    JWT_SECRET=$(openssl rand -base64 32)
    echo "Generated secure JWT_SECRET"
else
    JWT_SECRET=$JWT_SECRET_INPUT
fi

echo ""
echo "Please enter a secure ADMIN_PASSWORD (press Enter to generate one):"
read -s ADMIN_PASSWORD_INPUT

if [ -z "$ADMIN_PASSWORD_INPUT" ]; then
    ADMIN_PASSWORD=$(openssl rand -base64 16)
    echo "Generated secure ADMIN_PASSWORD"
else
    ADMIN_PASSWORD=$ADMIN_PASSWORD_INPUT
fi

# Set the secrets
echo ""
echo "Setting JWT_SECRET..."
echo "$JWT_SECRET" | npx wrangler secret put JWT_SECRET

echo "Setting ADMIN_PASSWORD..."
echo "$ADMIN_PASSWORD" | npx wrangler secret put ADMIN_PASSWORD

echo ""
echo "✅ Secrets configured successfully!"
echo ""
echo "⚠️  IMPORTANT: Save these credentials securely:"
echo "JWT_SECRET: $JWT_SECRET"
echo "ADMIN_PASSWORD: $ADMIN_PASSWORD"
echo ""
echo "These values are now stored securely in Cloudflare and will not be visible in the code."