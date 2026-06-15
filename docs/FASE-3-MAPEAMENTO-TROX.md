# Fase 3 - Mapeamento Formal da Planilha Trox

## Objetivo

Documentar formalmente o layout Trox para preparar a implementacao do parser na Fase 4.

Esta fase fecha:

- catalogacao das colunas da Trox
- destino esperado de cada campo
- politica de parse de datas e numeros
- classificacao dos campos exclusivos da Trox
- decisoes de negocio necessarias para o primeiro parser

Este documento complementa:

- [ESTUDO-IMPORTACAO-ROCKTEC-TROX.md](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/docs/ESTUDO-IMPORTACAO-ROCKTEC-TROX.md)
- [BACKLOG-IMPORTACAO-ROCKTEC-TROX.md](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/docs/BACKLOG-IMPORTACAO-ROCKTEC-TROX.md)

## Base de Observacao

Arquivo analisado:

- `RESUMO ORÇAMENTOS CONCORRENCIA.xlsx`

Validacao tecnica realizada diretamente no arquivo:

- aba encontrada: `Capa`
- cabecalho util na linha `1`
- total observado de colunas uteis: `14`
- valores de data observados como texto no formato `dd/MM/yyyy`
- valores monetarios observados como numericos com ponto decimal

Primeiras linhas observadas:

- `ROW 1`: `Orçamento | Revisão | Data de Emissão | Tipo | Status | Contato | Linha de produtos | Código Cliente | Nome Cliente | Obra | Vendedor | Instalador | Total do orçamento | Fator Médio`
- `ROW 2`: `477139 | 0 | 09/06/2026 | Consulta de preço | Informado | ELDER J. BONETTI | FILTROS | BR1007854 | ABECON ENGENHARIA E CLIMATIZACAO LT | DIVERSOS DE JUNHO | DECK - EMANUEL FERRI | ABECON ENGENHARIA E CLIMATIZACAO LT | 65515.83 | 0.8`

## Estrutura Formal da Trox

| Ordem | Coluna Trox | Obrigatoriedade inicial | Observacao |
| --- | --- | --- | --- |
| 1 | `Orçamento` | obrigatoria | identificador principal do orçamento |
| 2 | `Revisão` | obrigatoria | aparenta vir como inteiro |
| 3 | `Data de Emissão` | obrigatoria | data textual |
| 4 | `Tipo` | obrigatoria | semantica comercial da origem |
| 5 | `Status` | obrigatoria | status operacional da origem |
| 6 | `Contato` | recomendada | nome de pessoa |
| 7 | `Linha de produtos` | recomendada | classificacao comercial da Trox |
| 8 | `Código Cliente` | recomendada | identificador externo |
| 9 | `Nome Cliente` | recomendada | empresa cliente na origem |
| 10 | `Obra` | obrigatoria | nome textual da obra |
| 11 | `Vendedor` | obrigatoria | normalmente no padrao `DECK - NOME` |
| 12 | `Instalador` | obrigatoria | empresa instaladora |
| 13 | `Total do orçamento` | obrigatoria | valor monetario |
| 14 | `Fator Médio` | opcional | fator numerico |

## Mapeamento Aprovado Para o Modelo Normalizado

### Campos com destino direto no DTO normalizado

| Coluna Trox | Campo normalizado | Decisao |
| --- | --- | --- |
| `Orçamento` | `budgetNumber` | mapear diretamente como texto normalizado |
| `Revisão` | `revision` | converter para inteiro |
| `Data de Emissão` | `sentAt` | converter de `dd/MM/yyyy` para `time.Time` |
| `Data de Emissão` | `yearBudget` | derivar de `sentAt.Year()` |
| `Obra` | `projectName` | mapear diretamente |
| `Vendedor` | `salespersonName` | mapear com normalizacao de prefixo |
| `Instalador` | `installerName` | mapear diretamente |
| `Contato` | `contactName` | mapear diretamente |
| `Total do orçamento` | `grossValue` | converter para `float64` |

### Campos com destino controlado no primeiro parser

