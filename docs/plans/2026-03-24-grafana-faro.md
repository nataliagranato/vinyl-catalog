# Grafana Faro Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Adicionar rastreabilidade de erros frontend com Grafana Faro — capturando exceções JavaScript, erros de rede, Web Vitals e console errors, integrando nativamente com o Loki, Jaeger e Grafana já existentes.

**Architecture:** Grafana Alloy roda como Faro receiver (porta 12347); o browser SDK (`@grafana/faro-web-sdk`) envia erros/logs/medidas para o Alloy, que os encaminha para Loki (logs/erros) e OTel Collector (traces). O TracingInstrumentation do Faro substitui o WebTracerProvider manual de `lib/telemetry.ts`, eliminando o conflito de dois providers OTel.

**Tech Stack:** Grafana Alloy (River config), `@grafana/faro-web-sdk`, `@grafana/faro-web-tracing`, Next.js client component para inicialização, Grafana dashboard JSON.

---

### Contexto crítico para o implementador

- O projeto usa OTel Web SDK 2.x já instalado (`@opentelemetry/sdk-trace-web ^2.6.0`)
- `frontend/lib/telemetry.ts` registra um `WebTracerProvider` manualmente via `initTelemetry()` — isso vai CONFLITAR com o TracingInstrumentation do Faro; precisamos remover esse registro mas manter o `getTracer()` exportado
- `app/layout.tsx` é Server Component (sem `"use client"`); inicialização do Faro precisa de um componente client separado
- Alloy no Docker se conecta a `loki:3100` e `otel-collector:4317` pela rede interna do Docker Compose
- Porta 12347 precisa ser exposta no host para o browser alcançar (browser faz fetch direto para `localhost:12347`)
- Faro com `TracingInstrumentation` registra um OTel provider globalmente → `trace.getTracer()` continua funcionando sem modificação nos arquivos de uso

---

### Task 1: Criar Grafana Alloy config + adicionar ao Docker Compose

**Files:**
- Create: `alloy/config.alloy`
- Modify: `docker-compose.yml`

**Step 1: Criar diretório e arquivo de config**

```bash
mkdir -p /Users/natalia.granato/Downloads/vinyl-catalog/alloy
```

Criar `alloy/config.alloy` com o conteúdo abaixo. O River language usa blocos com `component_name "label" { }`.

```alloy
// Faro receiver — recebe sinais do browser SDK
faro.receiver "vinyl_frontend" {
  server {
    listen_address       = "0.0.0.0"
    listen_port          = 12347
    cors_allowed_origins = ["http://localhost:3001", "http://localhost:3000", "http://*"]
  }

  output {
    logs   = [loki.write.default.receiver]
    traces = [otelcol.exporter.otlp.default.input]
  }
}

// Encaminhar logs/erros para Loki
loki.write "default" {
  endpoint {
    url = "http://loki:3100/loki/api/v1/push"
  }
}

// Encaminhar traces para o OTel Collector existente
otelcol.exporter.otlp "default" {
  client {
    endpoint = "otel-collector:4317"
    tls {
      insecure = true
    }
  }
}
```

**Step 2: Adicionar serviço `alloy` ao `docker-compose.yml`**

Adicionar logo após o serviço `promtail` e antes de `grafana`:

```yaml
  alloy:
    image: grafana/alloy:latest
    ports:
      - "12347:12347"
    volumes:
      - ./alloy/config.alloy:/etc/alloy/config.alloy:ro
    command: run /etc/alloy/config.alloy
    depends_on:
      - loki
      - otel-collector
    restart: unless-stopped
```

Também adicionar `NEXT_PUBLIC_FARO_URL` ao serviço `frontend`:

```yaml
  frontend:
    environment:
      - API_URL=http://app:8080
      - NEXT_PUBLIC_API_URL=http://localhost:8080
      - NEXT_PUBLIC_FARO_URL=http://localhost:12347/collect
```

**Step 3: Verificar que o Alloy sobe**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
docker compose up -d alloy
docker compose logs alloy --tail=20
```

Esperado: logs mostrando `"Starting Faro receiver"` e `"listening on 0.0.0.0:12347"` sem erros.

**Step 4: Smoke test CORS**

```bash
curl -sv -X OPTIONS http://localhost:12347/collect \
  -H "Origin: http://localhost:3001" \
  -H "Access-Control-Request-Method: POST" 2>&1 | grep -iE "< HTTP|access-control"
```

Esperado: `HTTP/1.1 200` com `access-control-allow-origin`.

**Step 5: Commit**

```bash
git add alloy/config.alloy docker-compose.yml
git commit -m "feat(faro): add Grafana Alloy service as Faro receiver"
```

---

### Task 2: Instalar pacotes Faro no frontend

**Files:**
- Modify: `frontend/package.json` (via npm install)

**Step 1: Instalar pacotes**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npm install @grafana/faro-web-sdk @grafana/faro-web-tracing
```

