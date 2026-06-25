# Estudo: Fechamento Automatico Em Orcamentos Sem Cancelar Escopos Complementares

## Objetivo

Documentar o que precisa ser feito para atender o seguinte comportamento:

- Ao editar um orcamento e alterar o status para `Fechado`, o sistema deve agir diretamente na tela de `Orcamentos`, sem exigir a etapa manual de entrar em `Obras` para eleger vencedor.
- Ao fechar um orcamento, o sistema deve cancelar apenas os orcamentos em aberto de outros instaladores da mesma obra.
- Os orcamentos em aberto do mesmo instalador vencedor nao devem ser cancelados automaticamente, porque podem representar escopos complementares da mesma obra, como `Dutos`, `Damper` e `Difusao`.
- O vendedor deve receber um alerta informando que aquela obra teve um fechamento parcial ou um fechamento de escopo dentro do mesmo cliente/obra.

## Resumo Executivo

Hoje o sistema ja possui a logica de "eleger vencedor da obra", mas ela esta concentrada em um fluxo separado, disparado pela tela de `Obras`.

O problema do caso solicitado e que:

- editar um orcamento para `Fechado` na tela de `Orcamentos` nao executa a regra de vencedor da obra;
- o fluxo atual de vencedor cancela todos os demais orcamentos da obra, sem distinguir `instalador` e sem preservar escopos complementares;
- o sistema nao gera um alerta especifico ao vendedor quando a obra fecha apenas um escopo, mantendo outros itens do mesmo instalador ainda em aberto.

Portanto, o atendimento completo exige:

1. levar a logica de vencedor para o fluxo de edicao em `Orcamentos`;
2. refinar a regra de cancelamento para operar por `obra + instalador`;
3. introduzir um evento/aviso para o vendedor;
4. revisar os testes existentes, porque alguns validam exatamente o comportamento atual que deixara de ser verdadeiro.

## Comportamento Atual

### Fluxo De Edicao De Orcamento

Na edicao de orcamento, o frontend faz apenas `PUT /budgets/:id`:

