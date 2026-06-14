# Arquitetura Recomendada do Frontend

## Objetivo

Definir uma arquitetura simples, organizada e facil de manter para o frontend do sistema de gestao de orcamentos, considerando que:

- o backend principal ja existe em Go
- a API ja possui autenticacao, usuarios, orcamentos, follow-ups, historico de status e catalogos auxiliares
- o sistema tera perfil de uso mais administrativo do que institucional
- a prioridade e velocidade de entrega com boa legibilidade para manutencao manual
- a direcao visual desejada e moderna, limpa e profissional

## Direcao Visual

Para a camada visual, o frontend deve seguir um estilo:

- moderno
- limpo
- administrativo
- com alto foco em clareza operacional

Guia complementar:

- `docs/GUIA-VISUAL-FRONTEND.md`

## Stack Recomendada

### Linguagem

- `TypeScript`

Motivo:

- melhora a seguranca na integracao com a API
- facilita manutencao de formularios, tabelas e contratos HTTP
- reduz erros em refatoracoes

### Framework Base

- `React`
- `Vite`

Motivo:

- curva de entrada melhor que alternativas mais pesadas
- ecossistema muito amplo
- bom para painel administrativo e CRUD
- configuracao simples
- build e ambiente local rapidos

### Bibliotecas Principais

- UI: `@mui/material`
- icones: `@mui/icons-material`
- roteamento: `react-router-dom`
- requisicoes HTTP: `axios`
- cache e sincronizacao de dados: `@tanstack/react-query`
- formularios: `react-hook-form`
- validacao: `zod`
- integracao formulario + schema: `@hookform/resolvers`

## Tipo de Projeto

O frontend deve ser um painel administrativo web, separado do backend, consumindo a API REST ja existente.

Padrao sugerido:

- `backend/` para API Go
- `frontend/` para aplicacao React

## Estrutura de Pastas Recomendada

```text
frontend/
  src/
    app/
      providers/
      router/
      theme/
    assets/
    components/
      common/
      form/
      layout/
      feedback/
    features/
      auth/
        api/
        components/
        hooks/
        pages/
        schemas/
        types/
      budgets/
        api/
        components/
        hooks/
        pages/
        schemas/
        types/
      budget-follow-ups/
        api/
        components/
        hooks/
        pages/
        schemas/
        types/
      budget-status-history/
        api/
        components/
        hooks/
        pages/
        types/
      users/
        api/
        components/
        hooks/
        pages/
        schemas/
        types/
      catalogs/
        budget-statuses/
        loss-reasons/
        priorities/
        salespeople/
        installers/
        contacts/
        project-types/
        projects/
    lib/
      axios/
      query/
      utils/
    types/
    main.tsx
```

## Principios da Arquitetura

### 1. Separacao por feature

Cada modulo de negocio fica agrupado por funcionalidade e nao por tipo tecnico global.

Exemplo:

- `features/budgets/pages/BudgetListPage.tsx`
- `features/budgets/components/BudgetTable.tsx`
- `features/budgets/api/listBudgets.ts`
- `features/budgets/schemas/budgetFormSchema.ts`

Isso ajuda porque:

- facilita encontrar o codigo
- reduz acoplamento
- deixa o projeto mais simples para manutencao manual

### 2. Componentes compartilhados pequenos

Componentes reutilizaveis ficam em `components/`, mas apenas quando fizer sentido real.

Exemplos:

- `PageHeader`
- `AppSidebar`
- `AppDataTable`
- `ConfirmDialog`
- `FormTextField`
- `StatusChip`

Evitar abstrair cedo demais.

### 3. Regras de negocio fora dos componentes visuais

Os componentes devem focar em exibir tela e capturar interacao.

As responsabilidades devem ficar assim:

- `api/`: chamadas HTTP
- `hooks/`: composicao da regra de tela
- `schemas/`: validacao
- `pages/`: montagem da pagina
- `components/`: UI reutilizavel

### 4. Contratos tipados da API

Criar tipos claros para requests e responses.

Exemplo:

- `Budget`
- `BudgetListResponse`
- `CreateBudgetPayload`
- `UpdateBudgetPayload`

Sempre derivar os tipos com base no contrato real do backend.

## Camadas Recomendadas

### App

`src/app` concentra infraestrutura do frontend:

- providers globais
- tema
- router
- bootstrap da aplicacao

### Features

`src/features` concentra os modulos de negocio.

