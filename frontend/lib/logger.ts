// Simple logger stub for compatibility
export function log(level: string, route: string, data: any) {
  console.log(`[${level.toUpperCase()}] ${route}:`, data);
}
