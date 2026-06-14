# Backlog Tecnico De Implementacao

## Objetivo

Este documento organiza as tarefas tecnicas para construcao do backend de orcamentos em Go, utilizando como base o estudo de dominio registrado em `docs/ESTUDO-SISTEMA-BACKEND-ORCAMENTOS.md`.

O objetivo deste backlog e:

- definir a ordem de execucao do projeto
- separar tarefas por fase
- indicar dependencias
- sugerir prioridades iniciais
- facilitar o acompanhamento da implementacao

## Convencoes Deste Backlog

- codigo, banco, tabelas, colunas, DTOs e modulos em `ingles`
- documentacao em `portugues`
- banco de dados `PostgreSQL`
- ambiente local com `Docker Compose`
- arquitetura inspirada no projeto `go-delivery-routing-lab`
- autenticacao e token inspirados no projeto `go-tweets`

## Escala De Prioridade

- `P0`: bloqueante ou fundacional
- `P1`: muito importante para o fluxo principal
- `P2`: importante para consolidacao da API
- `P3`: melhoria, suporte ou evolucao posterior

## Criterio De Uso

Cada task abaixo possui:

- `ID`: identificador da tarefa
- `Prioridade`: urgencia inicial sugerida
- `Dependencias`: tarefas que precisam vir antes
- `Descricao`: o que precisa ser executado
- `Pronto quando`: criterio objetivo de conclusao

## Ordem Macro Recomendada

1. fundacao do projeto
2. autenticacao, usuarios e autorizacao
3. configuracao de banco e infraestrutura local
4. migrations base e cadastros auxiliares
5. implementacao de `budgets`
6. implementacao de `budget_follow_ups` e `budget_status_history`
7. testes automatizados
8. importacao e saneamento da planilha

## Fase 1 - Fundacao Do Projeto

### TASK-001 - Inicializar a estrutura base do projeto

- `Prioridade`: `P0`
- `Dependencias`: nenhuma
- `Descricao`: criar a estrutura inicial de pastas seguindo o padrao do projeto de referencia: `cmd`, `db/migrations`, `docs`, `internal`, `pkg`, `test`.
- `Pronto quando`: a estrutura base existir no repositorio com `go.mod` configurado e organizacao inicial coerente com o estudo.

### TASK-002 - Configurar arquivo de ambiente e loader de configuracao

- `Prioridade`: `P0`
- `Dependencias`: `TASK-001`
- `Descricao`: criar a camada `internal/config` para leitura e validacao das variaveis de ambiente do projeto.
- `Pronto quando`: a aplicacao conseguir carregar `PORT`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DATABASE_URL`, `SECRET_JWT` e demais variaveis obrigatorias.

### TASK-003 - Criar pacote de conexao com PostgreSQL

- `Prioridade`: `P0`
- `Dependencias`: `TASK-002`
- `Descricao`: implementar `pkg/internalsql/postgres.go` utilizando `database/sql` com driver de PostgreSQL.
- `Pronto quando`: a aplicacao conseguir abrir conexao, configurar pool e validar `PingContext`.

### TASK-004 - Configurar bootstrap HTTP da aplicacao

- `Prioridade`: `P0`
- `Dependencias`: `TASK-002`, `TASK-003`
- `Descricao`: montar `cmd/main.go` com inicializacao de config, banco, router, timeouts e graceful shutdown.
- `Pronto quando`: a API subir localmente e encerrar corretamente ao receber sinal de parada.

### TASK-005 - Implementar health check

- `Prioridade`: `P0`
- `Dependencias`: `TASK-004`
- `Descricao`: criar endpoint `GET /check-health` para validar disponibilidade do banco.
- `Pronto quando`: o endpoint responder `200` quando o banco estiver acessivel e `503` quando nao estiver.

## Fase 1A - Usuarios, Auth E Autorizacao

### TASK-005A - Criar migration de `users`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `users` com `name`, `email`, `username`, `password_hash`, `role`, `active`, timestamps e constraints `unique` para `email` e `username`.
- `Pronto quando`: existir persistencia basica de usuarios com perfis `admin` e `user`.

### TASK-005B - Criar migration de `refresh_tokens`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-005A`, `TASK-007`
- `Descricao`: criar tabela `refresh_tokens` vinculada a `users`, com `refresh_token`, `expired_at` e suporte a rotacao.
- `Pronto quando`: o banco suportar armazenamento e renovacao segura de sessao.

