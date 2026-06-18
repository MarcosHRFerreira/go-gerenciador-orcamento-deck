# Estudo Tecnico Das Alteracoes Solicitadas

## Objetivo

Consolidar o impacto tecnico das seguintes mudancas solicitadas a partir da planilha `RESUMO ORCAMENTOS CONCORRENCIA.xlsx` e do sistema atual:

- criar tabela auxiliar para `Linha de produtos` da Trox
- gravar o codigo dessa linha em `budgets`
- acrescentar em `budgets` o campo `Construtora`
- iniciar toda importacao com status `Em Negociacao`
- trocar em todo o sistema o nome `Projeto` para `Obra`
- criar tabela auxiliar de `Obra` com `codigo` e `descricao`
- criar tela para cadastro e manutencao de `Obra`
- criar filtro por periodo
- aumentar o limite de linhas por consulta
- trocar em todo o sistema e nas tabelas o nome `designer` para `projetista`
- permitir que o perfil `user` tambem edite orcamentos, mas sem permissao de excluir

## Resumo Executivo

- A planilha da Trox ja traz as colunas `Linha de produtos`, `Obra`, `Nome Cliente` e `Status`.
- Hoje a importacao Trox ignora `Linha de produtos`, `Codigo Cliente`, `Nome Cliente` e `Tipo` no dominio principal.
- Hoje a importacao Trox usa `Obra` para alimentar `projects` e grava em `budgets.project_id`.
- Hoje o status principal do orcamento importado e definido como `Nao informado`, enquanto o `Status` da planilha vai para `current_follow_up`.
- O backend ja possui filtro por periodo via `sent_at_from` e `sent_at_to`, mas o frontend ainda nao expoe esse filtro.
- O limite maximo atual no backend ja e `100` por consulta; nao foi encontrado limite atual de `50` no backend. No frontend, a listagem principal usa `20` por padrao.
- O sistema ja possui backend de `projects`, mas no frontend existe apenas detalhe de projeto e associacao de orcamentos, sem uma tela completa de cadastro e manutencao.
- Hoje a edicao e a exclusao de orcamentos estao restritas ao fluxo administrativo; a nova solicitacao exige abrir apenas a edicao para `user`, mantendo a exclusao somente para `admin`.

## Evidencias Da Planilha

Analise da aba `Capa`:

- total de linhas de dados observadas: `35`
- valores distintos em `Linha de produtos`: `4`
- amostras de `Linha de produtos`:
  - `DAMPER CORTA-FOGO`
  - `DIFUSAO A/F`
  - `FAN-COIL`
  - `FILTROS`
- valores distintos em `Obra`: `31`
- valores distintos em `Nome Cliente`: `30`
- valor observado em `Status`: `Informado`

Colunas relevantes da planilha:

| Coluna da planilha  | Uso atual                       | Proposta                                                                                                |
| ------------------- | ------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `Linha de produtos` | ignorada no dominio principal   | criar tabela auxiliar e gravar referencia em `budgets`                                                  |
| `Nome Cliente`      | ignorada no dominio principal   | usar como base do novo campo `Construtora`, sujeito a validacao de negocio                              |
| `Obra`              | usada para alimentar `projects` | renomear conceito para `Obra` e manter associacao em `budgets`                                          |
| `Status`            | vai para `current_follow_up`    | iniciar `status` principal como `Em Negociacao`; decidir se o valor da planilha continua como follow-up |
| `Contato`           | usado no dominio principal      | manter                                                                                                  |
| `Vendedor`          | usado no dominio principal      | manter                                                                                                  |
| `Instalador`        | usado no dominio principal      | manter                                                                                                  |

## Estado Atual Do Sistema

### Importacao Trox

Hoje a importacao Trox:

- reconhece a aba `Capa`
- usa `Obra` para preencher `projectName`
- define `projectTypeName` como `Nao informado`
- define `statusName` como `Nao informado`
- envia o valor da coluna `Status` da planilha para `currentFollowUp`
- ignora `Tipo`, `Linha de produtos`, `Codigo Cliente`, `Nome Cliente` e `Fator Medio` no dominio principal

