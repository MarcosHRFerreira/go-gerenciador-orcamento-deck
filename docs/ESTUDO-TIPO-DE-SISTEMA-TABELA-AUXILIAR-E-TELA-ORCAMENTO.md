# Estudo: Tipo de Sistema como Tabela Auxiliar e Exibicao na Tela de Orcamento

## Objetivo

Garantir que o campo `Tipo de Sistema` seja tratado como catalogo auxiliar oficial do sistema e que essa informacao fique visivel e editavel na tela de orcamento.

Valores iniciais previstos:

- `VRF`
- `Agua Gelada`
- `Exaustao`

## Resumo Executivo

O caminho recomendado e seguir o mesmo padrao ja adotado para `Linha de produtos`.

Em vez de manter `Tipo de Sistema` como texto livre dentro de `budgets`, a melhor abordagem e:

- criar uma tabela auxiliar propria, por exemplo `system_types`
- adicionar uma coluna de relacionamento em `budgets`, por exemplo `system_type_id`
- expor esse catalogo via API
- carregar esse catalogo no frontend da tela de orcamentos
- exibir e permitir selecao do campo no formulario e na consulta

Essa abordagem reduz inconsistencias, evita variacoes de texto e facilita evolucoes futuras no processo de importacao, filtros e governanca do catalogo.

## Estado Atual

## Referencia de implementacao existente

O caso mais parecido no projeto hoje e `Linha de produtos`, documentado em:

- `docs/ESTUDO-LINHA-DE-PRODUTOS-TABELA-AUXILIAR-E-TELA-ORCAMENTO.md`

Esse fluxo ja consolidou um padrao tecnico util para reaproveitar:

- tabela auxiliar dedicada
- chave estrangeira em `budgets`
- exposicao no backend
- consumo no frontend da tela de orcamentos

## Backend de orcamentos

Hoje o backend de orcamentos ja trafega varios catalogos auxiliares, como:

- `status`
- `priority`
- `installer`
- `product_line`
- `project`
- `salesperson`
- `estimator`
- `contact`
- `loss_reason`

Arquivo de referencia:

- `backend/internal/dto/budget_dto.go`

Ou seja, o sistema ja esta preparado conceitualmente para receber mais um catalogo no mesmo formato.

## Frontend da tela de orcamentos

O frontend da tela de orcamentos ja possui formulario com selects e catalogos carregados por API.

Arquivos relevantes:

- `frontend/src/features/budgets/components/BudgetForm.tsx`
- `frontend/src/features/budgets/api/budgets.ts`
- `frontend/src/features/budgets/types/budget.ts`
- `frontend/src/features/budgets/pages/BudgetListPage.tsx`

Isso significa que `Tipo de Sistema` pode ser encaixado sem criar um modelo novo de interface, apenas ampliando o padrao existente.

## Proposta de Modelagem

## Tabela auxiliar

Criar uma tabela propria:

```sql
system_types
- id
- code
- name
- description
- created_at
- updated_at
```

Valores iniciais recomendados:

```text
VRF
AGUA_GELADA
EXAUSTAO
```

Com nomes exibidos na interface:

- `VRF`
- `Agua Gelada`
- `Exaustao`

## Relacionamento com budgets

Adicionar em `budgets`:

```sql
system_type_id BIGINT NULL
```

Com relacao:

```sql
FOREIGN KEY (system_type_id) REFERENCES system_types(id) ON DELETE SET NULL
```

## Regra funcional recomendada

- a referencia oficial no sistema deve ser `budgets.system_type_id`
- o valor exibido na interface deve vir de `system_types.name`
- o campo deve ser opcional no primeiro momento, salvo se houver regra de negocio exigindo obrigatoriedade
- o catalogo deve nascer com os tres valores iniciais e poder crescer no futuro

## Proposta de Solucao

## Diretriz principal

Adotar `system_types` como fonte oficial de `Tipo de Sistema` no sistema inteiro.

Beneficios:

- evita variacoes de escrita
- melhora consistencia de relatorios e filtros
- facilita validacao visual da informacao
- reduz retrabalho no futuro ao integrar importacao e tela administrativa

## Fase 1 - Estrutura de banco

Objetivo:

- criar a base estrutural do novo catalogo

Escopo:

- criar migration para tabela `system_types`
- criar migration para adicionar `budgets.system_type_id`
- criar `FK` com `ON DELETE SET NULL`
- adicionar indice em `budgets.system_type_id`
- inserir os valores iniciais:
  - `VRF`
  - `Agua Gelada`
  - `Exaustao`

Resultado esperado:

- o banco passa a suportar oficialmente `Tipo de Sistema`

Risco:

- baixo

## Fase 2 - Backend

Objetivo:

- expor o novo catalogo e trafegar o campo no modulo de orcamentos

Escopo:

- criar model e repository de `system_types`, se ainda nao existir
- criar service e handler de `system_types`
- expor ao menos:
  - `GET /system-types`
- adicionar no fluxo de orcamentos:
  - `system_type_id`
  - `system_type_code`
  - `system_type_name`