### TASK-005C - Implementar pacote JWT e geracao de refresh token

- `Prioridade`: `P0`
- `Dependencias`: `TASK-002`
- `Descricao`: criar `pkg/internalsql/jwt` e `pkg/internalsql/refreshtoken` inspirados no `go-tweets`.
- `Pronto quando`: o projeto conseguir emitir e validar `access token` e gerar `refresh token`.

### TASK-005D - Implementar middleware de autenticacao

- `Prioridade`: `P0`
- `Dependencias`: `TASK-005C`
- `Descricao`: criar middleware para validar JWT, carregar `userID` no contexto e proteger rotas autenticadas.
- `Pronto quando`: handlers autenticados receberem a identidade do usuario pelo contexto.

### TASK-005E - Implementar middleware de autorizacao por perfil

- `Prioridade`: `P0`
- `Dependencias`: `TASK-005D`
- `Descricao`: criar middleware ou helper de autorizacao para restringir rotas por perfil `admin` e `user`.
- `Pronto quando`: rotas administrativas puderem ser protegidas por perfil.

### TASK-005F - Implementar modulo `auth`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-005A`, `TASK-005B`, `TASK-005C`
- `Descricao`: criar DTOs, repository, service e handler para `register`, `login` e `refresh`.
- `Pronto quando`: a API expor `POST /auth/register`, `POST /auth/login` e `POST /auth/refresh`.

### TASK-005G - Implementar modulo `users`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-005A`, `TASK-005E`
- `Descricao`: criar CRUD de usuarios do sistema, incluindo ativacao, desativacao e controle de perfil.
- `Pronto quando`: o sistema permitir manter usuarios com perfil `admin` e `user`.

## Fase 2 - Infraestrutura Local E Banco

### TASK-006 - Criar `docker-compose.yml` com PostgreSQL

- `Prioridade`: `P0`
- `Dependencias`: nenhuma
- `Descricao`: adicionar `docker-compose.yml` com servico `postgres`, volume persistente e porta exposta para desenvolvimento local.
- `Pronto quando`: o banco subir com `docker compose up -d` e aceitar conexao local.

### TASK-007 - Definir estrategia de migrations

- `Prioridade`: `P0`
- `Dependencias`: `TASK-001`
- `Descricao`: decidir e registrar a forma de execucao das migrations, mantendo os arquivos SQL versionados em `db/migrations`.
- `Pronto quando`: a equipe souber como criar, aplicar e versionar migrations do projeto.

### TASK-008 - Criar migration base de extensoes, padroes e convencoes

- `Prioridade`: `P1`
- `Dependencias`: `TASK-006`, `TASK-007`
- `Descricao`: criar a primeira migration com base de timestamps, padrao de `updated_at` e convencoes iniciais do banco.
- `Pronto quando`: existir um ponto de partida tecnico consistente para as demais migrations.

## Fase 3 - Cadastros Auxiliares Fundamentais

### TASK-009 - Criar migration de `budget_statuses`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `budget_statuses` com colunas de negocio e constraints `unique` para `code` e `name`.
- `Pronto quando`: a tabela existir com estrutura validada e pronta para seed inicial.

### TASK-010 - Criar seed inicial de `budget_statuses`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-009`
- `Descricao`: inserir os status iniciais do sistema, como `ORCAMENTO`, `FECHADO`, `CANCELADO`, `PERDIDO` e `COMPRA`.
- `Pronto quando`: a base local possuir os status minimos para operacao do fluxo principal.

