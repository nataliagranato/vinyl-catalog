#!/usr/bin/env bash
# Gera documentação (ADR, Runbook, README revisado) com IA (Verboo/glm-5.3-flash)
# analisando a branch alvo e abre um PR para ela.
# Uso: ./scripts/ai-docs.sh [branch-alvo]   (padrão: main)
set -euo pipefail

MODEL="glm-5.3-flash"
: "${VERBOO_API_KEY:?Defina VERBOO_API_KEY}"
: "${GH_TOKEN:?Defina GH_TOKEN}"

TARGET_BRANCH="${1:-main}"
HEAD_BRANCH="ai/docs"

git config user.name "ai-docs-bot"
git config user.email "actions@github.com"

SYSTEM_PROMPT='Você é um engenheiro de plataforma sênior. Analise o contexto do repositório fornecido e gere documentação técnica em português brasileiro (pt-BR). Responda APENAS com um JSON válido, sem blocos de código markdown, no formato:
{"files":[{"path":"caminho/arquivo.md","content":"conteúdo markdown completo"}]}
Inclua obrigatoriamente:
1. docs/adr/ADR-001-decisoes-arquiteturais.md — Architecture Decision Records com as decisões identificadas no código (formato: Contexto, Decisão, Consequências).
2. docs/runbooks/runbook.md — Runbook operacional: como subir, comandos úteis, pontos de atenção, troubleshooting básico.
3. docs/ai-analysis.md — Análise da IA sobre o repositório: stack, estrutura, riscos e sugestões de melhoria.
4. README.md — o README existente, revisado e aprimorado pela IA (mantenha as informações corretas, corrija o que estiver desatualizado, melhore organização e clareza). Se não houver README no contexto, crie um.
5. docs/swagger/swagger.yaml — a especificação OpenAPI atualizada: com base nos handlers/rotas encontrados no contexto, garanta que paths, schemas e descrições reflitam a API real. Mantenha o formato YAML do arquivo existente.
Regras: seja técnico e conciso; baseie-se APENAS no contexto fornecido; não invente módulos ou endpoints que não aparecem no contexto.'

echo "=== Branch alvo: $TARGET_BRANCH ==="
git fetch origin "$TARGET_BRANCH" --quiet
if gh pr list --head "$HEAD_BRANCH" --json state --jq '.[].state' | grep -q OPEN; then
  echo "PR já aberto para $HEAD_BRANCH, encerrando."
  exit 0
fi
git checkout -B "$HEAD_BRANCH" "origin/$TARGET_BRANCH" --quiet

CONTEXT=$( {
  echo "Branch: $TARGET_BRANCH"
  echo "=== Arquivos do repositório ==="
  git ls-files | head -120
  echo "=== README.md (até 150 linhas) ==="
  head -150 README.md 2>/dev/null || echo "(sem README)"
  echo "=== go.mod ==="
  head -40 go.mod 2>/dev/null || true
  echo "=== docker-compose.yml ==="
  head -60 docker-compose.yml 2>/dev/null || true
  echo "=== docs/swagger/swagger.yaml (especificação OpenAPI atual) ==="
  head -300 docs/swagger/swagger.yaml 2>/dev/null || echo "(sem swagger)"
  echo "=== Rotas registradas (busca em *.go) ==="
  grep -rn -E '(GET|POST|PUT|DELETE|PATCH|Group|Handle)\(' --include='*.go' . | head -40 || true
  echo "=== Principais arquivos Go (imports e declarações) ==="
  for f in $(git ls-files '*.go' | head -15); do
    echo "--- $f ---"; head -30 "$f"
  done
} )

PAYLOAD=$(jq -n \
  --arg model "$MODEL" \
  --arg sys "$SYSTEM_PROMPT" \
  --arg ctx "$CONTEXT" \
  '{model:$model, messages:[{role:"system",content:$sys},{role:"user",content:$ctx}], max_tokens:32000}')

CONTENT=""
for ATTEMPT in 1 2 3 4; do
  RESP=$(curl -s --max-time 600 https://code.verboo.ai/router/v1/chat/completions \
    -H "Authorization: Bearer $VERBOO_API_KEY" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD")
  CONTENT=$(echo "$RESP" | jq -r '.choices[0].message.content // empty')
  [ -z "$CONTENT" ] && { echo "Tentativa $ATTEMPT: sem conteúdo. Aguardando $((ATTEMPT*30))s..." >&2; sleep $((ATTEMPT*30)); continue; }
  # Remove cercas de código caso o modelo insista em markdown
  CONTENT=${CONTENT#\`\`\`json}
  CONTENT=${CONTENT#\`\`\`}
  CONTENT=${CONTENT%\`\`\`}
  # Valida o JSON completo (detecta saída truncada por max_tokens)
  if echo "$CONTENT" | jq -e '.files | type == "array"' > /dev/null 2>&1; then
    break
  fi
  echo "Tentativa $ATTEMPT: JSON inválido/truncado. Aguardando $((ATTEMPT*30))s..." >&2
  CONTENT=""
  sleep $((ATTEMPT*30))
done
if [ -z "$CONTENT" ]; then
  echo "ERRO: IA não retornou JSON válido após 4 tentativas" >&2
  exit 1
fi

echo "$CONTENT" | jq -r '.files[] | @base64' | while read -r ENC; do
  FILE=$(echo "$ENC" | base64 -d)
  PATH_=$(echo "$FILE" | jq -r '.path')
  mkdir -p "$(dirname "$PATH_")"
  echo "$FILE" | jq -j '.content' > "$PATH_"
  echo "  gerado: $PATH_"
done

git add -A
if git diff --cached --quiet; then
  echo "Nada a commitar."
  exit 0
fi
git commit -m "docs: documentação gerada por IA (ADR, runbook, README)" --quiet
# --force: branch efêmera gerada por esta action; pode existir de execuções anteriores
git push --force origin "$HEAD_BRANCH" --quiet
gh pr create --base "$TARGET_BRANCH" --head "$HEAD_BRANCH" \
  --title "docs(ai): documentação gerada por IA" \
  --body "Documentação gerada automaticamente pela action de IA (Verboo / $MODEL) analisando a branch \`$TARGET_BRANCH\`.

Inclui:
- ADR (\`docs/adr/\`)
- Runbook (\`docs/runbooks/\`)
- Análise da IA (\`docs/ai-analysis.md\`)
- README revisado

> Gerado na aula de Platform Engineering. Revise antes de aprovar." \
  > /dev/null && echo "PR aberto: $TARGET_BRANCH <- $HEAD_BRANCH"
echo "Concluído."
