# Backlog de Implementacao do Frontend

## Objetivo

Este documento continua o estudo iniciado em `docs/ARQUITETURA-FRONTEND.md` e transforma a arquitetura recomendada em um plano pratico de execucao para o frontend do sistema de gestao de orcamentos.

O foco aqui e:

- definir a ordem real de implementacao do frontend
- alinhar as telas com os endpoints ja disponiveis no backend
- reduzir ambiguidade antes do bootstrap do projeto React
- organizar o MVP em entregas pequenas e validaveis

## Estado Atual

Hoje o projeto possui:

- `backend/` com API Go funcional
- autenticacao com `JWT` e `refresh token`
- CRUD principal de `budgets`
- `follow-ups` e `status history`
- CRUDs auxiliares para catalogos operacionais
- `frontend/` ainda sem bootstrap de aplicacao
- `docs/ARQUITETURA-FRONTEND.md` com a direcao arquitetural

Conclusao:

- o estudo de frontend parou no nivel de arquitetura
- o proximo passo natural e iniciar o frontend com um backlog orientado a entregas

## Stack Confirmada

Seguir a stack recomendada no estudo anterior:

- `React`
- `TypeScript`
- `Vite`
- `@mui/material`
- `react-router-dom`
- `axios`
- `@tanstack/react-query`
- `react-hook-form`
- `zod`
- `@hookform/resolvers`
- `vitest`
- `@testing-library/react`
- `msw`

## Premissas do MVP

Para o primeiro ciclo do frontend, assumir:

- o backend roda localmente em `http://localhost:8080`
- o frontend sera um painel administrativo web
- o login sera por `email + password`
- tokens serao armazenados inicialmente em `localStorage`
- o menu lateral tera foco em operacao e nao em apresentacao institucional
- rotas administrativas devem respeitar o perfil retornado pela API em `/users/me`

## Modulos Ja Disponiveis no Backend

### Autenticacao e usuario

- `POST /auth/login`
- `POST /auth/refresh`
- `GET /users/me`
- `GET /users`
- `POST /users`

### Orcamentos

- `GET /budgets`
- `GET /budgets/:budget_id`
- `POST /budgets`
- `PUT /budgets/:budget_id`
- `DELETE /budgets/:budget_id`
- `PATCH /budgets/:budget_id/status`
- `GET /budgets/:budget_id/status-history`
- `GET /budgets/:budget_id/follow-ups`
- `POST /budgets/:budget_id/follow-ups`

### Catalogos auxiliares

- `GET|POST|GET by id|PUT|DELETE /budget-statuses`
- `GET|POST|GET by id|PUT|DELETE /loss-reasons`
- `GET|POST|GET by id|PUT|DELETE /priorities`
- `GET|POST|GET by id|PUT|DELETE /salespeople`
- `GET|POST|GET by id|PUT|DELETE /installers`
- `GET|POST|GET by id|PUT|DELETE /contacts`
- `GET|POST|GET by id|PUT|DELETE /project-types`
- `GET|POST|GET by id|PUT|DELETE /projects`

## Contratos Mais Importantes Para o Frontend

### Login

Request:

```json
{
  "email": "admin@local.dev",
  "password": "123456"
}
```

Response:

```json
{
  "token": "jwt",
  "refresh_token": "refresh-token"
}
```

### Usuario autenticado

Response esperada em `/users/me`:

```json
{
  "id": 1,
  "name": "Admin Local",
  "email": "admin@local.dev",
  "username": "admin",
  "role": "admin",
  "active": true,
  "created_at": "2026-06-13T10:00:00Z",
  "updated_at": "2026-06-13T10:00:00Z"
}
```

### Listagem de orcamentos

Filtros disponiveis:

- `budget_number`
- `year_budget`
- `status_id`
- `salesperson_id`
- `installer_id`
- `priority_id`
- `project_type_id`
- `designer_name`
- `competitor_name`
- `sent_at_from`
- `sent_at_to`
- `gross_value_min`
- `gross_value_max`
- `page`
- `page_size`
- `sort_by`
- `sort_order`