Isso significa que a solicitacao afeta diretamente o parser Trox e o fluxo de importacao.

### Dominio Atual

Hoje a tabela `budgets` possui:

- `project_id`
- `designer_name`
- nao possui `product_line_id`
- nao possui `construction_company` ou equivalente

Hoje a tabela `projects` possui:

- `id`
- `name`
- `project_type_id`
- `city`
- `state`
- `notes`

Ou seja:

- o conceito atual de `Projeto` ja funciona como a futura `Obra`
- a tabela auxiliar de `Obra` solicitada pode ser implementada reaproveitando a estrutura atual, com renomeacao e inclusao de `codigo`

### Filtros E Limites

Hoje:

- a API de listagem ja aceita `sent_at_from` e `sent_at_to`
- o frontend da listagem de orcamentos ainda nao expoe esses dois filtros
- o backend aplica `page_size` padrao `20`
- o backend permite no maximo `100`
- a listagem principal do frontend usa `20`
- a visao agrupada por projeto/obra usa carga paginada em blocos de `100`

### Permissoes Na Tela De Orcamento

Hoje:

- a rota de edicao de orcamento no frontend esta dentro de `AdminRoute`
- a exclusao de orcamento tambem esta dentro do fluxo administrativo
- no backend, as rotas administrativas de `budgets` usam restricao de perfil para operacoes sensiveis
- os testes de integracao atuais ja validam que o perfil `user` nao pode excluir

Conclusao:

- a permissao de exclusao para `user` ja esta corretamente bloqueada e deve permanecer assim
- a permissao de edicao para `user` precisa ser aberta de forma controlada
- a liberacao deve respeitar o escopo atual por vendedor para evitar que um `user` edite orcamentos fora do proprio alcance

## Proposta De Solucao

## 1. Linha De Produtos Trox

### Objetivo

Criar catalogo auxiliar para armazenar as linhas de produto da Trox e gravar a referencia em `budgets`.

### Proposta De Banco

Criar tabela:

```sql
product_lines
- id
- code
- name
- description
- created_at
- updated_at
```

Adicionar em `budgets`:

```sql
product_line_id BIGINT NULL
```

### Regra De Importacao

- ao importar da Trox, ler `Linha de produtos`
- normalizar o valor
- localizar ou criar o registro em `product_lines`
- gravar o `id` em `budgets.product_line_id`
- gerar `code` automatico a partir do nome normalizado, se a planilha nao trouxer codigo proprio

### Observacao

A solicitacao menciona "gravar o codigo na tabela de budgets". Tecnicamente, ha duas alternativas:

- gravar `product_line_id` em `budgets` e obter o codigo por join
- gravar tambem um campo denormalizado como `product_line_code`

Recomendacao:

- usar `product_line_id` como referencia principal
- expor `code` no DTO/resposta da API
- evitar duplicar `code` em `budgets`, salvo se houver necessidade explicita de historico imutavel

## 2. Campo Construtora Em Budgets

### Objetivo

Acrescentar em `budgets` a informacao de `Construtora`.

### Fonte Na Planilha

A melhor candidata na planilha analisada e `Nome Cliente`.

### Proposta De Banco

Adicionar em `budgets`:

```sql
construction_company VARCHAR(200) NOT NULL DEFAULT ''
```

### Regra De Importacao

- mapear `Nome Cliente` para `construction_company`
- manter como texto simples nesta primeira fase

### Alternativa Futura

Se houver necessidade de governanca de cadastros, evoluir depois para tabela auxiliar de construtoras:

```sql
construction_companies
- id
- code
- name
- created_at
- updated_at
```

Neste momento, para menor custo de implementacao, o campo textual em `budgets` parece suficiente.

## 3. Status Inicial Em Negociacao

### Objetivo

Toda importacao deve iniciar com status principal `Em Negociacao`.

### Impacto

Hoje o parser Trox seta `statusName = Nao informado`.

### Proposta

- alterar a importacao para definir `statusName = Em Negociacao`
- manter a coluna `Status` da planilha como `current_follow_up`, se isso continuar fazendo sentido para o processo comercial

### Dependencia

Garantir que exista registro em `budget_statuses` com nome `Em Negociacao`.