| Coluna Trox | Campo normalizado | Decisao da Fase 3 |
| --- | --- | --- |
| `Status` | `currentFollowUp` | usar como follow-up atual da origem |
| `Status` | `statusName` | usar `Nao informado` no primeiro parser |
| `Tipo` | `specificationDetails` | nao mapear para especificacao; manter fora do dominio principal nesta fase |
| `Tipo` | `rawType` conceitual | tratar como campo de origem da Trox para futura staging |
| `Linha de produtos` | `projectTypeName` | nao mapear diretamente |
| `Linha de produtos` | `productLine` conceitual | classificar como campo proprio da Trox |
| `Código Cliente` | `externalCustomerCode` conceitual | classificar como campo proprio da Trox |
| `Nome Cliente` | `externalCustomerName` conceitual | classificar como campo proprio da Trox |
| `Fator Médio` | `averageFactor` conceitual | classificar como campo proprio da Trox |

## Decisoes de Negocio da Fase 3

### 1. `Status` da Trox nao entra como `budget status` no primeiro parser

Decisao:

- o valor da coluna `Status` da Trox sera tratado inicialmente como `currentFollowUp`
- o campo `statusName` do DTO normalizado sera preenchido com `Nao informado`

Justificativa:

- na amostra observada, `Status` aparece com valor como `Informado`
- esse valor se parece mais com etapa operacional/comercial da origem do que com o catalogo principal de status do sistema
- mapear automaticamente isso para `budget_statuses` agora pode distorcer relatorios e regras atuais

Conclusao pratica:

- a Trox entra primeiro preservando o contexto operacional em `currentFollowUp`
- a definicao de mapeamento de `Status` para `budget_statuses` fica para uma fase posterior, se houver tabela de equivalencia aprovada

### 2. `Tipo` da Trox nao deve ser convertido em `project type`

Decisao:

- nao usar `Tipo` como `projectTypeName`

Justificativa:

- os valores observados, como `Consulta de preço`, representam natureza da negociacao, nao tipo de obra
- isso e semanticamente diferente da Rocktec, que tem `TIPO DE OBRA`

Conclusao pratica:

- `projectTypeName` deve ficar como `Nao informado` no primeiro parser da Trox
- `Tipo` deve ser preservado como campo de origem para futura staging ou campo adicional

### 3. `Linha de produtos` nao deve virar `project type`

Decisao:

- nao mapear `Linha de produtos` para `projectTypeName`

Justificativa:

- valores como `FILTROS`, `FAN-COIL` e `DIFUSÃO A/F` descrevem linha comercial ou linha de produto
- isso nao representa categoria de obra

Conclusao pratica:

- a linha de produtos deve ficar classificada como dado proprio da Trox
- o melhor destino futuro e `product_line` ou staging

### 4. `Nome Cliente` nao deve substituir `Instalador`

Decisao:

- manter `Instalador` como fonte de `installerName`
- nao sobrescrever com `Nome Cliente`

Justificativa:

- nos exemplos observados, `Instalador` e `Nome Cliente` podem coincidir, mas isso nao garante equivalencia sem regra formal
- `Nome Cliente` representa cliente externo da origem e deve ser preservado separadamente

Conclusao pratica:

- o parser Trox deve usar `Instalador` para `installerName`
- `Nome Cliente` deve ficar como campo de origem para futura staging ou extensao do modelo

### 5. `Vendedor` deve ser normalizado removendo prefixo organizacional

Decisao:

- remover o prefixo `DECK - ` quando existir antes da conciliacao com `salespeople`

Exemplos:

- `DECK - EMANUEL FERRI` -> `EMANUEL FERRI`
- `DECK - GUILHERME OLIVEIRA` -> `GUILHERME OLIVEIRA`

Justificativa:

- isso melhora o reaproveitamento do catalogo atual de vendedores
- reduz chance de duplicidade artificial entre Rocktec e Trox

Conclusao pratica:

- o parser deve preservar o valor original apenas para log bruto, se necessario
- o valor normalizado deve ser usado para lookup e persistencia

## Politica de Parse da Trox

### Data de Emissao

Formato observado:

- `09/06/2026`

Regra aprovada:

- aceitar `dd/MM/yyyy`
- interpretar em timezone neutra, preferencialmente `UTC`
- rejeitar valor vazio ou fora do formato esperado