- atualizar:
  - `CreateBudgetRequest`
  - `UpdateBudgetRequest`
  - `BudgetResponse`
- ajustar `repository` e `service` de budgets para gravar e retornar o relacionamento

Resultado esperado:

- a API fica pronta para a tela consumir o novo campo

Risco:

- medio e controlado

## Fase 3 - Frontend da tela de orcamentos

Objetivo:

- fechar a experiencia funcional de ponta a ponta

Escopo:

- adicionar `systemTypeId` nos types e payloads do frontend
- carregar `systemTypes` junto com os catalogos da tela
- adicionar o campo `Tipo de Sistema` no formulario de cadastro/edicao
- exibir `Tipo de Sistema`:
  - na listagem de orcamentos
  - no detalhe do orcamento

Recomendacao de UX:

- usar o mesmo padrao visual dos outros catalogos
- preferir `Autocomplete` ou `select`
- mostrar o valor ja selecionado ao editar

Resultado esperado:

- o usuario consegue visualizar e editar `Tipo de Sistema`

Risco:

- medio

## Fase 4 - Filtros e consulta

Objetivo:

- tornar o novo dado util para operacao diaria

Escopo:

- adicionar filtro por `Tipo de Sistema` na listagem de orcamentos
- incluir esse campo nas consultas do backend
- ajustar query params e pagina de listagem

Resultado esperado:

- o usuario consegue consultar orcamentos por tipo de sistema

Risco:

- baixo

## Fase 5 - Governanca do catalogo

Objetivo:

- permitir manutencao controlada do catalogo sem depender de banco ou importacao

Escopo:

- criar tela administrativa para `Tipos de Sistema`
- permitir criacao e edicao
- avaliar inativacao em vez de exclusao fisica
- padronizar `code` e `name`

Resultado esperado:

- o catalogo passa a ter vida propria e manutencao segura

Risco:

- medio

## Possivel Fase Extra - Importacao por planilha

Se a planilha atual ou futura trouxer a coluna `Tipo de Sistema`, recomenda-se incluir uma fase complementar:

- ler a coluna no parser
- localizar ou criar `system_types`
- gravar `budgets.system_type_id`
- mostrar no preview um contador como:
  - `system_types_to_create`

Observacao:

- se a planilha ainda nao usa esse campo, essa fase pode ficar para depois

## Impactos Tecnicos

## Banco

Mudancas esperadas:

- nova tabela `system_types`
- nova coluna `budgets.system_type_id`
- nova `FK`
- indice para performance

Impacto:

- baixo

## Backend

Mudancas esperadas:

- novos arquivos para o catalogo
- joins adicionais no modulo de budgets
- novos campos em DTOs e responses

Impacto:

- medio

## Frontend

Mudancas esperadas:

- ampliar catalogos carregados
- adicionar campo no formulario
- ampliar types e mapeamentos
- mostrar o valor na UI

Impacto:

- medio

## Riscos e Pontos de Atencao

## 1. Texto solto e duplicidade semantica

Sem tabela auxiliar, podem surgir variacoes como:

- `VRF`
- `vrf`
- `Agua gelada`
- `Água Gelada`
- `Exaustão`

Recomendacao:

- nao usar texto livre como referencia oficial
- manter o valor oficial no catalogo

## 2. Exclusao do catalogo

Se a relacao usar `ON DELETE SET NULL`, um orcamento historico pode perder a referencia caso o item seja removido.

Recomendacao:

- evitar exclusao fisica no uso operacional
- preferir inativacao em fase futura

## 3. Estrategia de codigo

Decisao recomendada:

- manter `code` tecnico padronizado
- manter `name` amigavel para exibicao

Exemplo:

- `VRF` / `VRF`
- `AGUA_GELADA` / `Agua Gelada`
- `EXAUSTAO` / `Exaustao`

## Recomendacao Final

Seguir em dois blocos:

## Bloco 1 - Entrega funcional rapida

- `Fase 1`
- `Fase 2`
- `Fase 3`

Resultado:

- `Tipo de Sistema` passa a existir de ponta a ponta no sistema

## Bloco 2 - Evolucao operacional

- `Fase 4`
- `Fase 5`
- fase extra de importacao, se necessario

Resultado:

- o campo deixa de ser apenas um dado exibido e passa a ser util para consulta, gestao e importacao

## Conclusao

`Tipo de Sistema` deve ser implementado como nova tabela auxiliar, no mesmo modelo de `Linha de produtos`.

Isso traz consistencia tecnica e funcional com baixo risco arquitetural, porque reaproveita um padrao ja conhecido no projeto.

Em resumo:

- banco: precisa criar a estrutura
- backend: precisa expor e trafegar o novo catalogo
- frontend: precisa consumir e exibir o campo
- importacao: opcional em fase posterior, se a planilha passar a trazer esse dado

O caminho mais eficiente e iniciar por `Fase 1`, `Fase 2` e `Fase 3`, porque isso ja entrega valor real de negocio rapidamente.
