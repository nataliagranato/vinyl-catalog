import { Router } from 'itty-router';
import { Env } from './types';
import { authRoutes } from './routes/auth';
import { vinylRoutes } from './routes/vinyls';
import { profileRoutes } from './routes/profile';
import { uploadRoutes } from './routes/upload';
import { authMiddleware } from './middleware/auth';

const router = Router();

// Root endpoint - API information (Vercel-inspired design, Swagger-style organization)
router.get('/', () => {
  const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Vinyl Catalog API</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #0a0a0a;
            color: #fafafa;
            line-height: 1.6;
        }
        
        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 0 40px;
        }
        
        .header {
            padding: 60px 0 40px;
            border-bottom: 1px solid #333;
        }
        
        .header h1 {
            font-size: 3rem;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 16px;
            background: linear-gradient(180deg, #fff 0%, #888 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .header p {
            color: #888;
            font-size: 1.125rem;
            max-width: 800px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            margin-bottom: 24px;
        }
        
        .version {
            display: none;
        }
        
        .status {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            background: #1a1a1a;
            border: 1px solid #333;
            padding: 12px 24px;
            border-radius: 8px;
            font-size: 0.875rem;
            color: #22c55e;
            transition: all 0.3s;
        }
        
        .status.error {
            background: #1a1a1a;
            border-color: #ef4444;
            color: #ef4444;
        }
        
        .status.warning {
            background: #1a1a1a;
            border-color: #f59e0b;
            color: #f59e0b;
        }
        
        .status.error .status-dot {
            background: #ef4444;
        }
        
        .status.warning .status-dot {
            background: #f59e0b;
        }
        
        .status-dot {
            width: 8px;
            height: 8px;
            background: #22c55e;
            border-radius: 50%;
            animation: pulse 2s infinite;
        }
        
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        
        .api-info {
            background: #1a1a1a;
            border: 1px solid #333;
            padding: 24px;
            border-radius: 8px;
            margin: 40px 0;
            font-family: monospace;
            font-size: 0.875rem;
        }
        
        .section {
            padding: 60px 0;
            border-bottom: 1px solid #333;
        }
        
        .section:last-child {
            border-bottom: none;
        }
        
        .section-title {
            font-size: 1.5rem;
            font-weight: 600;
            margin-bottom: 32px;
            color: #fff;
        }
        
        .subsection {
            margin-bottom: 40px;
            margin-top: 32px;
        }
        
        .subsection-title {
            font-size: 1.125rem;
            font-weight: 600;
            margin-bottom: 20px;
            color: #888;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        
        .endpoint-list {
            display: flex;
            flex-direction: column;
            gap: 12px;
        }
        
        .endpoint {
            display: flex;
            align-items: center;
            gap: 16px;
            padding: 16px;
            background: #1a1a1a;
            border: 1px solid #333;
            border-radius: 8px;
            transition: all 0.2s;
        }
        
        .endpoint:hover {
            background: #252525;
            border-color: #444;
        }
        
        .method {
            display: inline-block;
            min-width: 60px;
            padding: 4px 12px;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: 600;
            text-align: center;
        }
        
        .method.get {
            background: #0070f3;
            color: #fff;
        }
        
        .method.post {
            background: #22c55e;
            color: #fff;
        }
        
        .method.put {
            background: #f59e0b;
            color: #fff;
        }
        
        .method.delete {
            background: #ef4444;
            color: #fff;
        }
        
        .endpoint-path {
            font-family: monospace;
            font-size: 0.875rem;
            color: #fff;
            flex: 1;
        }
        
        .endpoint-desc {
            color: #888;
            font-size: 0.875rem;
            max-width: 400px;
        }
        
        .auth-required {
            background: rgba(239, 68, 68, 0.1);
            border: 1px solid rgba(239, 68, 68, 0.3);
            color: #ef4444;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 0.625rem;
            margin-left: auto;
        }
        
        .auth-not-required {
            background: rgba(34, 197, 94, 0.1);
            border: 1px solid rgba(34, 197, 94, 0.3);
            color: #22c55e;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 0.625rem;
            margin-left: auto;
        }
        
        .sidebar {
            position: sticky;
            top: 40px;
            max-height: calc(100vh - 80px);
            overflow-y: auto;
        }
        
        .nav-link {
            display: block;
            padding: 12px 16px;
            color: #888;
            text-decoration: none;
            font-size: 0.875rem;
            transition: color 0.2s;
            border-left: 2px solid transparent;
        }
        
        .nav-link:hover {
            color: #fff;
            border-left-color: #0070f3;
        }
        
        .nav-link.active {
            color: #fff;
            border-left-color: #0070f3;
            background: rgba(0, 112, 243, 0.1);
        }
        
        .main-content {
            display: grid;
            grid-template-columns: 250px 1fr;
            gap: 40px;
        }
        
        @media (max-width: 1024px) {
            .main-content {
                grid-template-columns: 1fr;
            }
            
            .sidebar {
                display: none;
            }
        }
        
        .footer {
            padding: 40px 0;
            border-top: 1px solid #333;
            text-align: center;
        }
        
        .footer-links {
            display: flex;
            justify-content: center;
            gap: 16px;
            margin-top: 0;
        }
        
        .footer-links a {
            display: inline-block;
            background: #1a1a1a;
            border: 1px solid #333;
            color: #fff;
            text-decoration: none;
            font-size: 0.875rem;
            padding: 12px 24px;
            border-radius: 8px;
            transition: all 0.2s;
        }
        
        .footer-links a:hover {
            background: #252525;
            border-color: #444;
            color: #fff;
        }
    </style>
    <script>
        async function checkSystemStatus() {
            try {
                const response = await fetch('/api/v1/health');
                const data = await response.json();
                
                const statusElement = document.querySelector('.status');
                const statusText = statusElement.querySelector('span:last-child');
                
                if (data.status === 'ok' && data.services.database === 'connected' && data.services.storage === 'connected') {
                    statusElement.classList.remove('error', 'warning');
                    statusText.textContent = 'All Systems Operational';
                } else {
                    statusElement.classList.add('error');
                    statusText.textContent = 'System Issues Detected';
                }
            } catch (error) {
                const statusElement = document.querySelector('.status');
                const statusText = statusElement.querySelector('span:last-child');
                statusElement.classList.add('error');
                statusText.textContent = 'System Unavailable';
            }
        }
        
        checkSystemStatus();
        setInterval(checkSystemStatus, 30000);
    </script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Vinyl Catalog API</h1>
            <p>A vinyl record cataloging system powered by Cloudflare Workers with zero hosting costs.</p>
            <div>
                <span class="status">
                    <div class="status-dot"></div>
                    <span>All Systems Operational</span>
                </span>
            </div>
        </div>
        
        <div class="api-info">
            <strong>Base URL:</strong> https://vinyl-catalog-api.nataliagranato.xyz
        </div>
        
        <div class="main-content">
            <div class="sidebar">
                <a href="#overview" class="nav-link">Overview</a>
                <a href="#system" class="nav-link">System</a>
                <a href="#authentication" class="nav-link">Authentication</a>
                <a href="#vinyls" class="nav-link">Vinyls</a>
                <a href="#tracks" class="nav-link">Tracks</a>
                <a href="#profile" class="nav-link">Profile</a>
                <a href="#uploads" class="nav-link">Uploads</a>
            </div>
            
            <div class="content">
                <div class="section" id="overview">
                    <h2 class="section-title">Overview</h2>
                    
                    <div class="subsection">
                        <p>The Vinyl Catalog API provides a RESTful interface for managing vinyl records, tracks, and user profiles. The API is built on Cloudflare Workers with full observability including structured logging and automatic tracing.</p>
                    </div>
                    
                    <div class="subsection">
                        <h3 class="subsection-title">Authentication</h3>
                        <p>Most endpoints require JWT authentication. Obtain a token by logging in with your credentials.</p>
                    </div>
                    
                    <div class="subsection">
                        <h3 class="subsection-title">Rate Limiting</h3>
                        <p>The API is currently running on Cloudflare Workers free tier with inherent rate limiting at the platform level.</p>
                    </div>
                    
                    <div class="subsection">
                        <h3 class="subsection-title">Response Format</h3>
                        <p>All responses are in JSON format with appropriate HTTP status codes.</p>
                    </div>
                </div>
                
                <div class="section" id="system">
                    <h2 class="section-title">System Endpoints</h2>
                    
                    <div class="subsection">
                        <h3 class="subsection-title">Health & Monitoring</h3>
                        <div class="endpoint-list">
                            <div class="endpoint">
                                <span class="method get">GET</span>
                                <span class="endpoint-path">/api/v1/health</span>
                                <span class="endpoint-desc">Health check endpoint with database and storage connectivity verification.</span>
                                <span class="auth-not-required">Public</span>
                            </div>
                            <div class="endpoint">
                                <span class="method get">GET</span>
                                <span class="endpoint-path">/api/v1/metrics</span>
                                <span class="endpoint-desc">API metrics including vinyl count, track count, and storage statistics.</span>
                                <span class="auth-not-required">Public</span>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="section" id="authentication">
                    <h2 class="section-title">Authentication</h2>
                    
                    <div class="subsection">
                        <h3 class="subsection-title">Login</h3>
                        <div class="endpoint-list">
                            <div class="endpoint">
                                <span class="method post">POST</span>
                                <span class="endpoint-path">/api/v1/auth/login</span>
                                <span class="endpoint-desc">Authenticate user and return JWT token for subsequent requests.</span>
                                <span class="auth-not-required">Public</span>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="section" id="vinyls">
                    <h2 class="section-title">Vinyls</h2>
                    
                    <div class="subsection">
                        <h3 class="subsection-title">Vinyl Operations</h3>
                        <div class="endpoint-list">
                            <div class="endpoint">
                                <span class="method get">GET</span>
                                <span class="endpoint-path">/api/v1/vinyls</span>
                                <span class="endpoint-desc">Retrieve a list of all vinyl records in the catalog.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method post">POST</span>
                                <span class="endpoint-path">/api/v1/vinyls</span>
                                <span class="endpoint-desc">Create a new vinyl record in the catalog.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method get">GET</span>
                                <span class="endpoint-path">/api/v1/vinyls/:id</span>
                                <span class="endpoint-desc">Retrieve a specific vinyl record by its unique identifier.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method put">PUT</span>
                                <span class="endpoint-path">/api/v1/vinyls/:id</span>
                                <span class="endpoint-desc">Update an existing vinyl record.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method delete">DELETE</span>
                                <span class="endpoint-path">/api/v1/vinyls/:id</span>
                                <span class="endpoint-desc">Delete a vinyl record from the catalog.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method post">POST</span>
                                <span class="endpoint-path">/api/v1/vinyls/:id/favorite</span>
                                <span class="endpoint-desc">Toggle the favorite status of a vinyl record.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="section" id="tracks">
                    <h2 class="section-title">Tracks</h2>
                    
                    <div class="subsection">
                        <h3 class="subsection-title">Track Operations</h3>
                        <div class="endpoint-list">
                            <div class="endpoint">
                                <span class="method get">GET</span>
                                <span class="endpoint-path">/api/v1/vinyls/:id/tracks</span>
                                <span class="endpoint-desc">Retrieve all tracks associated with a specific vinyl record.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method post">POST</span>
                                <span class="endpoint-path">/api/v1/vinyls/:id/tracks</span>
                                <span class="endpoint-desc">Add a new track to a vinyl record.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method put">PUT</span>
                                <span class="endpoint-path">/api/v1/vinyls/:id/tracks/:track_id</span>
                                <span class="endpoint-desc">Update an existing track within a vinyl record.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method delete">DELETE</span>
                                <span class="endpoint-path">/api/v1/vinyls/:id/tracks/:track_id</span>
                                <span class="endpoint-desc">Remove a track from a vinyl record.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="section" id="profile">
                    <h2 class="section-title">Profile</h2>
                    
                    <div class="subsection">
                        <h3 class="subsection-title">Profile Operations</h3>
                        <div class="endpoint-list">
                            <div class="endpoint">
                                <span class="method get">GET</span>
                                <span class="endpoint-path">/api/v1/profile</span>
                                <span class="endpoint-desc">Retrieve the public profile information.</span>
                                <span class="auth-not-required">Public</span>
                            </div>
                            <div class="endpoint">
                                <span class="method put">PUT</span>
                                <span class="endpoint-path">/api/v1/profile</span>
                                <span class="endpoint-desc">Update the user profile information.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method post">POST</span>
                                <span class="endpoint-path">/api/v1/profile/photo</span>
                                <span class="endpoint-desc">Upload a profile photo for the user account.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="section" id="uploads">
                    <h2 class="section-title">Uploads</h2>
                    
                    <div class="subsection">
                        <h3 class="subsection-title">File Operations</h3>
                        <div class="endpoint-list">
                            <div class="endpoint">
                                <span class="method post">POST</span>
                                <span class="endpoint-path">/api/v1/vinyls/:id/cover</span>
                                <span class="endpoint-desc">Upload a cover image for a vinyl record.</span>
                                <span class="auth-required">Auth Required</span>
                            </div>
                            <div class="endpoint">
                                <span class="method get">GET</span>
                                <span class="endpoint-path">/uploads/:filename</span>
                                <span class="endpoint-desc">Serve uploaded files from R2 storage.</span>
                                <span class="auth-not-required">Public</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="footer">
            <div class="footer-links">
                <a href="/api/v1/health">Health Check</a>
                <a href="/api/v1/metrics">Metrics</a>
                <a href="https://dash.cloudflare.com">Cloudflare Dashboard</a>
            </div>
        </div>
    </div>
</body>
</html>`;
  return new Response(html, {
    headers: { 'Content-Type': 'text/html; charset=utf-8' }
  });
});

// Health check endpoint
router.get('/api/v1/health', async (request: Request, env: Env) => {
  try {
    // Test database connection
    const dbResult = await env.DB.prepare('SELECT 1').first();
    
    // Test R2 connection
    const r2Buckets = await env.UPLOADS.list({ limit: 1 });
    
    return new Response(JSON.stringify({
      status: 'ok',
      timestamp: new Date().toISOString(),
      services: {
        database: dbResult ? 'connected' : 'disconnected',
        storage: r2Buckets ? 'connected' : 'disconnected',
        api: 'operational'
      },
      version: '1.0.0',
      environment: env.APP_ENV || 'production'
    }), {
    headers: { 'Content-Type': 'application/json' }
  });
  } catch (error) {
    return new Response(JSON.stringify({
      status: 'error',
      timestamp: new Date().toISOString(),
      error: error instanceof Error ? error.message : 'Unknown error'
    }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' }
    });
  }
});

// Metrics endpoint (basic Cloudflare Workers metrics)
router.get('/api/v1/metrics', async (request: Request, env: Env) => {
  try {
    // Get database stats
    const vinylCount = await env.DB.prepare('SELECT COUNT(*) as count FROM vinyls').first<{ count: number }>();
    const trackCount = await env.DB.prepare('SELECT COUNT(*) as count FROM tracks').first<{ count: number }>();
    const profileCount = await env.DB.prepare('SELECT COUNT(*) as count FROM profiles').first<{ count: number }>();
    
    // Get R2 stats
    const r2Objects = await env.UPLOADS.list();
    
    const metrics = {
      timestamp: new Date().toISOString(),
      database: {
        vinyls: vinylCount?.count || 0,
        tracks: trackCount?.count || 0,
        profiles: profileCount?.count || 0
      },
      storage: {
        total_objects: r2Objects.objects.length,
        truncated: r2Objects.truncated
      },
      api: {
        version: '1.0.0',
        environment: env.APP_ENV || 'production'
      }
    };
    
    return new Response(JSON.stringify(metrics, null, 2), {
      headers: { 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response(JSON.stringify({
      error: 'Failed to collect metrics',
      message: error instanceof Error ? error.message : 'Unknown error'
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' }
    });
  }
});

// Public routes
router.all('/api/v1/auth/*', authRoutes);

// Profile routes - mixed public/protected
router.get('/api/v1/profile', profileRoutes);
router.all('/api/v1/profile*', authMiddleware, profileRoutes);

// Protected routes
router.all('/api/v1/vinyls*', authMiddleware, vinylRoutes);
router.all('/api/v1/profile/photo', authMiddleware, profileRoutes);
router.all('/api/v1/uploads*', authMiddleware, uploadRoutes);

// Serve uploaded files from R2
router.get('/uploads/:filename', async (request: Request, env: Env) => {
  const url = new URL(request.url);
  const filename = url.pathname.split('/uploads/')[1];
  
  try {
    const object = await env.UPLOADS.get(filename);
    
    if (!object) {
      return new Response('File not found', { status: 404 });
    }
    
    const headers = new Headers();
    object.writeHttpMetadata(headers);
    headers.set('etag', object.httpEtag);
    
    return new Response(object.body, { headers });
  } catch (error) {
    return new Response('Error retrieving file', { status: 500 });
  }
});

// 404 handler
router.all('*', () => {
  return new Response('Not Found', { status: 404 });
});

// Cloudflare Workers fetch handler
export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    return router.handle(request, env);
  }
};