Response:

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 0
}
```

### Budget

Campos principais do formulario:

- `budget_number`
- `year_budget`
- `revision`
- `sent_at`
- `gross_value`
- `commission_value`
- `area_m2`
- `status_id`
- `priority_id`
- `installer_id`
- `project_id`
- `salesperson_id`
- `contact_id`
- `loss_reason_id`
- `competitor_name`
- `competitor_price`
- `designer_name`
- `specification_details`
- `current_follow_up`

### Follow-up

Request:

```json
{
  "notes": "Cliente pediu retorno na sexta",
  "follow_up_at": "2026-06-20T15:00:00Z"
}
```

### Mudanca de status

Request:

```json
{
  "status_id": 2,
  "notes": "Status alterado apos contato"
}
```

## Estrutura de Rotas do Frontend

### Publicas

- `/login`

### Protegidas

- `/`
- `/budgets`
- `/budgets/new`
- `/budgets/:budgetId`
- `/budgets/:budgetId/edit`
- `/catalogs/budget-statuses`
- `/catalogs/loss-reasons`
- `/catalogs/priorities`
- `/catalogs/salespeople`
- `/catalogs/installers`
- `/catalogs/contacts`
- `/catalogs/project-types`
- `/catalogs/projects`
- `/users`
- `/me`

## Estrutura Minima do Projeto

```text
frontend/
  src/
    app/
      providers/
      router/
      theme/
    components/
      common/
      feedback/
      form/
      layout/
    features/
      auth/
      budgets/
      budget-follow-ups/
      budget-status-history/
      catalogs/
      users/
    lib/
      axios/
      query/
      storage/
      utils/
    types/
    main.tsx
```

## Backlog por Fase

## Fase 1 - Bootstrap e infraestrutura

### FRONT-001 - Inicializar projeto React com Vite e TypeScript

- `Prioridade`: `P0`
- `Dependencias`: nenhuma
- `Descricao`: criar a base do frontend em `frontend/` usando `React + TypeScript + Vite`.
- `Pronto quando`: existir aplicacao inicial rodando em desenvolvimento com estrutura de `src/`.

### FRONT-002 - Instalar bibliotecas base do frontend

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-001`
- `Descricao`: instalar `MUI`, `React Router`, `Axios`, `React Query`, `React Hook Form`, `Zod` e stack de testes.
- `Pronto quando`: as dependencias do MVP estiverem declaradas e a aplicacao compilar.

### FRONT-003 - Configurar providers globais

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-001`, `FRONT-002`
- `Descricao`: criar `QueryClientProvider`, `ThemeProvider`, `CssBaseline` e infraestrutura base da aplicacao.
- `Pronto quando`: a app subir com tema e `react-query` configurados.

### FRONT-004 - Configurar variaveis de ambiente

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-001`
- `Descricao`: criar `.env.example` no frontend com `VITE_APP_NAME` e `VITE_API_URL`.
- `Pronto quando`: a URL da API estiver desacoplada do codigo.

### FRONT-005 - Criar cliente HTTP compartilhado

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-003`, `FRONT-004`
- `Descricao`: criar instancia `axios` com `baseURL`, header `Authorization` e interceptors.
- `Pronto quando`: as features consumirem a API por um cliente unico.

## Fase 2 - Autenticacao e protecao de rotas

### FRONT-006 - Implementar armazenamento de sessao

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-005`
- `Descricao`: criar camada para salvar, ler e limpar `token` e `refresh_token`.
- `Pronto quando`: a sessao puder ser persistida e restaurada.

### FRONT-007 - Implementar fluxo de login

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-006`
- `Descricao`: criar pagina `/login`, formulario com validacao e chamada para `POST /auth/login`.
- `Pronto quando`: usuario conseguir autenticar e entrar na area protegida.

### FRONT-008 - Implementar refresh token automatico

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-006`, `FRONT-007`
- `Descricao`: interceptar `401`, tentar `POST /auth/refresh` e repetir a requisicao original.
- `Pronto quando`: a sessao for renovada automaticamente quando possivel.

### FRONT-009 - Implementar carregamento do usuario autenticado

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-007`
- `Descricao`: consultar `GET /users/me` e disponibilizar os dados do usuario logado para o app.
- `Pronto quando`: layout e guards conhecerem `name`, `role` e estado autenticado.

### FRONT-010 - Implementar `ProtectedRoute` e `RoleGuard`

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-009`
- `Descricao`: proteger rotas autenticadas e esconder telas administrativas quando o perfil nao permitir.
- `Pronto quando`: rotas sensiveis forem controladas no frontend.

## Fase 3 - Layout e navegacao

