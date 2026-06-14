# Estudo Sistema Backend de Orcamentos

## Objetivo

Este documento consolida uma analise da planilha `RELATORIO DE ORCAMENTOS-25.xlsx` para orientar a construcao de um backend em Go inspirado no projeto `C:\Users\marco\OneDrive\Projects\go-delivery-routing-lab`, mas adaptado para `PostgreSQL` rodando em `Docker`.

O foco aqui e:

- identificar a entidade central do dominio
- propor entidades auxiliares e relacoes
- sugerir uma modelagem inicial de banco
- reaproveitar a estrutura e o estilo arquitetural do projeto de referencia
- indicar os primeiros modulos e endpoints da API

## Convencao De Idioma

Para este projeto, a decisao recomendada e:

- codigo-fonte, tabelas, colunas, DTOs, models, handlers, services e repositories em `ingles`
- documentacao funcional e tecnica principal em `portugues`
- manutencao de uma traducao da documentacao quando isso ajudar onboarding, compartilhamento ou padronizacao futura

Isso deixa o projeto alinhado com:

- o ecossistema Go
- o projeto de referencia `go-delivery-routing-lab`
- bibliotecas, exemplos e convencoes tecnicas do mercado

Exemplos:

- `budgets`
- `budget_number`
- `year_budget`
- `sent_at`
- `gross_value`
- `loss_reason_id`

## Autenticacao E Usuarios

Antes de iniciar a implementacao do dominio de orcamentos, eu recomendo incluir o modulo de autenticacao e usuarios como fundacao do sistema.

A referencia mais adequada para isso e o projeto `C:\Users\marco\OneDrive\Projects\go-tweets`, que ja trabalha com:

- cadastro de usuario
- login
- JWT de curta duracao
- refresh token persistido em banco
- middleware de autenticacao
- injecao de `userID` no contexto da requisicao

### Recomendacao Arquitetural

Para o projeto de orcamentos, a recomendacao e manter a mesma linha:

- `access token` via JWT
- `refresh token` persistido em banco
- middleware para validar token e preencher contexto autenticado
- uso de `SECRET_JWT` no `.env`
- protecao de endpoints por autenticacao e, quando necessario, por perfil

### Perfis De Usuario

Para o escopo atual, dois perfis sao suficientes:

- `admin`
- `user`

Sugestao de interpretacao:

- `admin`: pode manter cadastros auxiliares, usuarios, configuracoes e a operacao completa
- `user`: pode operar os fluxos principais do sistema conforme as permissoes definidas pela regra de negocio

### Entidades Recomendadas Para Autenticacao

#### `users`

Sugestao de colunas:

- `id`
- `name`
- `email`
- `username`
- `password_hash`
- `role`
- `active`
- `created_at`
- `updated_at`
- `deleted_at` opcional

Observacao:

- para este caso, eu nao criaria uma tabela `user_roles` separada neste primeiro momento
- como existem apenas dois perfis, `role` em `users` resolve melhor o inicio do projeto

#### `refresh_tokens`

Sugestao de colunas:

- `id`
- `user_id`
- `refresh_token`
- `expired_at`
- `created_at`
- `updated_at`
- `revoked_at` opcional

### Fluxo Recomendado

1. usuario e cadastrado
2. senha e persistida com hash usando `bcrypt`
3. login valida credenciais
4. sistema emite `access token`
5. sistema gera e persiste `refresh token`
6. middleware protege rotas autenticadas
7. endpoint de refresh emite novo `access token` e rotaciona o `refresh token`

## Leitura Da Planilha

Foram identificadas duas abas principais:

- `ORCAMENTOS`
- `FOLLOW-UP`

A aba `ORCAMENTOS` concentra o conjunto mais completo de dados. O cabecalho observado foi, em essencia:

- `DATA`
- `N DE ORCA`
- `REV`
- `INSTALADOR`
- `NOME DA OBRA`
- `TIPO DE OBRA`
- `VENDEDOR`
- `CONTATO`
- `VALOR BRUTO`
- `COMISSAO`
- `M2`
- `PRIORIDADE`
- `STATUS`
- `CONCORRENTE`
- `MOTIVO`
- `VALOR CONCORRENTE`
- `PROJETISTA`
- `ESPECIFICACOES`

