# Aula: GitHub Action de Documentação com IA

Workflow que varre as branches do repositório, usa IA (Verboo / `glm-5.3-flash`) para gerar documentação técnica e abre um Pull Request em cada branch analisada.

## O que a action faz

1. Disparada manualmente (`workflow_dispatch`), com opção de escolher as branches (padrão: todas, exceto as geradas por ela mesma).
2. Para cada branch, cria uma branch temporária `ai/docs-<branch>`.
3. Monta um contexto do repositório (lista de arquivos, README, `go.mod`, `docker-compose.yml`, principais arquivos Go) e envia para a API da Verboo.
4. A IA retorna JSON com 4 arquivos:
   - `docs/adr/ADR-001-decisoes-arquiteturais.md` — ADRs (Contexto, Decisão, Consequências)
   - `docs/runbooks/runbook.md` — runbook operacional
   - `docs/ai-analysis.md` — análise da stack, riscos e sugestões
   - `README.md` — README revisado
5. Commita, dá push e abre um PR `ai/docs-<branch>` → `<branch>`.

## Estrutura dos arquivos

```
.github/workflows/ai-docs.yml   # workflow (trigger, permissões, secrets)
scripts/ai-docs.sh              # lógica: loop de branches, chamada à API, PR
```

## Passo a passo

### 1. Criar a chave da API

Gere a API key no painel da Verboo (`https://code.verboo.ai`). Ela nunca vai para o código — só para os Secrets do GitHub.

### 2. Cadastrar o secret no repositório

Via CLI (ou web: Settings → Secrets and variables → Actions → New repository secret):

```bash
gh secret set VERBOO_API_KEY -R SEU_USUARIO/SEU_REPO --body "vbk_pro_..."
```

### 3. Criar o workflow

`.github/workflows/ai-docs.yml`:

```yaml
name: Documentação com IA

on:
  workflow_dispatch:
    inputs:
      branches:
        description: 'Branches a documentar (separadas por vírgula) ou "all"'
        required: false
        default: 'all'

permissions:
  contents: write      # push das branches ai/docs-*
  pull-requests: write # abrir os PRs
```

Pontos-chave:

- **`workflow_dispatch`**: execução manual — evita custo de API em cada push.
- **`permissions` explícitos**: sem isso o `GITHUB_TOKEN` não consegue push nem abrir PR.
- **`GITHUB_TOKEN`**: token temporário gerado pelo próprio GitHub na execução. Não precisa de PAT.

### 4. Criar o script `scripts/ai-docs.sh`

Responsabilidades do script (veja o arquivo completo no repo):

| Bloco | O que faz |
|---|---|
| Listagem de branches | `gh api repos/$REPO/branches` filtra as `ai/docs-*` |
| Contexto | `git ls-files` + trechos do README, go.mod, compose, fontes Go |
| Payload | `jq -n --arg` monta o JSON com system prompt + contexto |
| Chamada | `curl` → `POST /router/v1/chat/completions` (API compatível com OpenAI) |
| Parse | Extrai `.choices[0].message.content`, remove cercas ` ``` `, valida com `jq` |
| Git | `git checkout -B ai/docs-<branch>`, commit, push |
| PR | `gh pr create --base <branch> --head ai/docs-<branch>` |

O prompt exige resposta **apenas em JSON**: `{"files":[{"path":"...","content":"..."}]}` — isso torna a saída determinística e parseável.

### 5. Disparar

```bash
gh workflow run "Documentação com IA" -R SEU_USUARIO/SEU_REPO -f branches=all
# ou branches específicas:
gh workflow run "Documentação com IA" -R SEU_USUARIO/SEU_REPO -f branches="main,develop"
```

### 6. Verificar

- Aba **Actions**: acompanhe a execução.
- Aba **Pull requests**: um PR por branch analisada, com os 4 arquivos gerados.

## Pontos de discussão em aula

1. **Segurança de tokens**: a API key fica em Secret, nunca em código. O `GITHUB_TOKEN` do Actions tem escopo limitado ao repositório e permissões explícitas.
2. **Saída estruturada**: exigir JSON do LLM + validar com `jq` é o padrão para integrar IA em pipelines. Texto livre quebra o pipeline.
3. **Idempotência**: se já existe PR aberto para `ai/docs-<branch>`, o script pula. Rodar de novo não duplica PRs.
4. **Limites**: `max_tokens: 8000`; contexto truncado (`head`) para caber no prompt — em repos grandes, vale mandar só os arquivos-chave.
5. **IA que revisa IA**: os PRs são sugestões — revisão humana continua obrigatória antes do merge.

## Segurança da conta (após a aula)

Revogue qualquer PAT que tenha sido compartilhado em chat/screenshots: GitHub → Settings → Developer settings → Personal access tokens → Delete.