### TASK-011 - Criar migration de `loss_reasons`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `loss_reasons` com constraints `unique` em `code` e `name`.
- `Pronto quando`: a tabela estiver disponivel para relacionar perdas e cancelamentos.

### TASK-012 - Criar migration de `priorities`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `priorities` com `code`, `name`, `weight` e constraints `unique`.
- `Pronto quando`: a aplicacao puder classificar urgencia comercial de cada orcamento.

### TASK-013 - Criar migration de `salespeople`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `salespeople` com suporte a `email`, `phone`, `active` e `unique` em `email` quando houver.
- `Pronto quando`: existir cadastro estruturado dos vendedores.

### TASK-014 - Criar migration de `installers`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `installers` com dados cadastrais e `unique` por `document` quando disponivel.
- `Pronto quando`: existir o cadastro da empresa instaladora ou integradora.

### TASK-015 - Criar migration de `contacts`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-014`
- `Descricao`: criar tabela `contacts` vinculada a `installers`, com `unique (installer_id, email)` e `unique (installer_id, phone)` quando aplicavel.
- `Pronto quando`: o sistema puder representar varios contatos por instalador.

### TASK-016 - Criar migration de `project_types`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `project_types` com `unique` para `code` e `name`.
- `Pronto quando`: a classificacao de tipo de obra estiver normalizada.

### TASK-017 - Criar migration de `projects`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-016`
- `Descricao`: criar tabela `projects` para representar a obra, sem forcar `unique` por `name` neste primeiro momento.
- `Pronto quando`: a aplicacao puder associar multiplos orcamentos a uma mesma obra.

### TASK-018 - Criar migration de `designers`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `designers`, priorizando `unique` por `document` quando houver e, na ausencia, avaliando `name` conforme qualidade dos dados.
- `Pronto quando`: o cadastro de projetistas estiver estruturado.

### TASK-019 - Criar migration de `specifications`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `specifications` com `code`, `name`, `category` e constraint `unique` apropriada.
- `Pronto quando`: o sistema puder classificar especificacoes tecnicas de forma normalizada.

### TASK-020 - Criar migration de `competitors`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-007`
- `Descricao`: criar tabela `competitors` com `unique (name)`.
- `Pronto quando`: o sistema puder registrar concorrentes recorrentes de forma estruturada.

## Fase 4 - Modulo Principal De Budgets

### TASK-021 - Criar migration de `budgets`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-005A`, `TASK-009`, `TASK-011`, `TASK-012`, `TASK-013`, `TASK-014`, `TASK-015`, `TASK-017`, `TASK-018`, `TASK-019`, `TASK-020`
- `Descricao`: criar tabela `budgets` com os relacionamentos principais, campos monetarios, campo `year_budget` e constraint `unique (budget_number, year_budget)`.
- `Pronto quando`: a entidade central estiver persistida com integridade relacional e unicidade definida.

### TASK-022 - Criar DTOs e models de `budgets`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-021`
- `Descricao`: definir contratos de entrada e saida e representacoes internas do modulo `budgets`.
- `Pronto quando`: houver DTOs de criacao, listagem, detalhe e atualizacao.

### TASK-023 - Criar repository de `budgets`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-022`
- `Descricao`: implementar operacoes de persistencia para `create`, `get by id`, `get all`, `update` e `delete`.
- `Pronto quando`: o repository suportar o CRUD basico e consultas essenciais.

### TASK-024 - Criar service de `budgets`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-023`
- `Descricao`: concentrar regras de negocio como validacao de identidade, consistencia de status e obrigatoriedade de motivo de perda.
- `Pronto quando`: o service proteger as principais invariantes do modulo.

### TASK-025 - Criar handler HTTP de `budgets`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-024`
- `Descricao`: implementar rotas `POST /budgets`, `GET /budgets`, `GET /budgets/:budget_id`, `PUT /budgets/:budget_id` e `DELETE /budgets/:budget_id`.
- `Pronto quando`: o CRUD principal estiver acessivel via HTTP.

