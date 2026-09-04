#!/bin/bash

# Setup script for Cloudflare R2 storage

echo "Setting up Cloudflare R2 storage for Vinyl Catalog uploads..."

# Create R2 bucket
echo "Creating R2 bucket..."
npx wrangler r2 bucket create vinyl-catalog-uploads

echo "R2 bucket setup complete!"
echo "The bucket is now ready for file uploads"