### Recomendacao De Migracao

Criar migration idempotente para:

- inserir `Em Negociacao` em `budget_statuses`, se nao existir
- opcionalmente posicionar `sort_order`

## 4. Projeto Para Obra

### Objetivo

Trocar em todo o sistema o nome `Projeto` para `Obra`.

### Escopo Tecnico

A mudanca afeta:

- tabela `projects`
- rota `/projects`
- DTOs, models, repositories e services
- nomes de campos como `project_id`, `project_name`, `project_type_id`
- textos de tela, filtros, labels, mensagens e documentacao
- frontend em `features/projects`

### Duas Estrategias Possiveis

#### Estrategia A. Renomeacao Completa

Renomear estruturalmente:

- tabela `projects` para `works`
- `project_id` para `work_id`
- rotas `/projects` para `/works`
- tipos, DTOs e componentes para `work`

Vantagens:

- dominio fica coerente
- elimina ambiguidade futura

Riscos:

- custo alto
- impacto em testes, integracoes e queries
- maior risco de regressao

#### Estrategia B. Renomeacao Funcional Com Compatibilidade

Manter estrutura interna atual por enquanto e renomear:

- labels de tela para `Obra`
- documentacao e textos de API
- nomes novos apenas nas novas telas e contratos evoluidos

Vantagens:

- menor risco
- entrega mais rapida

Riscos:

- coexistencia temporaria de nomes tecnicos antigos e nomes funcionais novos

### Recomendacao

Executar em duas fases:

1. fase curta: renomeacao funcional na interface e documentacao
2. fase controlada: renomeacao estrutural de banco e API

## 5. Tabela Auxiliar De Obra Com Codigo E Descricao

### Objetivo

Ter tabela auxiliar de `Obra` com `codigo` e `descricao`.

### Reaproveitamento Do Modelo Atual

A tabela `projects` ja representa esse conceito. A melhor evolucao e:

- renomear semanticamente `projects` para `works`
- trocar `name` por `description`
- adicionar `code`

### Proposta De Estrutura

```sql
works
- id
- code VARCHAR(50) NOT NULL UNIQUE
- description VARCHAR(200) NOT NULL
- project_type_id BIGINT NULL
- city VARCHAR(100) NOT NULL DEFAULT ''
- state VARCHAR(50) NOT NULL DEFAULT ''
- notes TEXT NOT NULL DEFAULT ''
- created_at
- updated_at
```

### Decisao Importante

Definir como a importacao Trox vai preencher `code` da obra:

- opcao 1: gerar codigo automatico a partir da descricao
- opcao 2: manter `code` obrigatorio apenas no cadastro manual
- opcao 3: extrair padroes do texto da coluna `Obra` quando houver identificadores embutidos como `DVR 1121-2026`

Recomendacao:

- permitir codigo automatico na importacao
- permitir manutencao manual posterior

## 6. Tela De Cadastro E Manutencao De Obra

### Estado Atual

Hoje existe:

- backend CRUD de `projects`
- frontend apenas com detalhe da entidade e vinculacao de orcamentos

### Necessidade

Criar no frontend:

- tela de listagem de obras
- tela de criacao
- tela de edicao
- opcionalmente exclusao com validacao de vinculos

### Funcionalidades Minimas

- campos `codigo` e `descricao`
- filtros por `codigo` e `descricao`
- cadastro e edicao
- acesso restrito a admin
- navegacao a partir do menu lateral

### Recomendacao

Aproveitar o backend atual como base e evoluir:

- endpoints para suportar `code`
- frontend novo em `features/works` ou adaptar `features/projects`

## 7. Filtro Por Periodo

### Estado Atual

O backend ja suporta:

- `sent_at_from`
- `sent_at_to`

### Lacuna Atual

O frontend ainda nao envia esses filtros na tela principal de orcamentos.

### Proposta

Adicionar na listagem:

- filtro `Periodo inicial`
- filtro `Periodo final`
- persistencia em query string

### Conclusao

Aqui nao ha necessidade de alterar regra de negocio no backend, apenas:

- expor o filtro no frontend
- ajustar tipos e `buildBudgetListParams`

