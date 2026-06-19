# Cronograma Tecnico Do Cadastro De Orcamentista

## Objetivo

Transformar o estudo de `orcamentista` em um plano de implementacao executavel, considerando:

- manutencao do perfil tecnico `user`
- criacao de cadastro proprio de `orcamentista`
- separacao entre responsabilidade comercial e responsabilidade tecnica
- ajuste de permissao, escopo e telas do sistema

## Premissas

- `admin` continua como perfil administrativo completo
- `user` continua como perfil operacional
- `orcamentista` nao deve ser tratado como `vendedor`
- o `orcamentista` precisa poder trabalhar em orcamentos sem contaminar dashboards comerciais
- o dominio deve separar `salesperson` de `estimator`

## Decisao Funcional Base

Este cronograma assume a seguinte diretriz de negocio:

- `user` do tipo `estimator` pode criar e editar orcamentos
- `user` do tipo `estimator` nao pode excluir orcamentos
- `user` do tipo `estimator` nao participa do dashboard comercial
- `user` do tipo `estimator` deve ter escopo proprio por `orcamentista`
- `vendedor` e `orcamentista` permanecem como papeis distintos dentro do orcamento

## Entregavel Final

Ao fim da implementacao, o sistema deve permitir:

- cadastrar `orcamentistas`
- vincular um `user` a um `orcamentista`
- registrar o `orcamentista` responsavel no orcamento
- filtrar e listar orcamentos por `orcamentista`
- aplicar governanca correta para `user` vendedor e `user` orcamentista
- manter dashboards comerciais limpos, sem misturar papel tecnico e comercial

## Visao Geral Das Fases

| Fase | Nome | Objetivo principal | Dependencia |
| --- | --- | --- | --- |
| 1 | Modelagem | Preparar banco e contrato de dominio | nenhuma |
| 2 | Backend base | Criar CRUD e contratos de API | fase 1 |
| 3 | Governanca | Separar escopo de vendedor x orcamentista | fase 2 |
| 4 | Frontend cadastro | Criar telas e vinculos de cadastro | fase 2 |
| 5 | Orcamentos | Incluir orcamentista na operacao do orcamento | fases 2, 3 e 4 |
| 6 | Ajustes gerenciais | Revisar dashboards e relatorios | fase 5 |
| 7 | Homologacao | Validar fluxos e preparar entrada em producao | fases 1 a 6 |

## Fase 1. Modelagem

### Objetivo

Preparar a estrutura de banco e o contrato de dominio para suportar `orcamentista`.

### Tasks

- criar migration da tabela `estimators`
- criar migration para adicionar `budgets.estimator_id`
- criar migration para adicionar `users.user_kind`
- definir constraints de unicidade:
  - `estimators.code`
  - `estimators.user_id`, quando preenchido
- criar foreign keys:
  - `estimators.user_id -> users.id`
  - `budgets.estimator_id -> estimators.id`
- criar indices para:
  - `budgets.estimator_id`
  - `estimators.user_id`
  - `estimators.active`

### Entregaveis

- migrations versionadas
- banco preparado para o novo dominio

### Criterio De Aceite

- banco sobe sem erro
- migrations rodam em ambiente limpo e em ambiente ja existente
- rollback conceitual validado

### Risco

- risco medio por alteracao estrutural em `users` e `budgets`

## Fase 2. Backend Base

### Objetivo

Criar o modulo de `orcamentista` no backend e adaptar contratos de usuario.

### Tasks

- criar DTOs de `orcamentista`
- criar model de `orcamentista`
- criar repository de `orcamentista`
- criar service de `orcamentista`
- criar handler e rotas de `orcamentista`
- ajustar DTOs de `users` para aceitar `user_kind`
- ajustar service de `users` para validar:
  - `role = user` exige `user_kind`
  - `user_kind = estimator` nao pode ser tratado como vendedor
- expor `estimator_id` e `estimator_name` nos DTOs de `budgets`

### Endpoints Sugeridos

- `GET /estimators`
- `GET /estimators/:id`
- `POST /estimators`
- `PUT /estimators/:id`
- `DELETE /estimators/:id`

### Entregaveis

- API funcional de `orcamentistas`
- API de `users` preparada para tipo funcional

### Criterio De Aceite

- CRUD de `orcamentista` funcionando
- validacoes de negocio aplicadas no backend
- testes de unidade e integracao cobrindo fluxos principais

### Risco

- risco medio por mudar contrato de usuario

## Fase 3. Governanca

### Objetivo

Separar a logica de escopo operacional hoje acoplada a `salespeople`.

### Tasks

- revisar `ResolveRestrictedSalespersonID()`
- criar resolucao de escopo por tipo funcional
- implementar regra:
  - `salesperson` usa escopo por vendedor
  - `estimator` usa escopo por orcamentista
- ajustar services de `budgets`
- ajustar services de `dashboard` para ignorar `estimator` em metrica comercial
- revisar regras de acesso das rotas relacionadas a:
  - listagem de orcamentos
  - edicao de orcamentos
  - catalogos auxiliares

### Entregaveis

- governanca desacoplada de vendedor
- `user` operacional corretamente classificado por funcao

### Criterio De Aceite

- `user` vendedor so enxerga o proprio escopo comercial
- `user` orcamentista so enxerga o proprio escopo tecnico
- `admin` continua com visao total

### Risco

- risco alto, porque esta fase mexe no nucleo de permissao

## Fase 4. Frontend Cadastro

### Objetivo

Criar a estrutura visual para manutencao de `orcamentistas` e ajuste do cadastro de usuarios.