Cada feature deve conter somente o que faz parte daquele fluxo.

### Lib

`src/lib` concentra utilitarios tecnicos:

- instancia do `axios`
- configuracao do `react-query`
- helpers genericos

## Autenticacao

O frontend deve usar autenticacao baseada em:

- `token`
- `refresh_token`

Fluxo sugerido:

1. usuario faz login
2. frontend salva `token` e `refresh_token`
3. `axios` envia `Authorization: Bearer <token>`
4. se a API retornar erro de autenticacao, o frontend tenta `refresh`
5. se o refresh falhar, faz logout e redireciona para login

Armazenamento inicial recomendado:

- `localStorage`

Observacao:

- para um painel interno e MVP, `localStorage` e aceitavel
- no futuro, se quiser endurecer seguranca, da para revisar a estrategia

## Roteamento

Estrutura sugerida:

- rota publica:
  - `/login`
- rotas protegidas:
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

## Gerenciamento de Estado

Recomendacao:

- usar `React Query` para dados de servidor
- usar estado local do React para comportamento de tela
- evitar Redux neste primeiro momento

Motivo:

- a maior parte do estado vem da API
- reduz complexidade
- acelera desenvolvimento

## Formularios

Padrao recomendado:

- `React Hook Form` para controle do formulario
- `Zod` para validacao

Beneficios:

- menos boilerplate
- validacao clara
- boa integracao com TypeScript

## Tabelas e Listagens

As telas principais devem usar:

- filtros no topo
- tabela central
- paginacao no rodape
- acoes por linha

A listagem de orcamentos deve suportar:

- filtro por numero
- ano
- status
- vendedor
- instalador
- prioridade
- tipo de projeto
- projetista
- concorrente
- intervalo de data
- faixa de valor
- ordenacao
- paginacao

## Layout

Layout recomendado para o MVP:

- `AppShell` com menu lateral
- topo simples com nome do usuario e logout
- conteudo principal com `PageHeader`

Menu lateral sugerido:

- Dashboard
- Orcamentos
- Cadastros auxiliares
- Usuarios

## Modulos de Tela do MVP

### 1. Autenticacao

- Login

### 2. Orcamentos

- Listagem
- Criacao
- Edicao
- Detalhe

### 3. Fluxos do Orcamento

- Follow-ups
- Historico de status

### 4. Cadastros auxiliares

- Status
- Motivos de perda
- Prioridades
- Vendedores
- Instaladores
- Contatos
- Tipos de projeto
- Projetos

### 5. Usuarios

- Listagem
- Criacao
- Meu perfil basico

## Padrao de Consumo da API

Criar uma instancia unica do `axios`:

```ts
import axios from "axios";

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
});
```

Cada feature deve ter funcoes pequenas de API:

```ts
export const listBudgets = async (params: ListBudgetsParams) => {
  const response = await api.get<ListBudgetsResponse>("/budgets", { params });
  return response.data;
};
```

## Variaveis de Ambiente do Frontend

Arquivo sugerido:

- `frontend/.env`

Variaveis iniciais:

```env
VITE_APP_NAME=Gestao de Orcamentos
VITE_API_URL=http://localhost:8080
```

## Estrategia de Componentizacao

Prioridade:

1. componentes de pagina claros
2. componentes de formulario reaproveitaveis
3. componentes de tabela simples
4. evitar criar biblioteca interna grande cedo demais

## Estrategia de Testes do Frontend

Para a primeira fase:

- testes unitarios em componentes criticos e hooks principais
- testes de fluxo nas features mais importantes

Stack sugerida:

- `vitest`
- `@testing-library/react`
- `@testing-library/jest-dom`
- `msw`

## Sequencia Recomendada de Implementacao

### Fase 1

- bootstrap do frontend
- tema
- router
- layout base
- login

### Fase 2

- listagem de orcamentos
- filtros
- paginacao
- ordenacao

### Fase 3

- criacao e edicao de orcamento
- detalhe de orcamento

### Fase 4

- follow-ups
- historico de status

### Fase 5

- cadastros auxiliares
- usuarios

## Recomendacao Final

Para este projeto, a arquitetura mais indicada e:

- `frontend` separado do `backend`
- `React + TypeScript + Vite`
- `Material UI`
- `React Query`
- `React Hook Form`
- `Zod`

Essa combinacao equilibra:

- velocidade
- organizacao
- facilidade para manutencao
- grande disponibilidade de exemplos e suporte
