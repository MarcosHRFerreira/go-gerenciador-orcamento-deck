# Estudo Tecnico Do Cadastro De Orcamentista

## Objetivo

Definir uma solucao para cadastrar `orcamentistas` no sistema, considerando que:

- eles precisam ter login no sistema
- eles devem permanecer com perfil tecnico `user`
- eles fazem orcamentos, mas nao fazem vendas
- eles nao devem ser tratados como vendedores

## Resumo Executivo

- Hoje o sistema so diferencia `admin` e `user`.
- Hoje o perfil `user` esta fortemente acoplado ao conceito de `vendedor`.
- Esse acoplamento acontece porque o escopo de acesso do `user` e resolvido pelo `username` vinculado a `salespeople`.
- O sistema ainda nao possui um cadastro proprio para `orcamentista`.
- O dominio de `budgets` tambem nao possui um campo proprio para identificar o `orcamentista` responsavel.
- Portanto, apenas criar usuarios com role `user` nao resolve corretamente a necessidade de negocio.

## Estado Atual Do Sistema

### Perfis De Usuario

Hoje o backend e o frontend aceitam apenas dois perfis:

- `admin`
- `user`

Isso aparece na validacao dos DTOs e tambem na tela de manutencao de usuarios.

### Escopo Atual Do Perfil User

Hoje o perfil `user` nao e apenas um usuario operacional generico.
Na pratica, ele e tratado como um usuario associado a `salespeople`.

Regra atual observada:

- se o usuario for `admin`, ele nao sofre restricao por vendedor
- se o usuario for `user`, o sistema tenta localizar um vendedor pelo `username`
- se encontrar vendedor ativo, restringe o escopo desse usuario aos dados do proprio vendedor
- se nao encontrar vendedor ativo, o escopo fica zerado

Conclusao:

- hoje o `user` foi desenhado com comportamento de vendedor
- isso nao combina com o papel de `orcamentista`

### Dominio Atual De Orcamento

Hoje o orcamento ja possui referencias como:

- `salesperson_id`
- `project_id`
- `contact_id`
- `installer_id`
- `status_id`

Mas nao existe hoje um campo proprio como:

- `estimator_id`
- `budget_owner_id`
- `orcamentista_id`

Isso significa que o sistema nao consegue responder, de forma nativa:

- quem elaborou o orcamento
- quem vendeu o orcamento
- quem deve editar orcamentos tecnicos sem ser vendedor

## Problema De Negocio

O `orcamentista` e um ator diferente do `vendedor`.

Diferencas principais:

- o vendedor responde por carteira, negociacao, conversao e dashboard comercial
- o orcamentista responde pela elaboracao e manutencao tecnica do orcamento
- um mesmo vendedor pode trabalhar com um ou mais orcamentistas
- um orcamentista pode apoiar varios vendedores

Se o sistema continuar tratando o `orcamentista` como `salesperson`, surgem problemas:

- semantica errada no dominio
- distorcao no dashboard comercial
- dificuldade para filtrar responsabilidade tecnica x responsabilidade comercial
- regras de permissao ficam confusas

## Alternativas Possiveis

## 1. Usar Apenas User Sem Novo Cadastro

### Como seria

- criar usuarios com role `user`
- nao criar nenhuma estrutura nova
- usar apenas o login para permitir acesso

### Vantagens

- menor custo imediato
- nenhuma migration nova no banco

### Desvantagens

- nao cria o conceito de `orcamentista`
- nao permite vincular o orcamento a quem o elaborou
- nao separa responsabilidade tecnica de responsabilidade comercial
- continua dependendo da logica atual pensada para vendedor

### Conclusao

- nao recomendado

## 2. Reaproveitar Salespeople Para Orcamentista

### Como seria

- cadastrar o orcamentista como se fosse vendedor
- vincular o usuario `user` a `salespeople`

### Vantagens

- aproveita o fluxo atual de restricao por `username`
- exige poucas mudancas tecnicas

### Desvantagens

- semantica incorreta
- polui dashboard e relatorios de vendas
- mistura papeis tecnicos com papeis comerciais
- dificulta evolucao futura de indicadores por area

### Conclusao

- nao recomendado

## 3. Solucao Recomendada: Cadastro Proprio De Orcamentista Mantendo Role User

### Como seria