### TASK-026 - Implementar filtros de listagem para `budgets`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-025`
- `Descricao`: adicionar filtros por `status`, `salesperson`, `installer`, `designer`, `project_type`, periodo de envio, faixa de valor, `competitor` e `priority`.
- `Pronto quando`: a listagem suportar a consulta operacional basica da area comercial.

### TASK-027 - Implementar validacoes de `unique` e conflitos amigaveis

- `Prioridade`: `P1`
- `Dependencias`: `TASK-025`
- `Descricao`: mapear erros de constraint do banco para respostas HTTP claras, especialmente em `budgets`, `contacts`, `budget_statuses` e outras entidades com unicidade definida.
- `Pronto quando`: a API responder com mensagens de erro compreensiveis para conflitos de cadastro.

## Fase 5 - Historico E Follow-Up Comercial

### TASK-028 - Criar migration de `budget_follow_ups`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-021`, `TASK-005A`
- `Descricao`: criar tabela de atualizacoes comerciais com observacao, status associado, data de proximo contato e autor.
- `Pronto quando`: o sistema puder guardar a linha do tempo comercial do orcamento.

### TASK-029 - Criar migration de `budget_status_history`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-021`, `TASK-005A`
- `Descricao`: criar tabela para registrar mudancas de status com rastreabilidade.
- `Pronto quando`: cada mudanca relevante de status puder ser auditada.

### TASK-030 - Implementar append de follow-up

- `Prioridade`: `P1`
- `Dependencias`: `TASK-028`
- `Descricao`: implementar `POST /budgets/:budget_id/follow-ups` e `GET /budgets/:budget_id/follow-ups`.
- `Pronto quando`: for possivel registrar e consultar o historico de acompanhamento comercial.

### TASK-031 - Implementar alteracao de status com historico

- `Prioridade`: `P1`
- `Dependencias`: `TASK-029`
- `Descricao`: implementar `PATCH /budgets/:budget_id/status` com escrita automatica em `budget_status_history`.
- `Pronto quando`: toda alteracao de status gerar historico consistente.

### TASK-032 - Atualizar `current_follow_up` automaticamente

- `Prioridade`: `P2`
- `Dependencias`: `TASK-030`
- `Descricao`: ao registrar um novo follow-up, atualizar o campo resumido em `budgets`.
- `Pronto quando`: o ultimo resumo comercial estiver refletido no cadastro principal sem perder o historico.

## Fase 6 - CRUDs Auxiliares Da API

### TASK-033 - Implementar CRUD de `budget_statuses`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-009`
- `Descricao`: criar DTO, repository, service e handler de `budget_statuses`.
- `Pronto quando`: o cadastro de status estiver exposto via API.

### TASK-034 - Implementar CRUD de `loss_reasons`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-011`
- `Descricao`: criar o modulo HTTP completo de `loss_reasons`.
- `Pronto quando`: a API permitir manter motivos de perda.

### TASK-035 - Implementar CRUD de `salespeople`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-013`
- `Descricao`: criar o modulo HTTP completo de `salespeople`.
- `Pronto quando`: a API permitir manter o cadastro de vendedores.

### TASK-036 - Implementar CRUD de `installers`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-014`
- `Descricao`: criar o modulo HTTP completo de `installers`.
- `Pronto quando`: a API permitir manter o cadastro de instaladores.

### TASK-037 - Implementar CRUD de `contacts`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-015`
- `Descricao`: criar o modulo HTTP completo de `contacts`.
- `Pronto quando`: a API permitir manter contatos por instalador.

### TASK-038 - Implementar CRUD de `project_types`

- `Prioridade`: `P2`
- `Dependencias`: `TASK-016`
- `Descricao`: criar o modulo HTTP completo de `project_types`.
- `Pronto quando`: a API permitir manter tipos de obra.

### TASK-039 - Implementar CRUD de `projects`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-017`
- `Descricao`: criar o modulo HTTP completo de `projects`.
- `Pronto quando`: a API permitir manter o cadastro de obras.