**Step 2: Verificar que o TypeScript compila**

```bash
npx tsc --noEmit
```

Esperado: sem erros.

**Step 3: Commit**

```bash
git add package.json package-lock.json
git commit -m "feat(faro): install @grafana/faro-web-sdk and faro-web-tracing"
```

---

### Task 3: Atualizar lib/telemetry.ts — remover provider manual, manter getTracer()

**Files:**
- Modify: `frontend/lib/telemetry.ts`

**Contexto:** O arquivo atual registra um `WebTracerProvider` manualmente. O `TracingInstrumentation` do Faro (Task 4) vai registrar seu próprio provider. Ter dois providers causa conflito — o segundo sobrescreve o primeiro silenciosamente, mas o padrão de init com `__otelInit` bloquearia o Faro de registrar. Solução: remover o bloco de init daqui; manter apenas `getTracer()`.

**Step 1: Reescrever lib/telemetry.ts**

O arquivo novo é apenas o export de `getTracer()`. O Faro registra o provider; `trace.getTracer()` usa o provider global seja ele qual for.

```typescript
import { trace, type Tracer } from "@opentelemetry/api";

const SERVICE_NAME = "vinyl-catalog-frontend";

export function getTracer(): Tracer {
  return trace.getTracer(SERVICE_NAME);
}
```

Apagar todas as importações de `WebTracerProvider`, `SimpleSpanProcessor`, `OTLPTraceExporter`, `resourceFromAttributes`, `SEMRESATTRS_*`, e o bloco `initTelemetry` + singleton `__otelInit`.

**Step 2: Verificar que o TypeScript compila**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npx tsc --noEmit
```

Esperado: sem erros. Os arquivos que importam `getTracer()` (translate.ts, page.tsx, VinylListClient.tsx) continuam funcionando inalterados.

**Step 3: Commit**

```bash
git add lib/telemetry.ts
git commit -m "refactor(telemetry): remove manual OTel provider init, Faro handles it"
```

---

### Task 4: Criar lib/faro.ts com initializeFaro()

**Files:**
- Create: `frontend/lib/faro.ts`

**Step 1: Criar o arquivo**

```typescript
import {
  initializeFaro,
  getWebInstrumentations,
  type Faro,
} from "@grafana/faro-web-sdk";
import { TracingInstrumentation } from "@grafana/faro-web-tracing";

const SERVICE_NAME = "vinyl-catalog-frontend";

let faro: Faro | null = null;

export function initFaro(): void {
  if (typeof window === "undefined") return; // SSR guard
  if (faro) return;                           // já inicializado

  const url =
    process.env.NEXT_PUBLIC_FARO_URL ?? "http://localhost:12347/collect";

  faro = initializeFaro({
    url,
    app: {
      name:        SERVICE_NAME,
      version:     "1.0.0",
      environment: process.env.NODE_ENV ?? "production",
    },
    instrumentations: [
      ...getWebInstrumentations({
        captureConsole:           true,  // captura console.error / console.warn
        captureConsoleDisabledLevels: [], // captura todos os níveis
      }),
      new TracingInstrumentation(), // registra OTel provider + auto-instrumenta fetch
    ],
  });
}

/** Expõe o objeto faro para push manual de erros fora de React (ex: catch blocks) */
export function getFaro(): Faro | null {
  return faro;
}
```

**Step 2: Verificar que o TypeScript compila**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npx tsc --noEmit
```

Esperado: sem erros.

**Step 3: Commit**

```bash
git add lib/faro.ts
git commit -m "feat(faro): add initFaro() with error capture and OTel tracing"
```

---

### Task 5: Criar FaroInit client component + adicionar ao layout

**Files:**
- Create: `frontend/components/FaroInit.tsx`
- Modify: `frontend/app/layout.tsx`

**Contexto:** `app/layout.tsx` é Server Component. Inicialização do Faro precisa rodar no browser (client). Criar um componente mínimo `"use client"` que chama `initFaro()` no mount.

**Step 1: Criar FaroInit.tsx**

```typescript
"use client";

import { useEffect } from "react";
import { initFaro } from "@/lib/faro";

export default function FaroInit() {
  useEffect(() => {
    initFaro();
  }, []);

  return null; // não renderiza nada
}
```

**Step 2: Adicionar FaroInit ao layout.tsx**

```typescript
import type { Metadata } from "next";
import "./globals.css";
import { ToastProvider } from "@/components/ToastProvider";
import FaroInit from "@/components/FaroInit";

export const metadata: Metadata = {
  title: "Vinyl Catalog",
  description: "Your personal vinyl record collection",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <FaroInit />
        <ToastProvider>{children}</ToastProvider>
      </body>
    </html>
  );
}
```

