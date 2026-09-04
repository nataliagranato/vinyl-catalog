import { Router } from 'itty-router';
import { Env, Profile, AuthRequest } from '../types';

const router = Router();

export async function profileRoutes(request: AuthRequest, env: Env): Promise<Response> {
  const url = new URL(request.url);
  const method = request.method;
  const db = env.DB;

  // GET /api/v1/profile - Get public profile
  if (url.pathname === '/api/v1/profile' && method === 'GET') {
    const profile = await db.prepare('SELECT * FROM profiles WHERE username = ?')
      .bind(env.ADMIN_USERNAME).first<Profile>();
    
    if (!profile) {
      return new Response(JSON.stringify({ error: 'Profile not found' }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' }
      });
    }

    return new Response(JSON.stringify(profile), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  // PUT /api/v1/profile - Update profile
  if (url.pathname === '/api/v1/profile' && method === 'PUT') {
    try {
      const body = await request.json() as Partial<Profile>;
      
      await db.prepare(`
        INSERT INTO profiles (username, display_name, bio, photo_url, links, preferred_genres, favorite_vinyl_ids)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(username) DO UPDATE SET
          display_name = excluded.display_name,
          bio = excluded.bio,
          photo_url = excluded.photo_url,
          links = excluded.links,
          preferred_genres = excluded.preferred_genres,
          favorite_vinyl_ids = excluded.favorite_vinyl_ids
      `).bind(
        env.ADMIN_USERNAME,
        body.display_name || null,
        body.bio || null,
        body.photo_url || null,
        JSON.stringify(body.links || []),
        JSON.stringify(body.preferred_genres || []),
        JSON.stringify(body.favorite_vinyl_ids || [])
      ).run();

      const profile = await db.prepare('SELECT * FROM profiles WHERE username = ?')
        .bind(env.ADMIN_USERNAME).first<Profile>();
      
      return new Response(JSON.stringify(profile), {
        headers: { 'Content-Type': 'application/json' }
      });
    } catch (error) {
      return new Response(JSON.stringify({ error: 'Failed to update profile' }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  return new Response('Not Found', { status: 404 });
}