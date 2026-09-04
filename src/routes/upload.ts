import { Router } from 'itty-router';
import { Env, AuthRequest } from '../types';
import { v4 as uuidv4 } from 'uuid';

const router = Router();

export async function uploadRoutes(request: AuthRequest, env: Env): Promise<Response> {
  const url = new URL(request.url);
  const method = request.method;

  // POST /api/v1/vinyls/:id/cover - Upload cover image
  const coverMatch = url.pathname.match(/^\/api\/v1\/vinyls\/([^\/]+)\/cover$/);
  if (coverMatch && method === 'POST') {
    const vinylId = coverMatch[1];
    
    try {
      const formData = await request.formData();
      const file = formData.get('file');
      
      if (!file) {
        return new Response(JSON.stringify({ error: 'No file provided' }), {
          status: 400,
          headers: { 'Content-Type': 'application/json' }
        });
      }

      if (typeof file === 'string') {
        return new Response(JSON.stringify({ error: 'Invalid file' }), {
          status: 400,
          headers: { 'Content-Type': 'application/json' }
        });
      }

      const fileObj = file as unknown as File;
      const fileId = uuidv4();
      const extension = fileObj.name.split('.').pop();
      const filename = `${fileId}.${extension}`;

      await env.UPLOADS.put(filename, fileObj.stream(), {
        httpMetadata: {
          contentType: fileObj.type
        }
      });

      // Update vinyl with cover URL
      await env.DB.prepare('UPDATE vinyls SET cover_url = ? WHERE id = ?')
        .bind(`/uploads/${filename}`, vinylId).run();

      return new Response(JSON.stringify({ 
        cover_url: `/uploads/${filename}`,
        filename 
      }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      });
    } catch (error) {
      return new Response(JSON.stringify({ error: 'Failed to upload cover' }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  // POST /api/v1/profile/photo - Upload profile photo
  if (url.pathname === '/api/v1/profile/photo' && method === 'POST') {
    try {
      const formData = await request.formData();
      const file = formData.get('file');
      
      if (!file) {
        return new Response(JSON.stringify({ error: 'No file provided' }), {
          status: 400,
          headers: { 'Content-Type': 'application/json' }
        });
      }

      if (typeof file === 'string') {
        return new Response(JSON.stringify({ error: 'Invalid file' }), {
          status: 400,
          headers: { 'Content-Type': 'application/json' }
        });
      }

      const fileObj = file as unknown as File;
      const fileId = uuidv4();
      const extension = fileObj.name.split('.').pop();
      const filename = `profile-${fileId}.${extension}`;

      await env.UPLOADS.put(filename, fileObj.stream(), {
        httpMetadata: {
          contentType: fileObj.type
        }
      });

      // Update profile with photo URL
      await env.DB.prepare('UPDATE profiles SET photo_url = ? WHERE username = ?')
        .bind(`/uploads/${filename}`, env.ADMIN_USERNAME).run();

      return new Response(JSON.stringify({ 
        photo_url: `/uploads/${filename}`,
        filename 
      }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      });
    } catch (error) {
      return new Response(JSON.stringify({ error: 'Failed to upload photo' }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  return new Response('Not Found', { status: 404 });
}