# Estudo De Implementacao De Logs No Backend

## Objetivo

Este documento registra uma proposta pratica para introduzir um padrao de logs no backend do projeto `go-gerenciador-orcamento-deck`.

O objetivo e:

- definir por que agora faz sentido adicionar logs estruturados
- documentar o padrao recomendado
- indicar os pontos exatos da base onde a implantacao deve acontecer
- reduzir risco de implementacao desorganizada
- criar um roteiro incremental para execucao futura

## Momento Atual Do Projeto

O backend ja possui:

- autenticacao com `access token` e `refresh token`
- endurecimento de seguranca em auth, CORS, headers e rate limit
- escopo de acesso em `budgets`
- importacao de planilhas
- testes unitarios e de integracao
- mensagens da API padronizadas em portugues

Ao mesmo tempo, o sistema ainda nao possui um padrao central de observabilidade. Hoje, o bootstrap em `backend/cmd/main.go` usa `log.Printf`, mas nao existe:

- logger estruturado
- `request_id` por requisicao
- log padronizado de entrada e saida HTTP
- correlacao entre erro de negocio e requisicao
- estrategia unica para mascaramento de dados sensiveis

Isso torna este um momento bom para adicionar logs:

- a arquitetura principal ja esta consolidada
- os fluxos criticos ja existem
- o custo de introduzir observabilidade agora e menor do que depois de mais modulos

## Estado Atual Da Base

### Bootstrap

Hoje o `backend/cmd/main.go` registra apenas eventos de inicializacao e shutdown com `log` padrao da linguagem.

### Router

Hoje o `backend/internal/server/router.go` usa:

- `gin.New()`
- `gin.Recovery()`
- middlewares proprios de seguranca e CORS

Ainda nao existe middleware proprio de log HTTP.

### Camadas De Negocio

As camadas `handler`, `service` e `repository` retornam erros corretamente, mas nao produzem logs estruturados.

### Consequencia Pratica

Quando algo falha ou se comporta de forma inesperada, hoje depende-se de:

- reproducao manual
- mensagem retornada pela API
- leitura do codigo

Falta visibilidade operacional para responder perguntas como:

- qual rota foi chamada
- por quem
- com qual resultado
- em quanto tempo
- qual fluxo de negocio foi executado

## Recomendacao Tecnica

### Recomendacao Principal

Adotar logs estruturados em `JSON` no backend, usando `log/slog`.

### Por Que `slog`

Eu recomendo `log/slog` neste projeto porque:

- faz parte da biblioteca padrao moderna do Go
- reduz dependencia externa
- atende bem o nivel atual de complexidade
- suporta atributos estruturados por campo
- facilita evolucao futura para arquivo, stdout ou agregadores externos

### Quando Considerar `zap`

`zap` faria sentido se houver necessidade forte de:

- performance extrema de logging
- ecossistema mais maduro de wrappers
- padrao ja consolidado em outros projetos da equipe

Para o estado atual, `slog` e suficiente e mais simples.

## Principios Do Padrao De Logs

O padrao recomendado deve seguir estes principios:

- logs estruturados, nao texto solto
- campos consistentes entre modulos
- sem exposicao de senha, token ou cookie
- separacao entre log tecnico e resposta da API
- correlacao por `request_id`
- foco em eventos realmente uteis para operacao e diagnostico

## O Que Deve Ser Logado

### Requisicoes HTTP

Toda requisicao deveria gerar um log de inicio ou fim contendo pelo menos:

- `request_id`
- `method`
- `path`
- `status_code`
- `latency_ms`
- `client_ip`
- `user_agent`
- `user_id` quando autenticado
- `username` quando autenticado
- `role` quando autenticado

### Eventos De Autenticacao

Fluxos que merecem log dedicado:

- tentativa de login
- login concluido
- refresh de sessao
- logout
- troca de senha
- reset de senha por admin
- bloqueio por usuario inativo

### Eventos De Administracao

Fluxos que merecem log dedicado:

