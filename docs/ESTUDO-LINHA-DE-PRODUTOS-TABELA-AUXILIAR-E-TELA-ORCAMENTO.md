# Estudo: Linha de Produtos como Tabela Auxiliar e Exibicao na Tela de Orcamento

## Objetivo

Garantir que a coluna `Linha de produtos` da planilha seja tratada como catalogo auxiliar oficial do sistema e que essa informacao fique visivel e editavel na tela de orcamento.

## Resumo Executivo

O sistema ja possui uma implementacao parcial dessa ideia:

- a tabela `product_lines` ja existe no banco
- a tabela `budgets` ja possui a coluna `product_line_id`
- a importacao Trox/Rocktec ja le `Linha de produtos`
- a execucao da importacao ja localiza ou cria registros em `product_lines`
- o backend ja devolve `product_line_id`, `product_line_code` e `product_line_name` no `BudgetResponse`

O problema atual nao esta na modelagem principal do banco, mas sim na integracao incompleta com a camada de interface:

- nao existe endpoint dedicado para listar/cadastrar `product_lines`
- os catalogos usados pelo frontend da tela de orcamento nao carregam `product_lines`
- o formulario de orcamento nao possui campo para `Linha de produtos`
- os types e mapeamentos do frontend ignoram `product_line_id`, `product_line_code` e `product_line_name`
- o preview da importacao no frontend nao exibe o contador de `product_lines_to_create`, embora o backend ja envie esse dado

Conclusao: a base ja esta parcialmente pronta. O trabalho necessario agora e completar a exposicao desse catalogo na API e no frontend da tela de orcamento.

## Estado Atual

### Banco de dados

Ja existem migrations para suportar `Linha de produtos` como catalogo auxiliar:

- `backend/db/migrations/20260617100000_create_product_lines_table.sql`
- `backend/db/migrations/20260617101000_add_product_line_and_construction_company_to_budgets.sql`

Estrutura atual:

```sql
product_lines
- id
- code
- name
- description
- created_at
- updated_at

budgets
- product_line_id BIGINT NULL
- FK -> product_lines(id) ON DELETE SET NULL
```

Isso significa que a transformacao da coluna da planilha em tabela auxiliar ja foi iniciada e consolidada no schema.

### Importacao

O parser da Rocktec ja captura a coluna `Linha de produtos`:

- `backend/internal/service/budgetimport/rocktec_layout.go`

Na execucao da importacao, o sistema:

1. le `row.productLineName`
2. verifica se deve ignorar valores nao informados
3. busca o catalogo em memoria
4. cria o registro em `product_lines` quando necessario
5. grava o relacionamento em `budgets.product_line_id`

Arquivos centrais:

- `backend/internal/service/budgetimport/service.go`
- `backend/internal/service/budgetimport/execute.go`
- `backend/internal/repository/productline/repository.go`

Observacao importante:

- o preview do backend tambem contabiliza `product_lines_to_create`
- porem o frontend ainda nao mostra esse indicador

### Backend de orcamentos

O backend de orcamentos ja esta preparado para trafegar essa informacao:

- `backend/internal/dto/budget_dto.go`
- `backend/internal/service/budget/service.go`
- `backend/internal/repository/budget/repository.go`

O `BudgetResponse` ja contem:

- `product_line_id`
- `product_line_code`
- `product_line_name`

O repositorio de orcamentos ja faz `LEFT JOIN product_lines pl ON pl.id = b.product_line_id`.

### Frontend da tela de orcamento

Aqui esta a principal lacuna atual.

Arquivos relevantes:

- `frontend/src/features/budgets/types/budget.ts`
- `frontend/src/features/budgets/api/budgets.ts`
- `frontend/src/features/budgets/components/budgetFormValues.ts`
- `frontend/src/features/budgets/components/BudgetForm.tsx`
- `frontend/src/features/budgets/pages/BudgetListPage.tsx`
- `frontend/src/features/budgets/pages/BudgetImportPage.tsx`

Problemas identificados:

1. `BudgetApiItem` e `BudgetListItem` no frontend nao carregam `product_line_id`, `product_line_code` e `product_line_name`
2. `BudgetCreatePayload` nao possui `productLineId`
3. `BudgetFormValues` nao possui `productLineId`
4. `BudgetForm.tsx` nao renderiza nenhum campo `Linha de produtos`
5. `getBudgetCatalogsRequest()` nao busca `product_lines`
6. `BudgetCatalogsResult` nao possui colecao `productLines`
7. `BudgetImportPage.tsx` nao exibe `product_lines_to_create`

Em resumo: o backend ja suporta o dado, mas o frontend ainda nao o consome corretamente.

## Gap Funcional

Para atender ao objetivo completo, o sistema precisa chegar no seguinte estado:

1. a importacao continua alimentando `product_lines`
2. a tela de orcamento passa a exibir a `Linha de produtos`
3. o usuario consegue selecionar a `Linha de produtos` no cadastro/edicao do orcamento
4. a listagem e o detalhe do orcamento passam a mostrar essa informacao quando existir
5. opcionalmente, o sistema ganha um cadastro proprio de `Linhas de produtos`

Hoje apenas o item `1` esta efetivamente concluido de ponta a ponta.

## Proposta de Solucao

### Diretriz principal

Adotar `product_lines` como fonte oficial de `Linha de produtos` no sistema inteiro.

Regras recomendadas:

- a referencia oficial deve ser `budgets.product_line_id`
- o nome exibido na interface deve vir de `product_lines.name`
- o codigo deve continuar existindo em `product_lines.code`
- a importacao pode continuar criando catalogos ausentes quando a opcao `create_missing_catalogs` estiver ativa

### Etapa 1 - Completar integracao da API

Criar exposicao de `product_lines` para o frontend.

