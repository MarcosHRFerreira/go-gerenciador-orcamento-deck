# Especificacao Funcional: Comunicacao Interna

## Objetivo

Definir como deve funcionar o modulo de `Comunicacao Interna` do sistema, considerando:

- o envio de avisos entre usuarios
- o envio de comunicados gerais pelo `admin`
- a existencia de historico consultavel
- a separacao entre comunicacao institucional e conversas operacionais

Este documento complementa:

- `docs/ESTUDO-AVISOS-E-CONVERSAS-ENTRE-USUARIOS.md`

## Resultado Esperado

Ao final da implementacao completa, o sistema deve permitir:

- que o `admin` publique avisos gerais para todos os usuarios
- que o `admin` publique avisos direcionados para usuarios especificos
- que usuarios consultem sua caixa de avisos com status de leitura
- que usuarios troquem mensagens com historico dentro do sistema
- que o sistema mostre indicadores de nao lidos
- que a comunicacao fique centralizada na aplicacao

## Estrategia Recomendada

O modulo deve ser dividido em duas partes:

- `Avisos`
- `Conversas`

Motivo:

- `Avisos` resolvem comunicados oficiais
- `Conversas` resolvem troca operacional entre usuarios

## Escopo Recomendado

## Fase 1

Entregar o minimo funcional com maior valor:

- central de `Avisos`
- `admin` pode criar aviso geral
- `admin` pode criar aviso direcionado
- usuario pode listar avisos recebidos
- usuario pode marcar aviso como lido
- badge de avisos nao lidos

## Fase 2

Completar o nucleo conversacional:

- `Conversas` diretas entre usuarios
- historico de mensagens
- indicador de nao lidas
- lista de conversas recentes

## Fase 3

Melhorias e expansoes:

- anexos
- grupos
- vinculo com obra
- vinculo com orcamento
- tempo real
- auditoria mais detalhada

## Nomenclatura Recomendada

Para manter a linguagem clara no sistema:

- `Comunicacao`
- `Avisos`
- `Conversas`
- `Nao lido`
- `Lido`
- `Aviso geral`
- `Aviso direcionado`

Evitar usar na interface termos como:

- `notice`
- `message thread`
- `chat interno`

Esses nomes podem existir no backend, mas nao devem aparecer para o usuario final.

## Regras De Acesso

## Quem acessa o modulo

- usuarios com `role = admin`
- usuarios com `role = user`

## Avisos

### Quem pode criar

- `admin` pode criar `Aviso geral`
- `admin` pode criar `Aviso direcionado`

### Quem nao pode criar na primeira fase

- `user` nao cria `Aviso geral`
- `user` nao cria `Aviso direcionado`

### Quem pode visualizar

- todos os usuarios autenticados visualizam os avisos que receberam

### Quem pode marcar como lido

- todo usuario autenticado pode marcar como lido os avisos recebidos por ele

## Conversas

### Regra recomendada para a primeira versao

- `admin` pode iniciar conversa com qualquer usuario
- `user` pode iniciar conversa com `admin`

### Regra opcional para fase futura

- `user` pode iniciar conversa com outro `user`

Motivo para restringir no inicio:

- reduz ruido
- simplifica moderacao
- facilita o controle operacional

## Estrutura De Navegacao Recomendada

## Menu lateral

Adicionar um novo item:

- `Comunicacao`

Sugestao de rota principal:

- `/communication`

## Estrutura interna da tela

Usar abas:

- `Avisos`
- `Conversas`

Motivo:

- reduz poluicao do menu
- concentra tudo em uma central unica

## Topbar

Adicionar indicadores visuais:

- icone de sino para avisos nao lidos
- icone de mensagens para conversas nao lidas

Se a implementacao inicial precisar ser mais simples, pode comecar apenas com:

- badge numerico no item `Comunicacao`

## Estrutura De Rotas Recomendada

### Rotas protegidas

- `/communication`
- `/communication/notices`
- `/communication/conversations`
- `/communication/conversations/:conversationId`

### Observacao

Para a primeira versao, pode haver apenas uma rota:

- `/communication`

Com abas internas controladas pela interface.

## Estrutura Funcional Da Tela De Comunicacao

## Cabecalho

Deve conter:

- titulo `Comunicacao`
- subtitulo curto explicando que a area centraliza avisos e conversas

Exemplo de subtitulo:

- `Acompanhe comunicados internos e troque mensagens com historico.`

## Aba Avisos

Deve conter:

- lista de avisos recebidos
- filtros simples
- destaque para aviso fixado
- destaque para prioridade
- status de leitura

## Aba Conversas

Deve conter:

- lista de conversas recentes
- indicador de nao lidas
- painel de historico
- composer de mensagem

## Estado Vazio

### Avisos

Quando nao houver avisos:

- `Nenhum aviso encontrado no momento.`

