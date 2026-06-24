# Controle de Melhorias em Orcamentos e Grids

## Objetivo

Este documento consolida:

- o estudo funcional e tecnico da lista de melhorias solicitadas
- as fases recomendadas de execucao
- o que ja foi implementado
- o que ainda esta pendente

Ele serve como controle continuo para acompanhamento das entregas.

## Lista de Itens Solicitados

1. todas as grids do sistema com largura de colunas ajustavel
2. em orcamentos, a edicao deve trazer os campos relevantes da grid
3. prioridade por faixa de valor bruto, com validacao na importacao, exibicao amigavel e tabela auxiliar
4. trocar `Pedido` por `Fechado` e `Comissao` por `Fator`
5. filtro de valor da obra com minimo e maximo usando barra de scroll, carregando faixa do banco
6. retirar `Area m2` da grid e da edicao

## Diagnostico Consolidado

### Item 1. Grids com largura ajustavel

- O sistema usa majoritariamente `MUI Table` manual.
- Nao havia componente padrao para redimensionar colunas.
- A implementacao recomendada era criar um componente reutilizavel, sem trocar toda a base para outra biblioteca.

### Item 2. Edicao de orcamento incompleta

- A listagem mostrava `Construtora`, mas o formulario nao trazia esse campo.
- `Empresa` existe como origem do orcamento e faz sentido aparecer como informacao de leitura na edicao.

### Item 3. Prioridade por faixa

- Hoje a prioridade nao e calculada automaticamente por faixa de valor bruto.
- A importacao ainda nao aplica a classificacao desejada.
- Ja existe estrutura de prioridades, mas a regra por faixa precisa ser formalizada e aplicada de ponta a ponta.

### Item 4. Renomeacoes de negocio

- `Pedido` participa nao so da interface, mas tambem de regras de negocio.
- `Comissao` hoje e tratada como valor monetario.
- A troca de nome pode ser simples no label, mas pode exigir refatoracao maior se o significado mudar.

### Item 5. Filtro por valor com slider

- O backend ja suporta filtro minimo e maximo de valor bruto.
- O frontend ainda nao expunha essa capacidade em forma de barra de selecao.

### Item 6. Remocao de Area m2

- `Area m2` estava visivel na grid principal, na visao por obra e no formulario.
- A remocao na interface era segura, preservando compatibilidade no backend.

## Fases Recomendadas

### Fase 1. Ajustes visuais e consistencia da edicao

- implementar largura ajustavel nas grids de maior uso
- corrigir a edicao de orcamento com os campos relevantes
- remover `Area m2` da interface de orcamentos

### Fase 2. Prioridade por faixa

- definir oficialmente as faixas de valor bruto
- aplicar a classificacao na importacao
- aplicar a classificacao na criacao e edicao
- exibir o nome amigavel da faixa na tela
- usar tabela auxiliar para suporte a exibicao e governanca

### Fase 3. Filtro de valor com slider

- expor faixa minima e maxima a partir do banco
- criar controle visual com barra de selecao
- combinar slider com campos numericos para ajuste fino

### Fase 4. Renomeacoes de negocio

- trocar labels visuais de forma controlada
- revisar os pontos em que `Pedido` impacta regras
- revisar se `Fator` sera apenas novo nome ou mudanca semantica real

## Status Atual

### Fase 1

Status: `Concluida`

O que foi feito:

- criada infraestrutura reutilizavel de colunas ajustaveis
- aplicada largura ajustavel nas telas `Orcamentos` e `Acompanhamento de Entregas`
- corrigida a edicao de orcamento para incluir `Construtora`
- adicionada `Empresa` como informacao de leitura no formulario
- removido `Area m2` da grid principal
- removido `Area m2` da visao por obra
- removido `Area m2` do formulario de orcamento
- mantida compatibilidade com a API e com os fluxos existentes

Arquivos principais alterados nesta fase:

- `frontend/src/components/common/ResizableTable.tsx`
- `frontend/src/features/budgets/pages/BudgetListPage.tsx`
- `frontend/src/features/budgets/pages/BudgetDeliveryMonitorPage.tsx`
- `frontend/src/features/budgets/components/BudgetForm.tsx`
- `frontend/src/features/budgets/components/budgetFormValues.ts`
- `frontend/src/features/budgets/api/budgets.ts`
- `frontend/src/features/budgets/types/budget.ts`
- `frontend/src/features/projects/pages/ProjectDetailPage.tsx`

Validacao executada:

- diagnosticos dos arquivos alterados sem erros
- `yarn build` no frontend executado com sucesso

### Fase 2

Status: `Concluida`

O que foi feito:

- implementar regra de faixa:
- `Faixa 0 a 50k`
- `Faixa 50k a 250k`
- `Faixa acima de 250k`
- criada regra central reutilizavel para classificacao por `Valor bruto`
- aplicada a classificacao automatica na criacao e edicao de orcamentos
- aplicada a classificacao automatica na importacao
- ajustada a exibicao da prioridade na grid e no formulario para refletir a faixa
- criada migration para semear a tabela `priorities` com as tres faixas e fazer backfill dos orcamentos existentes

Arquivos principais alterados nesta fase:

- `backend/internal/budgetpriority/ranges.go`
- `backend/internal/service/budget/service.go`
- `backend/internal/service/budgetimport/service.go`
- `backend/internal/service/budgetimport/execute.go`
- `backend/internal/server/router.go`
- `backend/db/migrations/20260624103000_seed_budget_priority_ranges.sql`
- `backend/test/unit/budget_service_test.go`
- `backend/test/unit/budget_import_service_test.go`
- `frontend/src/features/budgets/utils/priorityRanges.ts`
- `frontend/src/features/budgets/components/BudgetForm.tsx`
- `frontend/src/features/budgets/pages/BudgetListPage.tsx`