A aba `FOLLOW-UP` parece representar uma visao comercial resumida dos mesmos orcamentos, com destaque para:

- data
- numero do orcamento
- cliente ou instalador
- obra
- vendedor
- contato
- valor
- status
- atualizacao
- especificacao

## Conclusoes De Dominio

A planilha nao representa apenas um cadastro simples de orcamentos. Ela mistura pelo menos quatro subdominios:

1. cadastro comercial do orcamento
2. acompanhamento do pipeline comercial
3. referencia tecnica da obra
4. inteligencia competitiva

Por isso, se tudo for colocado em uma unica tabela, o sistema tende a ficar fraco para evolucao, auditoria, relatorios e historico. O ideal e ter uma entidade central e algumas entidades auxiliares bem definidas.

## Entidade Central

A entidade central deve ser `budget` ou `orcamento`.

Ela representa o registro mestre do processo comercial. A maior parte das outras informacoes se conecta a ela.

### Campos sugeridos para `budgets`

Campos de negocio:

- `id`
- `code` ou `budget_number`
- `year_budget`
- `revision`
- `sent_at`
- `gross_value`
- `commission_value`
- `area_m2`
- `priority_id`
- `status_id`
- `installer_id`
- `project_id`
- `salesperson_id`
- `contact_id`
- `competitor_id`
- `loss_reason_id`
- `competitor_price`
- `designer_id`
- `specification_id`
- `current_follow_up`

Campos de controle:

- `created_at`
- `updated_at`
- `deleted_at` opcional para soft delete

### Mapeamento do que voce levantou

Os campos que voce sugeriu fazem bastante sentido e podem ser mapeados assim:

- `DATA_ENVIO` -> `sent_at`
- `ORCAMENTO` -> `budget_number`
- ano do orcamento -> `year_budget`
- `INSTALADOR` -> relacionamento com `installers`
- `OBRA` -> relacionamento com `projects`
- `TIPO_OBRA` -> relacionamento com `project_types`
- `VENDEDOR` -> relacionamento com `salespeople`
- `CONTATO` -> relacionamento com `contacts`
- `VALOR_BRUTO` -> `gross_value`
- `COMISSAO` -> `commission_value`
- `M2` -> `area_m2`
- `PRIORIDADE` -> relacionamento com `priorities`
- `STATUS` -> relacionamento com `budget_statuses`
- `CONCORRENTE` -> relacionamento com `competitors`
- `MOTIVO` -> relacionamento com `loss_reasons`
- `VALOR_CONCORRENTE` -> `competitor_price` ou campo textual auxiliar
- `PROJETISTA` -> relacionamento com `designers`
- `ESPECIFICACOES` -> relacionamento com `specifications`
- `DATACRIACAO` -> `created_at`
- `DATAALTERACAO` -> `updated_at`

## Entidades Auxiliares Recomendadas

### `users`

Essa entidade deve existir desde o inicio.

Ela sustenta autenticacao, autorizacao, auditoria e ownership das operacoes do sistema.

Sugestao de colunas:

- `id`
- `name`
- `email`
- `username`
- `password_hash`
- `role`
- `active`
- `created_at`
- `updated_at`
- `deleted_at` opcional

Observacao:

- `role` pode ser um `varchar` validado por regra de negocio ou um tipo enumerado no banco
- para este projeto, os valores iniciais recomendados sao `admin` e `user`

### `refresh_tokens`

Essa entidade tambem deve existir desde o inicio.

Ela viabiliza um fluxo mais seguro e controlado de renovacao de sessao, seguindo a referencia do `go-tweets`.

Sugestao de colunas:

- `id`
- `user_id`
- `refresh_token`
- `expired_at`
- `created_at`
- `updated_at`
- `revoked_at` opcional

### 1. `budget_statuses`

Essa tabela e obrigatoria.

