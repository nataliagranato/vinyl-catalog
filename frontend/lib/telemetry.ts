import { trace, SpanStatusCode } from "@opentelemetry/api";

export function getTracer() {
  return trace.getTracer("vinyl-catalog-frontend");
}

export { SpanStatusCode };
