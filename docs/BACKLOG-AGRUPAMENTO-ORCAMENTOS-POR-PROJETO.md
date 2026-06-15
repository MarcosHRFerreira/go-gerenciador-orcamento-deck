# Backlog De Implementacao Do Agrupamento De Orcamentos Por Projeto

## Objetivo

Este backlog organiza a implementacao da funcionalidade de agrupamento de orcamentos por projeto, com base no estudo registrado em `docs/ESTUDO-AGRUPAMENTO-ORCAMENTOS-POR-PROJETO.md`.

O objetivo e:

- quebrar a entrega em tarefas executaveis
- definir prioridade e dependencias
- reduzir risco de implementacao dispersa
- servir como guia para backend, frontend e testes

## Escopo Da Funcionalidade

Esta entrega cobre:

- associar varios orcamentos a um mesmo projeto
- visualizar os orcamentos vinculados ao projeto
- impedir mais de um `PEDIDO` por projeto
- ao definir um orcamento do projeto como `PEDIDO`, cancelar automaticamente os demais orcamentos ativos do mesmo grupo
- destacar no sistema que os itens cancelados por essa regra nao exigem mais atencao

## Premissas

- o agrupamento sera feito usando o `project_id` ja existente em `budgets`
- nao sera criada uma nova entidade de agrupamento nesta primeira versao
- a regra ficara concentrada no backend
- a acao que gera cancelamento automatico deve ser transacional
- o historico de status precisa refletir as mudancas feitas automaticamente

## Escala De Prioridade

- `P0`: bloqueante ou fundacional
- `P1`: necessario para o fluxo principal
- `P2`: importante para consolidacao e experiencia
- `P3`: melhoria futura

## Criterio De Uso

Cada task abaixo possui:

- `ID`: identificador da tarefa
- `Prioridade`: urgencia sugerida
- `Dependencias`: o que precisa vir antes
- `Descricao`: o que deve ser implementado
- `Pronto quando`: criterio objetivo de conclusao

## Ordem Macro Recomendada

1. alinhamento funcional da regra
2. suporte de dados e repository
3. regra transacional no backend
4. historico e auditoria
5. API de consulta do grupo
6. tela de projeto com orcamentos vinculados
7. refinamentos de UX
8. testes automatizados e validacao final

## Fase 1 - Alinhamento Funcional

### TASK-GRP-001 - Confirmar semantica do `project_id`

- `Prioridade`: `P0`
- `Dependencias`: nenhuma
- `Descricao`: validar com o negocio que o cadastro atual de `projects` representa corretamente a obra ou oportunidade que deve agrupar varios orcamentos.
- `Pronto quando`: houver decisao explicita de reutilizar `project_id` atual como agrupador na primeira versao.

### TASK-GRP-002 - Definir statuses concorrentes ativos

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-001`
- `Descricao`: listar quais statuses podem ser cancelados automaticamente quando outro orcamento do mesmo projeto virar `PEDIDO`.
- `Pronto quando`: existir uma regra fechada dizendo exatamente quais statuses entram na automacao.

### TASK-GRP-003 - Definir politica de reabertura

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-001`
- `Descricao`: decidir se um orcamento cancelado pelo grupo pode voltar a ficar ativo e em quais condicoes.
- `Pronto quando`: existir uma regra funcional documentada para reabertura ou bloqueio.

## Fase 2 - Estrutura De Dados

### TASK-GRP-004 - Revisar modelagem atual de `budgets.project_id`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-001`
- `Descricao`: verificar se a tabela `budgets` e os repositories atuais suportam sem ajustes a consulta de varios orcamentos por projeto.
- `Pronto quando`: houver confirmacao tecnica do que ja existe e do que falta adicionar.

### TASK-GRP-005 - Adicionar suporte de auditoria para cancelamento automatico

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-002`
- `Descricao`: avaliar e implementar campo auxiliar em `budgets`, como `canceled_by_group_rule`, e eventualmente `cancellation_reason`, para distinguir cancelamento manual de cancelamento automatico.
- `Pronto quando`: o banco conseguir diferenciar orcamentos cancelados por regra do grupo.

### TASK-GRP-006 - Criar migration da auditoria do grupo

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-005`
- `Descricao`: criar e versionar migration para os campos adicionais decididos na task anterior.
- `Pronto quando`: a migration existir e puder ser aplicada localmente sem quebrar dados existentes.

## Fase 3 - Repository E Consultas