A planilha mostra valores recorrentes como:

- `ORCAMENTO`
- `FECHADO`
- `CANCELADO`
- `PERDIDO`
- `COMPRA`

Sugestao de colunas:

- `id`
- `code`
- `name`
- `description`
- `is_final`
- `sort_order`
- `created_at`
- `updated_at`

Observacao importante:

`CANCELADO` e `PERDIDO` parecem semanticamente proximos, mas nao necessariamente iguais. Vale manter os dois status separados e deixar a regra de negocio decidir quando usar cada um.

### 2. `budget_status_history`

Essa tabela tambem e fortemente recomendada.

A planilha tem sinais claros de acompanhamento comercial, entao o sistema vai precisar de historico de mudanca de status. Sem isso, voce perde rastreabilidade.

Sugestao de colunas:

- `id`
- `budget_id`
- `from_status_id`
- `to_status_id`
- `changed_by`
- `change_reason`
- `created_at`

### 3. `loss_reasons`

Essa entidade tambem faz sentido.

Na planilha aparecem motivos como:

- `PRECO`
- `PRAZO`
- `PACOTE`
- `PACOTE + PRAZO`
- `PREFERENCIA`
- `QUALIDADE`
- `COND. PGTO`

Sugestao de colunas:

- `id`
- `code`
- `name`
- `description`
- `active`
- `created_at`
- `updated_at`

Observacao:

`loss_reasons` deve ser usada principalmente quando o orcamento entrar em estado como `PERDIDO` ou `CANCELADO`.

### 4. `designers`

Essa entidade faz bastante sentido.

Na planilha, `PROJETISTA` possui alta repeticao e se comporta como um cadastro de parceiro, escritorio ou referencia tecnica.

Sugestao de colunas:

- `id`
- `name`
- `document`
- `email`
- `phone`
- `notes`
- `created_at`
- `updated_at`

### 5. `project_types`

Essa entidade deve existir.

A planilha sugere categorias como:

- `HOSPITALAR/FARMACEUTICO`
- `COMERCIAL/LOJA`
- `ATACADISTA/SUPER`
- `RESIDENCIA`
- `ACADEMIA`
- `LOGISTICA/GALPAO`

Sugestao de colunas:

- `id`
- `code`
- `name`
- `description`
- `created_at`
- `updated_at`

### 6. `specifications`

Essa entidade e recomendada.

Os dados encontrados se comportam como classificacoes tecnicas padronizaveis, por exemplo:

- `MPU`
- `N/E`
- `ALUPIR`
- `CHAPA DE ACO GALVANIZADO`

Sugestao de colunas:

- `id`
- `code`
- `name`
- `description`
- `category`
- `created_at`
- `updated_at`

### 7. `competitors`

Essa entidade e recomendada.

Foram observados concorrentes recorrentes, como:

- `MULTIVAC`
- `ARCOTEC`

Sugestao de colunas:

- `id`
- `name`
- `notes`
- `created_at`
- `updated_at`

### 8. `priorities`

Mesmo que a planilha nao esteja totalmente consistente neste ponto, a entidade vale a pena se a aplicacao for usada operacionalmente.

Sugestao:

- `id`
- `code`
- `name`
- `weight`
- `created_at`
- `updated_at`

Exemplo de valores:

- `low`
- `medium`
- `high`
- `urgent`

### 9. `salespeople`

Essa entidade deve existir.

Na planilha existe um conjunto pequeno e bem repetido de vendedores. Isso e claramente um cadastro de usuario comercial ou responsavel de carteira.

Sugestao de colunas:

- `id`
- `name`
- `email`
- `phone`
- `active`
- `created_at`
- `updated_at`

### 10. `installers`

Essa entidade deve existir.

O campo `INSTALADOR` na pratica parece ser um cliente B2B, empresa instaladora ou integradora.

Sugestao de colunas:

- `id`
- `name`
- `document`
- `email`
- `phone`
- `city`
- `state`
- `notes`
- `created_at`
- `updated_at`

### 11. `contacts`

Essa entidade e importante.