### Tasks

- criar feature `estimators` no frontend
- criar lista de `orcamentistas`
- criar tela de criacao
- criar tela de edicao
- criar detalhe, se desejado
- ajustar tela de `users` para incluir:
  - `tipo funcional`
  - vinculo com `orcamentista`
  - vinculo com `vendedor`, se aplicavel
- ajustar menu e rotas para acesso administrativo

### Entregaveis

- CRUD visual de `orcamentista`
- manutencao de usuarios adaptada ao novo modelo

### Criterio De Aceite

- administrador consegue criar e vincular `user` a `orcamentista`
- formularios validam corretamente
- mensagens e labels ficam coerentes com o dominio

### Risco

- risco medio, principalmente por impacto em UX e consistencia de formularios

## Fase 5. Orcamentos

### Objetivo

Trazer `orcamentista` para o fluxo real de trabalho do orcamento.

### Tasks

- incluir campo `orcamentista` no cadastro de orcamento
- incluir campo `orcamentista` na edicao de orcamento
- incluir coluna `orcamentista` na lista de orcamentos
- incluir filtro por `orcamentista`
- ajustar payloads e queries de `budgets`
- garantir que `user` do tipo `estimator` consiga:
  - criar
  - editar
  - listar
- garantir que `user` do tipo `estimator` nao consiga excluir

### Entregaveis

- orcamento vinculado a `orcamentista`
- operacao do usuario tecnico funcionando ponta a ponta

### Criterio De Aceite

- `orcamentista` aparece no formulario e na lista
- filtros funcionam
- escopo operacional respeita o tipo funcional

### Risco

- risco medio, por impacto direto no fluxo principal do sistema

## Fase 6. Ajustes Gerenciais

### Objetivo

Evitar mistura entre papeis tecnicos e comerciais nos dashboards e relatorios.

### Tasks

- revisar dashboard comercial
- garantir que `orcamentista` nao alimente ranking de vendedores
- revisar exportacoes `CSV` e `XLSX`
- avaliar inclusao de visao tecnica futura por `orcamentista`

### Entregaveis

- dashboard comercial preservado
- relatorios consistentes

### Criterio De Aceite

- nenhum `orcamentista` aparece como vendedor
- dados tecnicos e comerciais ficam semanticamente corretos

### Risco

- risco baixo a medio

## Fase 7. Homologacao

### Objetivo

Fechar a entrega com validacao funcional e seguranca operacional.

### Tasks

- testar criacao de `orcamentista`
- testar vinculo `user -> orcamentista`
- testar criacao de orcamento por `orcamentista`
- testar edicao de orcamento por `orcamentista`
- testar bloqueio de exclusao
- testar que `vendedor` e `orcamentista` nao se confundem
- testar que dashboard comercial continua coerente
- revisar mensagens e documentacao

### Entregaveis

- pacote homologado
- checklist de entrada em producao

### Criterio De Aceite

- fluxos principais aprovados
- fluxos alternativos cobertos
- sem regressao nas regras atuais de `admin`

### Risco

- risco baixo, desde que fases anteriores estejam estabilizadas

## Ordem De Execucao Recomendada

1. Fase 1. Modelagem
2. Fase 2. Backend base
3. Fase 3. Governanca
4. Fase 4. Frontend cadastro
5. Fase 5. Orcamentos
6. Fase 6. Ajustes gerenciais
7. Fase 7. Homologacao

## Dependencias Criticas

- a Fase 3 depende da Fase 2, porque a governanca precisa conhecer `user_kind`
- a Fase 5 depende da Fase 3, porque o orcamento so fica seguro depois da nova governanca
- a Fase 6 depende da Fase 5, porque os dados do novo dominio precisam estar estabilizados

## Backlog Tecnico Resumido

### Banco

- tabela `estimators`
- campo `budgets.estimator_id`
- campo `users.user_kind`

### Backend

- CRUD de `estimators`
- vinculacao `user <-> estimator`
- resolucao de escopo funcional
- ajuste dos DTOs de `budgets`

### Frontend

- telas de `estimators`
- ajuste das telas de `users`
- ajuste da lista e formulario de `budgets`

### Relatorios

- separar visao comercial de visao tecnica

## Testes Recomendados

### Unidade

- validacao de `user_kind`
- validacao de vinculo com `orcamentista`
- validacao de escopo por tipo funcional

### Integracao

- CRUD de `estimators`
- acesso de `user estimator`
- bloqueio de exclusao de orcamento
- filtros por `orcamentista`

### E2E Manual

- cadastro do `orcamentista`
- vinculacao ao usuario
- login como `orcamentista`
- criacao e edicao de orcamento
- conferencias de permissoes

## Riscos Principais

- manter o acoplamento antigo de `user = vendedor`
- misturar `orcamentista` no dashboard comercial
- abrir permissao demais para `user`
- nao revisar todos os pontos onde hoje `salesperson` e assumido como unico papel operacional

## Recomendacao Final

O melhor caminho e executar esse tema em blocos curtos, mas na ordem certa:

- primeiro estrutura
- depois governanca
- depois telas
- por fim relatorios e homologacao

Isso reduz retrabalho e evita que o sistema crie um cadastro novo sem conseguir usar o novo ator de forma coerente no fluxo de orcamentos.

## Proximo Passo Sugerido

Se voce quiser, o proximo passo natural agora e eu transformar este cronograma em:

- `fase 1`
- `fase 2`
- `fase 3`

ja no formato de execucao tecnica, para irmos implementando por etapas no projeto.