**Step 3: Verificar que o TypeScript compila**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog/frontend
npx tsc --noEmit
```

**Step 4: Commit**

```bash
git add components/FaroInit.tsx app/layout.tsx
git commit -m "feat(faro): initialize Faro SDK on app mount via FaroInit component"
```

---

### Task 6: Criar dashboard Grafana para Faro

**Files:**
- Create: `grafana/dashboards/faro.json`

**Contexto:** Dados do Faro chegam no Loki com o label `service_name="vinyl-catalog-frontend"` (ou `app="vinyl-catalog-frontend"` dependendo da versão do Alloy). Cada entrada de log tem um campo JSON `kind` com valor `"exception"`, `"log"`, `"measurement"` ou `"event"`.

**Step 1: Criar o arquivo de dashboard**

```json
{
  "__inputs": [],
  "__requires": [],
  "annotations": { "list": [] },
  "description": "Grafana Faro — erros e performance do frontend",
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 1,
  "id": null,
  "links": [],
  "panels": [
    {
      "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
      "fieldConfig": {
        "defaults": {
          "color": { "mode": "thresholds" },
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "red", "value": 1 }
            ]
          },
          "unit": "short"
        },
        "overrides": []
      },
      "gridPos": { "h": 4, "w": 6, "x": 0, "y": 0 },
      "id": 1,
      "options": {
        "colorMode": "background",
        "graphMode": "none",
        "justifyMode": "center",
        "orientation": "auto",
        "reduceOptions": { "calcs": ["sum"], "fields": "", "values": false },
        "textMode": "auto"
      },
      "targets": [
        {
          "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
          "expr": "sum(count_over_time({service_name=\"vinyl-catalog-frontend\"} | json | kind=\"exception\" [24h]))",
          "refId": "A"
        }
      ],
      "title": "Exceções JS (24h)",
      "type": "stat"
    },
    {
      "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
      "fieldConfig": {
        "defaults": {
          "color": { "mode": "thresholds" },
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 5 },
              { "color": "red", "value": 20 }
            ]
          },
          "unit": "short"
        },
        "overrides": []
      },
      "gridPos": { "h": 4, "w": 6, "x": 6, "y": 0 },
      "id": 2,
      "options": {
        "colorMode": "background",
        "graphMode": "none",
        "justifyMode": "center",
        "orientation": "auto",
        "reduceOptions": { "calcs": ["sum"], "fields": "", "values": false },
        "textMode": "auto"
      },
      "targets": [
        {
          "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
          "expr": "sum(count_over_time({service_name=\"vinyl-catalog-frontend\"} | json | kind=\"log\" | level=\"error\" [24h]))",
          "refId": "A"
        }
      ],
      "title": "Console Errors (24h)",
      "type": "stat"
    },
    {
      "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
      "fieldConfig": {
        "defaults": {
          "color": { "mode": "thresholds" },
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 1000 },
              { "color": "red", "value": 2500 }
            ]
          },
          "unit": "ms"
        },
        "overrides": []
      },
      "gridPos": { "h": 4, "w": 6, "x": 12, "y": 0 },
      "id": 3,
      "options": {
        "colorMode": "background",
        "graphMode": "none",
        "justifyMode": "center",
        "orientation": "auto",
        "reduceOptions": { "calcs": ["lastNotNull"], "fields": "", "values": false },
        "textMode": "auto"
      },
      "targets": [
        {
          "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
          "expr": "last_over_time({service_name=\"vinyl-catalog-frontend\"} | json | kind=\"measurement\" | type=\"lcp\" | unwrap value [24h])",
          "refId": "A"
        }
      ],
      "title": "LCP — Largest Contentful Paint",
      "type": "stat"
    },
    {
      "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
      "fieldConfig": {
        "defaults": {
          "color": { "mode": "thresholds" },
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null },
              { "color": "yellow", "value": 0.1 },
              { "color": "red", "value": 0.25 }
            ]
          },
          "unit": "short",
          "decimals": 3
        },
        "overrides": []
      },
      "gridPos": { "h": 4, "w": 6, "x": 18, "y": 0 },
      "id": 4,
      "options": {
        "colorMode": "background",
        "graphMode": "none",
        "justifyMode": "center",
        "orientation": "auto",
        "reduceOptions": { "calcs": ["lastNotNull"], "fields": "", "values": false },
        "textMode": "auto"
      },
      "targets": [
        {
          "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
          "expr": "last_over_time({service_name=\"vinyl-catalog-frontend\"} | json | kind=\"measurement\" | type=\"cls\" | unwrap value [24h])",
          "refId": "A"
        }
      ],
      "title": "CLS — Cumulative Layout Shift",
      "type": "stat"
    },
    {
      "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "short" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 24, "x": 0, "y": 4 },
      "id": 5,
      "options": {
        "dedupStrategy": "none",
        "enableLogDetails": true,
        "prettifyLogMessage": true,
        "showTime": true,
        "showLabels": false,
        "sortOrder": "Descending",
        "wrapLogMessage": true
      },
      "targets": [
        {
          "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
          "expr": "{service_name=\"vinyl-catalog-frontend\"} | json | kind=\"exception\"",
          "refId": "A"
        }
      ],
      "title": "Exceções JavaScript (live)",
      "type": "logs"
    },
    {
      "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "short" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 24, "x": 0, "y": 12 },
      "id": 6,
      "options": {
        "dedupStrategy": "none",
        "enableLogDetails": true,
        "prettifyLogMessage": true,
        "showTime": true,
        "showLabels": false,
        "sortOrder": "Descending",
        "wrapLogMessage": true
      },
      "targets": [
        {
          "datasource": { "type": "loki", "uid": "${DS_LOKI}" },
          "expr": "{service_name=\"vinyl-catalog-frontend\"} | json | kind=~\"log|exception\"",
          "refId": "A"
        }
      ],
      "title": "Todos os Logs e Erros do Frontend",
      "type": "logs"
    }
  ],
  "refresh": "30s",
  "schemaVersion": 38,
  "tags": ["frontend", "faro", "errors"],
  "templating": {
    "list": [
      {
        "current": {},
        "hide": 0,
        "includeAll": false,
        "label": "Loki",
        "multi": false,
        "name": "DS_LOKI",
        "options": [],
        "query": "loki",
        "refresh": 1,
        "type": "datasource"
      }
    ]
  },
  "time": { "from": "now-3h", "to": "now" },
  "timepicker": {},
  "timezone": "browser",
  "title": "Frontend — Faro Errors",
  "uid": "faro-errors",
  "version": 1
}
```

**Step 2: Commit**

```bash
git add grafana/dashboards/faro.json
git commit -m "feat(faro): add Grafana Faro error tracking dashboard"
```

---

### Task 7: Rebuild frontend + smoke test end-to-end

**Step 1: Rebuildar e subir tudo**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
docker compose up -d --build frontend
```