O contato nao deve ficar solto como texto diretamente em `budgets`, porque o mesmo instalador pode ter varios contatos.

Sugestao de colunas:

- `id`
- `installer_id`
- `name`
- `email`
- `phone`
- `role`
- `is_primary`
- `created_at`
- `updated_at`

### 12. `projects`

Essa entidade e importante e deve representar a obra.

Sugestao de colunas:

- `id`
- `name`
- `project_type_id`
- `city`
- `state`
- `notes`
- `created_at`
- `updated_at`

Observacao:

Em alguns casos a mesma obra pode gerar mais de um orcamento ou revisoes. Por isso, `projects` deve ser separada de `budgets`.

## Constraints `Unique` Recomendadas

Para esse dominio, vale definir algumas unicidades logo no inicio para evitar duplicidade cadastral e inconsistencias operacionais.

### `budgets`

Sugestao principal:

- `unique (budget_number, year_budget)`

Observacao:

- essa composicao atende ao seu criterio de negocio para identificar o numero do orcamento dentro do ano
- `revision` continua sendo um atributo funcional, mas nao entra na regra principal de unicidade neste momento

### `users`

Sugestao:

- `unique (email)`
- `unique (username)`

Observacao:

- `role` nao precisa ser `unique`
- se o login for apenas por email, `username` ainda assim vale como identificador interno opcional

### `refresh_tokens`

Sugestao:

- `unique (refresh_token)`

Observacao:

- se voce quiser manter apenas um refresh token ativo por usuario, tambem pode avaliar `unique (user_id)` com rotacao simples

### `budget_statuses`

Sugestao:

- `unique (code)`
- `unique (name)`

### `loss_reasons`

Sugestao:

- `unique (code)`
- `unique (name)`

### `designers`

Sugestao:

- `unique (name)`

Observacao:

- em cenarios com alto risco de homonimo, pode ser melhor trocar para `unique (document)` quando houver documento confiavel

### `project_types`

Sugestao:

- `unique (code)`
- `unique (name)`

### `specifications`

Sugestao:

- `unique (code)`

Observacao:

- se `code` nao existir no inicio, pode usar `unique (name)` temporariamente

### `competitors`

Sugestao:

- `unique (name)`

### `priorities`

Sugestao:

- `unique (code)`
- `unique (name)`

### `salespeople`

Sugestao:

- `unique (email)` quando existir
- `unique (name)` apenas se o cadastro for pequeno e controlado

Observacao:

- como vendedor pode ter homonimo, a unicidade por `name` depende da sua operacao

### `installers`

Sugestao:

- `unique (document)` quando houver CNPJ ou CPF confiavel

Observacao:

- eu nao recomendo `unique (name)` aqui porque a planilha tende a ter variacao de grafia e nomes comerciais parecidos

### `contacts`

Sugestao:

- `unique (installer_id, email)` quando email existir
- `unique (installer_id, phone)` quando telefone existir

Observacao:

- nao faz sentido `unique (name)` global, porque nomes de contato se repetem entre empresas

### `projects`

Sugestao:

- nao criar `unique` por `name` sozinho no inicio

Observacao:

- a mesma obra pode aparecer com pequenas variacoes de nome, etapas diferentes ou unidades diferentes
- se futuramente voce tiver dados melhores, pode avaliar `unique (name, city, state)` ou uma chave de negocio mais robusta

### 13. `budget_follow_ups`

Essa e uma das entidades adicionais mais importantes.

A aba `FOLLOW-UP` indica que a operacao comercial precisa registrar atualizacoes sucessivas, nao apenas o status atual.

Sugestao de colunas:

- `id`
- `budget_id`
- `status_id`
- `note`
- `next_contact_at`
- `created_by`
- `created_at`

Isso evita guardar tudo em um unico campo de texto dentro de `budgets`.

### 14. `budget_revisions`

Como a planilha tem coluna `REV`, vale separar revisoes se isso fizer parte real do processo.

Sugestao de colunas:

- `id`
- `budget_id`
- `revision_number`
- `gross_value`
- `commission_value`
- `area_m2`
- `specification_id`
- `sent_at`
- `created_at`

Se a revisao nao for relevante na operacao, esse conceito pode ficar apenas como coluna simples em `budgets`. Mas eu recomendaria manter no backlog.

## Modelo Relacional Inicial

Sugestao de modelagem inicial:

```text
salespeople 1---n budgets
installers 1---n contacts
installers 1---n budgets
projects 1---n budgets
project_types 1---n projects
budget_statuses 1---n budgets
budget_statuses 1---n budget_status_history
loss_reasons 1---n budgets
competitors 1---n budgets
designers 1---n budgets
specifications 1---n budgets
priorities 1---n budgets
budgets 1---n budget_follow_ups
budgets 1---n budget_status_history
budgets 1---n budget_revisions
```

## Tabelas Minimas Para O MVP

Se o objetivo for comecar rapido sem perder estrutura, eu faria o MVP com estas tabelas:

- `budgets`
- `budget_statuses`
- `budget_follow_ups`
- `installers`
- `contacts`
- `projects`
- `project_types`
- `salespeople`
- `loss_reasons`
- `designers`
- `specifications`
- `competitors`
- `priorities`

## Tabela `budgets` Sugerida

Exemplo conceitual:

```sql
create table budgets (
  id bigserial primary key,
  budget_number varchar(50) not null,
  year_budget integer not null,
  revision varchar(20),
  sent_at date,
  installer_id bigint not null references installers(id),
  project_id bigint references projects(id),
  salesperson_id bigint not null references salespeople(id),
  contact_id bigint references contacts(id),
  gross_value numeric(14,2) not null default 0,
  commission_value numeric(14,2) not null default 0,
  area_m2 numeric(14,2),
  priority_id bigint references priorities(id),
  status_id bigint not null references budget_statuses(id),
  competitor_id bigint references competitors(id),
  loss_reason_id bigint references loss_reasons(id),
  competitor_price_text varchar(255),
  competitor_price numeric(14,2),
  designer_id bigint references designers(id),
  specification_id bigint references specifications(id),
  current_follow_up text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  constraint uq_budgets_budget_number_year unique (budget_number, year_budget)
);
```

Observacoes:

- `competitor_price_text` e util porque a planilha mistura valores numericos e observacoes como `PLACA 145,00`
- `competitor_price` pode ser preenchido apenas quando houver valor limpo
- `budget_number` sozinho nao deve ser unico
- a unicidade recomendada passa a ser `budget_number + year_budget`

## Regras De Negocio Relevantes

### 1. Identidade do orcamento

Sugestao:

- unicidade por `budget_number + year_budget`
- `year_budget` pode ser derivado de `sent_at` ou informado explicitamente, conforme a sua regra operacional
- se revisao vier vazia, assumir `0` ou `A` conforme a regra comercial definida

Observacao:

- se no futuro o mesmo numero puder existir duas vezes no mesmo ano por causa de revisao formal, ai vale migrar a regra para `budget_number + year_budget + revision`

### 2. Identidade do usuario

Sugestao:

- usuario deve ter `email` unico
- usuario deve ter `username` unico
- senha nunca deve ser persistida em texto puro
- `password_hash` deve ser gerado com `bcrypt`

### 3. Autenticacao

Sugestao:

- login gera `access token` JWT de curta duracao
- refresh token deve ser persistido em banco
- refresh token deve ser rotacionado a cada renovacao
- endpoints protegidos devem depender de middleware de autenticacao

### 4. Autorizacao por perfil

Sugestao:

- rotas administrativas devem aceitar apenas `admin`
- rotas operacionais podem aceitar `admin` e `user`
- toda operacao sensivel deve conhecer o `user_id` autenticado

### 5. Mudanca de status

Toda mudanca importante deve gravar historico em `budget_status_history`.

### 6. Motivo de perda

`loss_reason_id` deve ser obrigatorio quando o status final indicar perda, cancelamento ou substituicao por concorrente.

### 7. Follow-up