### TASK-040 - Implementar CRUD de `designers`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-018`
- `Descricao`: criar o modulo HTTP completo de `designers`.
- `Pronto quando`: a API permitir manter o cadastro de projetistas.

### TASK-041 - Implementar CRUD de `specifications`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-019`
- `Descricao`: criar o modulo HTTP completo de `specifications`.
- `Pronto quando`: a API permitir manter especificacoes tecnicas.

### TASK-042 - Implementar CRUD de `competitors`

- `Prioridade`: `P2`
- `Dependencias`: `TASK-020`
- `Descricao`: criar o modulo HTTP completo de `competitors`.
- `Pronto quando`: a API permitir manter o cadastro de concorrentes.

### TASK-043 - Implementar CRUD de `priorities`

- `Prioridade`: `P2`
- `Dependencias`: `TASK-012`
- `Descricao`: criar o modulo HTTP completo de `priorities`.
- `Pronto quando`: a API permitir manter prioridades comerciais.

### TASK-043A - Restringir CRUDs auxiliares para `admin`

- `Prioridade`: `P1`
- `Dependencias`: `TASK-005E`, `TASK-033`, `TASK-034`, `TASK-035`, `TASK-036`, `TASK-037`, `TASK-038`, `TASK-039`, `TASK-040`, `TASK-041`, `TASK-042`, `TASK-043`
- `Descricao`: aplicar autorizacao por perfil nas rotas de manutencao dos cadastros auxiliares.
- `Pronto quando`: apenas usuarios `admin` puderem alterar cadastros estruturais do sistema.

## Fase 7 - Testes Automatizados

### TASK-044 - Criar base de testes de unidade

- `Prioridade`: `P0`
- `Dependencias`: `TASK-001`
- `Descricao`: preparar a organizacao de `test/unit` com helpers, stubs e convencoes.
- `Pronto quando`: existir estrutura minima reutilizavel para testes unitarios.

### TASK-045 - Criar base de testes de integracao HTTP

- `Prioridade`: `P0`
- `Dependencias`: `TASK-004`, `TASK-006`
- `Descricao`: preparar `test/integration` com helpers HTTP e montagem do router de testes.
- `Pronto quando`: existir estrutura minima reutilizavel para testes de integracao.

### TASK-046 - Testar `health check`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-005`, `TASK-045`
- `Descricao`: cobrir cenarios de sucesso e indisponibilidade do banco.
- `Pronto quando`: o endpoint de saude estiver coberto por integracao.

### TASK-047 - Testar use cases de `budgets`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-024`, `TASK-044`
- `Descricao`: criar testes unitarios cobrindo fluxo principal e alternativos do service de `budgets`.
- `Pronto quando`: as principais regras de negocio estiverem cobertas.

### TASK-048 - Testar endpoints de `budgets`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-025`, `TASK-045`
- `Descricao`: criar testes de integracao para fluxo principal e principais erros HTTP do modulo `budgets`.
- `Pronto quando`: o CRUD principal estiver validado ponta a ponta.

### TASK-049 - Testar follow-up e historico de status

- `Prioridade`: `P1`
- `Dependencias`: `TASK-030`, `TASK-031`, `TASK-044`, `TASK-045`
- `Descricao`: testar criacao de follow-up, consulta de timeline e mudanca de status com auditoria.
- `Pronto quando`: os fluxos de acompanhamento comercial estiverem cobertos.

### TASK-050 - Testar modulos auxiliares criticos

- `Prioridade`: `P1`
- `Dependencias`: `TASK-033`, `TASK-036`, `TASK-037`, `TASK-039`, `TASK-040`, `TASK-041`
- `Descricao`: criar testes de unidade e integracao para os cadastros que sustentam o fluxo principal de orcamentos.
- `Pronto quando`: as entidades auxiliares essenciais tiverem cobertura minimamente robusta.

### TASK-050A - Testar `auth` e `users`

