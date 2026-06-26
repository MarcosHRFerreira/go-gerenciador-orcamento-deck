# Plano Tecnico Executavel: Fechamento Automatico Em Orcamentos Sem Cancelar Escopos Complementares

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

## Plano Tecnico Executavel Por Fases

## Visao Geral De Execucao

Objetivo da implementacao:

- permitir que a alteracao para `Fechado` na tela de `Orcamentos` execute a regra automatica de fechamento da obra;
- cancelar somente os orcamentos concorrentes de outros instaladores na mesma obra;
- preservar os demais orcamentos do mesmo instalador;
- registrar historico coerente;
- enviar aviso ao vendedor;
- alinhar a semantica entre `Orcamentos` e `Obras`.

Estrategia recomendada de entrega:

- primeiro fechar as decisoes de negocio;
- depois implementar o nucleo de regra no backend;
- em seguida acoplar aviso e ajustes de UI;
- por fim validar regressao e homologacao.

Sequencia recomendada de commits/PRs:

1. backend: regra de fechamento por instalador;
2. backend: avisos e historico refinado;
3. frontend: mensagens e alinhamento de telas;
4. testes e ajustes finais.

## Fase 0 - Fechamento Funcional

### Objetivo

Eliminar ambiguidades antes de alterar regra de negocio sensivel.

### Entradas

- este documento;
- exemplos reais de obras com multiplos orcamentos;
- validacao do usuario de negocio.

### Decisoes Que Precisam Ser Confirmadas

1. A chave de preservacao sera somente `installer_id`.
2. Orcamentos sem `installer_id` nao sofrerao cancelamento automatico.
3. O aviso sera enviado para `vendedor + admin`.
4. A tela de `Obras` passara a seguir a mesma semantica de `Orcamentos`.
5. E permitido haver mais de um orcamento `Fechado` para a mesma obra desde que sejam do mesmo instalador e representem escopos complementares.
6. O aviso da primeira versao sera apenas informativo, sem link ou acao adicional.

### Tarefas

- registrar resposta para cada decisao acima;
- transformar respostas em regras finais do dominio;
- revisar texto de negocio para `Fechado`, `Cancelado` e aviso ao vendedor.

### Entregaveis

- checklist de decisoes respondidas;
- texto final da regra funcional;
- mensagem padrao de historico automatico;
- mensagem padrao do aviso.

### Decisoes Consolidadas Da Fase 0

- grupo preservado: somente `installer_id`
- sem `installer_id`: manter apenas o orcamento atual como `Fechado`, sem cancelar terceiros
- destinatarios do aviso: `vendedor + admin`
- mesma semantica para `Orcamentos` e `Obras`
- multiplos `Fechado` do mesmo instalador na mesma obra: permitido
- aviso da primeira versao: apenas informativo

### Criterios De Aceite

- nao restar decisao aberta sobre `installer_id`, aviso e comportamento sem instalador;
- negocio aprovar exemplos concretos de `Dutos`, `Damper` e `Difusao`.

## Fase 1 - Backend Nucleo Da Regra

### Objetivo

Fazer o `PUT /budgets/:id` aplicar a nova regra quando o status for alterado para `Fechado`.

### Escopo Tecnico

- detectar transicao para `PEDIDO`/`Fechado` no fluxo de `Update`;
- substituir a logica de cancelamento total por cancelamento seletivo por instalador;
- preservar orcamentos ativos do mesmo instalador;
- manter o comportamento atual para atualizacoes que nao mudem para `Fechado`.

### Tarefas Tecnicas

1. Revisar [Update](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go) para identificar o ponto exato da transicao de status.
2. Evoluir `buildStatusChangeParams` para carregar contexto suficiente da obra e do instalador vencedor.
3. Criar funcao dedicada para cancelar apenas orcamentos da mesma obra com `installer_id` diferente.
4. Garantir que orcamentos ja finalizados nao sejam alterados indevidamente.
5. Garantir que orcamentos sem `project_id` ou sem `installer_id` sigam a regra decidida na Fase 0.
6. Manter compatibilidade com o fluxo atual de historico de status.

