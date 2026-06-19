# Estudo: Avisos E Conversas Entre Usuarios

## Objetivo

Estudar como o sistema pode oferecer um modulo de comunicacao interna para:

- permitir que usuarios enviem avisos ou mensagens entre si
- permitir que o `admin` envie comunicados gerais
- manter historico consultavel
- encaixar esse recurso na arquitetura atual do backend e do frontend

O pedido funcional pode ser entendido de duas formas proximas:

1. um sistema simples de `avisos`
2. um sistema mais completo com `conversas`

Minha recomendacao e nao escolher apenas um dos dois.

O melhor desenho para este sistema e um modulo unico de `Comunicacao Interna` com dois recursos:

- `Avisos`: comunicados gerais, operacionais ou direcionados
- `Conversas`: troca de mensagens entre usuarios com historico

## Resumo Executivo

Minha recomendacao final e:

- criar um modulo chamado `Comunicacao`
- separar tecnicamente `avisos` de `conversas`
- permitir que apenas `admin` envie aviso geral para todos
- permitir que `admin` e `user` enviem mensagens diretas conforme escopo aprovado
- manter historico completo das conversas e dos avisos
- comecar por uma fase simples sem WebSocket, usando atualizacao por consulta HTTP
- preparar o modelo para notificacoes nao lidas, leitura e auditoria

Em termos praticos:

- `Avisos` resolvem o caso de comunicacao institucional
- `Conversas` resolvem o caso de troca entre usuarios
- juntar tudo em uma tabela unica aumenta ambiguidade e complica regras

## Estado Atual Do Sistema

Com base no codigo atual, o sistema ja possui alguns pontos favoraveis para este modulo:

### Backend

- autenticacao por JWT
- tabela `users`
- papeis `admin` e `user`
- arquitetura organizada em `handler`, `service`, `repository` e `model`
- padrao consolidado de migrations versionadas

Referencias:

- [router.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/server/router.go)
- [user_model.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/model/user_model.go)

### Frontend

- shell principal com menu lateral
- sessao autenticada carregando o usuario atual
- separacao por features
- controle de permissao por perfil administrativo

Referencias:

- [AppShell.tsx](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/components/layout/AppShell.tsx)
- [AppRouter.tsx](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/app/router/AppRouter.tsx)
- [auth.ts](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/features/auth/types/auth.ts)

### O Que Nao Existe Hoje

- tabela de avisos
- tabela de conversas
- tabela de mensagens
- indicador de nao lidos
- inbox ou central de comunicacao
- notificacoes visuais no topo da aplicacao

## Problema De Negocio

Hoje o sistema e forte na operacao de orcamentos, mas fraco em comunicacao interna.

Isso gera alguns gaps comuns:

- o `admin` nao tem um canal oficial para avisar todos os usuarios
- usuarios nao conseguem trocar informacoes dentro do proprio sistema
- decisoes e alinhamentos acabam indo para WhatsApp, e-mail ou ligacoes
- o historico operacional fica fora da plataforma

Na pratica, o modulo pedido ajudaria em cenarios como:

- `admin` avisar parada de sistema, mudanca de processo ou prazo importante
- vendedor pedir ajuda para orcamentista
- orcamentista responder duvida sobre um atendimento
- registrar contexto de uma solicitacao sem depender de ferramenta externa

## Opcoes De Solucao

## Opcao 1. Apenas Avisos

Descricao:

- criar somente avisos
- admin pode publicar para todos
- opcionalmente permitir aviso para usuario especifico

Vantagens:

- implementacao mais simples
- resolve comunicados institucionais
- baixo impacto inicial

Limites:

- nao resolve conversa entre usuarios
- nao atende bem perguntas e respostas
- historico fica unilateral, sem troca

Quando usar:

- se o objetivo for apenas mural interno

## Opcao 2. Apenas Chat Ou Conversas

Descricao:

- criar apenas canal de conversa entre usuarios
- toda comunicacao, inclusive geral, seria modelada como mensagem

Vantagens:

- modelo mais flexivel
- tudo fica em historico conversacional

Limites:

- comunicados gerais ficam menos claros
- mistura conversa operacional com anuncio institucional
- exige mais cuidado de UX para nao virar um chat confuso

Quando usar:

- se a prioridade for troca frequente entre usuarios

## Opcao 3. Solucao Hibrida

Descricao:

- criar `Avisos`
- criar `Conversas`
- cada recurso resolve um tipo de problema diferente

Vantagens:

- melhor aderencia ao negocio
- experiencia mais clara para o usuario
- regras mais simples de permissao
- historico melhor organizado

Limites:

- um pouco mais de implementacao que a opcao 1
- exige cuidado de navegacao no frontend

## Recomendacao

Minha recomendacao e a `Opcao 3`.

Em resumo:

- `Avisos` para comunicacao oficial
- `Conversas` para troca entre usuarios

## Desenho Recomendado

## 1. Avisos

`Avisos` seriam comunicados com comportamento de mural.

