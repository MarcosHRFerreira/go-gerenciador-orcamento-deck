# go-gerenciador-orcamento-deck

Sistema de gestao de orcamentos com backend em Go e frontend em React.

## Visao Geral

- `backend/`: API REST em Go com `Gin`, `PostgreSQL`, `JWT` e testes de unidade e integracao
- `frontend/`: aplicacao web em `React`, `TypeScript`, `Vite`, `Material UI`, `React Query`, `React Hook Form` e `Zod`
- `docs/`: estudos, backlog e especificacoes funcionais

## Funcionalidades Entregues

- autenticacao com `login`, `refresh token` e obtencao do usuario logado
- troca obrigatoria de senha no primeiro acesso
- reset de senha por administrador com nova troca obrigatoria no proximo login
- politica de senha forte no backend e no frontend
- gestao de usuarios com perfis `admin` e `user`
- bloqueio de acesso administrativo para usuario comum
- importacao de orcamentos por planilha
- listagem, criacao, edicao e exclusao de orcamentos
- historico de status e follow-ups de orcamentos
- escopo por vendedor:
  - `admin` acessa todos os orcamentos
  - `user` acessa apenas os orcamentos vinculados ao vendedor resolvido pelo `username`
- melhorias visuais no frontend:
  - menu lateral recolhivel com tooltip nos icones
  - redirecionamento inicial para `Orcamentos`
  - tabela de orcamentos com cabecalho fixo
  - area fixa com colunas `ID` e `Orcamento`

## Estrutura

```text
go-gerenciador-orcamento-deck/
  backend/
  frontend/
  docs/
  RELATORIO DE ORCAMENTOS-25.xlsx
```

## Documentacao

- `docs/ESTUDO-SISTEMA-BACKEND-ORCAMENTOS.md`
- `docs/BACKLOG-TECNICO-IMPLEMENTACAO.md`
- `docs/ARQUITETURA-FRONTEND.md`
- `docs/BACKLOG-FRONTEND-IMPLEMENTACAO.md`
- `docs/GUIA-VISUAL-FRONTEND.md`
- `docs/ESTUDO-CARGA-PLANILHA-ORCAMENTOS.md`
- `docs/ESPECIFICACAO-API-IMPORTACAO-ORCAMENTOS.md`
- `docs/ESTUDO-CADASTRO-USUARIOS-E-PERFIS.md`
- `docs/ESPECIFICACAO-TELA-USUARIOS.md`

## Como Rodar

### Backend

Execute os comandos a partir de `backend/`:

1. copie `.env.example` para `.env`
2. suba o banco com `docker compose up -d`
3. aplique as migrations da pasta `db/migrations`
4. rode a API com `go run ./cmd`

### Frontend

Execute os comandos a partir de `frontend/`:

1. instale as dependencias com `yarn`
2. rode o projeto com `yarn dev`

## Testes E Validacao

### Backend

- `go test ./...`
- `go test ./test/unit/...`
- `go test ./test/integration/...`

### Frontend

- `yarn lint`
- `yarn build`

## Observacoes

- a rota inicial autenticada do frontend aponta para `Orcamentos`
- o fluxo de primeiro acesso redireciona o usuario para troca de senha antes de liberar o restante do sistema
- o escopo por vendedor usa o `username` do usuario autenticado para resolver o vendedor correspondente
