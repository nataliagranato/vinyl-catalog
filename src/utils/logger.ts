// Structured logging utility for Cloudflare Workers
// Uses JSON format for better indexing in Workers Logs

export enum LogLevel {
  DEBUG = 'debug',
  INFO = 'info',
  WARN = 'warn',
  ERROR = 'error'
}

interface LogContext {
  userId?: string;
  requestId?: string;
  endpoint?: string;
  method?: string;
  userAgent?: string;
  [key: string]: any;
}

interface LogEntry {
  level: LogLevel;
  message: string;
  timestamp: string;
  context?: LogContext;
  error?: {
    message: string;
    stack?: string;
    name?: string;
  };
}

export class Logger {
  private defaultContext: LogContext;

  constructor(defaultContext: LogContext = {}) {
    this.defaultContext = defaultContext;
  }

  private log(level: LogLevel, message: string, context: LogContext = {}, error?: Error) {
    const logEntry: LogEntry = {
      level,
      message,
      timestamp: new Date().toISOString(),
      context: { ...this.defaultContext, ...context }
    };

    if (error) {
      logEntry.error = {
        message: error.message,
        stack: error.stack,
        name: error.name
      };
    }

    console.log(JSON.stringify(logEntry));
  }

  debug(message: string, context: LogContext = {}) {
    this.log(LogLevel.DEBUG, message, context);
  }

  info(message: string, context: LogContext = {}) {
    this.log(LogLevel.INFO, message, context);
  }

  warn(message: string, context: LogContext = {}) {
    this.log(LogLevel.WARN, message, context);
  }

  error(message: string, error?: Error, context: LogContext = {}) {
    this.log(LogLevel.ERROR, message, context, error);
  }

  // Convenience methods for common operations
  authAttempt(username: string, success: boolean) {
    this.info('Authentication attempt', {
      username,
      success,
      timestamp: new Date().toISOString()
    });
  }

  apiRequest(method: string, endpoint: string, statusCode: number, duration: number) {
    this.info('API request', {
      method,
      endpoint,
      statusCode,
      duration: `${duration}ms`
    });
  }

  databaseQuery(query: string, duration: number, error?: Error) {
    if (error) {
      this.error('Database query failed', error, { query, duration: `${duration}ms` });
    } else {
      this.debug('Database query executed', { query, duration: `${duration}ms` });
    }
  }
}

// Create a default logger instance
export const logger = new Logger();