Esperado: container frontend reconstrói e sobe sem erros.

**Step 2: Verificar que Alloy está recebendo**

```bash
docker compose logs alloy --tail=20
```

Esperado: sem erros de inicialização. O Alloy fica escutando na porta 12347.

**Step 3: Enviar um evento de teste manualmente (simula o browser)**

```bash
curl -s -X POST http://localhost:12347/collect \
  -H "Content-Type: application/json" \
  -H "Origin: http://localhost:3001" \
  -d '{
    "traces":{"resourceSpans":[]},
    "logs":[{
      "timestamp":"'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'",
      "level":"error",
      "message":"test error from smoke test",
      "context":{},
      "trace":{"trace_id":"","span_id":""}
    }],
    "exceptions":[],
    "measurements":[],
    "events":[],
    "meta":{
      "sdk":{"name":"@grafana/faro-web-sdk","version":"1.0.0"},
      "app":{"name":"vinyl-catalog-frontend","version":"1.0.0","environment":"test"},
      "session":{"id":"smoke-test"},
      "user":{}
    }
  }' -w "\nHTTP %{http_code}"
```

Esperado: `HTTP 202`.

**Step 4: Verificar que o log chegou no Loki**

```bash
sleep 5
curl -s -G "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={service_name="vinyl-catalog-frontend"}' \
  --data-urlencode 'limit=5' \
  | python3 -c "
import sys, json
d = json.load(sys.stdin)
results = d.get('data',{}).get('result',[])
if results:
    for r in results:
        for v in r.get('values',[]):
            print(v[1][:120])
else:
    print('nenhum resultado ainda')
"
```

Esperado: linha de log com `"test error from smoke test"`.

**Step 5: Abrir o browser e verificar Faro inicializa**

Abrir http://localhost:3001 no browser. Abrir DevTools → Network → filtrar por `12347` ou `collect`. Deve aparecer um POST para `http://localhost:12347/collect` com status 202.

**Step 6: Verificar dashboard no Grafana**

Abrir http://localhost:3000 → Dashboards → "Frontend — Faro Errors". Verificar que o dashboard carregou sem erros de datasource.

**Step 7: Commit final**

```bash
cd /Users/natalia.granato/Downloads/vinyl-catalog
git add -A
git commit -m "feat(faro): complete Grafana Faro integration — error tracking live"
```

---

### Referências

- [Grafana Alloy Faro receiver docs](https://grafana.com/docs/alloy/latest/reference/components/faro.receiver/)
- [Faro Web SDK README](https://github.com/grafana/faro-web-sdk)
- [TracingInstrumentation](https://github.com/grafana/faro-web-sdk/tree/main/packages/web-tracing)
