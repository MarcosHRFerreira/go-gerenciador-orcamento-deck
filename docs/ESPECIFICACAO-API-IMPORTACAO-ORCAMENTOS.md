# Especificacao da API de Importacao de Orcamentos

## Objetivo

Definir o contrato da API para importar a aba `ORCAMENTOS` da planilha `RELATORIO DE ORCAMENTOS-25.xlsx`, com fluxo seguro em duas etapas:

- `preview` para leitura, validacao e simulacao
- `importacao` para persistencia efetiva

Esta especificacao considera o modelo atual do sistema, especialmente:

- `budgets`
- `budget_statuses`
- `priorities`
- `installers`
- `projects`
- `project_types`
- `salespeople`
- `contacts`
- `loss_reasons`

## Principios da Solucao

- processar somente a aba `ORCAMENTOS`
- obrigar uma etapa de preview antes da importacao
- permitir criacao automatica de catalogos auxiliares ausentes
- aplicar `Nao informado` quando o valor estiver ausente e houver campo correspondente
- registrar relatorio detalhado por linha
- manter compatibilidade com o padrao atual de rotas REST do backend

## Escopo da Primeira Versao

Entram na V1:

- upload do arquivo `.xlsx`
- leitura da aba `ORCAMENTOS`
- identificacao do cabecalho na linha `10`
- normalizacao de colunas e valores
- preview com avisos e erros
- importacao com estrategia de duplicidade
- retorno de resumo e itens processados

Ficam fora da V1:

- importacao da aba `FOLLOW-UP`
- processamento em background com fila externa
- importacao incremental por multiplos arquivos em lote
- reversao automatica da carga

## Autenticacao e Autorizacao

Todos os endpoints devem exigir:

- `Authorization: Bearer <token>`

Perfil recomendado para acesso:

- `admin`
- ou usuario com permissao especifica de importacao, caso o controle de permissao seja expandido depois

## Fluxo Recomendado

1. cliente envia o arquivo para `preview`
2. API valida tipo, aba, cabecalho e linhas
3. API normaliza valores e resolve catalogos
4. API retorna um `preview_id` temporario com resumo e alertas
5. cliente exibe os resultados e solicita confirmacao
6. cliente chama a importacao usando o `preview_id`
7. API grava os dados e gera um relatorio final

## Endpoints Propostos

### `POST /budget-imports/preview`

Objetivo:

- receber o arquivo `.xlsx`
- ler somente a aba `ORCAMENTOS`
- validar estrutura e dados
- montar o plano de importacao sem persistir `budgets`

Formato recomendado:

- `multipart/form-data`

Campos do form-data:

- `file`: arquivo Excel obrigatorio
- `duplicate_strategy`: opcional, valores `ignore` ou `update`
- `create_missing_catalogs`: opcional, `true` por padrao
- `use_default_not_informed`: opcional, `true` por padrao

Comportamento:

- valida extensao `.xlsx`
- valida existencia da aba `ORCAMENTOS`
- considera a linha `10` como cabecalho
- ignora linhas vazias
- normaliza placeholders vazios como `-`, `N/E`, `N/I`
- identifica campos ausentes que deverao usar `Nao informado`
- calcula impacto de registros novos, atualizaveis e ignorados

Resposta de sucesso:

```json
{
  "preview_id": "imp_prev_01jz8lq4k7kgx3m5w0n8x9a1b2",
  "file_name": "RELATORIO DE ORCAMENTOS-25.xlsx",
  "sheet_name": "ORCAMENTOS",
  "header_row": 10,
  "expires_at": "2026-06-13T22:00:00Z",
  "options": {
    "duplicate_strategy": "update",
    "create_missing_catalogs": true,
    "use_default_not_informed": true
  },
  "summary": {
    "rows_read": 3156,
    "rows_valid": 3102,
    "rows_with_warning": 54,
    "rows_with_error": 0,
    "rows_empty_ignored": 21,
    "new_budgets": 2970,
    "existing_budgets": 132
  },
  "catalog_actions": {
    "budget_statuses_to_create": 4,
    "loss_reasons_to_create": 12,
    "installers_to_create": 18,
    "projects_to_create": 203,
    "project_types_to_create": 7,
    "salespeople_to_create": 2,
    "contacts_to_create": 126
  },
  "warnings": [
    {
      "code": "COMMISSION_INTERPRETATION_ASSUMED",
      "message": "A coluna COMISSAO foi tratada como valor numerico simples."
    }
  ],
  "errors": [],
  "sample_rows": [
    {
      "row_number": 11,
      "budget_number": "1",
      "status": "ready",
      "action": "create",
      "messages": []
    }
  ]
}
```

Erros esperados:

- `400 Bad Request`: arquivo ausente, formato invalido, cabecalho invalido
- `401 Unauthorized`: token ausente ou invalido
- `403 Forbidden`: usuario sem permissao
- `422 Unprocessable Entity`: arquivo valido, mas sem dados importaveis