- manter o papel tecnico de autenticacao como `user`
- criar um cadastro proprio para `orcamentista`
- vincular o usuario ao cadastro de orcamentista
- acrescentar o campo `orcamentista_id` em `budgets`

### Vantagens

- separa autenticacao de funcao de negocio
- preserva o uso de `admin` e `user` sem quebrar o modelo atual
- cria um conceito correto para o dominio
- permite filtros, relatorios e trilhas futuras por orcamentista
- nao contamina dashboards de vendedores

### Conclusao

- esta e a alternativa recomendada

## Proposta De Modelagem

## 1. Novo Cadastro De Orcamentista

### Tabela sugerida

```sql
estimators
- id
- code
- name
- email
- phone
- active
- notes
- user_id
- created_at
- updated_at
```

### Regras sugeridas

- `user_id` deve ser unico quando informado
- um `user` pode estar vinculado a um unico `orcamentista`
- um `orcamentista` pode existir antes do usuario, se necessario
- `code` pode ser gerado automaticamente, por exemplo `ORC-000001`

### Observacao

Se voce preferir um nome mais aderente ao sistema em portugues, a tabela tambem pode se chamar:

- `orcamentistas`

Tecnicamente eu recomendo manter consistencia com o padrao atual do projeto e escolher um unico idioma no dominio. Como o projeto hoje usa muitos nomes em ingles no backend, `estimators` e uma opcao coerente. Mas, se a prioridade for legibilidade do negocio, `orcamentistas` tambem e valida.

## 2. Novo Campo Em Budgets

### Proposta

Adicionar em `budgets`:

```sql
estimator_id BIGINT NULL
```

### Objetivo

Permitir registrar quem elaborou ou esta responsavel tecnicamente pelo orcamento.

### Beneficios

- separar `vendedor` de `orcamentista`
- permitir filtro por orcamentista
- permitir trilha de responsabilidade tecnica
- sustentar regras futuras de permissao

## 3. Vinculo Do Usuario Com Funcao De Negocio

### Proposta

Manter `users.role` apenas com:

- `admin`
- `user`

E acrescentar um subtipo funcional no proprio usuario ou no vinculo com o cadastro:

Alternativa A:

```sql
users.user_kind
- salesperson
- estimator
```

Alternativa B:

- inferir o tipo funcional pelo vinculo existente:
  - se estiver vinculado a `salespeople`, e vendedor
  - se estiver vinculado a `estimators`, e orcamentista

### Recomendacao

Recomendo a alternativa A, com um campo explicito como:

```sql
user_kind VARCHAR(30) NOT NULL
```

Valores iniciais:

- `salesperson`
- `estimator`

Motivos:

- deixa a regra mais clara
- reduz ambiguidade
- facilita validacoes no backend
- simplifica exibicao no frontend

## Regras De Negocio Recomendadas

## 1. Perfil Tecnico

- `admin` continua com acesso administrativo completo
- `user` continua sendo perfil operacional
- o que muda e o `tipo funcional` desse `user`

## 2. Comportamento Do User Orcamentista

Sugestao para `user` do tipo `estimator`:

- pode listar orcamentos
- pode editar orcamentos
- pode criar orcamentos, se isso fizer parte do processo
- nao pode excluir orcamentos
- nao pode acessar dashboard administrativo
- nao pode manter usuarios
- nao pode manter vendedores
- nao pode importar arquivos administrativos, salvo decisao explicita

## 3. Escopo De Acesso

Hoje o `user` usa escopo por vendedor.

Para `orcamentista`, ha duas alternativas:

### Opcao A. Escopo Pelo Proprio Orcamentista

- o `user` do tipo `estimator` enxerga apenas orcamentos com `estimator_id` vinculado a ele

### Opcao B. Escopo Mais Amplo De Operacao

- o `user` do tipo `estimator` enxerga todos os orcamentos, mas com acoes limitadas

### Recomendacao

Recomendo a `Opcao A` como padrao inicial:

- menor risco de seguranca
- responsabilidade mais clara
- mais facil de auditar

Se depois houver necessidade operacional, pode-se evoluir para:

- times de orcamento
- grupos de acesso
- compartilhamento por filial, empresa ou carteira

## Impactos Tecnicos

## 1. Banco De Dados

Mudancas sugeridas:

- criar tabela `estimators`
- adicionar `budgets.estimator_id`
- adicionar `users.user_kind`
- criar indices para busca por `estimator_id`
- criar constraints de unicidade e integridade