- `Prioridade`: `P0`
- `Dependencias`: `TASK-005F`, `TASK-005G`, `TASK-044`, `TASK-045`
- `Descricao`: criar testes de unidade e integracao para cadastro, login, refresh token e autorizacao por perfil.
- `Pronto quando`: os fluxos de autenticacao e gestao de usuarios estiverem cobertos.

## Fase 8 - Importacao E Saneamento Da Planilha

### TASK-051 - Mapear regras de transformacao da planilha

- `Prioridade`: `P1`
- `Dependencias`: nenhuma
- `Descricao`: definir como cada coluna da planilha sera transformada para as tabelas normalizadas do banco.
- `Pronto quando`: existir um mapeamento fechado da origem para o destino.

### TASK-052 - Definir regras de saneamento e consolidacao

- `Prioridade`: `P1`
- `Dependencias`: `TASK-051`
- `Descricao`: estabelecer tratamento para encoding, grafias diferentes, valores mistos, duplicidades e dados ausentes.
- `Pronto quando`: existir uma estrategia documentada de limpeza dos dados legados.

### TASK-053 - Criar rotina de importacao inicial

- `Prioridade`: `P2`
- `Dependencias`: `TASK-021`, `TASK-028`, `TASK-029`, `TASK-051`, `TASK-052`
- `Descricao`: desenvolver processo de importacao da planilha para o banco, respeitando normalizacao e constraints.
- `Pronto quando`: a base conseguir receber uma carga inicial controlada a partir da planilha.

### TASK-054 - Criar relatorio de inconsistencias da importacao

- `Prioridade`: `P2`
- `Dependencias`: `TASK-053`
- `Descricao`: registrar linhas rejeitadas, dados ambiguos e conflitos de cadastro durante a carga.
- `Pronto quando`: a importacao gerar saida auditavel para revisao manual.

## Fase 9 - Documentacao Operacional

### TASK-055 - Documentar setup local

- `Prioridade`: `P1`
- `Dependencias`: `TASK-006`, `TASK-004`
- `Descricao`: documentar como subir banco, configurar `.env`, aplicar migrations e iniciar a API.
- `Pronto quando`: uma pessoa nova conseguir subir o ambiente local sem apoio externo.

### TASK-056 - Documentar endpoints principais

- `Prioridade`: `P1`
- `Dependencias`: `TASK-025`, `TASK-030`, `TASK-031`
- `Descricao`: documentar payloads, respostas, erros esperados e filtros do fluxo principal.
- `Pronto quando`: os endpoints essenciais estiverem descritos para uso funcional e tecnico.

### TASK-057 - Manter traducao da documentacao

- `Prioridade`: `P3`
- `Dependencias`: `TASK-055`, `TASK-056`
- `Descricao`: manter versao traduzida ou complementar da documentacao para facilitar compartilhamento e onboarding.
- `Pronto quando`: o projeto tiver documentacao consistente conforme a estrategia de idioma definida.

## Sequencia Inicial Sugerida

Se o objetivo for iniciar logo com menor risco, eu sugiro executar primeiro:

1. `TASK-001`
2. `TASK-002`
3. `TASK-006`
4. `TASK-003`
5. `TASK-004`
6. `TASK-005`
7. `TASK-007`
8. `TASK-005A`
9. `TASK-005B`
10. `TASK-005C`
11. `TASK-005D`
12. `TASK-005E`
13. `TASK-005F`
14. `TASK-005G`
15. `TASK-009`
16. `TASK-010`
17. `TASK-014`
18. `TASK-015`
19. `TASK-016`
20. `TASK-017`
21. `TASK-013`
22. `TASK-018`
23. `TASK-019`
24. `TASK-020`
25. `TASK-011`
26. `TASK-012`
27. `TASK-021`

## Recorte Por MVP

### MVP 1 - Fundacao E Operacao Basica

Este deve ser o primeiro recorte real de implementacao.

Objetivo:

- subir a API com autenticacao
- manter usuarios do sistema
- manter cadastros auxiliares essenciais
- permitir criar e consultar orcamentos
- garantir testes do fluxo principal

