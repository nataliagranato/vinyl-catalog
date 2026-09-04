# OpenTelemetry Export Setup Guide
## Vinyl Catalog API - Cloudflare Workers

**Status:** ✅ Wrangler v4.129.0 atualizado  
**Status:** ✅ Traces habilitados (10% sampling)  
**Status:** ⏳ OpenTelemetry export configurado (aguarda destino)

---

## 🎯 **O que foi implementado**

### ✅ **Wrangler Atualizado**
- De v3.114.17 → v4.129.0
- @cloudflare/workers-types atualizado para v5.20260903.1
- Suporte nativo para Traces e OpenTelemetry

### ✅ **Traces Habilitados**
```toml
[observability.traces]
enabled = true
head_sampling_rate = 0.1  # 10% das requisições
```

### ✅ **Logs Habilitados**
```toml
[observability.logs]
enabled = true
head_sampling_rate = 1.0  # 100% das requisições
```

---

## 📊 **Opções de OpenTelemetry Export**

Para habilitar o OpenTelemetry export, você precisa:

1. **Criar um destino** no dashboard da Cloudflare
2. **Configurar o wrangler.toml** com o nome do destino
3. **Re-deployar** o Worker

### **Plataformas Suportadas**

#### **1. Grafana Cloud (Recomendado)**
- Custo: Plano gratuito disponível
- Features: Traces, Logs, Metrics
- Endpoint: `https://otlp-gateway-prod-us-east-2.grafana.net/otlp`

#### **2. Sentry**
- Custo: Plano gratuito disponível
- Features: Error tracking, Performance monitoring
- Endpoint: `https://{HOST}/api/{PROJECT_ID}/integration/otlp/v1/traces`

#### **3. Honeycomb**
- Custo: Plano gratuito disponível
- Features: High-cardinality observability
- Endpoint: `https://api.honeycomb.io/v1/traces`

#### **4. Datadog**
- Custo: Plano pago
- Features: Full observability platform
- Endpoint: `https://ingest.{REGION}.datadoghq.com/api/v2/otlp`

#### **5. Outros (OTLP-compatible)**
- Qualquer plataforma com endpoint OTLP
- Splunk, New Relic, etc.

---

## 🛠️ **Passo a Passo para Configurar**

### **Passo 1: Acessar Dashboard Cloudflare**

1. Acesse: https://dash.cloudflare.com
2. Vá em: **Workers & Pages** → **Observability**
3. Clique em: **Add destination**

### **Passo 2: Configurar Destino**

Para cada tipo (Traces e Logs), você precisa criar um destino:

**Exemplo para Grafana Cloud:**

**Traces Destination:**
- **Destination Name:** `grafana-traces`
- **Destination Type:** Traces
- **OTLP Endpoint:** `https://otlp-gateway-prod-us-east-2.grafana.net/otlp/v1/traces`
- **Custom Headers:** 
  - Header name: `Authorization`
  - Header value: `Basic MTMxxx...` (seu token Grafana)

**Logs Destination:**
- **Destination Name:** `grafana-logs`
- **Destination Type:** Logs
- **OTLP Endpoint:** `https://otlp-gateway-prod-us-east-2.grafana.net/otlp/v1/logs`
- **Custom Headers:** 
  - Header name: `Authorization`
  - Header value: `Basic MTMxxx...` (seu token Grafana)

### **Passo 3: Atualizar wrangler.toml**

```toml
[observability.traces]
enabled = true
head_sampling_rate = 0.1
destinations = ["grafana-traces"]  # Nome do destino no dashboard
persist = true  # Armazena também no dashboard Cloudflare

[observability.logs]
enabled = true
head_sampling_rate = 1.0
destinations = ["grafana-logs"]  # Nome do destino no dashboard
persist = true  # Armazena também no dashboard Cloudflare
```

### **Passo 4: Re-deployar**

```bash
npm run deploy
```

---

## 📝 **Instruções Específicas por Plataforma**

### **Grafana Cloud**

1. **Criar conta Grafana Cloud** (se não tiver)
2. **Configurar OTLP:**
   - Acesse Connections → Add new connection
   - Search "OpenTelemetry" → Select OpenTelemetry (OTLP)
   - Create token e salve as credenciais
3. **Criar destinos Cloudflare** (conforme Passo 2 acima)
4. **Atualizar wrangler.toml**

### **Sentry**

1. **Criar projeto Sentry**
2. **Configurar OTLP:**
   - Settings → Projects → Settings → Client Keys (DSN)
   - OTLP endpoints:
     - Traces: `https://{HOST}/api/{PROJECT_ID}/integration/otlp/v1/traces`
     - Logs: `https://{HOST}/api/{PROJECT_ID}/integration/otlp/v1/logs`
3. **Criar destinos Cloudflare** com header `x-sentry-auth`
4. **Atualizar wrangler.toml**

### **Honeycomb**

1. **Criar conta Honeycomb**
2. **Configurar OTLP:**
   - Encontrar seu API Key em Account Settings
   - OTLP endpoints:
     - Traces: `https://api.honeycomb.io/v1/traces`
     - Logs: `https://api.honeycomb.io/v1/logs`
3. **Criar destinos Cloudflare** com header `x-honeycomb-team`
4. **Atualizar wrangler.toml**

---

## 🔧 **Configuração Atual (wrangler.toml)**

```toml
[observability]
enabled = true

[observability.logs]
enabled = true
head_sampling_rate = 1.0
# OpenTelemetry export for logs (configure destination in Cloudflare dashboard first)
# destinations = ["logs-destination-name"]
# persist = true

[observability.traces]
enabled = true
head_sampling_rate = 0.1
# OpenTelemetry export for traces (configure destination in Cloudflare dashboard first)
# destinations = ["tracing-destination-name"]
# persist = true
```

---

## ✅ **Status Atual**

- ✅ **Wrangler v4** atualizado
- ✅ **Traces habilitados** (dashboard Cloudflare)
- ✅ **Logs habilitados** (dashboard Cloudflare)
- ⏳ **OpenTelemetry export** aguarda configuração de destino

---

## 🚀 **Próximos Passos**

**Você precisa escolher:**

1. **Qual plataforma de observabilidade?**
   - Grafana Cloud (recomendado, plano gratuito)
   - Sentry (foco em errors)
   - Honeycomb (high-cardinality)
   - Outra plataforma OTLP-compatible

2. **Prefere configurar agora ou depois?**
   - **Agora:** Eu guio você pelo setup específico
   - **Depois:** O logging/dashboard Cloudflare já é funcional

---

**Nota:** Mesmo sem OpenTelemetry export, você já tem:
- ✅ Logging estruturado no dashboard Cloudflare
- ✅ Traces automáticos no dashboard Cloudflare
- ✅ Real-time logs via `wrangler tail`

OpenTelemetry export é para integração com plataformas externas, que é opcional para observabilidade básica.