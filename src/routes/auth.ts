import { Router } from 'itty-router';
import { Env } from '../types';
import { AuthService } from '../services/auth';
import { logger } from '../utils/logger';

const router = Router();

export async function authRoutes(request: Request, env: Env): Promise<Response> {
  const authService = new AuthService(env);
  const url = new URL(request.url);
  const method = request.method;

  if (url.pathname === '/api/v1/auth/login' && method === 'POST') {
    const startTime = Date.now();
    
    try {
      const body = await request.json() as { username: string; password: string };
      
      logger.info('Login attempt', {
        username: body.username,
        endpoint: '/api/v1/auth/login',
        method: 'POST'
      });

      const token = await authService.login(body.username, body.password);
      
      if (!token) {
        logger.authAttempt(body.username, false);
        
        return new Response(JSON.stringify({ error: 'Invalid credentials' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' }
        });
      }

      logger.authAttempt(body.username, true);
      logger.apiRequest('POST', '/api/v1/auth/login', 200, Date.now() - startTime);

      return new Response(JSON.stringify({ token }), {
        headers: { 'Content-Type': 'application/json' }
      });
    } catch (error) {
      logger.error('Login request failed', error as Error, {
        endpoint: '/api/v1/auth/login',
        method: 'POST'
      });
      
      return new Response(JSON.stringify({ error: 'Invalid request' }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' }
      });
    }
  }

  return new Response('Not Found', { status: 404 });
}