- criacao de usuario
- alteracao de perfil
- ativacao e desativacao de usuario
- reset de senha

### Eventos De Orcamento

Fluxos que merecem log dedicado:

- criacao de orcamento
- edicao de orcamento
- exclusao de orcamento
- tentativa de acesso negada por escopo

### Importacao De Planilha

Fluxos que merecem log dedicado:

- inicio do preview
- preview concluido
- inicio da importacao
- resumo da importacao
- falhas de parse
- criacao automatica de catalogos durante importacao

## O Que Nao Deve Ser Logado

Nunca deve ser logado:

- senha em texto puro
- `access token`
- `refresh token`
- cookie de refresh
- corpo bruto completo de requisicoes sensiveis
- stack trace cru em resposta ao cliente

Deve haver cuidado extra com:

- e-mail completo em cenarios sensiveis
- dados de contatos
- dados de vendedores
- payloads de importacao grandes

## Estrutura Recomendada Dos Logs

### Campos Base

Campos minimos recomendados:

- `timestamp`
- `level`
- `message`
- `service`
- `environment`
- `request_id`

### Campos HTTP

- `method`
- `path`
- `route`
- `status_code`
- `latency_ms`
- `client_ip`

### Campos De Usuario

- `user_id`
- `username`
- `role`

### Campos De Negocio

Exemplos:

- `budget_id`
- `budget_number`
- `import_preview_id`
- `import_rows_read`
- `catalog_type`
- `target_user_id`

## Niveis De Log

Padrao recomendado:

- `DEBUG`: diagnostico local e detalhamento eventual
- `INFO`: fluxo normal importante
- `WARN`: falha de regra de negocio, acesso negado, input invalido
- `ERROR`: erro interno, dependencia falha, excecao inesperada

### Regra Pratica

- resposta `2xx` importante: `INFO`
- resposta `4xx` por regra de negocio: `WARN`
- resposta `5xx`: `ERROR`

## Arquitetura Recomendada

### 1. Pacote Central De Logger

Criar um pacote dedicado, por exemplo:

- `backend/internal/logger`

Responsabilidades:

- construir o logger principal
- configurar saida `JSON`
- aplicar campos fixos de servico e ambiente
- expor helpers para colocar e recuperar logger do contexto

### 2. Middleware De Request Log

Criar um middleware, por exemplo:

- `backend/internal/middleware/request_logger.go`

Responsabilidades:

- gerar `request_id`
- anexar `request_id` na resposta HTTP
- medir tempo de execucao
- coletar status final
- enriquecer com dados autenticados quando existirem
- produzir log unico por requisicao

### 3. Logger No Contexto

Sempre que possivel, handlers e services deveriam conseguir recuperar um logger contextualizado com:

- `request_id`
- usuario autenticado
- identificadores de negocio

### 4. Eventos De Dominio

Nos services principais, adicionar logs curtos e relevantes de negocio.

Importante:

- nao logar tudo
- nao logar toda validacao trivial
- priorizar eventos operacionais importantes

## Pontos De Entrada Recomendados Na Base Atual

### Fase 1

Arquivos candidatos:

- `backend/cmd/main.go`
- `backend/internal/server/router.go`
- `backend/internal/middleware`

Objetivo:

- trocar `log` padrao por logger central
- adicionar middleware global de request

### Fase 2

Arquivos candidatos:

- `backend/internal/service/auth/service.go`
- `backend/internal/service/user/service.go`
- `backend/internal/service/budgetimport/service.go`
- `backend/internal/service/budgetimport/execute.go`

Objetivo:

- cobrir auth e importacao, que sao fluxos de maior sensibilidade operacional

### Fase 3

Arquivos candidatos:

- `backend/internal/service/budget/service.go`
- `backend/internal/service/budgetfollowup/service.go`
- `backend/internal/service/budgetstatushistory/service.go`

Objetivo:

- cobrir operacao principal de negocio

## Estrategia De Implantacao

### Etapa 1 - Fundacao

Implementar:

- pacote `logger`
- configuracao inicial em `main.go`
- injecao do logger no bootstrap