## 2. Backend

Impactos principais:

- DTOs de usuario precisam aceitar `user_kind`
- service de usuario precisa validar combinacoes entre `role` e `user_kind`
- novo modulo de cadastro de orcamentista:
  - handler
  - service
  - repository
  - DTOs
- DTOs e queries de orcamento precisam expor `estimator_id` e `estimator_name`
- a logica de escopo hoje baseada em `salespeople` precisa ser expandida

### Ajuste importante de governanca

Hoje o `ResolveRestrictedSalespersonID()` foi pensado so para vendedor.

Com orcamentista, sera necessario evoluir essa camada para algo como:

- resolver escopo por vendedor quando `user_kind = salesperson`
- resolver escopo por orcamentista quando `user_kind = estimator`

Ou seja:

- a governanca deve sair de um modelo "todo user e vendedor"
- e passar para um modelo "user operacional com subtipo funcional"

## 3. Frontend

Impactos principais:

- tela de usuarios deve permitir escolher:
  - `Perfil tecnico`: `admin` ou `user`
  - `Tipo funcional`: `vendedor` ou `orcamentista`
- criar tela de cadastro/manutencao de `orcamentistas`
- incluir campo `orcamentista` no formulario de orcamento
- incluir coluna e filtro por `orcamentista` na lista de orcamentos
- ajustar exibicao de menus conforme o tipo funcional

## 4. Dashboard E Relatorios

Como o orcamentista nao e vendedor:

- ele nao deve entrar nas metricas comerciais de conversao
- ele nao deve contaminar ranking de vendedores
- eventualmente pode existir um dashboard proprio de produtividade tecnica

Exemplos futuros de indicadores de orcamentista:

- quantidade de orcamentos elaborados
- tempo medio de preparacao
- orcamentos revisados por periodo
- orcamentos por vendedor atendido

## Sugestao De Telas

## 1. Cadastro De Orcamentistas

Campos sugeridos:

- codigo
- nome
- e-mail
- telefone
- ativo
- observacoes
- usuario vinculado

## 2. Ajuste Na Tela De Usuarios

Campos sugeridos:

- nome
- username
- e-mail
- role
- tipo funcional
- vinculo de negocio

Exemplo:

- role = `user`
- tipo funcional = `orcamentista`
- vinculo = `ORC-000012 - Joao Silva`

## 3. Ajuste Na Tela De Orcamento

Adicionar campo:

- `Orcamentista`

Esse campo deve ser separado de:

- `Vendedor`

## Regras De Validacao Recomendadas

- `admin` pode existir sem `user_kind`
- `user` deve obrigatoriamente possuir `user_kind`
- `user_kind = salesperson` exige vinculo com vendedor
- `user_kind = estimator` exige vinculo com orcamentista
- nao permitir o mesmo usuario operacional vinculado simultaneamente como vendedor e orcamentista

## Estrategia De Implantacao

## Fase 1. Estrutura De Dominio

- criar tabela de orcamentistas
- adicionar `budgets.estimator_id`
- adicionar `users.user_kind`
- ajustar DTOs e APIs

## Fase 2. Cadastro E Vinculo

- criar CRUD de orcamentistas
- ajustar tela de usuarios para vinculo funcional

## Fase 3. Orcamento

- incluir campo `orcamentista` no cadastro/edicao/lista
- incluir filtro por `orcamentista`
- ajustar permissoes por subtipo funcional

## Fase 4. Governanca

- evoluir o escopo operacional para suportar vendedor e orcamentista
- revisar regras de visibilidade e edicao

## Fase 5. Relatorios

- separar visao comercial e visao tecnica

## Recomendacao Final

Para atender corretamente a necessidade do cliente:

- manter `orcamentista` com role tecnico `user`
- nao reaproveitar `salespeople` para isso
- criar um cadastro proprio de `orcamentista`
- criar um subtipo funcional para `user`
- vincular o orcamento ao `orcamentista` por campo proprio

Essa abordagem resolve o problema atual e deixa o sistema preparado para crescer sem misturar:

- autenticacao
- permissao tecnica
- papel comercial
- responsabilidade de elaboracao do orcamento

## Proximo Passo Sugerido

Se voce quiser, o proximo passo natural e transformar este estudo em um plano de implementacao com:

- migrations
- endpoints
- telas
- regras de permissao
- cronograma por fases
