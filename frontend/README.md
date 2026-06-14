# Frontend

Esta pasta agora contem a base inicial da aplicacao web do sistema.

## Stack Recomendada

- `React`
- `TypeScript`
- `Vite`
- `Material UI`
- `React Router`
- `Axios`
- `React Query`
- `React Hook Form`
- `Zod`

## Referencia de Arquitetura

Antes de iniciar a implementacao, use como base:

- `../docs/ARQUITETURA-FRONTEND.md`
- `../docs/BACKLOG-FRONTEND-IMPLEMENTACAO.md`
- `../docs/GUIA-VISUAL-FRONTEND.md`

## Objetivo Inicial do Frontend

Construir um painel administrativo com:

- login
- listagem de orcamentos
- criacao e edicao de orcamentos
- follow-ups
- historico de status
- cadastros auxiliares
- usuarios

## Proximo Passo Pratico

O estudo arquitetural ja foi consolidado. A continuidade recomendada agora e:

- seguir a ordem definida em `../docs/BACKLOG-FRONTEND-IMPLEMENTACAO.md`
- usar `../docs/GUIA-VISUAL-FRONTEND.md` como referencia do estilo moderno e limpo
- evoluir a autenticacao real com `POST /auth/login` e `POST /auth/refresh`
- conectar a listagem de orcamentos ao backend
- expandir o layout base para os modulos do MVP

## Estado Atual

O frontend ja possui:

- bootstrap com `React + TypeScript + Vite`
- `Material UI` configurado
- tema inicial alinhado ao visual moderno e limpo
- router base
- `AppShell` inicial
- pagina de login placeholder
- dashboard inicial
- tela inicial de listagem de orcamentos

## Comandos

- desenvolvimento: `yarn dev`
- build de producao: `yarn build`
- lint: `yarn lint`