## 8. Aumentar O Limite De Linhas Por Consulta

### Estado Atual

Nao foi identificado limite atual de `50` no backend.

Estado encontrado:

- padrao no backend: `20`
- maximo permitido no backend: `100`
- padrao no frontend principal: `20`
- visoes auxiliares usam `100`

### Interpretacao Possivel Do Pedido

O pedido provavelmente se refere a:

- aumentar o volume padrao exibido por consulta
- ou aumentar o maximo permitido na experiencia do usuario

### Recomendacao

Definir explicitamente:

- novo padrao da listagem: `50` ou `100`
- novo maximo permitido: `200`, se houver necessidade operacional

Sugestao equilibrada:

- padrao no frontend: `50`
- maximo no backend: `200`

Isso exige:

- alterar validacao do backend
- alterar valor padrao do frontend
- incluir seletor de quantidade por pagina

## 8.1. Permitir Que User Edite Orcamento, Sem Excluir

### Objetivo

Permitir que o perfil `user` tambem consiga editar a linha do orcamento, mantendo a exclusao restrita a `admin`.

### Estado Atual

Hoje a edicao e a exclusao estao concentradas no fluxo administrativo.

No comportamento atual:

- o frontend protege a rota de edicao com `AdminRoute`
- a exclusao aparece nas acoes administrativas da listagem
- o backend possui protecao de perfil nas rotas administrativas
- os testes de integracao ja garantem que `user` nao pode excluir

### Proposta

Separar claramente as permissoes:

- `GET /budgets`: `admin` e `user`
- `GET /budgets/:id`: `admin` e `user`, respeitando escopo
- `PUT /budgets/:id`: `admin` e `user`, respeitando escopo
- `DELETE /budgets/:id`: somente `admin`

### Regras De Negocio Recomendadas

- `user` pode editar apenas orcamentos dentro do seu escopo por vendedor
- `user` nao pode excluir orcamentos
- o botao de excluir deve continuar oculto ou desabilitado para `user`
- o botao de editar deve ficar disponivel para `user` quando o item estiver no escopo permitido

### Impacto Tecnico

Backend:

- revisar `handler` de `budgets` para permitir `PUT` fora do bloco exclusivo de `admin`, mantendo validacao de escopo
- garantir que o `service` de update valide o alcance do `user`
- manter `DELETE` como operacao exclusiva de `admin`

Frontend:

- mover a rota `/budgets/:budgetId/edit` para area autenticada comum, e nao mais exclusiva de `AdminRoute`
- ajustar botoes e acoes da listagem para `user` ver `Editar` e nao ver `Excluir`
- revisar a tela de edicao para que catálogos e campos continuem funcionando para o perfil `user`

Testes:

- criar teste de integracao para garantir que `user` consegue editar orcamento do proprio escopo
- criar teste de integracao para garantir que `user` nao consegue editar orcamento fora do proprio escopo
- manter teste de integracao garantindo que `user` nao exclui

## 9. Designer Para Projetista

### Objetivo

Trocar em todo o sistema e tambem nas tabelas o nome `designer` para `projetista`.

### Escopo Tecnico

Afeta:

- coluna `budgets.designer_name`
- DTOs com `designer_name`
- filtros por `designer_name`
- importacao Trox
- testes unitarios e integracao
- frontend inteiro onde aparece `designerName`

### Estrategia Recomendada

Fazer em duas etapas:

1. compatibilidade de API
2. renomeacao fisica

#### Etapa 1. Compatibilidade

- manter coluna fisica temporariamente
- expor novo campo `project_designer_name` ou `projetista_name` na API
- aceitar tanto `designer_name` quanto `projetista_name` por um periodo
- trocar labels de tela para `Projetista`

#### Etapa 2. Renomeacao Fisica

- renomear coluna `designer_name` para `project_designer_name` ou `projetista_name`
- renomear filtros, repositorios, models e testes

### Recomendacao De Nome

Como o sistema esta em portugues, a melhor opcao funcional e:

```sql
projetista_name
```

## Proposta De Backlog Tecnico

### Fase 1. Banco E Backend