Pronto quando:

- aplicacao subir com logger `JSON`
- logs de start e shutdown estiverem estruturados

### Etapa 2 - Middleware HTTP

Implementar:

- `request_id`
- middleware de log por requisicao
- header `X-Request-Id`

Pronto quando:

- toda requisicao produzir um log consistente

### Etapa 3 - Auth E Seguranca

Implementar logs em:

- login
- refresh
- logout
- troca de senha
- reset de senha
- negacoes por auth ou role

Pronto quando:

- eventos de seguranca estiverem rastreaveis sem expor dados sensiveis

### Etapa 4 - Importacao

Implementar logs em:

- preview
- parse
- execucao
- resumo final

Pronto quando:

- uma importacao puder ser auditada por `request_id` e `preview_id`

### Etapa 5 - Orcamentos

Implementar logs em:

- create
- update
- delete
- alteracao de status
- follow-up

Pronto quando:

- operacoes de orcamento estiverem rastreaveis por identificador

## Campos Recomendados Por Fluxo

### Auth

- `email` apenas se houver real necessidade operacional
- `user_id`
- `username`
- `role`
- `auth_action`
- `result`

### Importacao

- `preview_id`
- `file_name`
- `sheet_name`
- `rows_read`
- `rows_imported`
- `rows_updated`
- `rows_skipped`
- `catalogs_created`

### Orcamentos

- `budget_id`
- `budget_number`
- `year_budget`
- `actor_user_id`
- `actor_username`

## Exemplo De Log HTTP

```json
{
  "timestamp": "2026-06-14T15:20:10Z",
  "level": "INFO",
  "message": "request completed",
  "service": "go-gerenciador-orcamento-backend",
  "request_id": "req_01JXYZ",
  "method": "POST",
  "path": "/auth/login",
  "status_code": 200,
  "latency_ms": 18,
  "client_ip": "127.0.0.1"
}
```

## Exemplo De Log De Negocio

```json
{
  "timestamp": "2026-06-14T15:20:11Z",
  "level": "WARN",
  "message": "login blocked for inactive user",
  "service": "go-gerenciador-orcamento-backend",
  "request_id": "req_01JXYZ",
  "user_id": 17,
  "username": "guilherme",
  "auth_action": "login"
}
```

## Riscos Se Nao Houver Padrao

- logs duplicados e inconsistentes
- texto solto sem campo estruturado
- exposicao acidental de dados sensiveis
- dificuldade de correlacionar erro com requisicao
- alto custo de manutencao futura

## Riscos Da Implantacao

- excesso de logs sem criterio
- ruido em fluxos muito frequentes
- vazamento de dados sensiveis
- acoplamento indevido da camada de negocio ao framework HTTP

## Mitigacoes

- definir campos obrigatorios e proibidos
- revisar logs de auth com foco em seguranca
- manter logger injetado por contexto, nao por variavel global espalhada
- usar `INFO` e `WARN` com parcimonia
- nao logar payloads completos por padrao

## Testes Recomendados

Quando a implementacao comecar, recomendo validar:

- middleware gera `request_id`
- resposta carrega `X-Request-Id`
- request log inclui status e duracao
- logs de erro interno nao expõem detalhes ao cliente
- logs de auth nao incluem senha nem token

## Checklist De Implementacao

- criar `internal/logger`
- configurar logger em `cmd/main.go`
- adicionar middleware de request log
- adicionar `request_id`
- padronizar campos base
- instrumentar auth
- instrumentar importacao
- instrumentar orcamentos
- revisar redacao de dados sensiveis
- adicionar testes de middleware e contexto

## Decisao Recomendada

A recomendacao e seguir com a implantacao de logs estruturados no backend agora, em etapas, com esta ordem:

1. fundacao do logger
2. middleware HTTP com `request_id`
3. auth e seguranca
4. importacao
5. orcamentos e modulos auxiliares

Essa abordagem entrega valor rapido, documenta o padrao e evita que o projeto cresca sem observabilidade minima.