### TASK-GRP-007 - Listar orcamentos por projeto

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-004`
- `Descricao`: adicionar no repository de `budgets` consulta para listar todos os orcamentos de um `project_id`, com os dados necessarios para tela e regra de negocio.
- `Pronto quando`: o backend conseguir recuperar o grupo completo de orcamentos do projeto.

### TASK-GRP-008 - Localizar pedido existente no projeto

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-002`, `TASK-GRP-007`
- `Descricao`: adicionar consulta que localiza se ja existe outro orcamento com status `PEDIDO` no mesmo projeto.
- `Pronto quando`: a camada de servico puder bloquear mais de um `PEDIDO` por projeto.

### TASK-GRP-009 - Atualizar status dos demais orcamentos em lote

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-002`, `TASK-GRP-007`
- `Descricao`: adicionar operacao de repository para cancelar, em lote, os demais orcamentos ativos do mesmo projeto, exceto o vencedor.
- `Pronto quando`: existir uma operacao unica e segura para cancelar os concorrentes do grupo.

## Fase 4 - Regra De Negocio No Backend

### TASK-GRP-010 - Implementar regra transacional de vencedor do projeto

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-008`, `TASK-GRP-009`
- `Descricao`: no service de `budgets`, ao alterar um orcamento para `PEDIDO`, validar concorrencia, impedir duplicidade de vencedor e cancelar os demais em transacao.
- `Pronto quando`: a mudanca para `PEDIDO` aplicar a regra automaticamente sem inconsistencias.

### TASK-GRP-011 - Bloquear segundo `PEDIDO` no mesmo projeto

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-010`
- `Descricao`: garantir retorno de erro de negocio claro quando o usuario tentar criar ou atualizar um segundo `PEDIDO` para o mesmo projeto.
- `Pronto quando`: o backend rejeitar esse estado com mensagem consistente.

### TASK-GRP-012 - Ignorar a regra para orcamentos sem projeto

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-010`
- `Descricao`: garantir que a automacao nao ocorra quando o orcamento nao tiver `project_id`.
- `Pronto quando`: orcamentos sem projeto continuarem independentes.

## Fase 5 - Historico E Rastreabilidade

### TASK-GRP-013 - Registrar historico do vencedor

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-010`
- `Descricao`: registrar no `budget_status_history` a mudanca do orcamento vencedor para `PEDIDO`.
- `Pronto quando`: a auditoria do item vencedor estiver persistida.

### TASK-GRP-014 - Registrar historico dos cancelamentos automaticos

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-010`, `TASK-GRP-013`
- `Descricao`: registrar no historico que os demais itens foram para `CANCELADO` por regra automatica do grupo.
- `Pronto quando`: os orcamentos cancelados automaticamente puderem ser auditados e explicados.

### TASK-GRP-015 - Instrumentar logs da regra do grupo

- `Prioridade`: `P2`
- `Dependencias`: `TASK-GRP-010`
- `Descricao`: aproveitar o padrao de logs ja criado para registrar evento de definicao de vencedor do projeto e cancelamento dos concorrentes.
- `Pronto quando`: houver log com `project_id`, orcamento vencedor e quantidade de itens impactados.

## Fase 6 - API E Contratos

### TASK-GRP-016 - Expor endpoint para listar orcamentos de um projeto

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-007`
- `Descricao`: criar endpoint especifico para retornar os orcamentos vinculados a um projeto, por exemplo `GET /projects/:project_id/budgets`.
- `Pronto quando`: o frontend conseguir montar a tela de grupo do projeto usando API dedicada.

### TASK-GRP-017 - Enriquecer resposta com metadados do grupo

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-016`
- `Descricao`: adicionar na resposta informacoes como vencedor atual, total de orcamentos, quantidade ativa e quantidade cancelada.
- `Pronto quando`: a tela do projeto puder montar cards-resumo sem calculo excessivo no cliente.

### TASK-GRP-018 - Ajustar contrato de atualizacao de status

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-010`
- `Descricao`: revisar se a API atual de update de orcamento precisa retornar informacao adicional quando a mudanca para `PEDIDO` impactar outros itens do grupo.
- `Pronto quando`: o frontend puder exibir feedback claro sobre quantos itens foram cancelados automaticamente.

## Fase 7 - Frontend

### TASK-GRP-019 - Exibir projeto e estado do grupo na listagem de orcamentos

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-016`
- `Descricao`: enriquecer a listagem atual de orcamentos com indicacao mais clara do projeto vinculado e, se viavel, do estado do grupo.
- `Pronto quando`: o usuario conseguir perceber que o orcamento pertence a um agrupamento.