### Sequencia Tecnica Recomendada

1. Ajustar [Update](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go#L357-L376) para distinguir:
   - atualizacao comum sem troca de status;
   - troca de status comum;
   - troca para `PEDIDO`/`Fechado` com regra especial por instalador.
2. Evoluir [buildStatusChangeParams](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go#L909-L945) para carregar:
   - `project_id` do orcamento atual;
   - `installer_id` do orcamento atual;
   - `is_final` do status de destino;
   - `cancelled_status_id` necessario para cancelamentos automaticos;
   - flag explicita da nova politica por instalador.
3. Ampliar `ChangeStatusParams` no repositório para comportar a nova regra, sem quebrar o fluxo legado de historico.
4. Alterar [changeBudgetStatusExecutor](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go#L1134-L1183) para:
   - nao depender apenas de `EnforceProjectWinnerRule`;
   - aplicar cancelamento seletivo quando o destino for `Fechado` e houver `project_id + installer_id`;
   - pular cancelamento quando faltar `project_id` ou `installer_id`.
5. Criar uma nova funcao de repositório, separada da atual `cancelOtherProjectBudgets(...)`, por exemplo:
   - `cancelOtherProjectBudgetsByDifferentInstaller(...)`
6. Fazer a nova funcao cancelar apenas orcamentos da mesma obra:
   - com `installer_id` diferente;
   - com status ainda nao final;
   - excluindo o proprio orcamento fechado.
7. Manter os orcamentos da mesma obra e do mesmo instalador intactos, inclusive se ainda estiverem em aberto.
8. Revisar o fluxo legado de [ElectProjectWinner](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go#L379-L459) para futura convergencia com a mesma regra, sem misturar isso na primeira entrega da Fase 1 se aumentar o risco.

### Contrato Tecnico Sugerido Para A Fase 1

O objetivo da Fase 1 nao e criar novo endpoint, e sim reutilizar o `PUT /budgets/:id`.

Comportamento esperado do backend nessa fase:

- se o status nao mudar, `Update()` continua igual;
- se o status mudar para algo diferente de `Fechado`, o fluxo continua igual;
- se o status mudar para `Fechado`:
  - salva a alteracao do orcamento;
  - registra historico do proprio orcamento;
  - cancela automaticamente apenas concorrentes de outros instaladores da mesma obra;
  - registra historico automatico dos cancelados.

### Regras Operacionais Da Fase 1

- com `project_id` e `installer_id`: aplica cancelamento seletivo
- sem `project_id`: nao cancela ninguem
- sem `installer_id`: nao cancela ninguem
- com outro orcamento ja `Fechado` do mesmo instalador: permitido
- com outro orcamento ja `Fechado` de outro instalador: a Fase 1 deve seguir a nova regra por instalador, nao a restricao antiga de vencedor global

### Arquivos Principais

- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go)
- [repository.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go)

### Pontos Exatos Do Codigo Mapeados

- entrada da troca de status: [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go#L357-L376)
- montagem dos parametros da troca: [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go#L909-L945)
- update transacional com historico: [repository.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go#L1022-L1048)
- executor atual da troca de status: [repository.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go#L1134-L1183)
- fluxo legado de eleger vencedor: [repository.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/budget/repository.go#L1185-L1260)

### Entregaveis

- regra de update adaptada para `Fechado`;
- metodo de cancelamento seletivo por instalador;
- observacao automatica de cancelamento revisada;
- preservacao dos orcamentos do mesmo instalador validada.

### Criterios De Aceite

- ao fechar um orcamento, os concorrentes de outros instaladores da mesma obra sao cancelados;
- orcamentos do mesmo instalador permanecem ativos;
- orcamentos sem obra ou sem instalador nao geram efeito colateral fora da regra aprovada;
- historico permanece consistente para todos os registros afetados.

### Testes Minimos Da Fase

1. `PUT /budgets/:id` fechando orcamento cancela apenas outros instaladores.
2. `PUT /budgets/:id` fechando orcamento nao cancela itens do mesmo instalador.
3. `PUT /budgets/:id` sem `project_id` nao cancela terceiros.
4. `PUT /budgets/:id` sem `installer_id` segue a regra fechada na Fase 0.

### Sequencia De Testes Recomendada

1. Ajustar primeiro o teste que hoje garante o comportamento antigo de nao cancelar ningem ao editar para `PEDIDO`.
2. Criar um cenario novo com:
   - obra unica;
   - dois orcamentos do mesmo instalador;
   - dois orcamentos de instaladores concorrentes.
3. Validar que, ao fechar um dos orcamentos:
   - o segundo do mesmo instalador permanece;
   - os de outros instaladores sao cancelados.
4. Validar cenario sem `installer_id`.
5. Validar cenario sem `project_id`.

### Risco Principal Da Fase 1

O maior risco desta fase e misturar a nova regra por instalador com a regra antiga de vencedor global ainda embutida no fluxo `elect-winner`.

Mitigacao recomendada:

- implementar primeiro a nova regra somente no `PUT /budgets/:id`;
- manter o endpoint legado isolado nesta fase;
- alinhar `Obras` na fase seguinte ou em uma etapa controlada, caso a convergencia imediata gere regressao desnecessaria.

## Fase 2 - Avisos E Auditoria

### Objetivo

Informar o vendedor que houve fechamento de escopo na obra e manter rastreabilidade da automacao.

### Escopo Tecnico

- criar emissao de aviso no backend ao final do fechamento bem-sucedido;
- reutilizar a infraestrutura existente de `notices`;
- evitar notificacao duplicada no mesmo evento funcional, se necessario.

### Tarefas Tecnicas

1. Mapear o ponto ideal do fluxo para emissao do aviso, preferencialmente apos consolidacao da transacao principal.
2. Definir payload do aviso com obra, orcamento fechado e contexto resumido.
3. Implementar servico de dominio ou helper de notificacao para fechamento de orcamento.
4. Definir se sera necessario mecanismo de deduplicacao.
5. Garantir que falha no aviso nao corrompa a consistencia da alteracao principal, caso essa seja a regra desejada.

### Sequencia Tecnica Recomendada

1. Nao usar diretamente [noticeService.Create](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/notice/service.go#L38-L93) para a automacao, porque esse fluxo exige autor autenticado com papel `admin` e foi desenhado para criacao manual via API.
2. Reaproveitar o padrao ja usado em [deliveryalert/service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/deliveryalert/service.go#L156-L183), que cria avisos automaticos direto pelo `noticeRepo`.
3. Criar um servico dedicado, por exemplo:
   - `budgetclosingnotice.Service`
   - ou um helper interno no dominio de `budget`
4. Esse servico deve:
   - resolver o `admin` autor tecnico do aviso;
   - montar lista de destinatarios `vendedor + admins`;
   - remover duplicidade de IDs antes de persistir;
   - criar aviso com `scope_type = users`;
   - usar prioridade inicial `info` ou `warning`, conforme definicao final.
5. Disparar o aviso somente depois que a Fase 1 concluir a alteracao principal com sucesso.
6. Nao acoplar a emissao do aviso na mesma transacao da troca de status nesta primeira versao, para evitar rollback do fechamento por falha de notificacao.

### Contrato Tecnico Sugerido Para A Fase 2

Comportamento esperado:

- ao concluir um fechamento com efeito automatico na obra, o backend gera um aviso informativo;
- o aviso vai para:
  - vendedor responsavel pelo orcamento fechado;
  - todos os usuarios `admin` ativos;
- o aviso nao depende de chamada HTTP adicional;
- a operacao principal de fechamento permanece bem-sucedida mesmo se a notificacao falhar, desde que o erro seja logado adequadamente.

### Payload Funcional Sugerido Do Aviso

Titulo sugerido:

- `Fechamento automatico aplicado na obra do orcamento {budget_number}`

Corpo sugerido:

- informar o numero do orcamento fechado;
- identificar a obra por codigo/nome quando existir;
- informar o instalador vencedor do fechamento;
- informar que concorrentes de outros instaladores foram cancelados automaticamente;
- informar que escopos do mesmo instalador podem continuar ativos.

Exemplo de texto:

- `Aviso automatico do sistema: o orcamento 12345 foi marcado como Fechado na obra ABC - Cliente X. Orcamentos em aberto de outros instaladores foram cancelados automaticamente. Orcamentos do mesmo instalador podem permanecer ativos por representarem escopos complementares da obra.`

### Estrategia De Destinatarios

Regra recomendada para a primeira versao:

- destinatario primario: vendedor do orcamento fechado;
- destinatarios secundarios: todos os administradores ativos;
- nao enviar para todos os usuarios da obra;
- nao enviar para orcamentistas na primeira versao.

### Estrategia De Autor Tecnico

Como o modulo de avisos exige `created_by_user_id`, a recomendacao e repetir a mesma estrategia do alerta de entrega:

- tentar usar um `admin` configurado como autor tecnico;
- se nao houver configuracao, usar o primeiro `admin` ativo disponivel;
- se nao existir `admin` ativo, registrar erro de infraestrutura e nao bloquear o fechamento principal.

### Estrategia De Duplicidade

Para a primeira versao, a recomendacao e nao criar tabela nova de deduplicacao ainda.

Motivo:

- o gatilho de emissao sera uma transicao de status bem definida;
- a chance de duplicidade acidental e menor do que em jobs recorrentes;
- o custo de complexidade de uma tabela de eventos agora pode ser evitado.

Mitigacao minima:

- emitir aviso apenas quando houver transicao real para `Fechado`;
- nao emitir aviso quando o orcamento ja estiver `Fechado` e sofrer apenas nova edicao;
- garantir que a lista final de destinatarios esteja sem IDs repetidos.

### Arquivos Principais

- [notice service](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/notice/service.go)
- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budget/service.go)

### Pontos Exatos Do Codigo Mapeados

- servico manual de avisos: [notice/service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/notice/service.go#L38-L93)
- restricao de criacao manual para admin: [notice/handler.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/handler/notice/handler.go#L41-L45)
- persistencia de avisos e destinatarios: [notice/repository.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/repository/notice/repository.go#L35-L65)
- modelo de dados `notices` e `notice_recipients`: [20260619100000_create_notices_tables.sql](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/db/migrations/20260619100000_create_notices_tables.sql#L1-L47)
- referencia de implementacao automatica: [deliveryalert/service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/deliveryalert/service.go#L156-L203)

### Entregaveis

- aviso funcional ao vendedor;
- aviso funcional aos admins;
- texto padronizado do aviso;
- registro de auditoria suficiente para suporte e depuracao.

### Criterios De Aceite

- ao fechar um orcamento com sucesso, o vendedor e os admins recebem aviso visivel no modulo de `Avisos`;
- o aviso identifica a obra e informa que outros escopos do mesmo instalador podem permanecer em aberto;
- repeticoes indevidas do mesmo aviso nao ocorrem no mesmo fluxo.
- falha na criacao do aviso nao desfaz o fechamento principal da obra.

### Testes Minimos Da Fase

1. fechamento gera aviso para o vendedor correto;
2. fechamento gera aviso tambem para os admins ativos;
3. aviso nao e enviado para usuario indevido;
4. fluxo nao cria duplicidade indevida no mesmo evento;
5. falha de aviso nao desfaz a alteracao principal.

### Sequencia De Testes Recomendada

1. Criar stub do `noticeRepo` ou servico dedicado para validar destinatarios.
2. Validar que o vendedor do orcamento fechado esta na lista final.
3. Validar que os admins ativos entram na lista final.
4. Validar que o vendedor nao recebe em duplicidade quando tambem for admin.
5. Validar que nao ha emissao quando a transicao real para `Fechado` nao ocorrer.
6. Validar politica de erro: notificacao falha, fechamento permanece.

### Risco Principal Da Fase 2

O maior risco desta fase e acoplar demais a automacao de aviso ao fluxo principal de fechamento, tornando a operacao de negocio dependente de infraestrutura secundaria.

Mitigacao recomendada:

- isolar a criacao do aviso em um servico proprio;
- executar o aviso somente depois da conclusao do fechamento;
- registrar falhas em log estruturado sem reverter a alteracao principal;
- reutilizar o padrao ja estabilizado do modulo de alertas de entrega.

## Fase 3 - Frontend E Alinhamento De UX

### Objetivo

Explicar o novo comportamento ao usuario e alinhar `Orcamentos`, `Obras` e `Avisos`.

### Escopo Tecnico

- ajustar mensagem da edicao de orcamento ao selecionar `Fechado`;
- revisar textos da tela de `Obras` para a nova semantica;
- validar exibicao do aviso no frontend existente.

### Tarefas Tecnicas

1. Em [BudgetForm](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/components/BudgetForm.tsx), adicionar mensagem contextual quando o status for `Fechado`.
2. Em [ProjectDetailPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/projects/pages/ProjectDetailPage.tsx), ajustar texto de confirmacao para refletir regra por instalador.
3. Validar se [CommunicationPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/communication/pages/CommunicationPage.tsx) ja exibe o aviso sem mudancas adicionais.
4. Se necessario, incluir texto de apoio ou tooltip para reduzir erro operacional.

### Sequencia Tecnica Recomendada

1. Em [BudgetForm](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/components/BudgetForm.tsx), reaproveitar o padrao ja existente de `Alert` acima do formulario para exibir a nova mensagem de impacto do status `Fechado`.
2. Usar o `selectedStatusId` e o `selectedStatus` que o componente ja calcula para detectar quando o usuario escolheu `Fechado`.
3. Mostrar mensagem somente no contexto relevante:
   - preferencialmente quando houver `projectId`;
   - com texto alternativo mais neutro quando faltar `projectId` ou `installerId`.
4. Revisar em [ProjectDetailPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/projects/pages/ProjectDetailPage.tsx#L1321-L1378) o texto do dialogo de eleicao de vencedor para remover a promessa de que "todos os demais orcamentos" serao cancelados.
5. Trocar a linguagem de "vencedor global da obra" por texto aderente a regra nova, sem quebrar a compreensao do usuario final.
6. Validar a exibicao dos avisos na [CommunicationPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/communication/pages/CommunicationPage.tsx), sem criar nova UX nesta fase, porque a tela atual ja lista avisos com prioridade, titulo e corpo.

### Contrato De UX Sugerido Para A Fase 3

Na tela de `Orcamentos`, ao selecionar `Fechado`, a UI deve comunicar que:

- o proprio orcamento sera marcado como `Fechado`;
- orcamentos em aberto de outros instaladores da mesma obra podem ser cancelados automaticamente;
- orcamentos do mesmo instalador podem permanecer ativos por representarem escopos complementares;
- um aviso informativo sera enviado para `vendedor + admin`.

Na tela de `Obras`, a confirmacao deve comunicar que:

- a acao aplica a mesma regra da tela de `Orcamentos`;
- concorrentes de outros instaladores podem ser cancelados;
- orcamentos do mesmo instalador podem permanecer ativos.

Na tela de `Avisos`, nao e necessario novo layout nesta fase:

- basta garantir que o titulo e o corpo gerados no backend sejam claros o suficiente para leitura imediata.

### Textos Sugeridos

Mensagem de alerta em `Orcamentos`:

- `Ao marcar este orçamento como Fechado, o sistema poderá cancelar automaticamente os orçamentos em aberto de outros instaladores da mesma obra. Orçamentos do mesmo instalador podem permanecer ativos por representarem escopos complementares. Um aviso informativo será enviado ao vendedor e aos administradores.`

Mensagem alternativa quando faltar obra ou instalador:

- `Este orçamento será marcado como Fechado. Como a obra ou o instalador não estão totalmente definidos, o sistema não aplicará cancelamento automático em outros orçamentos.`

Mensagem sugerida para o dialogo de `Obras`:

- `Ao confirmar, este orçamento será definido como Fechado para a obra. Orçamentos em aberto de outros instaladores poderão ser alterados para Cancelado automaticamente. Orçamentos do mesmo instalador podem permanecer ativos quando representarem escopos complementares.`

### Estrategia De UX

Recomendacao para a primeira versao:

- usar apenas `Alert` informativo;
- nao adicionar modal extra na edicao de `Orcamentos`;
- nao exigir checkbox de confirmacao;
- nao criar novo centro de notificacoes no frontend.

Motivo:

- a regra ja e operacionalmente importante, mas ainda precisa ser clara e leve;
- um `Alert` contextual resolve a comunicacao sem friccao extra;
- o modulo de `Avisos` ja cobre a comunicacao posterior.

### Entregaveis

- mensagem explicativa na edicao de orcamento;
- tela de `Obras` semanticamente alinhada;
- visibilidade do aviso ao vendedor.

### Pontos Exatos Do Codigo Mapeados

- deteccao de status selecionado em [BudgetForm](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/components/BudgetForm.tsx#L702-L707)
- area atual de `Alert` no formulario em [BudgetForm](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/components/BudgetForm.tsx#L779-L830)
- seletor de status em [BudgetForm](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/components/BudgetForm.tsx#L919-L933)
- dialogo de vencedor em [ProjectDetailPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/projects/pages/ProjectDetailPage.tsx#L1321-L1378)
- listagem de avisos em [CommunicationPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/communication/pages/CommunicationPage.tsx)

### Criterios De Aceite

- o usuario entende, na propria tela de edicao, que apenas outros instaladores serao cancelados;
- a tela de `Obras` nao contradiz a nova regra;
- o aviso pode ser lido no modulo existente.
- a UX nao introduz passos extras desnecessarios para salvar o orcamento.

### Testes Minimos Da Fase

1. mensagem visual aparece ao selecionar `Fechado`;
2. texto de confirmacao em `Obras` fica coerente com a regra nova;
3. fluxo visual de aviso e validado manualmente.

### Sequencia De Testes Recomendada

1. Validar manualmente a edicao de um orcamento com status diferente de `Fechado` para garantir ausencia de ruido visual.
2. Selecionar `Fechado` em orcamento com `projectId` e `installerId` preenchidos e verificar a mensagem completa.
3. Selecionar `Fechado` em orcamento sem `projectId` ou sem `installerId` e verificar a mensagem alternativa.
4. Abrir o dialogo da tela de `Obras` e validar que o texto nao promete mais cancelamento total da obra.
5. Validar recebimento do aviso na `CommunicationPage` com titulo e corpo legiveis.

### Risco Principal Da Fase 3

O maior risco desta fase e a UX continuar comunicando a regra antiga mesmo com backend novo, gerando operacao incorreta por expectativa errada.

Mitigacao recomendada:

- atualizar primeiro os textos de `Orcamentos` e `Obras`;
- evitar linguagem ambigua como "todos os demais orcamentos";
- homologar manualmente com usuarios que conhecem o processo real.

## Fase 4 - Testes, Regressao E Homologacao

### Objetivo

Garantir que a nova regra nao quebre fluxos existentes e que o caso de negocio esteja coberto ponta a ponta.

### Escopo Tecnico

- atualizar testes que codificam a regra antiga;
- criar novos cenarios de integracao;
- executar validacao manual com exemplos reais.

### Tarefas Tecnicas

1. Revisar e ajustar [budget_flow_test.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go).
2. Criar cenarios cobrindo:
   - mesmo instalador com multiplos escopos;
   - outros instaladores concorrentes;
   - ausencia de `project_id`;
   - ausencia de `installer_id`;
   - emissao de aviso.
3. Executar teste manual com obra de exemplo contendo `Dutos`, `Damper` e `Difusao`.
4. Validar tambem o fluxo legado da tela de `Obras`.

### Sequencia Tecnica Recomendada

1. Atualizar primeiro os testes de integracao que hoje codificam a regra antiga.
2. Manter os testes antigos apenas quando ainda cobrirem comportamento valido; caso contrario, renomear e reescrever para a nova semantica.
3. Criar novos testes de integracao backend antes de ajustar o fluxo legado de `Obras`, para evitar regressao invisivel.
4. Reaproveitar a infraestrutura existente de `notice_test.go` para validar a presenca do aviso nos destinatarios corretos.
5. Finalizar com checklist manual de homologacao ponta a ponta, porque a UX do frontend nao aparenta ter uma cobertura de testes automatizados consolidada nesta area.

### Estrategia De Testes Automatizados

Prioridade recomendada:

1. testes de integracao backend do fluxo de status;
2. testes de integracao backend do fluxo de aviso;
3. validacao manual do frontend;
4. testes de frontend apenas se houver infraestrutura proxima que agregue valor sem custo alto.

Justificativa:

- a regra critica mora no backend;
- os comportamentos atuais que vao mudar ja estao refletidos em `budget_flow_test.go`;
- o frontend desta area nao demonstra uma malha de testes dedicada que justifique criar grande volume de testes novos agora;
- a comunicacao visual e melhor validada com checklist manual controlado nesta entrega.

### Casos Automatizados Obrigatorios

#### Backend - Fluxo De Update

1. `PUT /budgets/:id` para `Fechado` cancela apenas orcamentos de outros instaladores da mesma obra.
2. `PUT /budgets/:id` para `Fechado` preserva outros orcamentos do mesmo instalador.
3. `PUT /budgets/:id` para `Fechado` sem `project_id` nao cancela terceiros.
4. `PUT /budgets/:id` para `Fechado` sem `installer_id` nao cancela terceiros.
5. `PUT /budgets/:id` sem mudanca real para `Fechado` nao dispara aviso.

#### Backend - Historico

1. o orcamento fechado gera historico proprio correto;
2. os orcamentos cancelados automaticamente geram historico com a nova observacao;
3. orcamentos do mesmo instalador que permanecem ativos nao recebem historico indevido.

#### Backend - Avisos

1. o vendedor responsavel recebe aviso;
2. administradores ativos recebem aviso;
3. vendedor que tambem for admin nao recebe duplicado;
4. usuario nao destinatario nao recebe aviso;
5. falha no aviso nao desfaz o fechamento principal, conforme politica definida.

#### Backend - Fluxo De Obras

1. o fluxo legado de `elect-winner` deve ser reavaliado conforme a estrategia da entrega;
2. se ele for convergido nesta mesma liberacao, precisa seguir a mesma regra por instalador;
3. se ele ficar temporariamente isolado, isso deve estar explicitamente validado e documentado para nao gerar expectativa errada.

### Testes Existentes Que Precisam Ser Revisados

Arquivo principal:

- [budget_flow_test.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go)

Casos mapeados:

- [TestBudgetStatusHistoryShouldCancelOtherProjectBudgetsWhenOneBecomesPedido](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L248-L385)
- [TestBudgetUpdateShouldNotCancelOtherProjectBudgetsWhenOneBecomesPedido](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L387-L507)
- [TestBudgetElectWinnerShouldCancelOtherProjectBudgets](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L509-L620)
- [TestBudgetElectWinnerShouldReplacePreviousWinnerFromSameProject](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go#L622-L759)

Arquivo de apoio para avisos:

- [notice_test.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/notice_test.go)

### Checklist De Homologacao Manual

#### Cenario 1 - Mesmo Instalador Com Escopos Complementares

- criar ou usar uma obra com `Dutos`, `Damper` e `Difusao` no mesmo `installer_id`;
- fechar apenas um dos orcamentos;
- validar que os outros do mesmo instalador permanecem ativos;
- validar que nenhum deles foi cancelado indevidamente.

#### Cenario 2 - Instaladores Concorrentes

- criar ou usar uma obra com ao menos dois instaladores diferentes;
- fechar um orcamento de um instalador;
- validar que orcamentos em aberto dos outros instaladores foram cancelados;
- validar historico dos cancelados.

#### Cenario 3 - Sem Instalador

- fechar um orcamento sem `installer_id`;
- validar que o proprio orcamento foi salvo como `Fechado`;
- validar que nenhum terceiro foi cancelado.

#### Cenario 4 - Sem Obra

- fechar um orcamento sem `project_id`;
- validar que nao houve efeito colateral em outros orcamentos.

#### Cenario 5 - Avisos

- validar que o vendedor recebe o aviso;
- validar que administradores recebem o aviso;
- validar legibilidade do titulo e corpo na [CommunicationPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/communication/pages/CommunicationPage.tsx).

#### Cenario 6 - Tela De Obras

- abrir a confirmacao da tela de `Obras`;
- validar que o texto comunica a mesma semantica da tela de `Orcamentos`;
- se o fluxo de `Obras` estiver convergido na entrega, validar o comportamento real ponta a ponta.

### Entregaveis

- testes de integracao atualizados;
- evidencias de homologacao manual;
- checklist de regressao concluido.

### Pontos Exatos Do Codigo Mapeados

- suite principal de integracao do dominio: [budget_flow_test.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/budget_flow_test.go)
- suite de integracao do modulo de avisos: [notice_test.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/test/integration/notice_test.go)
- tela de `Orcamentos` para homologacao: [BudgetForm](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/budgets/components/BudgetForm.tsx)
- tela de `Obras` para homologacao: [ProjectDetailPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/projects/pages/ProjectDetailPage.tsx)
- tela de `Avisos` para homologacao: [CommunicationPage](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/communication/pages/CommunicationPage.tsx)

### Criterios De Aceite

- todos os testes automatizados relevantes passam;
- o caso de negocio principal funciona do inicio ao fim;
- nao ha cancelamento indevido de escopos do mesmo instalador;
- `Obras` e `Orcamentos` permanecem coerentes entre si.
- a trilha de historico e aviso fica consistente para suporte e auditoria.

### Sequencia De Testes Recomendada

1. Ajustar primeiro os testes de integracao backend afetados pela regra antiga.
2. Criar os novos cenarios de mesma obra com mesmo instalador e com outros instaladores.
3. Cobrir avisos com base na infraestrutura ja validada em `notice_test.go`.
4. Executar `go test` nas suites relevantes.
5. Executar `yarn build` no frontend.
6. Rodar checklist manual de homologacao nos seis cenarios acima.

### Risco Principal Da Fase 4

O maior risco desta fase e liberar uma combinacao parcialmente atualizada, em que backend, mensagens de UI e homologacao manual apontem para semanticas diferentes.

Mitigacao recomendada:

- tratar a Fase 4 como gate de liberacao;
- nao considerar a entrega concluida apenas por testes unitarios;
- exigir validacao manual real em `Orcamentos`, `Obras` e `Avisos`;
- registrar evidencias minimas da homologacao para futura referencia.

## Checklist De Execucao

### Antes De Comecar

- validar decisoes da Fase 0;
- separar exemplos reais para homologacao;
- alinhar mensagem funcional do aviso.

### Antes De Subir Backend

- regra de cancelamento seletivo coberta por testes;
- historico automatico revisado;
- comportamento sem instalador decidido e implementado.

### Antes De Subir Frontend

- backend estabilizado e homologado;
- contrato do aviso confirmado;
- textos aprovados pelo negocio.

### Antes De Liberar

- `yarn build` no frontend;
- testes unitarios/integracao relevantes no backend;
- validacao manual completa com obra real;
- conferencia do modulo de `Avisos`.

## Definicao De Pronto

Esta implementacao so deve ser considerada concluida quando:

1. editar um orcamento para `Fechado` produzir o mesmo efeito funcional esperado sem precisar entrar em `Obras`;
2. somente orcamentos de outros instaladores forem cancelados;
3. orcamentos do mesmo instalador permanecerem disponiveis;
4. historicos automaticos ficarem corretos;
5. vendedor receber aviso;
6. telas de `Orcamentos`, `Obras` e `Avisos` ficarem semanticamente alinhadas;
7. testes automatizados e homologacao manual forem aprovados.

## Recomendacao Final

A recomendacao tecnica e executar a implementacao em quatro fases tecnicas, com a Fase 0 funcionando como gate obrigatorio de definicao funcional.

Melhor ordem pratica:

1. fechar regra funcional;
2. entregar backend do nucleo;
3. entregar aviso;
4. alinhar frontend;
5. concluir testes e homologacao.

Essa sequencia minimiza retrabalho, controla risco de regressao e permite validar o comportamento critico no backend antes de investir em polimento visual.

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
