export interface Env {
  DB: D1Database;
  UPLOADS: R2Bucket;
  JWT_SECRET?: string; // Can be from env vars or secrets
  ADMIN_PASSWORD?: string; // Can be from env vars or secrets
  JWT_EXPIRATION_HOURS: string;
  ADMIN_USERNAME: string;
  APP_ENV: string;
}

export interface Vinyl {
  id: string;
  title: string;
  artist: string;
  year: number;
  genre: string;
  label: string;
  description?: string;
  cover_url?: string;
  created_at: string;
  updated_at: string;
}

export interface Track {
  id: string;
  vinyl_id: string;
  title: string;
  position: number;
  lyrics?: string;
  created_at: string;
  updated_at: string;
}

export interface Profile {
  username: string;
  display_name?: string;
  bio?: string;
  photo_url?: string;
  links?: string[];
  preferred_genres?: string[];
  favorite_vinyl_ids?: string[];
}

export interface AuthRequest extends Request {
  user?: {
    username: string;
  };
}