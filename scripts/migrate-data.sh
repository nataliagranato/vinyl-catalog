#!/bin/bash

# Data migration script from PostgreSQL to Cloudflare D1
# This script exports data from PostgreSQL and imports it to D1

set -e

echo "🔄 Starting data migration from PostgreSQL to Cloudflare D1..."

# Check if PostgreSQL is running
if ! docker ps | grep -q vinyl-catalog-db; then
    echo "❌ PostgreSQL container not found. Starting docker-compose..."
    docker-compose up -d db
    sleep 5
fi

# Export data from PostgreSQL
echo "📤 Exporting data from PostgreSQL..."
docker exec vinyl-catalog-db-1 pg_dump -U postgres -d vinyl_catalog --data-only --disable-triggers > postgres_export.sql

# Convert PostgreSQL export to SQLite format
echo "🔄 Converting to SQLite format..."
# This is a simplified conversion - you may need to adjust based on your actual data
cat > sqlite_import.sql << 'EOF'
-- Clear existing data
DELETE FROM tracks;
DELETE FROM vinyls;
DELETE FROM profiles;

-- Import vinyls (you'll need to manually map the data)
-- Copy the INSERT statements from postgres_export.sql and modify for SQLite
-- Example: 
-- INSERT INTO vinyls (id, title, artist, year, genre, label, description, cover_url, created_at, updated_at)
-- VALUES ('uuid-here', 'Album Title', 'Artist', 2024, 'Genre', 'Label', 'Description', '/uploads/cover.jpg', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z');

-- Import tracks (similarly modify INSERT statements)
-- Import profiles (similarly modify INSERT statements)
EOF

echo "⚠️  Manual step required:"
echo "1. Review postgres_export.sql"
echo "2. Copy INSERT statements to sqlite_import.sql"
echo "3. Adjust syntax for SQLite (remove PostgreSQL-specific functions)"
echo "4. Run: wrangler d1 execute vinyl-catalog-db --file=sqlite_import.sql"

echo "📝 PostgreSQL export saved to postgres_export.sql"
echo "📝 SQLite template saved to sqlite_import.sql"