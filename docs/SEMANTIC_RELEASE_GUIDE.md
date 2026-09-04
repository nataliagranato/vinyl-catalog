# Semantic Release Guide

## O que é Semantic Release?

**Semantic Release** é uma ferramenta de automação que usa os princípios do **SemVer (Semantic Versioning)** para determinar automaticamente a próxima versão do projeto baseado nas mensagens de commit.

## Como Funciona

```
Commit → Semantic Release Analisa → Bump Version → Cria Tag → Cria Release
```

## Versão Atual

- **Projeto não publicado no NPM** (é uma API Cloudflare Workers)
- **Semantic Release configurado apenas para GitHub Releases**
- **Versão gerenciada via tags no GitHub**

## Configuração Atual

**Workflow:** `.github/workflows/release.yml`
**Config:** `.releaserc.json`
**Branches:** `main` e `develop`

## Como Usar

### Conventional Commits (Obrigatório)

As mensagens de commit DEVEM seguir o formato:

```
<tipo>(<escopo>): <descrição>

[opcional body]

[opcional footer]
```

### Tipos de Commit e Impacto na Versão

| Tipo | Exemplo | Impacto |
|------|---------|---------|
| `feat` | `feat: adicionar autenticação JWT` | **MINOR** (v1.0.0 → v1.1.0) |
| `fix` | `fix: corrigir bug no login` | **PATCH** (v1.0.0 → v1.0.1) |
| `feat!` | `feat!: mudar estrutura da API` | **MAJOR** (v1.0.0 → v2.0.0) |
| `BREAKING CHANGE` | `feat: nova feature\n\nBREAKING CHANGE: remove endpoint antigo` | **MAJOR** (v1.0.0 → v2.0.0) |
| `docs` | `docs: atualizar README` | Nenhum |
| `style` | `style: formatar código` | Nenhum |
| `refactor` | `refactor: otimizar função` | Nenhum |
| `perf` | `perf: melhorar performance` | **PATCH** |
| `test` | `test: adicionar testes` | Nenhum |
| `build` | `build: atualizar dependências` | Nenhum |
| `ci` | `ci: atualizar workflow` | Nenhum |
| `chore` | `chore: limpeza de código` | Nenhum |

### Exemplos Práticos

**Nova Feature (MINOR):**
```bash
git commit -m "feat: adicionar endpoint de busca por artista"
```

**Bug Fix (PATCH):**
```bash
git commit -m "fix: corrigir validação de email no login"
```

**Breaking Change (MAJOR):**
```bash
git commit -m "feat!: mudar formato de resposta da API"
```

Ou com footer:
```bash
git commit -m "feat: adicionar autenticação

BREAKING CHANGE: removido suporte para API key antiga"
```

**Documentação (Sem impacto):**
```bash
git commit -m "docs: atualizar guia de instalação"
```

**Refatoração (Sem impacto):**
```bash
git commit -m "refactor: simplificar lógica de autenticação"
```

## Fluxo de Trabalho

### 1. Criar Branch Feature
```bash
git checkout -b feature/nova-autenticacao
```

### 2. Fazer Commits com Conventional Commits
```bash
git add .
git commit -m "feat: implementar autenticação JWT"
git commit -m "fix: corrigir validação de token"
git commit -m "docs: atualizar documentação de auth"
```

### 3. Push e PR
```bash
git push origin feature/nova-autenticacao
```
Criar PR no GitHub

### 4. Merge para Main/Develop
```bash
git checkout main
git merge feature/nova-autenticacao
git push origin main
```

### 5. Semantic Release Atua Automaticamente
- Analisa os commits
- Determina a nova versão
- Cria tag (ex: v1.1.0)
- Cria GitHub Release
- Gera changelog automaticamente

## Workflow GitHub Actions

O workflow `.github/workflows/release.yml` executa:

1. **Checkout** do código
2. **Setup Node.js** v18
3. **Install dependencies**
4. **Run tests** (opcional, não falha se falhar)
5. **Semantic Release** analisa commits e cria release

## Como Verificar a Próxima Versão

### Localmente (Dry Run)
```bash
npx semantic-release --dry-run
```

### Após Merge
```bash
git log --oneline --no-decorate
```
Ver os commits recentes para prever a versão

## Exemplo de Fluxo Real

