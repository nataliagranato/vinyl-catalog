-- Vinyls table
CREATE TABLE IF NOT EXISTS vinyls (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  artist TEXT NOT NULL,
  year INTEGER NOT NULL,
  genre TEXT NOT NULL,
  label TEXT NOT NULL,
  description TEXT,
  cover_url TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- Tracks table
CREATE TABLE IF NOT EXISTS tracks (
  id TEXT PRIMARY KEY,
  vinyl_id TEXT NOT NULL,
  title TEXT NOT NULL,
  position INTEGER NOT NULL,
  lyrics TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (vinyl_id) REFERENCES vinyls(id) ON DELETE CASCADE
);

-- Profiles table
CREATE TABLE IF NOT EXISTS profiles (
  username TEXT PRIMARY KEY,
  display_name TEXT,
  bio TEXT,
  photo_url TEXT,
  links TEXT, -- JSON array
  preferred_genres TEXT, -- JSON array
  favorite_vinyl_ids TEXT -- JSON array
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_vinyls_created_at ON vinyls(created_at);
CREATE INDEX IF NOT EXISTS idx_tracks_vinyl_id ON tracks(vinyl_id);
CREATE INDEX IF NOT EXISTS idx_tracks_position ON tracks(vinyl_id, position);