### Conversas

Quando nao houver conversas:

- `Nenhuma conversa iniciada ainda.`

## Especificacao Funcional De Avisos

## Objetivo De Negocio

Permitir que o `admin` publique comunicados oficiais para um ou mais usuarios dentro do sistema.

## Tipos De Aviso

### Aviso geral

Enviado para todos os usuarios ativos.

Exemplos:

- parada programada
- mudanca de processo
- novo fluxo implantado

### Aviso direcionado

Enviado para usuarios especificos.

Exemplos:

- pedido de revisao
- orientacao operacional
- comunicacao a uma equipe restrita

## Estrutura Da Lista De Avisos

## Colunas ou informacoes principais

- titulo
- resumo da mensagem
- prioridade
- criado por
- data de envio
- status `Lido` ou `Nao lido`
- indicacao de fixado

## Como exibir

- `Nao lido`: destaque visual mais forte
- `Lido`: destaque mais neutro
- `Prioridade critica`: cor de alerta
- `Fixado`: aparece antes na lista

## Filtros recomendados

- todos
- nao lidos
- lidos
- prioridade

Para a primeira entrega, os filtros minimos recomendados sao:

- todos
- nao lidos
- lidos

## Tela Ou Painel De Detalhe Do Aviso

Ao abrir um aviso, exibir:

- titulo completo
- mensagem completa
- autor
- data
- prioridade
- se e geral ou direcionado
- data de expiracao, se houver

Comportamento esperado:

- ao abrir o aviso, ele pode ser automaticamente marcado como `Lido`
- ou o usuario pode clicar em `Marcar como lido`

Minha recomendacao:

- marcar automaticamente ao abrir o detalhe

## Tela De Criacao De Aviso

Disponivel apenas para `admin`.

## Campos recomendados

- titulo
- mensagem
- tipo de destinatario
- usuarios destinatarios
- prioridade
- fixado
- data de expiracao

## Tipo de campo

- `titulo`: texto curto
- `mensagem`: texto longo
- `tipo de destinatario`: select
- `usuarios destinatarios`: autocomplete multiplo
- `prioridade`: select
- `fixado`: checkbox
- `data de expiracao`: data e hora opcional

## Valores padrao recomendados

- `tipo de destinatario = geral`
- `prioridade = informativo`
- `fixado = false`

## Validacoes

- titulo obrigatorio
- mensagem obrigatoria
- prioridade obrigatoria
- se for direcionado, deve haver pelo menos um destinatario

## Comportamento ao salvar

1. criar o aviso
2. registrar os destinatarios
3. exibir mensagem de sucesso
4. voltar para a lista de avisos
5. atualizar badge de nao lidos para os destinatarios

## Acoes Recomendadas Sobre O Aviso

Na primeira versao:

- `Marcar como lido`
- `Ver detalhe`

Na versao do `admin`:

- `Publicar novo aviso`
- `Editar`, opcional
- `Encerrar` ou `Expirar`, opcional

Minha recomendacao para o inicio:

- nao permitir edicao de aviso ja publicado
- se precisar corrigir, publicar novo aviso

Motivo:

- simplifica auditoria
- evita duvida sobre o que foi alterado depois do envio

## Regras Operacionais De Avisos

### 1. Aviso geral so pode ser criado por admin

### 2. Aviso direcionado da primeira fase tambem deve ficar restrito a admin

### 3. Avisos expirados nao devem aparecer por padrao

### 4. Avisos fixados aparecem antes dos demais

### 5. Avisos nao devem ser apagados fisicamente na primeira fase

Recomendacao:

- manter historico
- ocultar ou expirar em vez de excluir

## Especificacao Funcional De Conversas

## Objetivo De Negocio

Permitir troca de mensagens com historico dentro do proprio sistema.

## Tipo De Conversa Na Primeira Versao

- `Conversa direta`

Sem grupos na primeira versao.

## Estrutura Da Tela De Conversas

A tela pode ser dividida em duas areas:

- lista lateral de conversas
- painel principal da conversa atual

## Lista lateral de conversas

Deve mostrar:

- nome da outra pessoa
- trecho da ultima mensagem
- data da ultima mensagem
- indicador de nao lidas

## Painel principal da conversa

Deve mostrar:

- nome do participante
- historico completo
- mensagens ordenadas por data
- composer para enviar nova mensagem

## Composer

Campos minimos:

- textarea de mensagem
- botao `Enviar`

Validacoes:

- mensagem obrigatoria
- remover espacos vazios no inicio e no fim
- nao permitir envio vazio

## Comportamento ao enviar

1. registrar a mensagem
2. atualizar a lista de conversas
3. manter a conversa aberta
4. atualizar o indicador de nao lidas do outro participante

## Comportamento ao abrir conversa

Ao abrir uma conversa:

- carregar o historico
- marcar como lida para o usuario atual

## Como iniciar uma conversa

Na primeira versao, recomendar duas entradas:

- botao `Nova conversa` dentro da aba `Conversas`
- atalho contextual em telas futuras

## Tela De Nova Conversa

Campos recomendados:

- selecionar usuario
- escrever primeira mensagem

Validacoes:

- participante obrigatorio
- mensagem obrigatoria
- nao permitir selecionar a si proprio

## Regras Operacionais De Conversas

### 1. Nao criar conversa duplicada entre os mesmos dois usuarios

Se ja existir conversa direta entre os dois:

- abrir a conversa existente
- nao criar outra

### 2. Usuario so pode ver conversas onde e participante

### 3. Historico nao deve ser apagado fisicamente na primeira fase

### 4. Conversa sem nova mensagem continua disponivel no historico

### 5. Nao permitir edicao de mensagem na primeira fase

Motivo:

- simplifica implementacao
- preserva rastreabilidade

## Badge E Indicadores

## Indicadores recomendados

- total de avisos nao lidos
- total de conversas com mensagens nao lidas

## Onde exibir

Opcao recomendada:

- no item `Comunicacao` do menu

Opcao futura:

- tambem no topo da aplicacao

## Atualizacao

Na primeira versao:

- atualizar por consulta HTTP periodica
- ou atualizar ao navegar e ao executar acoes

Nao e necessario tempo real na primeira entrega.

## Integracao Com Outras Areas

## Integracao futura com Obra

Permitir, no futuro:

- abrir conversa a partir da obra
- avisos com referencia de obra

## Integracao futura com Orcamento

Permitir, no futuro:

- abrir conversa a partir de um orcamento
- mensagem contextual ligada a orcamento

Minha recomendacao:

- isso deve entrar apenas depois que a central de comunicacao basica estiver estabilizada

## Contratos De API Recomendados

## Avisos

### Criar aviso

- `POST /notices`

### Listar avisos do usuario

- `GET /notices`

### Obter detalhe do aviso

- `GET /notices/:noticeId`

### Marcar como lido

- `PATCH /notices/:noticeId/read`

## Conversas

### Listar conversas

- `GET /conversations`

### Criar conversa direta

- `POST /conversations`

### Listar mensagens

- `GET /conversations/:conversationId/messages`

### Enviar mensagem

- `POST /conversations/:conversationId/messages`

### Marcar como lida

- `PATCH /conversations/:conversationId/read`

## Criterio De Pronto Da Primeira Entrega

A primeira entrega pode ser considerada pronta quando:

- o `admin` consegue publicar aviso geral
- o `admin` consegue publicar aviso direcionado
- o usuario autenticado consegue listar seus avisos
- o usuario autenticado consegue abrir e marcar aviso como lido
- o sistema mostra a quantidade de avisos nao lidos
- o usuario comum nao consegue publicar aviso geral

## Criterio De Pronto Da Segunda Entrega

A segunda entrega pode ser considerada pronta quando:

- `admin` consegue iniciar conversa com qualquer usuario
- `user` consegue iniciar conversa com `admin`
- o historico fica disponivel
- mensagens nao lidas sao identificadas
- a abertura da conversa marca leitura corretamente

## Recomendacoes Tecnicas

### Recomendacao 1

Implementar primeiro `Avisos`.

Motivo:

- gera valor rapido
- resolve o caso mais claro de negocio
- exige menos regras que conversas

### Recomendacao 2

Implementar `Conversas` em seguida, mas sem tempo real.

### Recomendacao 3

Criar uma feature dedicada:

```text
frontend/src/features/communication/
  api/
  components/
  pages/
  types/
```

### Recomendacao 4

No backend, manter separacao por dominio:

```text
internal/model/
internal/repository/notice/
internal/repository/conversation/
internal/service/notice/
internal/service/conversation/
internal/handler/notice/
internal/handler/conversation/
```

## Sequencia Recomendada De Implementacao

### Etapa 1

- migrations de avisos
- endpoints de avisos
- listagem de avisos
- badge de nao lidos

### Etapa 2

- criacao de aviso pelo admin
- detalhe e leitura

### Etapa 3

- migrations de conversas
- listagem de conversas
- envio de mensagens

### Etapa 4

- regras de leitura
- indicadores de nao lidas

### Etapa 5

- melhorias de UX
- filtros mais completos
- contexto com obra e orcamento

## Conclusao

O melhor caminho para este modulo e criar uma `Central de Comunicacao` com duas partes:

- `Avisos`
- `Conversas`

Para a primeira entrega, o ideal e focar em `Avisos`, porque isso ja resolve o envio geral do `admin` e cria a base de leitura, historico e notificacao.

Depois disso, o sistema pode evoluir com seguranca para `Conversas` sem precisar refazer a experiencia principal.
