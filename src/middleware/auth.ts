import { Env, AuthRequest } from '../types';
import { JWTService } from '../services/jwt';

export async function authMiddleware(request: Request, env: Env): Promise<Response> {
  const authHeader = request.headers.get('Authorization');
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return new Response(JSON.stringify({ error: 'Unauthorized' }), {
      status: 401,
      headers: { 'Content-Type': 'application/json' }
    });
  }

  const token = authHeader.substring(7);
  const jwtSecret = env.JWT_SECRET || 'default-insecure-secret';
  const jwtService = new JWTService(jwtSecret);
  
  const decoded = await jwtService.verify(token);
  
  if (!decoded) {
    return new Response(JSON.stringify({ error: 'Invalid token' }), {
      status: 401,
      headers: { 'Content-Type': 'application/json' }
    });
  }

  (request as AuthRequest).user = { username: decoded.username };
  return null as any; // Continue to next handler
}