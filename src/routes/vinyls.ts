import { Router } from 'itty-router';
import { Env, Vinyl, Track, AuthRequest } from '../types';
import { v4 as uuidv4 } from 'uuid';

const router = Router();

export async function vinylRoutes(request: AuthRequest, env: Env): Promise<Response> {
  const url = new URL(request.url);
  const method = request.method;
  const db = env.DB;

  // GET /api/v1/vinyls - List all vinyls
  if (url.pathname === '/api/v1/vinyls' && method === 'GET') {
    const { results } = await db.prepare('SELECT * FROM vinyls ORDER BY created_at DESC').all<Vinyl>();
    return new Response(JSON.stringify(results), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  // GET /api/v1/vinyls/:id - Get vinyl by ID
  const vinylMatch = url.pathname.match(/^\/api\/v1\/vinyls\/([^\/]+)$/);
  if (vinylMatch && method === 'GET') {
    const vinylId = vinylMatch[1];
    const vinyl = await db.prepare('SELECT * FROM vinyls WHERE id = ?').bind(vinylId).first<Vinyl>();
    
    if (!vinyl) {
      return new Response(JSON.stringify({ error: 'Vinyl not found' }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' }
      });
    }

    return new Response(JSON.stringify(vinyl), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  // POST /api/v1/vinyls - Create vinyl
  if (url.pathname === '/api/v1/vinyls' && method === 'POST') {
    try {
      const body = await request.json() as Partial<Vinyl>;
      const id = uuidv4();
      const now = new Date().toISOString();

      await db.prepare(`
        INSERT INTO vinyls (id, title, artist, year, genre, label, description, cover_url, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `).bind(
        id,
        body.title,
        body.artist,
        body.year,
        body.genre,
        body.label,
        body.description || null,
        body.cover_url || null,
        now,
        now
      ).run();

      const vinyl = await db.prepare('SELECT * FROM vinyls WHERE id = ?').bind(id).first<Vinyl>();
      return new Response(JSON.stringify(vinyl), {
        status: 201,
        headers: { 'Content-Type': 'application/json' }
      });
    } catch (error) {
      return new Response(JSON.stringify({ error: 'Failed to create vinyl' }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  // PUT /api/v1/vinyls/:id - Update vinyl
  const vinylUpdateMatch = url.pathname.match(/^\/api\/v1\/vinyls\/([^\/]+)$/);
  if (vinylUpdateMatch && method === 'PUT') {
    const vinylId = vinylUpdateMatch[1];
    try {
      const body = await request.json() as Partial<Vinyl>;
      const now = new Date().toISOString();

      await db.prepare(`
        UPDATE vinyls 
        SET title = ?, artist = ?, year = ?, genre = ?, label = ?, description = ?, cover_url = ?, updated_at = ?
        WHERE id = ?
      `).bind(
        body.title,
        body.artist,
        body.year,
        body.genre,
        body.label,
        body.description || null,
        body.cover_url || null,
        now,
        vinylId
      ).run();

      const vinyl = await db.prepare('SELECT * FROM vinyls WHERE id = ?').bind(vinylId).first<Vinyl>();
      return new Response(JSON.stringify(vinyl), {
        headers: { 'Content-Type': 'application/json' }
      });
    } catch (error) {
      return new Response(JSON.stringify({ error: 'Failed to update vinyl' }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  // DELETE /api/v1/vinyls/:id - Delete vinyl
  const vinylDeleteMatch = url.pathname.match(/^\/api\/v1\/vinyls\/([^\/]+)$/);
  if (vinylDeleteMatch && method === 'DELETE') {
    const vinylId = vinylDeleteMatch[1];
    await db.prepare('DELETE FROM vinyls WHERE id = ?').bind(vinylId).run();
    return new Response(null, { status: 204 });
  }

  // POST /api/v1/vinyls/:id/favorite - Toggle favorite
  const favoriteMatch = url.pathname.match(/^\/api\/v1\/vinyls\/([^\/]+)\/favorite$/);
  if (favoriteMatch && method === 'POST') {
    const vinylId = favoriteMatch[1];
    // Implement favorite logic based on user
    return new Response(JSON.stringify({ message: 'Favorite toggled' }), {
      headers: { 'Content-Type': 'application/json' }
    });
  }

  // Tracks routes
  const tracksMatch = url.pathname.match(/^\/api\/v1\/vinyls\/([^\/]+)\/tracks$/);
  if (tracksMatch) {
    const vinylId = tracksMatch[1];

    // GET tracks
    if (method === 'GET') {
      const { results } = await db.prepare('SELECT * FROM tracks WHERE vinyl_id = ? ORDER BY position')
        .bind(vinylId).all<Track>();
      return new Response(JSON.stringify(results), {
        headers: { 'Content-Type': 'application/json' }
      });
    }

    // POST track
    if (method === 'POST') {
      try {
        const body = await request.json() as Partial<Track>;
        const id = uuidv4();
        const now = new Date().toISOString();

        await db.prepare(`
          INSERT INTO tracks (id, vinyl_id, title, position, lyrics, created_at, updated_at)
          VALUES (?, ?, ?, ?, ?, ?, ?)
        `).bind(
          id,
          vinylId,
          body.title,
          body.position,
          body.lyrics || null,
          now,
          now
        ).run();

        const track = await db.prepare('SELECT * FROM tracks WHERE id = ?').bind(id).first<Track>();
        return new Response(JSON.stringify(track), {
          status: 201,
          headers: { 'Content-Type': 'application/json' }
        });
      } catch (error) {
        return new Response(JSON.stringify({ error: 'Failed to create track' }), {
          status: 500,
          headers: { 'Content-Type': 'application/json' }
        });
      }
    }
  }

  // Individual track routes
  const trackMatch = url.pathname.match(/^\/api\/v1\/vinyls\/([^\/]+)\/tracks\/([^\/]+)$/);
  if (trackMatch) {
    const vinylId = trackMatch[1];
    const trackId = trackMatch[2];

    // PUT track
    if (method === 'PUT') {
      try {
        const body = await request.json() as Partial<Track>;
        const now = new Date().toISOString();

        await db.prepare(`
          UPDATE tracks 
          SET title = ?, position = ?, lyrics = ?, updated_at = ?
          WHERE id = ? AND vinyl_id = ?
        `).bind(
          body.title,
          body.position,
          body.lyrics || null,
          now,
          trackId,
          vinylId
        ).run();

        const track = await db.prepare('SELECT * FROM tracks WHERE id = ?').bind(trackId).first<Track>();
        return new Response(JSON.stringify(track), {
          headers: { 'Content-Type': 'application/json' }
        });
      } catch (error) {
        return new Response(JSON.stringify({ error: 'Failed to update track' }), {
          status: 500,
          headers: { 'Content-Type': 'application/json' }
        });
      }
    }

    // DELETE track
    if (method === 'DELETE') {
      await db.prepare('DELETE FROM tracks WHERE id = ? AND vinyl_id = ?').bind(trackId, vinylId).run();
      return new Response(null, { status: 204 });
    }
  }

  return new Response('Not Found', { status: 404 });
}