### FRONT-011 - Criar `AppShell`

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-010`
- `Descricao`: implementar menu lateral, topo com usuario e botao de logout.
- `Pronto quando`: as rotas protegidas compartilharem um layout unico.

### FRONT-012 - Criar componentes base reutilizaveis

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-011`
- `Descricao`: criar `PageHeader`, `AppSidebar`, `AppEmptyState`, `LoadingView`, `ErrorState`, `ConfirmDialog`.
- `Pronto quando`: as primeiras telas nao repetirem estrutura de layout e feedback.

### FRONT-013 - Montar router principal

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-011`
- `Descricao`: registrar todas as rotas do MVP com lazy loading quando fizer sentido.
- `Pronto quando`: o frontend navegar entre login, budgets, catalogos e usuarios.

## Fase 4 - Orcamentos

### FRONT-014 - Tipar contratos de `budgets`

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-005`
- `Descricao`: criar types de request, response, filtros, paginação e ordenacao.
- `Pronto quando`: a feature `budgets` tiver contratos alinhados com o backend.

### FRONT-015 - Implementar API de `budgets`

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-014`
- `Descricao`: criar funcoes `listBudgets`, `getBudgetById`, `createBudget`, `updateBudget`, `deleteBudget`, `changeBudgetStatus`, `listBudgetStatusHistory`, `listBudgetFollowUps`, `createBudgetFollowUp`.
- `Pronto quando`: toda a feature usar uma camada `api/` pequena e tipada.

### FRONT-016 - Criar tela de listagem de orcamentos

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-015`, `FRONT-012`
- `Descricao`: montar tabela com colunas principais, loading, empty state e acoes por linha.
- `Pronto quando`: `/budgets` listar dados reais da API.

### FRONT-017 - Criar barra de filtros de orcamentos

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-016`
- `Descricao`: implementar filtros do backend com sincronizacao em query string.
- `Pronto quando`: o usuario conseguir filtrar e compartilhar a URL da listagem.

### FRONT-018 - Criar pagina de detalhe de orcamento

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-015`
- `Descricao`: montar pagina com dados principais, follow-ups e historico de status.
- `Pronto quando`: `/budgets/:budgetId` exibir os dados consolidados do orcamento.

### FRONT-019 - Criar formulario de criacao e edicao de orcamento

- `Prioridade`: `P0`
- `Dependencias`: `FRONT-015`
- `Descricao`: criar schema `zod`, integrar `react-hook-form` e suportar `POST` e `PUT`.
- `Pronto quando`: o usuario conseguir criar e editar orcamentos pelo frontend.

### FRONT-020 - Criar acoes de delete e alteracao de status

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-018`
- `Descricao`: implementar confirmacao de exclusao e formulario de mudanca de status.
- `Pronto quando`: as operacoes administrativas principais estiverem acessiveis na UI.

## Fase 5 - Follow-ups e historico

### FRONT-021 - Criar timeline de follow-ups

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-018`
- `Descricao`: listar `GET /budgets/:budget_id/follow-ups` em formato de timeline ou lista cronologica.
- `Pronto quando`: o historico comercial estiver visivel no detalhe do orcamento.

### FRONT-022 - Criar formulario de novo follow-up

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-021`
- `Descricao`: permitir registrar `notes` e `follow_up_at`.
- `Pronto quando`: a tela atualizar a timeline apos cadastro.

### FRONT-023 - Criar bloco de historico de status

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-018`
- `Descricao`: consumir `GET /budgets/:budget_id/status-history` e exibir mudancas com data, usuario e observacao.
- `Pronto quando`: o usuario enxergar a auditoria de status no frontend.

## Fase 6 - Catalogos auxiliares

### FRONT-024 - Criar feature generica de catalogo simples

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-012`, `FRONT-005`
- `Descricao`: criar base reaproveitavel para listagem, formulario e exclusao de catalogos.
- `Pronto quando`: pelo menos dois catalogos reaproveitarem a mesma estrutura visual.

### FRONT-025 - Implementar `budget-statuses`

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-024`
- `Descricao`: criar tela CRUD completa de status.
- `Pronto quando`: `/catalogs/budget-statuses` estiver funcional.

### FRONT-026 - Implementar `loss-reasons`

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-024`
- `Descricao`: criar tela CRUD completa de motivos de perda.
- `Pronto quando`: `/catalogs/loss-reasons` estiver funcional.

### FRONT-027 - Implementar `priorities`

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-024`
- `Descricao`: criar tela CRUD completa de prioridades.
- `Pronto quando`: `/catalogs/priorities` estiver funcional.