### `POST /budget-imports`

Objetivo:

- confirmar e executar a importacao a partir de um `preview_id`

Formato recomendado:

- `application/json`

Payload:

```json
{
  "preview_id": "imp_prev_01jz8lq4k7kgx3m5w0n8x9a1b2",
  "duplicate_strategy": "update",
  "create_missing_catalogs": true,
  "use_default_not_informed": true
}
```

Regras:

- o `preview_id` precisa estar valido e nao expirado
- as opcoes finais devem ser compatíveis com o preview
- a importacao deve usar exatamente a massa validada no preview
- a API pode recriar a validacao rapidamente antes da persistencia

Resposta de sucesso:

```json
{
  "import_id": "imp_01jz8m4f1j4w1kq2r6v8m3c7p9",
  "preview_id": "imp_prev_01jz8lq4k7kgx3m5w0n8x9a1b2",
  "status": "completed",
  "started_at": "2026-06-13T21:05:10Z",
  "finished_at": "2026-06-13T21:05:28Z",
  "summary": {
    "rows_processed": 3102,
    "budgets_created": 2970,
    "budgets_updated": 132,
    "budgets_ignored": 0,
    "rows_failed": 0,
    "catalogs_created": 372
  },
  "result": {
    "message": "Importacao concluida com sucesso."
  }
}
```

Erros esperados:

- `400 Bad Request`: payload inconsistente
- `404 Not Found`: `preview_id` inexistente
- `409 Conflict`: preview expirado ou ja consumido
- `422 Unprocessable Entity`: ha erros impeditivos nas linhas

### `GET /budget-imports/:import_id`

Objetivo:

- consultar o resumo final de uma importacao executada

Resposta de sucesso:

```json
{
  "import_id": "imp_01jz8m4f1j4w1kq2r6v8m3c7p9",
  "status": "completed",
  "file_name": "RELATORIO DE ORCAMENTOS-25.xlsx",
  "sheet_name": "ORCAMENTOS",
  "created_by_user_id": 1,
  "created_at": "2026-06-13T21:05:10Z",
  "finished_at": "2026-06-13T21:05:28Z",
  "summary": {
    "rows_processed": 3102,
    "budgets_created": 2970,
    "budgets_updated": 132,
    "budgets_ignored": 0,
    "rows_failed": 0,
    "catalogs_created": 372
  }
}
```

### `GET /budget-imports/:import_id/rows`

Objetivo:

- listar o relatorio detalhado por linha

Query params sugeridos:

- `page`
- `page_size`
- `status`

Resposta de sucesso:

```json
{
  "items": [
    {
      "row_number": 11,
      "budget_number": "1",
      "action": "create",
      "status": "completed",
      "messages": [
        "Instalador criado automaticamente: KEEVA TEIC"
      ]
    },
    {
      "row_number": 12,
      "budget_number": "2",
      "action": "update",
      "status": "completed",
      "messages": []
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 3102
}
```

## Estrategia de Duplicidade

Como a tabela `budgets` possui unicidade em `budget_number + year_budget`, a API deve suportar:

- `ignore`: mantem o registro existente e marca a linha como ignorada
- `update`: atualiza o registro existente com base na linha importada

Padrao recomendado:

- `update`

## Regras de Negocio

### Arquivo e aba

- aceitar somente arquivos `.xlsx`
- usar somente a aba `ORCAMENTOS`
- rejeitar arquivos sem essa aba
- rejeitar arquivo sem cabecalho valido na linha `10`

### Linhas

- ignorar linhas totalmente vazias
- ignorar linhas sem `Nº DE ORCA`
- considerar `DATA` obrigatoria
- converter `REV.` em inteiro, usando `0` quando ausente ou invalido
- converter `VALOR BRUTO`, `COMISSAO`, `M2` e `VALOR CONCORRENTE` para decimal

### Texto ausente

Valores considerados ausentes:

- vazio
- espacos
- `-`
- `N/E`
- `N/I`

Campos da tabela `budgets` que devem receber `Nao informado` quando vazios:

- `competitor_name`
- `designer_name`
- `specification_details`
- `current_follow_up`

### Tabelas auxiliares

As tabelas auxiliares devem ter item padrao `Nao informado`, criado automaticamente quando ainda nao existir:

- `budget_statuses`
- `priorities`
- `installers`
- `projects`
- `project_types`
- `salespeople`
- `contacts`
- `loss_reasons`

Quando a coluna da planilha estiver vazia:

- o relacionamento deve apontar para o item `Nao informado`

Quando a coluna vier preenchida e o valor ainda nao existir:

- se `create_missing_catalogs = true`, a API cria o item automaticamente
- se `create_missing_catalogs = false`, a API registra erro impeditivo na linha

### Observacao importante de mapeamento

Com base na analise da planilha:

- `PRIORIDADE` deve alimentar `status_id`
- `STATUS` deve alimentar `current_follow_up`