Resultado esperado:

- preencher `sentAt`
- derivar `yearBudget`

### Revisao

Formato observado:

- `0`

Regra aprovada:

- converter para inteiro
- se vier vazio, usar `0` com aviso
- se vier texto nao numerico, marcar erro da linha

### Total do orcamento

Formato observado:

- `65515.83`
- `76313.070000000007`

Regra aprovada:

- converter para `float64`
- aceitar ponto decimal
- aplicar trim de espacos
- se valor vier vazio, marcar erro da linha

Observacao:

- a Trox pode expor artefatos de precisao binaria do Excel
- o parser deve aceitar isso normalmente e deixar eventual arredondamento para a camada de exibicao ou regra financeira posterior

### Fator Medio

Formato observado:

- `0.8`
- `1.05`
- `0.75`

Regra aprovada:

- tratar como numero decimal opcional
- no primeiro parser, nao usar no dominio principal
- preservar para futura staging ou extensao do modelo

## Politica de Valores Ausentes

Para a primeira versao do parser Trox:

- `projectTypeName` -> `Nao informado`
- `priorityName` -> `Nao informado`
- `lossReasonName` -> `Nao informado`
- `competitorName` -> `Nao informado`
- `designerName` -> `Nao informado`
- `specification` -> `Nao informado`
- `statusName` -> `Nao informado`

Observacao:

- `currentFollowUp` deve receber o `Status` da Trox

## Mapeamento Final Para Implementacao da Fase 4

| Coluna Trox | Campo do DTO normalizado | Regra |
| --- | --- | --- |
| `Orçamento` | `budgetNumber` | obrigatorio |
| `Revisão` | `revision` | inteiro |
| `Data de Emissão` | `sentAt` | parse `dd/MM/yyyy` |
| `Data de Emissão` | `yearBudget` | derivado |
| `Tipo` | fora do dominio principal | preservar para futuro campo de origem |
| `Status` | `currentFollowUp` | mapear diretamente |
| `Contato` | `contactName` | mapear diretamente |
| `Linha de produtos` | fora do dominio principal | preservar para futuro campo de origem |
| `Código Cliente` | fora do dominio principal | preservar para futuro campo de origem |
| `Nome Cliente` | fora do dominio principal | preservar para futuro campo de origem |
| `Obra` | `projectName` | mapear diretamente |
| `Vendedor` | `salespersonName` | remover prefixo `DECK - ` |
| `Instalador` | `installerName` | mapear diretamente |
| `Total do orçamento` | `grossValue` | float obrigatorio |
| `Fator Médio` | fora do dominio principal | preservar para futuro campo de origem |

## Campos da Trox Classificados por Destino

### Entram no dominio principal ja na Fase 4

- `Orçamento`
- `Revisão`
- `Data de Emissão`
- `Obra`
- `Vendedor`
- `Instalador`
- `Contato`
- `Total do orçamento`
- `Status` como `currentFollowUp`

### Ficam fora do dominio principal no primeiro parser

- `Tipo`
- `Linha de produtos`
- `Código Cliente`
- `Nome Cliente`
- `Fator Médio`

### Recomendacao para esses campos fora do dominio

- manter em memoria durante o parse para futura staging
- nao perder a modelagem conceitual desses campos
- evitar criar colunas novas no dominio principal antes da Fase 5

## Critério de Pronto da Fase 3

Considera-se a Fase 3 concluida quando:

- o layout da Trox estiver formalmente descrito
- cada coluna estiver classificada com destino claro
- as regras de parse de data e numero estiverem decididas
- houver uma decisao explicita para os campos exclusivos da Trox
- a equipe puder iniciar a implementacao do parser Trox sem abrir novas discussoes conceituais basicas

## Proximo Passo

Com esta fase fechada, o proximo passo tecnico recomendado e:

- implementar o `TroxImportLayout`
- adicionar parse de data textual `dd/MM/yyyy`
- normalizar vendedor removendo prefixo `DECK - `
- preencher o DTO normalizado conforme as decisoes acima
- adicionar testes unitarios e de preview para a Trox
