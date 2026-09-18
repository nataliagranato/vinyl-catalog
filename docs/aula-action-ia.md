# GitHub Action reutilizável de documentação com IA

O `vinyl-catalog` usa a Action TypeScript `Tech-Preta/actions/ai-docs@v1` para analisar o projeto, gerar documentação técnica e abrir ou atualizar um pull request para a `main`.

## Responsabilidades

O fluxo está dividido entre dois repositórios:

- `Tech-Preta/actions`: implementa a coleta segura de contexto, integração com APIs compatíveis com OpenAI, validação da resposta e publicação do PR via API do GitHub.
- `vinyl-catalog`: define o prompt específico, os arquivos que podem ser alterados e o disparo manual.

Não há script local para gerar a documentação. O bundle JavaScript versionado na Action contém toda a lógica de runtime.

## Arquivos locais

```text
.github/ai-docs-prompt.md       # requisitos de documentação deste projeto
.github/workflows/ai-docs.yml   # workflow manual consumidor
```

## Documentos autorizados

A allowlist do workflow permite que a IA altere somente:

- `docs/adr/ADR-001-decisoes-arquiteturais.md`
- `docs/runbooks/runbook.md`
- `docs/ai-analysis.md`
- `README.md`
- `docs/swagger/swagger.yaml`

Qualquer outro caminho retornado pelo modelo faz a execução falhar antes da criação de branch ou commit.

## Credenciais

Cadastre estes Actions secrets no repositório:

- `VERBOO_API_KEY`: chave do provedor de IA.
- `GH_TOKEN`: token com acesso ao repositório e permissão para criar branches e pull requests.

O workflow declara:

```yaml
permissions:
  contents: write
  pull-requests: write
```

As credenciais são fornecidas como variáveis de ambiente e mascaradas pela Action. Nunca devem ser colocadas no prompt ou commitadas.

## Disparo manual

Na interface do GitHub, acesse **Actions → Documentação com IA → Run workflow**.

Pela CLI:

```bash
gh workflow run ai-docs.yml -R nataliagranato/vinyl-catalog --ref main
gh run watch -R nataliagranato/vinyl-catalog --exit-status
```

## Fluxo da execução

1. `actions/checkout` disponibiliza o estado da `main`.
2. A Action lê `.github/ai-docs-prompt.md`.
3. Apenas arquivos rastreados e seguros entram no contexto; segredos, binários, dependências, artefatos e arquivos grandes são excluídos.
4. A API Verboo recebe o prompt e o contexto usando o modelo padrão `glm-5.3-flash`.
5. A resposta JSON é validada quanto ao schema, volume, caminhos e allowlist.
6. A Action compara o conteúdo com a `main`.
7. Se houver mudanças, um único commit é publicado em `ai/docs`.
8. Um PR para `main` é criado ou o PR aberto existente é atualizado.

Se o conteúdo gerado for idêntico ao existente, não há novo commit. A Action nunca faz merge automático.

## Troca de provedor ou modelo

A Verboo é o padrão, mas a Action aceita qualquer endpoint de chat completions compatível com OpenAI:

```yaml
with:
  api-base-url: https://provedor.example/v1
  model: modelo-documentacao
```

Por compatibilidade, a chave continua sendo fornecida em `VERBOO_API_KEY`.

## Troubleshooting

- **Action não encontrada:** `Tech-Preta/actions` precisa estar público para consumo por um repositório fora da organização, ou a política de compartilhamento precisa autorizar o consumidor.
- **401/403 da IA:** revise `VERBOO_API_KEY`, o endpoint e o acesso ao modelo.
- **403 do GitHub:** revise `GH_TOKEN`, as permissões do workflow e as configurações de Actions da organização.
- **429:** a Action realiza retentativas; persistência indica limite de cota.
- **JSON inválido ou truncado:** reduza o contexto ou revise `max-output-tokens`.
- **Caminho rejeitado:** confirme que o arquivo está na allowlist do workflow.

## Revisão humana

O PR é uma proposta. Antes do merge, confira principalmente:

- se os endpoints do Swagger existem no código;
- se comandos e variáveis do runbook são reais;
- se o README não remove informações relevantes;
- se riscos e recomendações estão claramente separados do estado atual do projeto.