- [BudgetEditPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/pages/BudgetEditPage.tsx)
- [updateBudgetRequest](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/api/budgets.ts#L666-L670)

No backend, a atualizacao chama:

- [Update](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go#L357-L376)

Quando o status muda, o fluxo usa:

- [buildStatusChangeParams](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go#L909-L945)
- [UpdateAndChangeStatus](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go#L1022-L1081)
- [changeBudgetStatusExecutor](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go#L1134-L1183)

Ponto importante:

- no fluxo comum de edicao, `EnforceProjectWinnerRule` nao e habilitado;
- isso significa que alterar para `Fechado` pela edicao nao cancela automaticamente outros orcamentos da obra.

Isso tambem esta refletido em testes que preservam explicitamente esse comportamento atual:

- [TestBudgetUpdateShouldNotCancelOtherProjectBudgetsWhenOneBecomesPedido](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L387-L507)

### Fluxo Atual De Eleger Vencedor

Hoje a regra de vencedor da obra esta no endpoint separado:

- [electBudgetWinnerRequest](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/api/budgets.ts#L673-L680)
- [ElectProjectWinner handler](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/handler/budget/handler.go#L173-L194)
- [ElectProjectWinner service](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go#L379-L459)
- [electProjectWinnerExecutor](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go#L1185-L1260)

Esse fluxo hoje:

- garante que o orcamento escolhido fique com status `PEDIDO` (`Fechado` na UI);
- cancela os demais orcamentos da mesma obra;
- restaura vencedores anteriores quando ha troca;
- restaura automaticamente orcamentos cancelados em troca de vencedor;
- grava historico automatico de status.

Na UI, isso fica centralizado na tela de `Obras`, com confirmacao explicita:

- [ProjectDetailPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/projects/pages/ProjectDetailPage.tsx#L1321-L1349)

### Testes Que Confirmam O Fluxo Atual

Os testes atuais mostram que:

- mudar status para `PEDIDO` via endpoint de historico/cambio de status cancela os demais orcamentos da obra;
- editar um orcamento para `PEDIDO` nao cancela os demais;
- eleger vencedor pela acao dedicada cancela os demais.

Referencias:

- [TestBudgetStatusHistoryShouldCancelOtherProjectBudgetsWhenOneBecomesPedido](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L248-L385)
- [TestBudgetUpdateShouldNotCancelOtherProjectBudgetsWhenOneBecomesPedido](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L387-L507)
- [TestBudgetElectWinnerShouldCancelOtherProjectBudgets](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L509-L620)
- [TestBudgetElectWinnerShouldReplacePreviousWinnerFromSameProject](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L622-L759)

## Gap Em Relacao Ao Caso Solicitado

O caso solicitado exige uma semantica diferente da atual.

Hoje, a regra de vencedor opera com a premissa:

- "uma obra possui um unico vencedor global"

O novo caso exige algo mais proximo de:

- "uma obra pode ter varios orcamentos do mesmo instalador ainda em aberto, representando escopos complementares"
- "ao fechar um item, somente os concorrentes de outros instaladores devem ser cancelados"

Exemplo solicitado:

- `Dutos` fechado
- `Damper` em aberto
- `Difusao` em aberto
- todos relacionados ao mesmo cliente/obra

Resultado esperado:

- orcamentos em aberto de outros instaladores na mesma obra devem ser cancelados;
- orcamentos do mesmo instalador vencedor devem permanecer ativos;
- vendedor deve receber aviso de que houve fechamento naquela obra.

## Regra De Negocio Recomendada

### Regra Principal

Ao alterar um orcamento para `Fechado` pela tela de `Orcamentos`, o sistema deve:

1. validar se o orcamento esta vinculado a uma obra;
2. validar se o orcamento possui `installer_id`;
3. manter o orcamento atual como `Fechado`;
4. cancelar apenas os demais orcamentos da mesma `project_id` cujo `installer_id` seja diferente do instalador vencedor e cujo status ainda nao seja final;
5. manter os demais orcamentos da mesma obra com o mesmo `installer_id` do vencedor sem cancelamento automatico;
6. registrar historico automatico para todos os orcamentos afetados;
7. emitir um aviso direcionado ao vendedor responsavel pela obra/orcamento.

### Regra De Preservacao

Nao cancelar automaticamente:

- orcamentos da mesma obra e do mesmo instalador;
- orcamentos ja finalizados;
- orcamentos sem obra vinculada;
- orcamentos sem instalador definido, ate que a regra seja explicitamente decidida.

### Regra De Alerta

Quando houver fechamento e o sistema detectar outros orcamentos ativos do mesmo instalador na mesma obra:

- criar um aviso para o vendedor do orcamento vencedor;
- informar que houve fechamento de um escopo da obra, mantendo outros escopos do mesmo instalador em andamento;
- permitir que o vendedor acompanhe esses itens remanescentes sem perde-los por cancelamento automatico.

## Perguntas De Negocio Que Precisam Ser Fechadas

Antes da implementacao, alguns pontos precisam ser confirmados:

1. A chave correta para preservar orcamentos complementares e apenas `installer_id`, ou tambem deve considerar `product_line_id` e/ou `system_type_id`?
2. Se houver dois orcamentos do mesmo instalador e ambos forem marcados como `Fechado`, isso e permitido?
3. Se um orcamento estiver sem `installer_id`, o que deve acontecer ao fechar?
4. O aviso deve ir apenas para o vendedor do orcamento vencedor, para todos os vendedores da obra, ou tambem para administradores?
5. O texto do aviso deve ser apenas informativo ou precisa conter link/acao sugerida?

## Recomendacao De Modelo

Com base no caso descrito, a melhor regra inicial e:

- `project_id` define a obra;
- `installer_id` define o grupo economico concorrente a ser mantido ou cancelado;
- `product_line_id` e `system_type_id` devem ser mantidos como dados informativos e, por enquanto, nao devem ser usados como chave primaria de decisao.

Motivo:

- o proprio pedido do usuario diferencia explicitamente "outros instaladores" versus "mesmo instalador";
- `Dutos`, `Damper` e `Difusao` aparecem como escopos complementares que podem coexistir;
- usar `product_line_id` ou `system_type_id` como chave central neste primeiro momento tende a complicar a regra e aumentar risco de falso positivo.

## Mudancas Necessarias No Backend

### 1. Alterar O Fluxo De Update

Hoje o `Update` comum nao aplica a regra de vencedor.

Sera necessario:

- detectar na atualizacao se o status destino e `PEDIDO`/`Fechado`;
- enriquecer `buildStatusChangeParams` para carregar a regra especial;
- permitir que `UpdateAndChangeStatus` execute a nova politica por instalador.

Arquivos mais provaveis:

- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go)
- [repository.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go)

### 2. Criar Nova Estrategia De Cancelamento

Hoje o metodo usado por vencedor cancela "todos os outros da obra".

Sera necessario substituir ou complementar isso com uma regra como:

- cancelar outros orcamentos `WHERE project_id = ? AND id <> vencedor AND installer_id <> installer_id_vencedor AND status not final`

Isso deve ficar em uma funcao dedicada, para nao quebrar o fluxo antigo sem controle.

Nome sugerido:

- `cancelOtherProjectBudgetsByDifferentInstaller(...)`

### 3. Manter Historico Coerente

Cada cancelamento automatico deve continuar gerando historico de status.

Tambem sera preciso padronizar uma nova observacao automatica, por exemplo:

- `Cancelado automaticamente porque outro instalador da obra foi definido como Fechado`

Isso e diferente da nota atual, que presume um unico vencedor global da obra.

### 4. Criar Emissao De Aviso

O projeto ja possui modulo pronto de avisos:

- [notice handler](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/handler/notice/handler.go)
- [notice service](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/notice/service.go)
- [notices migration](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/db/migrations/20260619100000_create_notices_tables.sql)

Recomendacao:

- criar um servico de dominio para `budget closing notices`;
- reutilizar `notices` para enviar aviso direcionado ao vendedor responsavel;
- opcionalmente criar uma tabela de auditoria deduplicadora, nos moldes de `delivery_alert_events`, se houver risco de reenviar o mesmo aviso varias vezes.

### 5. Decidir O Futuro Do Endpoint `elect-winner`

Existem duas alternativas:

1. manter o endpoint atual para `Obras`, mas mudar sua regra interna para tambem respeitar `installer_id`;
2. separar os conceitos:
   - `Fechado por instalador` na tela de `Orcamentos`
   - `Vencedor global da obra` na tela de `Obras`

A recomendacao deste estudo e evitar dois conceitos concorrentes para o mesmo dominio.

Melhor caminho:

- alinhar o comportamento de `Obras` e `Orcamentos` para a mesma regra por instalador;
- mudar os textos de UI para refletir essa semantica.

## Mudancas Necessarias No Frontend

### 1. Orcamentos

Na edicao de orcamento:

- quando o usuario selecionar `Fechado`, a tela deve exibir uma mensagem clara informando o efeito:
  - outros orcamentos de outros instaladores serao cancelados;
  - orcamentos do mesmo instalador permanecerao ativos;
  - um aviso sera enviado ao vendedor.

Arquivo principal:

- [BudgetForm](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/components/BudgetForm.tsx)

### 2. Obras

A tela de `Obras` hoje ainda fala em "vencedor da obra" e "os demais orcamentos ficarao cancelados":

- [ProjectDetailPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/projects/pages/ProjectDetailPage.tsx#L1321-L1349)

Esse texto precisara ser revisado para evitar contradicao com a nova regra.

### 3. Comunicacao / Avisos

O frontend ja tem a tela de `Avisos`:

- [CommunicationPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/communication/pages/CommunicationPage.tsx)

Se o backend criar o aviso corretamente, a UI existente tende a ser suficiente para a primeira versao.

## Mudancas Necessarias Em Testes

### Backend

Os testes abaixo precisarao ser ajustados ou duplicados:

- [TestBudgetUpdateShouldNotCancelOtherProjectBudgetsWhenOneBecomesPedido](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L387-L507)
- [TestBudgetElectWinnerShouldCancelOtherProjectBudgets](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L509-L620)

Novos testes recomendados:

1. fechar um orcamento via `PUT /budgets/:id` deve cancelar apenas outros instaladores da mesma obra;
2. fechar um orcamento via `PUT /budgets/:id` nao deve cancelar outro orcamento da mesma obra e do mesmo instalador;
3. fechar um orcamento sem `project_id` nao deve tentar cancelar terceiros;
4. fechar um orcamento sem `installer_id` deve seguir a regra decidida pelo negocio;
5. deve gerar historico automatico correto para cancelamentos;
6. deve criar aviso para o vendedor;
7. trocar vencedor em `Obras` deve respeitar a mesma regra por instalador.

### Frontend

Se houver cobertura de comportamento em tela:

- adicionar testes para a mensagem/alerta visual quando o status virar `Fechado`;
- validar o texto de confirmacao atualizado em `Obras`.

## Riscos E Pontos De Atencao

1. Risco de ambiguidade entre "vencedor global da obra" e "fechamento de escopo por instalador".
2. Risco de cancelar orcamentos errados se `installer_id` estiver ausente ou inconsistente.
3. Risco de regressao em `Obras`, porque hoje a tela foi desenhada assumindo cancelamento total da obra.
4. Risco de historico automatico ficar sem clareza se a nota de cancelamento nao for ajustada.
5. Risco de notificar usuario errado caso o criterio de destinatario do aviso nao seja fechado.

## Recomendacao De Implementacao Em Fases

### Fase 1 - Fechamento Das Regras

Confirmar com o negocio:

- chave de preservacao por `installer_id`;
- comportamento para ausencia de instalador;
- destinatarios do aviso;
- semantica oficial de `Fechado`.

### Fase 2 - Backend

- adaptar `Update` para disparar a nova regra ao mudar para `Fechado`;
- criar cancelamento seletivo por instalador;
- ajustar historico automatico;
- criar emissao de aviso ao vendedor.

### Fase 3 - Frontend

- ajustar texto/alerta da edicao em `Orcamentos`;
- ajustar texto da tela de `Obras`;
- validar integracao com `Avisos`.

### Fase 4 - Testes

- atualizar testes de integracao existentes;
- criar cenarios cobrindo o caso `mesmo instalador x outro instalador`;
- validar que `Dutos`, `Damper` e `Difusao` podem coexistir quando pertencem ao mesmo instalador.

## Recomendacao Final

A recomendacao tecnica e implementar o fechamento automatico na propria tela de `Orcamentos`, mas com regra por `obra + instalador`, e nao mais com cancelamento total de todos os orcamentos da obra.

Essa abordagem atende melhor o caso real informado:

- reduz operacao manual;
- preserva escopos complementares;
- continua eliminando concorrencia aberta de outros instaladores;
- permite avisar o vendedor de forma proativa.

## Arquivos Mais Relevantes Para A Futura Implementacao

### Backend

- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go)
- [repository.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go)
- [budget_flow_test.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go)
- [notice service](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/notice/service.go)

### Frontend

- [BudgetEditPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/pages/BudgetEditPage.tsx)
- [BudgetForm](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/components/BudgetForm.tsx)
- [ProjectDetailPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/projects/pages/ProjectDetailPage.tsx)
- [CommunicationPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/communication/pages/CommunicationPage.tsx)