Existem 2 caminhos possiveis:

#### Opcao A: endpoint dedicado

Criar endpoints como:

- `GET /product-lines`
- `POST /product-lines`
- `PUT /product-lines/:id`
- `DELETE /product-lines/:id` se fizer sentido de negocio

Vantagens:

- segue o mesmo padrao dos demais catalogos
- facilita futura tela administrativa de cadastro
- desacopla a tela de orcamento do detalhe de implementacao dos catalogos

Desvantagem:

- exige criar handler e service especificos

#### Opcao B: incluir no endpoint de catalogos de orcamento

Adicionar `productLines` no retorno agregado da tela de orcamento.

Vantagem:

- menor esforco inicial

Desvantagem:

- nao resolve o problema de manutencao administrativa do catalogo
- continua sem uma API propria para o recurso

Recomendacao:

- usar a `Opcao A`
- no curto prazo, se quiser acelerar a entrega da tela de orcamento, pode-se combinar:
  - `GET /product-lines` para leitura
  - manutencao administrativa em fase posterior

### Etapa 2 - Completar types e mapeamentos do frontend

Atualizar:

- `BudgetApiItem`
- `BudgetListItem`
- `BudgetDetailItem`
- `BudgetCreatePayload`
- `BudgetFormValues`
- `BudgetCatalogsResult`

Campos esperados:

```ts
productLineId: number | null
productLineCode: string | null
productLineName: string | null
```

E no catalogo:

```ts
productLines: BudgetCatalogItem[]
```

### Etapa 3 - Exibir na tela de orcamento

Atualizar `BudgetForm.tsx` para incluir um campo `Linha de produtos`.

Recomendacao de UX:

- usar `Autocomplete` ou `TextField select`, como ja e feito com outros catalogos
- manter comportamento opcional
- quando o orcamento vier da importacao, mostrar o valor selecionado automaticamente

Tambem e recomendavel exibir `Linha de produtos`:

- na listagem de orcamentos
- no detalhe do orcamento

Isso facilita validacao visual apos a importacao.

### Etapa 4 - Ajustar preview da importacao

Atualizar `BudgetImportPage.tsx` para mostrar:

- `Linhas de produtos` = `previewResult.catalogActions.productLinesToCreate`

Isso deixa claro para o usuario que a importacao tambem esta alimentando esse catalogo auxiliar.

## Impactos Tecnicos

### Backend

Implementacoes provaveis:

- criar `internal/service/productline`
- criar `internal/handler/productline`
- registrar rotas em `internal/server/router.go`
- expor `List` para consumo do frontend
- opcionalmente expor `Create`, `Update` e `Delete`

Como o repositorio ja existe, o custo principal esta na camada de servico/handler.

### Frontend

Implementacoes provaveis:

- incluir `productLines` no carregamento de catalogos
- adicionar campo no formulario
- ajustar mapping de API
- mostrar `productLineName` na listagem
- atualizar a tela de preview da importacao

### Banco

Nenhuma mudanca estrutural obrigatoria imediata, porque o schema principal ja existe.

Mudancas opcionais futuras:

- backfill de codigos padronizados em `product_lines.code`
- regras de normalizacao para evitar duplicidades semanticas

## Riscos e Pontos de Atencao

### 1. Duplicidade semantica

Exemplos:

- `Filtros`
- `filtros`
- `FILTROS`
- `Filtro`

Hoje a tabela possui `UNIQUE(name)`, mas isso nao resolve todas as variacoes semanticas.

Recomendacao:

- manter normalizacao no fluxo de importacao
- definir uma regra oficial de nome e codigo
- avaliar normalizacao case-insensitive no futuro

### 2. Estrategia de codigo

Hoje o repositorio e a importacao trabalham com `code`, mas a origem da planilha entrega principalmente `name`.

Decisao recomendada:

- manter `code` gerado automaticamente a partir do nome, como ja vem sendo feito no fluxo da importacao
- permitir ajuste manual posterior em eventual tela administrativa

### 3. Exclusao de catalogo

O relacionamento atual usa `ON DELETE SET NULL`.

Isso e bom para seguranca referencial, mas permite que um orcamento historico perca a referencia caso uma linha de produtos seja excluida.

Recomendacao:

- evitar exclusao fisica no uso operacional
- preferir inativacao no futuro, caso a tela administrativa seja criada

## Recomendacao Final

Seguir com a demanda em 2 fases:

### Fase 1 - Entrega funcional rapida

- expor leitura de `product_lines`
- carregar `productLines` na tela de orcamento
- adicionar campo `Linha de produtos` no formulario
- exibir `productLineName` na listagem e detalhe
- mostrar `product_lines_to_create` no preview da importacao

Resultado:

- a coluna da planilha passa a ser visivel e utilizavel de ponta a ponta
- baixo risco, porque o schema e a importacao ja estao prontos

### Fase 2 - Governanca do catalogo

- criar cadastro administrativo de `Linhas de produtos`
- permitir criacao/edicao controlada
- estudar inativacao em vez de exclusao
- padronizar estrategia de codigo/nome

Resultado:

- o catalogo deixa de depender apenas da importacao para ser mantido

## Conclusao

Tecnicamente, a transformacao de `Linha de produtos` em tabela auxiliar ja foi feita no nucleo do sistema. O que falta agora e completar a exposicao desse dado no fluxo funcional da tela de orcamento e, opcionalmente, criar uma gestao administrativa dedicada para esse catalogo.

Ou seja:

- banco: pronto
- importacao: pronta
- backend de orcamentos: quase pronto
- frontend de orcamentos: incompleto

O proximo passo mais eficiente e implementar a Fase 1, porque ela fecha rapidamente a experiencia de ponta a ponta com baixo impacto estrutural.