Tipos recomendados:

- `geral`: enviado pelo `admin` para todos
- `direcionado`: enviado para um ou mais usuarios
- `operacional`: ligado a manutencao, processo ou alerta interno

Campos principais sugeridos:

- titulo
- mensagem
- criado por
- tipo
- prioridade
- escopo de destinatarios
- data de expiracao opcional
- fixado no topo opcional
- status de leitura por usuario

Exemplos:

- `Sistema ficara indisponivel hoje das 18h as 19h`
- `Novo fluxo de importacao Trox entrou em producao`
- `Favor revisar orcamentos da obra 90`

## 2. Conversas

`Conversas` seriam threads com mensagens.

Tipos recomendados:

- `direta`: entre dois usuarios
- `grupo administrativo`: opcional para fase futura
- `contextual`: opcional, vinculada a obra ou orcamento

Minha sugestao para a primeira fase:

- suportar apenas conversa direta entre dois usuarios

Campos principais:

- participantes
- ultima mensagem
- data da ultima mensagem
- total de nao lidas por participante
- historico ordenado por data

## 3. Historico

O historico deve existir nos dois recursos, mas com papeis diferentes:

- em `Avisos`, historico de publicacao, leitura e eventual expiracao
- em `Conversas`, historico de mensagens trocadas

Isso permite:

- rastrear quem enviou
- saber quem leu
- auditar o que foi comunicado

## Modelo De Dados Recomendado

## 1. Tabelas Para Avisos

### `notices`

```sql
notices
- id
- title
- body
- scope_type          -- all | users
- priority            -- info | warning | critical
- pinned              -- boolean
- expires_at          -- nullable
- created_by_user_id
- created_at
- updated_at
```

### `notice_recipients`

```sql
notice_recipients
- id
- notice_id
- user_id
- read_at             -- nullable
- hidden_at           -- nullable
- created_at
- updated_at
```

Observacao:

- para aviso geral, o backend pode materializar destinatarios na criacao
- isso simplifica leitura, badge e auditoria por usuario

## 2. Tabelas Para Conversas

### `conversations`

```sql
conversations
- id
- type                -- direct
- created_by_user_id
- created_at
- updated_at
```

### `conversation_participants`

```sql
conversation_participants
- id
- conversation_id
- user_id
- last_read_message_id -- nullable
- joined_at
- created_at
- updated_at
```

### `conversation_messages`

```sql
conversation_messages
- id
- conversation_id
- sender_user_id
- body
- created_at
- updated_at
- deleted_at          -- nullable, se futuramente quiser soft delete
```

## 3. Tabelas Opcionais Futuras

### `notice_events`

Para auditoria detalhada:

```sql
notice_events
- id
- notice_id
- user_id
- event_type          -- created | read | hidden
- created_at
```

### `conversation_message_attachments`

Para anexos futuros:

```sql
conversation_message_attachments
- id
- message_id
- file_name
- mime_type
- storage_path
- created_at
```

## Regras De Permissao Recomendadas

## Avisos

- `admin` pode criar aviso geral para todos
- `admin` pode criar aviso direcionado
- `user` nao cria aviso geral
- opcionalmente `user` pode criar aviso direcionado apenas em fase futura
- todos podem listar os avisos que receberam
- todos podem marcar aviso como lido

## Conversas

- `admin` pode iniciar conversa com qualquer usuario
- `user` pode iniciar conversa com `admin`
- `user` pode iniciar conversa com outro `user` apenas se isso fizer sentido de negocio

Minha recomendacao para a fase inicial:

- `user` fala com `admin`
- `admin` fala com todos
- `user` com `user` fica para fase 2

Isso reduz risco de uso indevido e simplifica moderacao.

## Regras De UX Recomendadas

## Navegacao

Sugestao de novo item no menu lateral:

- `Comunicacao`

Subsecoes visuais:

- `Avisos`
- `Conversas`

Alternativa mais enxuta:

- um unico item `Comunicacao`
- abas internas `Avisos` e `Conversas`

## Topbar

Sugestao:

- icone de sino para avisos nao lidos
- icone de mensagens para conversas com nao lidas

Como o `AppShell` atual ja tem area de topo, esse encaixe e natural.

Referencia:

- [AppShell.tsx](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/frontend/src/components/layout/AppShell.tsx)

## Telas Recomendadas

### Tela 1. Central De Comunicacao

Com:

- resumo de avisos nao lidos
- resumo de conversas recentes
- filtros por lido, nao lido e prioridade

### Tela 2. Detalhe Do Aviso

Com:

- titulo
- mensagem completa
- autor
- data
- prioridade
- status de leitura

### Tela 3. Caixa De Conversas

Com:

- lista lateral de conversas
- painel central do historico
- composer de nova mensagem
- badge de nao lidas

## API Recomendada

## Avisos

### Criar aviso

`POST /notices`

Payload sugerido:

```json
{
  "title": "Parada programada",
  "body": "O sistema ficara indisponivel das 18h as 19h.",
  "scope_type": "all",
  "priority": "warning",
  "pinned": true,
  "expires_at": "2026-06-20T23:59:59Z",
  "recipient_user_ids": []
}
```

### Listar avisos do usuario

`GET /notices`

Filtros sugeridos:

- `status=unread|read|all`
- `priority=info|warning|critical`

### Marcar aviso como lido

`PATCH /notices/:noticeId/read`

### Ocultar aviso

`PATCH /notices/:noticeId/hide`

## Conversas

### Listar conversas do usuario

`GET /conversations`

### Criar conversa direta

`POST /conversations`

Payload sugerido:

```json
{
  "participant_user_ids": [12]
}
```

### Listar mensagens

`GET /conversations/:conversationId/messages`

### Enviar mensagem

`POST /conversations/:conversationId/messages`

Payload sugerido:

```json
{
  "body": "Preciso revisar o orcamento da obra 90."
}
```

### Marcar conversa como lida

`PATCH /conversations/:conversationId/read`

## Solucao Tecnica Recomendada Por Fases

## Fase 1. Avisos Simples

Entregas:

- admin envia aviso geral
- usuario visualiza avisos recebidos
- status lido e nao lido
- badge de avisos no topo

Vantagens:

- entrega rapida
- alto valor operacional
- baixo risco

## Fase 2. Conversa Direta Simples

Entregas:

- conversa direta entre admin e usuario
- historico de mensagens
- lista de conversas recentes
- indicador de nao lidas

Tecnologia recomendada:

- somente HTTP
- polling curto ou refetch manual

Observacao:

- nao precisa de WebSocket no inicio
- isso reduz custo e risco

## Fase 3. Contexto E Evolucoes

Entregas futuras:

- mensagem vinculada a `orcamento`
- mensagem vinculada a `obra`
- anexos
- grupos
- WebSocket ou SSE para tempo real

## Integracao Com O Dominio Atual

A comunicacao pode crescer melhor se depois aceitar contexto opcional:

- `budget_id`
- `project_id`

Exemplo:

- conversa aberta a partir do detalhe da obra
- conversa aberta a partir do orcamento
- aviso operacional referenciando uma obra

Mas minha recomendacao e nao depender disso no MVP.

Primeiro:

- central de comunicacao independente

Depois:

- contextualizacao com obra e orcamento

## Riscos E Cuidados

## 1. Misturar aviso com conversa em uma unica entidade

Risco:

- regras confusas
- UI ambigua
- filtros complexos

Recomendacao:

- manter recursos separados

## 2. Fazer tempo real logo de inicio

Risco:

- mais complexidade no backend
- mais cuidado com infra e conexoes
- maior custo de manutencao

Recomendacao:

- iniciar com HTTP e polling

## 3. Liberar conversa irrestrita entre todos os usuarios na primeira versao

Risco:

- aumento de ruido
- dificuldade de governanca
- uso indevido

Recomendacao:

- comecar com regras controladas

## 4. Nao materializar leitura por usuario

Risco:

- badge de nao lidas mais dificil
- auditoria pior
- consultas mais complexas

Recomendacao:

- registrar recebimento e leitura por usuario

## Minha Sugestao Final

Se o objetivo for ter algo realmente util e sustentavel, eu sugiro:

1. criar a `Central de Comunicacao`
2. implementar primeiro `Avisos`
3. depois implementar `Conversas diretas`
4. deixar grupos, anexos e tempo real para depois

Em linguagem simples:

- para aviso do admin para todos: `Avisos`
- para troca entre pessoas: `Conversas`

Isso atende exatamente o que voce descreveu, sem transformar o sistema em um chat pesado demais logo de inicio.

## Backlog Tecnico Sugerido

### Backend

- criar migration de `notices`
- criar migration de `notice_recipients`
- criar migration de `conversations`
- criar migration de `conversation_participants`
- criar migration de `conversation_messages`
- criar models
- criar repositories
- criar services
- criar handlers
- registrar novas rotas no router

### Frontend

- criar `features/communication`
- criar tipos de aviso e conversa
- criar APIs de avisos
- criar APIs de conversas
- adicionar item `Comunicacao` no menu
- criar tela central
- criar badge de nao lidos

### Testes

- testes de permissao para admin e user
- testes de criacao de aviso geral
- testes de listagem apenas dos avisos recebidos
- testes de conversa direta
- testes de leitura e nao lidas

## Conclusao

Sim, vale muito a pena implementar isso.

Mas eu recomendaria fazer do jeito abaixo:

- nao apenas `aviso`
- nao apenas `chat`
- e sim um modulo de `Comunicacao` com:
- `Avisos`
- `Conversas`

Essa abordagem entrega:

- comunicacao institucional
- conversa entre usuarios
- historico consultavel
- base boa para evolucao futura

Se a decisao for seguir, o melhor primeiro passo e:

- `Fase 1: Avisos gerais e direcionados`

Depois:

- `Fase 2: Conversas diretas com historico`
