# Homologacao Da Fase 4

## Status

- Backend principal validado para `PUT /budgets/:id`
- Fluxo legado de `Obras` validado para `POST /budgets/:id/elect-winner`
- Frontend validado com `yarn build`
- Endpoint tecnico `PATCH /budgets/:id/status` mantido como fluxo isolado, sem automacao de cancelamento por instalador nesta entrega

## Evidencias Automatizadas

- `go test ./test/unit/... -run "BudgetService(UpdateShouldNotCreateClosingNoticeWithoutRealStatusTransition|ElectProjectWinnerShouldEnsureStatusesAndCallRepository)"`
- `go test ./test/integration/... -run "Budget(StatusHistory|Update|ElectWinner)|Notices"`
- `go test ./... -run ^$`
- `yarn build`

## Resultado Esperado Da Entrega

- Fechar um orcamento em `Orcamentos` cancela apenas concorrentes de outros instaladores da mesma obra
- Orcamentos do mesmo instalador podem permanecer ativos como escopos complementares
- O vendedor e os administradores recebem aviso automatico quando houver fechamento com `project_id` e `installer_id`
- A tela de `Obras` segue a mesma semantica por instalador

## Checklist Manual

- Cenario 1: mesma obra com `Dutos`, `Damper` e `Difusao` no mesmo instalador
- Acao: marcar apenas um dos orcamentos como `Fechado` em `Orcamentos`
- Validar: os demais do mesmo instalador permanecem ativos

- Cenario 2: mesma obra com instaladores concorrentes
- Acao: marcar um orcamento como `Fechado` em `Orcamentos`
- Validar: orcamentos em aberto dos outros instaladores ficam `Cancelado`

- Cenario 3: aviso automatico
- Acao: concluir um fechamento com `project_id` e `installer_id`
- Validar: vendedor e admins visualizam o aviso em `Comunicacao > Avisos`

- Cenario 4: tela de `Obras`
- Acao: abrir a confirmacao de `eleger vencedor`
- Validar: o texto nao promete mais cancelamento total da obra

- Cenario 5: troca de vencedor entre instaladores
- Acao: eleger um vencedor em `Obras` para um instalador diferente do atual
- Validar: o vencedor anterior deixa de prevalecer, os complementares do novo instalador podem ser restaurados e os concorrentes ficam cancelados

## Observacao

- O endpoint `PATCH /budgets/:id/status` continua util para troca tecnica de status e historico, mas nao foi promovido a fluxo principal de automacao desta entrega.