- criar tabela `product_lines`
- adicionar `product_line_id` em `budgets`
- adicionar `construction_company` em `budgets`
- garantir status `Em Negociacao`
- adicionar `code` na entidade de obra/projeto
- criar compatibilidade inicial para `projetista`
- preparar ajuste de permissao para `user` editar orcamento com validacao de escopo

### Fase 2. Importacao Trox

- mapear `Linha de produtos` para catalogo auxiliar
- mapear `Nome Cliente` para `construction_company`
- iniciar status principal como `Em Negociacao`
- manter `Status` da planilha como follow-up atual, salvo regra futura contraria
- ajustar testes de importacao

### Fase 3. Frontend

- trocar labels de `Projeto` para `Obra`
- trocar labels de `Designer` para `Projetista`
- expor filtro por periodo
- alterar tamanho padrao de pagina
- criar telas de listagem, criacao e edicao de obra
- liberar a edicao de orcamento para `user`
- manter a exclusao visivel apenas para `admin`

### Fase 4. Refatoracao Estrutural

- renomear rotas, tabelas e contratos internos de `project` para `obra/work`
- renomear coluna fisica `designer_name`
- revisar toda documentacao

## Impacto Em Banco De Dados

### Novas Estruturas Minimas

```sql
product_lines
- id
- code
- name
- description
- created_at
- updated_at
```

### Alteracoes Em Budgets

```sql
budgets
- add product_line_id BIGINT NULL
- add construction_company VARCHAR(200) NOT NULL DEFAULT ''
- rename designer_name -> projetista_name
```

### Alteracoes Em Projects/Obras

```sql
projects
- add code VARCHAR(50) NULL/UNIQUE
- avaliar rename name -> description
```

## Impacto Em Testes

Sera necessario revisar:

- testes unitarios da importacao Trox
- testes unitarios do servico de orcamentos
- testes de integracao de listagem com filtro
- testes de integracao de CRUD da futura tela de obra
- testes de serializacao e compatibilidade de DTO para `projetista`

## Riscos

- renomeacao estrutural ampla de `Projeto` para `Obra` pode quebrar rotas, joins, filtros e telas existentes
- renomeacao fisica de `designer_name` exige migracao cuidadosa e revisao total do frontend/backend
- gravar `Construtora` a partir de `Nome Cliente` depende de validacao funcional, pois o campo pode representar cliente comercial e nao necessariamente construtora em todos os casos
- a geracao automatica de `code` para obra e linha de produto precisa de regra deterministica para evitar duplicidades

## Recomendacao Final

Executar a implementacao nesta ordem:

1. `Linha de produtos` como catalogo auxiliar + referencia em `budgets`
2. `Construtora` em `budgets` a partir de `Nome Cliente`
3. status inicial `Em Negociacao` na importacao
4. filtro por periodo no frontend
5. aumento do padrao de itens por consulta
6. liberar `user` para editar orcamento sem permitir exclusao
7. tela de cadastro/manutencao de `Obra`
8. renomeacao funcional de `Projeto` para `Obra`
9. renomeacao funcional de `Designer` para `Projetista`
10. somente depois, renomeacoes fisicas de tabelas e colunas

## Cronograma Proposto

Cronograma sugerido em sprints curtas, considerando entregas incrementais e com menor risco de regressao.

### Sprint 1. Base De Dados E Contratos

Tasks:

- criar migration da tabela `product_lines`
- criar migration para adicionar `product_line_id` em `budgets`
- criar migration para adicionar `construction_company` em `budgets`
- criar migration para garantir `Em Negociacao` em `budget_statuses`
- criar migration para adicionar `code` na entidade atual de `projects`
- definir estrategia de compatibilidade para `designer` e `projetista`
- revisar DTOs e contratos de API impactados

Entregaveis:

- banco preparado para as novas informacoes
- contratos de backend definidos para a proxima fase

### Sprint 2. Importacao Trox

Tasks:

- alterar parser Trox para ler `Linha de produtos`
- localizar ou criar catalogo de linha de produto na importacao
- gravar referencia da linha de produto em `budgets`
- mapear `Nome Cliente` para `construction_company`
- alterar status principal importado para `Em Negociacao`
- manter ou revisar uso do `Status` da planilha como `current_follow_up`
- atualizar testes unitarios da importacao
- atualizar testes de integracao relacionados a importacao

