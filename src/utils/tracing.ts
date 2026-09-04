// Custom span utilities for Cloudflare Workers tracing
// Requires tracing to be enabled in wrangler.toml

// Cloudflare Workers v4 supports automatic tracing
// Custom spans can be created using the runtime trace API

export function createSpan(name: string, fn: () => Promise<any>): Promise<any> {
  // Note: Custom spans require tracing to be enabled
  // In Workers v4, automatic tracing is enabled by default
  // For custom spans, you can use the runtime trace API
  
  // For now, this is a placeholder for future custom span implementation
  // The automatic tracing will capture:
  // - HTTP requests
  // - Database queries (D1)
  // - R2 operations
  // - Subrequests
  
  return fn();
}

// Example of how to use custom spans in the future:
/*
export async function withCustomSpan<T>(
  name: string,
  fn: () => Promise<T>
): Promise<T> {
  const span = // get current span from context
  const childSpan = span.startChild(name);
  
  try {
    const result = await fn();
    childSpan.finish();
    return result;
  } catch (error) {
    childSpan.recordException(error);
    childSpan.finish();
    throw error;
  }
}
*/