Cada atualizacao comercial deve ser gravada em `budget_follow_ups`, e o campo `current_follow_up` em `budgets` deve representar apenas o ultimo resumo.

### 8. Contato

Um `contact` deve pertencer a um `installer`. O orcamento referencia um contato especifico usado naquela negociacao.

### 9. Obra

Uma `project` pode existir sem todos os dados completos no inicio. O cadastro deve aceitar informacoes progressivas.

## Estrutura De Projeto Recomendada

A base deve seguir de perto o `go-delivery-routing-lab`, mantendo a separacao por camadas:

```text
cmd/
  main.go
db/
  migrations/
docs/
internal/
  apperror/
  config/
  dto/
  handler/
    auth/
    budget/
    installer/
    project/
    salesperson/
    user/
    health/
  httpresponse/
  middleware/
  model/
  repository/
    auth/
    budget/
    installer/
    project/
    salesperson/
    user/
  server/
  service/
    auth/
    budget/
    installer/
    project/
    salesperson/
    user/
pkg/
  internalsql/
    jwt/
    refreshtoken/
test/
  integration/
  unit/
```

## Bibliotecas Recomendadas

Com base no projeto de referencia, eu manteria:

- `github.com/gin-gonic/gin`
- `github.com/go-playground/validator/v10`
- `github.com/golang-jwt/jwt/v5`
- `github.com/joho/godotenv`
- `golang.org/x/crypto/bcrypt`

Para PostgreSQL, eu substituiria a camada de acesso ao banco para:

- `github.com/jackc/pgx/v5/stdlib`

Opcionalmente, para migrations:

- `github.com/golang-migrate/migrate/v4`

Se quiser seguir o estilo mais proximo do projeto de referencia, ainda da para continuar com `database/sql`, mudando apenas o driver.

## Adaptacao Do Projeto De Referencia Para PostgreSQL

No `go-delivery-routing-lab`, hoje existe:

- `Gin` para HTTP
- `validator` para validacao
- `database/sql` com driver MySQL
- `docker-compose.yml` para banco local
- migrations SQL em `db/migrations`

Para o novo projeto, a adaptacao recomendada e:

1. manter `cmd/main.go`
2. manter `internal/server/router.go`
3. manter `internal/config/config.go`
4. adicionar `internal/middleware` para autenticacao e autorizacao
5. adicionar `pkg/internalsql/jwt` e `pkg/internalsql/refreshtoken` inspirados no `go-tweets`
6. trocar `pkg/internalsql/mysql.go` por `pkg/internalsql/postgres.go`
7. trocar `docker-compose.yml` para `postgres`
8. criar migrations novas do dominio de orcamentos e autenticacao

## Exemplo De Variaveis De Ambiente

```env
PORT=8080
SECRET_JWT=local-dev-secret
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=budget_management
DATABASE_URL=postgres://postgres:postgres@localhost:5432/budget_management?sslmode=disable
```

## Exemplo De Docker Compose Para PostgreSQL

```yaml
services:
  db:
    image: postgres:17
    container_name: db_budget_management
    environment:
      POSTGRES_DB: budget_management
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## Modulos Prioritarios Da API

Eu sugiro implementar nesta ordem:

1. `health`
2. `budget_statuses`
3. `salespeople`
4. `installers`
5. `contacts`
6. `project_types`
7. `projects`
8. `designers`
9. `specifications`
10. `competitors`
11. `loss_reasons`
12. `budgets`
13. `budget_follow_ups`
14. `budget_status_history`
15. `users`
16. `auth`

## Endpoints Iniciais Sugeridos

### Health

- `GET /check-health`

### Auth

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`

### Users

- `POST /users`
- `GET /users`
- `GET /users/:user_id`
- `PUT /users/:user_id`
- `PATCH /users/:user_id/status`

### Budgets

- `POST /budgets`
- `GET /budgets`
- `GET /budgets/:budget_id`
- `PUT /budgets/:budget_id`
- `PATCH /budgets/:budget_id/status`
- `DELETE /budgets/:budget_id`

### Follow-ups