Entregaveis:

- importacao Trox aderente ao novo mapeamento
- cobertura de testes atualizada

### Sprint 3. Orcamentos E Permissoes

Tasks:

- revisar backend para permitir `PUT` de orcamento por `user` no proprio escopo
- manter `DELETE` exclusivo para `admin`
- mover rota de edicao do frontend para area autenticada comum
- ajustar botoes da listagem para `user` editar e nao excluir
- revisar mensagens e comportamento da tela de edicao
- criar testes de integracao para editar com `user`
- criar testes negativos para bloquear exclusao por `user`

Entregaveis:

- `user` pode editar orcamentos permitidos
- exclusao continua protegida

### Sprint 4. Filtros E Experiencia De Consulta

Tasks:

- expor filtro por periodo na listagem de orcamentos
- persistir filtros de periodo na query string
- ajustar `buildBudgetListParams` para enviar `sent_at_from` e `sent_at_to`
- alterar tamanho padrao de pagina para o valor aprovado
- aumentar limite maximo do backend, se confirmado
- adicionar seletor de quantidade por pagina
- validar performance da listagem

Entregaveis:

- tela de orcamentos com filtro por periodo
- consulta com paginacao revisada

### Sprint 5. Obra

Tasks:

- definir se a renomeacao sera funcional ou estrutural nesta fase
- adicionar `code` e ajustar cadastro da entidade atual de `projects`
- trocar labels de `Projeto` para `Obra` no frontend
- criar tela de listagem de obras
- criar tela de cadastro de obra
- criar tela de edicao de obra
- incluir entrada de menu para obra
- revisar associacao de orcamentos com obra

Entregaveis:

- cadastro e manutencao de obra disponiveis
- navegacao da aplicacao atualizada

### Sprint 6. Projetista E Padronizacao Terminologica

Tasks:

- trocar labels de `Designer` para `Projetista`
- atualizar filtros, formularios e listagens no frontend
- implementar compatibilidade de API para `projetista`
- revisar impacto em repositorios, DTOs e responses
- atualizar testes automatizados
- revisar documentacao funcional e tecnica

Entregaveis:

- sistema apresentado com nomenclatura de `Projetista`
- compatibilidade preservada durante a transicao

### Sprint 7. Refatoracao Estrutural Final

Tasks:

- renomear fisicamente tabela e contratos de `projects` para `works` ou equivalente aprovado
- renomear fisicamente coluna `designer_name`
- ajustar rotas e repositorios
- atualizar toda a suite de testes
- executar validacao completa de regressao

Entregaveis:

- base tecnica alinhada ao dominio final
- reducao de nomes legados no codigo e no banco

## Lista Consolidada De Tasks

- validar semanticamente se `Nome Cliente` representa `Construtora`
- definir estrategia do codigo da `Linha de produtos`
- definir estrategia do codigo da `Obra`
- criar migrations de banco
- adaptar importacao Trox
- adaptar DTOs e responses
- liberar edicao de orcamento para `user` com escopo
- manter exclusao exclusiva de `admin`
- expor filtro por periodo no frontend
- revisar paginacao e limite de consulta
- criar CRUD completo de `Obra`
- renomear labels de `Projeto` para `Obra`
- renomear labels de `Designer` para `Projetista`
- implementar compatibilidade de nomenclatura no backend
- revisar e expandir testes unitarios e de integracao
- atualizar documentacao

## Itens Que Precisam De Confirmacao

- `Nome Cliente` deve mesmo virar `Construtora` em todos os casos?
- o valor da coluna `Status` da planilha continua indo para `follow-up atual` ou deve alimentar outra estrutura?
- o codigo da `Obra` sera manual, automatico ou extraido da propria descricao?
- o codigo de `Linha de produtos` sera gerado automaticamente ou havera tabela oficial de codigos?
- o nome fisico desejado para a coluna de `designer` e `projetista_name`?
- o novo limite por consulta deve ser `50`, `100` ou `200`?