Validacao executada:

- diagnosticos dos arquivos alterados sem erros
- `go test ./test/unit -run "Budget(Service|Import)"` executado com sucesso
- `yarn build` no frontend executado com sucesso

### Fase 3

Status: `Concluida`

O que foi feito:

- criado endpoint para retornar a faixa minima e maxima de `Valor bruto` com o mesmo escopo dos filtros atuais
- adicionado DTO de resposta e integracao no backend de orcamentos
- ligado o frontend ao novo endpoint `/budgets/gross-value-range`
- criado filtro visual de `Valor da obra` na tela de `Orcamentos`
- adicionados os campos `Valor minimo` e `Valor maximo`
- adicionado slider de duas alcas com faixa carregada dinamicamente do banco
- mantida a integracao com `URLSearchParams`, filtros ativos e listagem atual
- ajustada a normalizacao para evitar inversao entre minimo e maximo

Arquivos principais alterados nesta fase:

- `backend/internal/dto/budget_dto.go`
- `backend/internal/repository/budget/repository.go`
- `backend/internal/service/budget/service.go`
- `backend/internal/handler/budget/handler.go`
- `frontend/src/features/budgets/types/budget.ts`
- `frontend/src/features/budgets/api/budgets.ts`
- `frontend/src/features/budgets/pages/BudgetListPage.tsx`
- `frontend/src/features/dashboard/api/dashboard.ts`
- `frontend/src/features/projects/pages/ProjectDetailPage.tsx`
- `backend/test/unit/budget_service_test.go`

Validacao executada:

- diagnosticos dos arquivos alterados sem erros
- `go test ./test/unit` executado com sucesso
- `yarn build` no frontend executado com sucesso
- `go test ./...` bloqueado nas integracoes por ausencia de PostgreSQL local em `127.0.0.1:5433`

### Fase 4

Status: `Concluida`

O que foi feito:

- aplicada troca visual controlada de `Pedido` para `Fechado` nas telas principais do frontend
- aplicada troca visual controlada de `Comissao` para `Fator` nos pontos de entrada e exibicao do frontend
- criado helper central para padronizar a exibicao de termos de negocio no frontend
- mantidos os contratos tecnicos e persistencia com `commission_value` para evitar quebra de API e banco
- mantido o identificador operacional `PEDIDO` no backend, mas com tolerancia a `Fechado` em consultas dependentes do status vencedor
- atualizadas mensagens e labels do backend que chegam ao usuario, como `Fechado sem data de entrega`
- ajustados dashboard, acompanhamento de entregas, listagem de orcamentos, formulario e detalhe de obra para refletir a nova nomenclatura

Arquivos principais alterados nesta fase:

- `frontend/src/features/budgets/utils/businessTerms.ts`
- `frontend/src/features/budgets/components/BudgetForm.tsx`
- `frontend/src/features/budgets/pages/BudgetListPage.tsx`
- `frontend/src/features/budgets/pages/BudgetDeliveryMonitorPage.tsx`
- `frontend/src/features/projects/pages/ProjectDetailPage.tsx`
- `frontend/src/features/dashboard/pages/DashboardPage.tsx`
- `frontend/src/features/dashboard/api/dashboard.ts`
- `backend/internal/service/budget/service.go`
- `backend/internal/repository/budget/repository.go`
- `backend/internal/repository/deliveryalert/repository.go`
- `backend/internal/repository/dashboard/repository.go`
- `backend/test/unit/budget_service_test.go`

Validacao executada:

- diagnosticos dos arquivos alterados sem erros
- `go test ./test/unit` executado com sucesso
- `yarn build` no frontend executado com sucesso

### Onda complementar. Expansao de grids com largura ajustavel

Status: `Concluida parcialmente`

O que foi feito:

- expandida a infraestrutura de colunas ajustaveis para novas grids administrativas
- aplicada largura ajustavel nas telas `Obras`, `Usuarios`, `Vendedores`, `Orcamentistas` e `Tipos de sistema`
- mantida a persistencia local das larguras por tela, seguindo o mesmo padrao usado em `Orcamentos`
- preservado o layout e o comportamento funcional das acoes existentes em cada grid

Arquivos principais alterados nesta onda:

- `frontend/src/features/projects/pages/ProjectListPage.tsx`
- `frontend/src/features/users/pages/UserListPage.tsx`
- `frontend/src/features/salespeople/pages/SalespersonListPage.tsx`
- `frontend/src/features/estimators/pages/EstimatorListPage.tsx`
- `frontend/src/features/system-types/pages/SystemTypeListPage.tsx`

Validacao executada:

- diagnosticos dos arquivos alterados sem erros
- `yarn build` no frontend executado com sucesso

## Pendencias Abertas

- avaliar se ainda vale expandir largura ajustavel para grids restantes, como tabelas auxiliares e fluxos de importacao
- decidir futuramente se `Fator` deixara de ser apenas label visual e passara a representar mudanca semantica real no modelo

## Proximo Passo Recomendado

Revisar as grids restantes fora do nucleo principal para decidir se a proxima onda incluira tabelas auxiliares e importacao ou se o foco passa a ser a avaliacao estrutural de `Fator`.

## Historico de Execucao

### Registro Atual

- estudo consolidado da lista criado
- fases de execucao definidas
- fase 1 registrada como concluida
- fase 2 registrada como concluida
- fase 3 registrada como concluida
- fase 4 registrada como concluida
- migration de prioridade aplicada no banco externo
- onda complementar de grids administrativas executada
- pendencias e proximos passos documentados