### FRONT-028 - Implementar `salespeople`

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-024`
- `Descricao`: criar tela CRUD completa de vendedores.
- `Pronto quando`: `/catalogs/salespeople` estiver funcional.

### FRONT-029 - Implementar `installers`

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-024`
- `Descricao`: criar tela CRUD completa de instaladores.
- `Pronto quando`: `/catalogs/installers` estiver funcional.

### FRONT-030 - Implementar `contacts`

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-024`
- `Descricao`: criar tela CRUD completa de contatos.
- `Pronto quando`: `/catalogs/contacts` estiver funcional.

### FRONT-031 - Implementar `project-types`

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-024`
- `Descricao`: criar tela CRUD completa de tipos de projeto.
- `Pronto quando`: `/catalogs/project-types` estiver funcional.

### FRONT-032 - Implementar `projects`

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-024`
- `Descricao`: criar tela CRUD completa de projetos.
- `Pronto quando`: `/catalogs/projects` estiver funcional.

## Fase 7 - Usuarios

### FRONT-033 - Criar listagem de usuarios

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-009`
- `Descricao`: montar tabela com `GET /users`.
- `Pronto quando`: admins conseguirem listar usuarios do sistema.

### FRONT-034 - Criar formulario de criacao de usuario

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-033`
- `Descricao`: implementar `POST /users` com validacao.
- `Pronto quando`: admins conseguirem cadastrar novos usuarios.

### FRONT-035 - Criar pagina de perfil basico

- `Prioridade`: `P2`
- `Dependencias`: `FRONT-009`
- `Descricao`: exibir dados do usuario autenticado em `/me`.
- `Pronto quando`: o usuario visualizar nome, email e papel atual.

## Fase 8 - Qualidade

### FRONT-036 - Configurar testes do frontend

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-002`
- `Descricao`: configurar `vitest`, `testing-library` e `msw`.
- `Pronto quando`: existir ambiente pronto para teste de componentes e fluxos.

### FRONT-037 - Testar autenticacao e guards

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-010`, `FRONT-036`
- `Descricao`: cobrir login, sessao restaurada, logout e redirecionamento quando token expirar.
- `Pronto quando`: o fluxo de autenticacao estiver coberto.

### FRONT-038 - Testar listagem e formulario de budgets

- `Prioridade`: `P1`
- `Dependencias`: `FRONT-016`, `FRONT-019`, `FRONT-036`
- `Descricao`: testar carregamento da listagem, filtros principais e submissao do formulario.
- `Pronto quando`: a feature central do MVP tiver cobertura minima relevante.

## Sequencia Recomendada de Implementacao

Se o objetivo for entregar valor rapido com menor risco, executar nesta ordem:

1. `FRONT-001`
2. `FRONT-002`
3. `FRONT-003`
4. `FRONT-004`
5. `FRONT-005`
6. `FRONT-006`
7. `FRONT-007`
8. `FRONT-008`
9. `FRONT-009`
10. `FRONT-010`
11. `FRONT-011`
12. `FRONT-012`
13. `FRONT-013`
14. `FRONT-014`
15. `FRONT-015`
16. `FRONT-016`
17. `FRONT-017`
18. `FRONT-019`
19. `FRONT-018`
20. `FRONT-020`
21. `FRONT-021`
22. `FRONT-022`
23. `FRONT-023`
24. `FRONT-024`
25. `FRONT-025` ate `FRONT-032`
26. `FRONT-033`
27. `FRONT-034`
28. `FRONT-035`
29. `FRONT-036`
30. `FRONT-037`
31. `FRONT-038`

## Recorte Minimo Para o Primeiro Sprint

Se a meta for sair do zero e ter algo navegavel rapidamente, o sprint inicial ideal e:

- bootstrap da aplicacao
- tema e providers
- cliente HTTP
- login
- sessao persistida
- protecao de rotas
- layout base
- listagem de orcamentos com filtros iniciais

Entregas esperadas:

- usuario consegue fazer login
- usuario navega no painel autenticado
- usuario lista orcamentos reais da API
- base do frontend fica pronta para crescimento organizado

## Proximo Passo Recomendado

Depois deste backlog, o proximo passo ideal e um destes dois caminhos:

1. inicializar efetivamente o projeto em `frontend/` com `Vite + React + TypeScript`
2. detalhar os contratos TypeScript do frontend antes do bootstrap, criando um documento de `types` por feature