### TASK-GRP-020 - Criar tela de detalhe do projeto com orcamentos vinculados

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-016`, `TASK-GRP-017`
- `Descricao`: criar a tela ou aba de projeto com tabela de todos os orcamentos vinculados e destaque visual do vencedor.
- `Pronto quando`: o usuario conseguir abrir um projeto e ver claramente o grupo de orcamentos.

### TASK-GRP-021 - Adicionar acao de definir como `PEDIDO` com confirmacao

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-020`, `TASK-GRP-018`
- `Descricao`: antes de concluir a mudanca para `PEDIDO`, exibir confirmacao informando que os demais orcamentos ativos do projeto serao cancelados automaticamente.
- `Pronto quando`: o usuario confirmar conscientemente a decisao de vencedor do grupo.

### TASK-GRP-022 - Destacar cancelamento automatico no frontend

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-014`, `TASK-GRP-020`
- `Descricao`: exibir badge, texto ou estado visual indicando que determinado orcamento foi cancelado automaticamente por outro item do projeto ter virado `PEDIDO`.
- `Pronto quando`: a interface explicar com clareza por que o item saiu de atencao ativa.

### TASK-GRP-023 - Permitir associacao de orcamentos existentes a um projeto

- `Prioridade`: `P2`
- `Dependencias`: `TASK-GRP-020`
- `Descricao`: adicionar fluxo de vinculo de orcamentos existentes a partir da tela do projeto ou da tela de orcamentos.
- `Pronto quando`: o usuario conseguir reorganizar grupos sem precisar recriar orcamentos.

## Fase 8 - Testes

### TASK-GRP-024 - Criar testes unitarios da regra do grupo

- `Prioridade`: `P0`
- `Dependencias`: `TASK-GRP-010`, `TASK-GRP-011`, `TASK-GRP-012`
- `Descricao`: cobrir no service os cenarios de sucesso e falha da regra de um `PEDIDO` por projeto com cancelamento automatico dos demais.
- `Pronto quando`: os cenarios principais e alternativos estiverem cobertos por testes unitarios.

### TASK-GRP-025 - Criar testes de historico da automacao

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-013`, `TASK-GRP-014`
- `Descricao`: validar que a mudanca do vencedor e o cancelamento dos concorrentes ficam registrados corretamente.
- `Pronto quando`: o historico automatico estiver testado.

### TASK-GRP-026 - Criar testes de integracao da API do grupo

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-016`, `TASK-GRP-018`
- `Descricao`: testar o fluxo principal de listar grupo do projeto e definir um item como `PEDIDO`, verificando o cancelamento dos demais.
- `Pronto quando`: a API do agrupamento estiver validada em integracao.

### TASK-GRP-027 - Validar frontend do fluxo principal

- `Prioridade`: `P1`
- `Dependencias`: `TASK-GRP-020`, `TASK-GRP-021`, `TASK-GRP-022`
- `Descricao`: validar o fluxo visual de projeto com varios orcamentos, incluindo confirmacao de vencedor e refletindo cancelamento automatico.
- `Pronto quando`: a tela operar sem quebrar filtros, navegacao e mensagens ao usuario.

## Fase 9 - Refinamentos

### TASK-GRP-028 - Remover itens cancelados da fila principal por padrao

- `Prioridade`: `P2`
- `Dependencias`: `TASK-GRP-022`
- `Descricao`: ajustar listagens e filtros para que itens cancelados automaticamente deixem de poluir a fila operacional principal, mantendo possibilidade de consulta.
- `Pronto quando`: o sistema priorizar naturalmente o que ainda exige atencao.

### TASK-GRP-029 - Adicionar resumo do grupo na listagem geral

- `Prioridade`: `P2`
- `Dependencias`: `TASK-GRP-019`, `TASK-GRP-017`
- `Descricao`: exibir indicadores como quantidade de orcamentos do projeto e existencia de vencedor na listagem geral de orcamentos.
- `Pronto quando`: a listagem principal ajudar a identificar grupos sem abrir o detalhe.

### TASK-GRP-030 - Reavaliar necessidade de entidade dedicada de agrupamento

- `Prioridade`: `P3`
- `Dependencias`: todas as anteriores
- `Descricao`: apos o uso da primeira versao, reavaliar se `project_id` continua suficiente ou se faz sentido criar entidade propria para oportunidade comercial.
- `Pronto quando`: houver decisao clara sobre manter ou evoluir a modelagem.

## Entrega Minima Recomendada

Se for necessario reduzir escopo para a primeira entrega funcional, recomendo considerar como minimo:

- `TASK-GRP-001`
- `TASK-GRP-002`
- `TASK-GRP-007`
- `TASK-GRP-008`
- `TASK-GRP-009`
- `TASK-GRP-010`
- `TASK-GRP-011`
- `TASK-GRP-012`
- `TASK-GRP-016`
- `TASK-GRP-020`
- `TASK-GRP-021`
- `TASK-GRP-024`
- `TASK-GRP-026`

Com esse conjunto, o sistema ja entrega a regra central e a tela principal do agrupamento.
