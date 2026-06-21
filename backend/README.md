# go-gerenciador-orcamento-deck

Backend em Go para gestao de orcamentos com PostgreSQL.

> Execute os comandos deste arquivo a partir da pasta `backend/`.

## Primeira Base Implementada

Nesta primeira etapa o projeto ja possui:

- bootstrap HTTP com `Gin`
- carga de configuracao via `.env`
- conexao com `PostgreSQL`
- `docker-compose.yml` para banco local
- endpoint `GET /check-health`
- autenticacao com `JWT`
- `refresh token` persistido em banco
- modulo inicial de `users`
- migrations iniciais de `users` e `refresh_tokens`

## Documentacao

- estudo do dominio: `../docs/ESTUDO-SISTEMA-BACKEND-ORCAMENTOS.md`
- backlog tecnico: `../docs/BACKLOG-TECNICO-IMPLEMENTACAO.md`
- arquitetura recomendada do frontend: `../docs/ARQUITETURA-FRONTEND.md`

## Como Rodar

1. Copie `.env.example` para `.env`
2. Ajuste `DB_USER`, `DB_PASSWORD`, `DATABASE_URL`, `SECRET_JWT` e `INITIAL_ADMIN_SETUP_TOKEN` com valores locais
3. Suba o banco com `docker compose up -d`
4. Aplique as migrations da pasta `db/migrations`
5. Rode `go run ./cmd`

## Testes

- a suite atual esta concentrada em `test/integration`
- os testes criam um banco temporario por execucao usando a conexao base definida no `.env`
- o banco temporario aplica as migrations de `db/migrations` automaticamente
- a base principal `budget_management` nao e reutilizada pelos testes de integracao

### Pre-requisitos

- `.env` configurado com acesso valido ao PostgreSQL local
- banco `postgres` acessivel pelo usuario configurado, pois os testes criam e removem databases temporarios
- container do PostgreSQL em execucao

### Comandos

- rodar toda a suite: `go test ./...`
- rodar apenas integracao: `go test ./test/integration/...`
- validar formatacao e compilacao antes dos testes:
  - `gofmt -w ./cmd ./internal ./pkg ./test`
  - `go test ./...`

### Cobertura Atual

- `GET /check-health`
- fluxo de `auth`:
  - `POST /auth/register`
  - `POST /auth/login`
  - `POST /auth/refresh`
- fluxo de `users`:
  - `GET /users/me`
  - `GET /users`
  - `POST /users`
- fluxo principal de `budgets`:
  - `POST /budgets`
  - `GET /budgets`
  - `GET /budgets/:budget_id`
  - `PUT /budgets/:budget_id`
  - `DELETE /budgets/:budget_id`
- regras principais de `budgets`:
  - filtros, paginacao e ordenacao
  - unicidade por `budget_number + year_budget`
- fluxos derivados de `budgets`:
  - `POST /budgets/:budget_id/follow-ups`
  - `GET /budgets/:budget_id/follow-ups`
  - `PATCH /budgets/:budget_id/status`
  - `GET /budgets/:budget_id/status-history`

## Banco Local

- host: `127.0.0.1`
- porta: `5433`
- database: `budget_management`
- usuario: definido no seu `.env` local
- senha: definida no seu `.env` local

## Endpoints Iniciais

- `GET /check-health`
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`
- `GET /users/me`
- `GET /users`
- `POST /users`
- `GET /budget-statuses`
- `POST /budget-statuses`
- `GET /budget-statuses/:status_id`
- `PUT /budget-statuses/:status_id`
- `DELETE /budget-statuses/:status_id`
- `GET /loss-reasons`
- `POST /loss-reasons`
- `GET /loss-reasons/:reason_id`
- `PUT /loss-reasons/:reason_id`
- `DELETE /loss-reasons/:reason_id`
- `GET /priorities`
- `POST /priorities`
- `GET /priorities/:priority_id`
- `PUT /priorities/:priority_id`
- `DELETE /priorities/:priority_id`
- `GET /salespeople`
- `POST /salespeople`
- `GET /salespeople/:salesperson_id`
- `PUT /salespeople/:salesperson_id`
- `DELETE /salespeople/:salesperson_id`
- `GET /installers`
- `POST /installers`
- `GET /installers/:installer_id`
- `PUT /installers/:installer_id`
- `DELETE /installers/:installer_id`
- `GET /contacts`
- `POST /contacts`
- `GET /contacts/:contact_id`
- `PUT /contacts/:contact_id`
- `DELETE /contacts/:contact_id`
- `GET /budgets`
- `GET /budgets/:budget_id`
- `POST /budgets`
- `PUT /budgets/:budget_id`
- `DELETE /budgets/:budget_id`
- `PATCH /budgets/:budget_id/status`
- `GET /budgets/:budget_id/status-history`
- `GET /budgets/:budget_id/follow-ups`
- `POST /budgets/:budget_id/follow-ups`
- `GET /project-types`
- `POST /project-types`
- `GET /project-types/:project_type_id`
- `PUT /project-types/:project_type_id`
- `DELETE /project-types/:project_type_id`
- `GET /projects`
- `POST /projects`
- `GET /projects/:project_id`
- `PUT /projects/:project_id`
- `DELETE /projects/:project_id`

## Observacao

- `GET /budgets` aceita filtros por query string: `budget_number`, `year_budget`, `status_id`, `salesperson_id`, `installer_id`, `priority_id`, `project_type_id`, `projetista_name`, `competitor_name`, `sent_at_from`, `sent_at_to`, `gross_value_min`, `gross_value_max`, `page`, `page_size`, `sort_by` e `sort_order`
- `POST /auth/register` fica disponivel apenas enquanto nao existir nenhum usuario cadastrado
- o primeiro usuario criado por esse endpoint nasce com perfil `admin`
