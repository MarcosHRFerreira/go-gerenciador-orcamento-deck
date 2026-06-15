# Fase 7 - Operacao e Governanca da Importacao

## Objetivo

Consolidar as decisoes operacionais do importador multi-layout apos a entrada de Rocktec e Trox.

Esta fase fecha tres temas:

- estrategia de defaults e seeds por layout
- politica de duplicidade no contexto multiempresa
- observabilidade minima do processo de importacao

## Resumo Executivo

As decisoes adotadas nesta fase sao:

- manter `Nao informado` como fallback operacional padrao para os dois layouts
- tratar a duplicidade pelo trio `source_company + budget_number + year_budget`
- manter compatibilidade transitoria com registros legados sem `source_company`
- registrar logs estruturados no preview e na execucao da importacao

## Politica de Defaults por Layout

### Regra comum

Os dois layouts passam a operar com a mesma base de defaults catalogados:

- `Status`
- `Prioridade`
- `Instalador`
- `Projeto`
- `Tipo de obra`
- `Vendedor`
- `Contato`
- `Motivo de perda`

Quando a opcao `Usar Nao informado` estiver ativa no preview:

- campos ausentes ou nao conciliados podem cair no item padrao `Nao informado`
- se o item padrao nao existir e `Criar catalogos ausentes` estiver ativo, ele pode ser criado automaticamente

### Rocktec

Na Rocktec:

- a aderencia ao dominio principal continua alta
- os defaults entram principalmente para campos vazios ou marcados como ausentes na planilha
- `PRIORIDADE` continua com tratamento operacional atual, usando `Nao informado` quando necessario

### Trox

Na Trox:

- varios campos sem aderencia direta ao dominio continuam usando `Nao informado`
- `Status` da origem continua sendo tratado como `current_follow_up`
- `status_name`, `priority_name`, `project_type_name`, `loss_reason_name`, `competitor_name`, `designer_name` e `specification` permanecem como `Nao informado`

## Politica de Duplicidade

## Regra adotada

A duplicidade do importador multi-layout passa a ser avaliada por:

- `source_company`
- `budget_number`
- `year_budget`

Em termos praticos:

- Rocktec compara com Rocktec
- Trox compara com Trox
- um numero de orçamento igual entre empresas diferentes nao deve mais colidir automaticamente

## Compatibilidade com legados

Existe uma regra de transicao importante:

- registros antigos sem `source_company` continuam elegiveis como correspondencia legado

Motivo:

- evitar duplicacao artificial apos a implantacao da rastreabilidade de origem
- permitir que novas importacoes reconciliem dados historicos que foram gravados antes da Fase 5

## Impacto pratico

Antes:

- a chave operacional de duplicidade era `budget_number + year_budget`

Agora:

- a chave operacional recomendada e `source_company + budget_number + year_budget`
- com fallback legado para `source_company = ''`

## Observabilidade

## Logs estruturados adicionados

O importador passa a registrar logs estruturados para:

- inicio do preview
- conclusao do preview
- inicio da execucao da importacao
- conclusao da execucao da importacao

Campos esperados nos logs:

- `import_action`
- `source_layout`
- `source_company`
- `file_name`
- `preview_id`
- `import_id`
- `rows_read`
- `rows_valid`
- `rows_with_warning`
- `rows_with_error`
- `rows_processed`
- `budgets_created`
- `budgets_updated`
- `budgets_ignored`
- `rows_failed`
- `catalogs_created`

## Governanca exposta no preview

O preview da importacao passa a exibir:

- layout identificado
- empresa de origem
- escopo de duplicidade
- politica de valores ausentes
- catalogos que dependem de `Nao informado`
- observacao sobre conciliacao com legados

Isso permite que o usuario entenda a governanca aplicada antes de confirmar a carga.

## Decisoes Fechadas

- `Nao informado` continua sendo o fallback padrao dos catalogos operacionais
- a Trox continua com mapeamento conservador para o dominio principal
- a duplicidade deixa de ser global entre empresas
- a reconciliacao com legados sem origem continua ativa nesta fase
- preview e execucao passam a ter logs estruturados do processo

## Proximo Passo

Com a Fase 7 concluida, o proximo passo recomendado e:

- avaliar se os registros legados devem receber retroativamente `source_company`
- decidir se a compatibilidade com `source_company = ''` continuara indefinidamente ou sera removida em migracao futura
- ampliar analytics operacionais da importacao caso o volume de uso cresca