Na V1, `priority_id` deve apontar para o item `Nao informado`, salvo se houver regra futura mais clara para preencher prioridade operacional.

## Mapeamento de Colunas

| Coluna da planilha | Campo destino | Regra |
| --- | --- | --- |
| `DATA` | `sent_at` | converter serial do Excel para `timestamp` |
| `Nº DE ORCA` | `budget_number` | texto normalizado |
| `REV.` | `revision` | inteiro, padrao `0` |
| `INSTALADOR` | `installer_id` | resolver em `installers` |
| `NOME DA OBRA` | `project_id` | resolver em `projects` |
| `TIPO DE OBRA` | `project_type_id` | resolver em `project_types` para vinculo do projeto |
| `VENDEDOR` | `salesperson_id` | resolver em `salespeople` |
| `CONTATO` | `contact_id` | resolver em `contacts` associado ao instalador |
| `VALOR BRUTO` | `gross_value` | decimal |
| `COMISSAO` | `commission_value` | decimal |
| `M2` | `area_m2` | decimal |
| `PRIORIDADE` | `status_id` | resolver em `budget_statuses` |
| `STATUS` | `current_follow_up` | texto livre |
| `CONCORRENTE` | `competitor_name` | texto ou `Nao informado` |
| `MOTIVO` | `loss_reason_id` | resolver em `loss_reasons` |
| `VALOR CONCORRENTE` | `competitor_price` | decimal ou `null` |
| `PROJETISTA` | `designer_name` | texto ou `Nao informado` |
| `ESPECIFICACOES` | `specification_details` | texto ou `Nao informado` |

## Modelo de Persistencia Sugerido

Para suportar auditoria e consulta posterior, a API pode introduzir tabelas tecnicas:

### `budget_imports`

Campos sugeridos:

- `id`
- `file_name`
- `sheet_name`
- `status`
- `duplicate_strategy`
- `create_missing_catalogs`
- `use_default_not_informed`
- `created_by_user_id`
- `started_at`
- `finished_at`
- `created_at`
- `updated_at`

### `budget_import_rows`

Campos sugeridos:

- `id`
- `budget_import_id`
- `row_number`
- `budget_number`
- `action`
- `status`
- `raw_payload`
- `normalized_payload`
- `messages`
- `created_at`
- `updated_at`

Observacao:

- se a V1 precisar ser mais simples, `preview` pode ficar temporariamente em memoria ou cache, mas `importacao` concluida deve ser persistida

## Padrao de Status

### Status do preview

- `ready`
- `expired`
- `consumed`

### Status da importacao

- `processing`
- `completed`
- `completed_with_errors`
- `failed`

### Status da linha

- `ready`
- `warning`
- `error`
- `completed`
- `ignored`
- `failed`

## Padrao de Erro

Formato recomendado:

```json
{
  "message": "Nao foi possivel validar a planilha.",
  "errors": [
    {
      "code": "INVALID_HEADER",
      "field": "sheet.ORCAMENTOS.header",
      "detail": "A linha 10 nao corresponde ao layout esperado."
    }
  ]
}
```

Codigos de erro sugeridos:

- `FILE_REQUIRED`
- `INVALID_FILE_TYPE`
- `SHEET_NOT_FOUND`
- `INVALID_HEADER`
- `PREVIEW_NOT_FOUND`
- `PREVIEW_EXPIRED`
- `PREVIEW_ALREADY_CONSUMED`
- `ROW_BUDGET_NUMBER_REQUIRED`
- `ROW_SENT_AT_INVALID`
- `ROW_NUMERIC_VALUE_INVALID`
- `MISSING_CATALOG_VALUE`

## Impacto no Frontend

Para a tela de carga, o frontend deve ter:

- seletor de arquivo
- preview com resumo
- tabela de avisos e erros
- confirmacao da importacao
- tela final com relatorio consolidado

Fluxo de consumo:

1. enviar arquivo para `POST /budget-imports/preview`
2. exibir `summary`, `warnings`, `errors` e `sample_rows`
3. ao confirmar, chamar `POST /budget-imports`
4. consultar `GET /budget-imports/:import_id`
5. carregar `GET /budget-imports/:import_id/rows` para detalhamento

## Recomendacao de Implementacao

Ordem recomendada:

1. criar seed ou rotina de garantia dos itens `Nao informado`
2. criar parser da aba `ORCAMENTOS`
3. criar servico de normalizacao e resolucao de catalogos
4. implementar endpoint de `preview`
5. implementar endpoint de `importacao`
6. persistir relatorio por linha
7. implementar a tela no frontend

## Recomendacao Final

O melhor desenho para a V1 e:

- `POST /budget-imports/preview`
- `POST /budget-imports`
- `GET /budget-imports/:import_id`
- `GET /budget-imports/:import_id/rows`

Esse conjunto entrega seguranca operacional, permite conferência antes de gravar e combina com a necessidade de importar uma planilha real com muitos registros e dados auxiliares incompletos.