### Cenário 1: Nova Feature
```bash
# Branch develop
git commit -m "feat: adicionar suporte a upload de imagens"
git push origin develop

# Merge para main
git checkout main
git merge develop
git push origin main

# Resultado: v1.0.0 → v1.1.0
```

### Cenário 2: Bug Fix
```bash
git commit -m "fix: corrigir erro 500 no endpoint de busca"
git push origin main

# Resultado: v1.1.0 → v1.1.1
```

### Cenário 3: Breaking Change
```bash
git commit -m "feat!: remover endpoint /api/v1/legacy

BREAKING CHANGE: endpoint legacy não é mais suportado"
git push origin main

# Resultado: v1.1.1 → v2.0.0
```

## Boas Práticas

### 1. Use Conventional Commits Sempre
❌ **Errado:**
```bash
git commit -m "mudar coisas"
git commit -m "fix bug"
git commit -m "atualizar"
```

✅ **Correto:**
```bash
git commit -m "feat: adicionar validação de email"
git commit -m "fix: corrigir erro no login"
git commit -m "docs: atualizar README"
```

### 2. Seja Específico
❌ **Errado:**
```bash
git commit -m "fix: corrigir bug"
```

✅ **Correto:**
```bash
git commit -m "fix: corrigir validação de senha no login"
```

### 3. Use Scopo Quando Aplicável
```bash
git commit -m "feat(auth): adicionar refresh token"
git commit -m "fix(upload): corrigir upload de arquivos grandes"
git commit -m "feat(vinyls): adicionar busca por gênero"
```

### 4. Breaking Changes Declarados
```bash
git commit -m "feat!: mudar formato de resposta

BREAKING CHANGE: API agora retorna objetos com camelCase em vez de snake_case"
```

## Changelog Automático

O Semantic Release gera automaticamente o changelog na GitHub Release com base nos commits:

```markdown
## [1.1.0](github.com/nataliagranato/vinyl-catalog/compare/v1.0.0...v1.1.0) (2024-01-15)

### Features

* adicionar suporte a upload de imagens ([abc123](https://github.com/nataliagranato/vinyl-catalog/commit/abc123))
* adicionar endpoint de busca por artista ([def456](https://github.com/nataliagranato/vinyl-catalog/commit/def456))

### Bug Fixes

* corrigir validação de token ([ghi789](https://github.com/nataliagranato/vinyl-catalog/commit/ghi789))
```

## Troubleshooting

### Release Não Foi Criado

**Verifique:**
1. Commit segue Conventional Commits?
2. Branch é `main` ou `develop`?
3. Workflow rodou com sucesso?
4. `GITHUB_TOKEN` tem permissões de escrita?

### Versão Não Subiu

**Verifique:**
1. Há commits com `feat` ou `fix` desde a última release?
2. Commit anterior criou release corretamente?
3. Verifique logs do workflow no GitHub Actions

### Erro de Permissão

**Certifique-se:**
```yaml
permissions:
  contents: write
```
No workflow `.github/workflows/release.yml`

## Ferramentas Auxiliares

### Commitlint (Validação de Commits)
```bash
npm install --save-dev @commitlint/cli @commitlint/config-conventional
echo "module.exports = {extends: ['@commitlint/config-conventional']}" > commitlint.config.js
```

### Commitizen (Prompt de Commit)
```bash
npm install --save-dev commitizen cz-conventional-changelog
echo "package.json" | npx commitizen init cz-conventional-changelog --save-dev --save-exact
```

## Configuração Atual do Projeto

**Workflow:** `.github/workflows/release.yml`
**Config:** `.releaserc.json`
**Branches:** `main`, `develop`
**Plugins:**
- `@semantic-release/commit-analyzer`
- `@semantic-release/release-notes-generator`
- `@semantic-release/github`
- `@semantic-release/git`

## Resumo

| Ação | Comando | Resultado |
|------|---------|-----------|
| Nova feature | `git commit -m "feat: descrição"` | Minor bump |
| Bug fix | `git commit -m "fix: descrição"` | Patch bump |
| Breaking change | `git commit -m "feat!: descrição"` | Major bump |
| Documentação | `git commit -m "docs: descrição"` | Sem bump |
| Refatoração | `git commit -m "refactor: descrição"` | Sem bump |

**Simples:** Use Conventional Commits, push para main/develop, e o Semantic Release cuida do resto!