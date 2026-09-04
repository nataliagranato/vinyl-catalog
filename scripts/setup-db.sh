#!/bin/bash

# Setup script for Cloudflare D1 database

echo "Setting up Cloudflare D1 database for Vinyl Catalog..."

# Create D1 database
echo "Creating D1 database..."
npx wrangler d1 create vinyl-catalog-db

# Note: The command above will output a database_id. 
# You need to copy that ID and update wrangler.toml

# Create the database tables
echo "Creating database schema..."
npx wrangler d1 execute vinyl-catalog-db --file=./schema.sql

echo "Database setup complete!"
echo "Please update wrangler.toml with the database_id from the create command output"