- `POST /budgets/:budget_id/follow-ups`
- `GET /budgets/:budget_id/follow-ups`

### Cadastros auxiliares

- `GET /budget-statuses`
- `POST /budget-statuses`
- `GET /loss-reasons`
- `POST /loss-reasons`
- `GET /designers`
- `POST /designers`
- `GET /installers`
- `POST /installers`
- `GET /projects`
- `POST /projects`
- `GET /users`
- `POST /users`

## Filtros Importantes Para Listagem

O modulo de orcamentos vai precisar nascer com filtros desde o inicio. Os principais sao:

- por status
- por vendedor
- por instalador
- por projetista
- por tipo de obra
- por periodo de envio
- por faixa de valor
- por concorrente
- por prioridade

## Cuidados De Modelagem

### 1. Nomes inconsistentes

A planilha mostra variacoes de escrita, por exemplo:

- `ARCO +` e `ARCO+`
- `PRECO` e `PREAO` por problema de encoding
- `Cliente Negociando` e `CLIENTE NEGOCIANDO`

Sera necessario um processo de saneamento no momento da carga inicial.

### 2. Dados misturados

`VALOR_CONCORRENTE` mistura numero e texto. Portanto o sistema deve aceitar ambos durante a migracao de dados.

### 3. Historico perdido em coluna unica

A coluna de atualizacao da aba `FOLLOW-UP` nao deveria permanecer como dado unico dentro de `budgets`. Isso deve virar uma tabela historica.

### 4. Cadastro duplicado

Instaladores, projetistas, especificacoes e contatos provavelmente possuem duplicidade por variacao de grafia. Isso exigira regras de consolidacao.

## Proposta De Fases

### Fase 1 - Fundacao

- bootstrap da API
- config
- conexao PostgreSQL
- autenticacao JWT
- refresh token persistido em banco
- health check
- migrations base

### Fase 2 - Cadastros auxiliares

- budget statuses
- loss reasons
- users
- salespeople
- installers
- contacts
- project types
- projects
- designers
- specifications
- competitors

### Fase 3 - Orcamentos

- CRUD de budgets
- filtros
- alteracao de status
- historico de status

### Fase 4 - Follow-up comercial

- CRUD ou append-only de follow-up
- timeline por orcamento
- proxima data de contato

### Fase 5 - Importacao da planilha

- normalizacao dos valores
- carga inicial
- conciliacao de duplicidades

## Recomendacao Final De Modelagem

Se eu tivesse que resumir a decisao principal:

- `budgets` deve ser a entidade central
- `users` e `refresh_tokens` devem nascer juntos com a fundacao do projeto
- `budget_statuses`, `loss_reasons`, `designers` e `budget_follow_ups` devem existir como entidades proprias
- `installers`, `contacts`, `projects`, `project_types`, `salespeople`, `competitors`, `specifications` e `priorities` tambem valem muito a pena
- o sistema deve guardar historico, nao apenas o estado atual

## Melhor Recorte Para Comecar

Para um primeiro ciclo de implementacao, eu recomendo subir este escopo:

1. `health`
2. `budget_statuses`
3. `loss_reasons`
4. `designers`
5. `users`
6. `auth`
7. `salespeople`
8. `installers`
9. `contacts`
10. `project_types`
11. `projects`
12. `specifications`
13. `budgets`
14. `budget_follow_ups`

Com isso, voce ja consegue:

- cadastrar a estrutura base do dominio
- criar orcamentos de forma consistente
- acompanhar andamento comercial
- preparar o terreno para importacao da planilha

## Proximo Passo Recomendado

O proximo passo ideal e transformar este estudo em backlog tecnico com:

- definicao das tabelas SQL
- ordem de migrations
- DTOs
- models
- repositories
- services
- handlers
- testes unitarios e de integracao

Se quiser, o proximo passo pode ser eu gerar:

1. o `backlog tecnico completo`
2. as `migrations iniciais PostgreSQL`
3. o `esqueleto do projeto Go` seguindo o padrao do `go-delivery-routing-lab`