Tasks recomendadas:

- `TASK-001`
- `TASK-002`
- `TASK-003`
- `TASK-004`
- `TASK-005`
- `TASK-006`
- `TASK-007`
- `TASK-005A`
- `TASK-005B`
- `TASK-005C`
- `TASK-005D`
- `TASK-005E`
- `TASK-005F`
- `TASK-005G`
- `TASK-009`
- `TASK-010`
- `TASK-011`
- `TASK-012`
- `TASK-013`
- `TASK-014`
- `TASK-015`
- `TASK-016`
- `TASK-017`
- `TASK-018`
- `TASK-019`
- `TASK-020`
- `TASK-021`
- `TASK-022`
- `TASK-023`
- `TASK-024`
- `TASK-025`
- `TASK-027`
- `TASK-044`
- `TASK-045`
- `TASK-046`
- `TASK-047`
- `TASK-048`
- `TASK-050A`
- `TASK-055`

Entregas esperadas:

- ambiente local funcional com PostgreSQL
- autenticacao com JWT e refresh token
- usuarios com perfil `admin` e `user`
- `health check`
- CRUD principal de `budgets`
- cadastros auxiliares minimos para operacao
- cobertura de testes do fluxo principal

### MVP 2 - Operacao Comercial Completa

Este recorte entra depois que o MVP 1 estiver estavel.

Objetivo:

- evoluir o fluxo comercial
- registrar historico e acompanhamento
- expor CRUDs auxiliares completos
- melhorar listagens e seguranca por perfil

Tasks recomendadas:

- `TASK-026`
- `TASK-028`
- `TASK-029`
- `TASK-030`
- `TASK-031`
- `TASK-032`
- `TASK-033`
- `TASK-034`
- `TASK-035`
- `TASK-036`
- `TASK-037`
- `TASK-038`
- `TASK-039`
- `TASK-040`
- `TASK-041`
- `TASK-042`
- `TASK-043`
- `TASK-043A`
- `TASK-049`
- `TASK-050`
- `TASK-056`

Entregas esperadas:

- timeline de follow-up por orcamento
- historico de mudanca de status
- filtros operacionais mais ricos
- manutencao completa dos cadastros auxiliares
- autorizacao por perfil aplicada nas rotas administrativas

### Pos-MVP - Importacao, Saneamento E Evolucao

Este bloco pode esperar ate a operacao principal estar estabilizada.

Tasks recomendadas:

- `TASK-008`
- `TASK-051`
- `TASK-052`
- `TASK-053`
- `TASK-054`
- `TASK-057`

Entregas esperadas:

- importacao controlada da planilha legada
- tratamento de inconsistencias de dados
- traducao e amadurecimento da documentacao

## Bloco Minimo Para Entregar Valor

O menor recorte que ja entrega valor funcional relevante para o projeto seria:

- fundacao da API
- PostgreSQL em Docker
- usuarios do sistema
- autenticacao com JWT
- refresh token persistido
- `health check`
- `budget_statuses`
- `installers`
- `contacts`
- `project_types`
- `projects`
- `salespeople`
- `users`
- `designers`
- `specifications`
- `loss_reasons`
- `competitors`
- `priorities`
- `budgets`
- testes do fluxo principal

## Recomendacao Pratica

Se voce quiser comecar com o menor risco possivel, eu recomendo atacar apenas o `MVP 1` neste momento.

Ordem pragmatica dentro do `MVP 1`:

1. fundacao e infraestrutura
2. autenticacao e usuarios
3. cadastros auxiliares essenciais
4. `budgets`
5. testes e documentacao minima

Os itens de `follow-up`, historico detalhado e importacao da planilha podem entrar depois, quando a base principal estiver estavel.

## Proximo Passo

Depois de revisar este backlog, o passo ideal e transformar as tasks priorizadas em:

- ordem real de execucao
- arquivo de acompanhamento de status
- migrations SQL iniciais
- esqueleto do projeto Go
