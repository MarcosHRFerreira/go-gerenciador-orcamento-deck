# Estudo de Viabilidade de Importacao Rocktec x Trox

## Objetivo

Avaliar a viabilidade de utilizar a mesma estrutura de dados e o mesmo fluxo de importacao para duas origens diferentes de planilhas:

- Rocktec
- Trox

O foco deste estudo e responder:

- se as duas empresas podem compartilhar a mesma estrutura principal do sistema
- se o importador atual pode ser reaproveitado
- quais ajustes arquiteturais sao recomendados
- qual o impacto esperado no banco de dados
- qual backlog tecnico seria necessario para evolucao

## Resumo Executivo

Conclusao recomendada:

- manter a mesma estrutura principal de negocio para as duas empresas
- nao criar tabelas principais separadas de orcamentos por empresa neste momento
- evoluir o importador para um modelo multi-layout
- registrar a origem da importacao e o layout utilizado
- considerar uma camada de staging ou log de importacao para auditoria e rastreabilidade

Em termos praticos:

- mesma base de dominio: sim
- mesmo importador atual sem adaptacao: nao
- mesmas tabelas de negocio para Rocktec e Trox: sim
- parsers e mapeamentos separados por empresa: sim

## Contexto Atual

O sistema atual foi construindo inicialmente com base na planilha da Rocktec.

No importador atual, a implementacao esta rigidamente acoplada ao layout da Rocktec:

- aba fixa `ORCAMENTOS`
- linha de cabecalho fixa na linha `10`
- validacao de cabecalho baseada em posicoes esperadas
- leitura orientada por indice de coluna
- interpretacao de campos seguindo a semantica observada na planilha Rocktec

Trechos relevantes no backend:

- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/service.go#L23-L29)
- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/service.go#L201-L228)
- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/service.go#L307-L379)
- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/service.go#L410-L427)
- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/service.go#L975-L991)

O estudo funcional da Rocktec ja existe em:

- [ESTUDO-CARGA-PLANILHA-ORCAMENTOS.md](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/docs/ESTUDO-CARGA-PLANILHA-ORCAMENTOS.md)

## Estrutura Observada da Rocktec

Com base no estudo anterior, a planilha Rocktec possui:

- arquivo de referencia com aba `ORCAMENTOS`
- linha de cabecalho util na linha `10`
- cerca de `23` colunas observadas
- forte aderencia ao modelo de negocio atual do sistema

Principais colunas mapeadas:

- `DATA`
- `Nº DE ORCA`
- `REV.`
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

Essa estrutura se encaixa bem nas tabelas atuais:

- `budgets`
- `projects`
- `project_types`
- `salespeople`
- `installers`
- `contacts`
- `loss_reasons`
- `budget_statuses`
- `priorities`

## Estrutura Observada da Trox

Arquivo analisado:

- `RESUMO ORÇAMENTOS CONCORRENCIA.xlsx`

Observacoes tecnicas encontradas:

- aba encontrada: `Capa`
- cabecalho util ja aparece na linha `1`
- quantidade observada de colunas uteis: `14`
- formato da data observado como texto, por exemplo `09/06/2026`

Cabecalhos observados na linha `1`:

1. `Orçamento`
2. `Revisão`
3. `Data de Emissão`
4. `Tipo`
5. `Status`
6. `Contato`
7. `Linha de produtos`
8. `Código Cliente`
9. `Nome Cliente`
10. `Obra`
11. `Vendedor`
12. `Instalador`
13. `Total do orçamento`
14. `Fator Médio`

Exemplos de valores observados:

- `Orçamento`: `477139`
- `Tipo`: `Consulta de preço`
- `Status`: `Informado`
- `Linha de produtos`: `FILTROS`, `FAN-COIL`, `DIFUSÃO A/F`
- `Código Cliente`: `BR1007854`
- `Nome Cliente`: razao social ou nome do cliente
- `Obra`: nome textual da obra
- `Vendedor`: nomes no padrao `DECK - NOME`
- `Instalador`: nome da empresa instaladora
- `Total do orçamento`: valor monetario
- `Fator Médio`: fator numerico

## Comparacao Rocktec x Trox

### Similaridades

As duas planilhas representam o mesmo conceito central:

- orcamentos comerciais
- revisao
- data
- obra
- vendedor
- instalador
- contato
- valor do orcamento

Essas similaridades indicam que ambas podem alimentar a mesma estrutura principal de dominio.

### Diferencas

As diferencas sao significativas no layout e na semantica:

- Rocktec usa aba `ORCAMENTOS`; Trox usa aba `Capa`
- Rocktec usa cabecalho na linha `10`; Trox usa cabecalho na linha `1`
- Rocktec usa cerca de `23` colunas; Trox apresenta `14`
- Rocktec tem `TIPO DE OBRA`; Trox tem `Tipo`
- Rocktec tem `NOME DA OBRA`; Trox tem `Obra`
- Rocktec tem `VALOR BRUTO`; Trox tem `Total do orçamento`
- Rocktec traz `MOTIVO`, `VALOR CONCORRENTE`, `PROJETISTA`, `ESPECIFICACOES`; Trox nao traz essas colunas no mesmo layout
- Trox traz `Linha de produtos`, `Código Cliente`, `Nome Cliente`, `Fator Médio`, que nao existem no mapeamento principal atual da Rocktec
- Rocktec trabalha com data serial do Excel; Trox aparenta trabalhar com data textual

### Comparacao por aderencia ao modelo atual

Campos com aderencia clara:

- numero do orcamento
- revisao
- data de emissao
- obra
- vendedor
- instalador
- contato
- valor do orcamento

Campos que pedem avaliacao de regra:

- `Tipo` da Trox
- `Status` da Trox
- `Linha de produtos`
- `Código Cliente`
- `Nome Cliente`
- `Fator Médio`

## Opiniao Tecnica

Minha recomendacao e **nao separar em tabelas principais diferentes por empresa**.

Separar completamente o modelo em algo como:

- `rocktec_budgets`
- `trox_budgets`

traria desvantagens importantes:

- duplicacao de regras de negocio
- duplicacao de consultas e filtros
- duplicacao de APIs e componentes de tela
- maior custo de manutencao
- pior escalabilidade do produto para futuras empresas

O melhor desenho e:

- uma estrutura principal unica de dominio
- layouts de importacao separados
- normalizacao para um modelo comum

## Recomendacao de Arquitetura

### Diretriz principal

Adotar uma arquitetura de importacao em duas etapas:

1. parser especifico por layout
2. normalizacao para um modelo comum

Fluxo sugerido:

1. identificar a origem da planilha
2. escolher o parser correto
3. mapear para um DTO intermediario comum
4. aplicar validacoes e regras de negocio
5. persistir no modelo atual

### Modelo conceitual sugerido

Parsers especificos:

- `RocktecImportLayout`
- `TroxImportLayout`

Objeto intermediario comum:

- `NormalizedBudgetImportRow`

Persistencia:

- tabelas atuais do dominio

### Exemplo de campos do DTO intermediario

- `sourceCompany`
- `sourceLayout`
- `budgetNumber`
- `yearBudget`
- `revision`
- `sentAt`
- `projectName`
- `projectTypeName`
- `salespersonName`
- `installerName`
- `contactName`
- `grossValue`
- `statusName`
- `currentFollowUp`
- `lossReasonName`
- `competitorName`
- `competitorPrice`
- `designerName`
- `specificationDetails`
- `externalCustomerCode`
- `externalCustomerName`
- `productLine`
- `averageFactor`
- `rawType`
- `rawStatus`

### Vantagens dessa abordagem

- reaproveita o modelo principal
- reduz impacto nas telas existentes
- facilita onboarding de novas empresas
- preserva rastreabilidade da origem
- evita refazer regras de dominio ja prontas

## Impacto no Banco

### O que pode permanecer igual

As tabelas principais atuais podem continuar sendo o centro do sistema:

- `budgets`
- `projects`
- `project_types`
- `salespeople`
- `installers`
- `contacts`
- `loss_reasons`
- `budget_statuses`
- `priorities`

### O que recomendo adicionar

#### 1. Rastreabilidade de origem

Adicionar no dominio principal ou em tabela vinculada:

- `source_company`
- `source_layout`
- `import_batch_id`

Isso permite saber de qual empresa e de qual importacao cada registro veio.

#### 2. Tabela de lote de importacao

Sugestao:

- `budget_import_batches`

Campos sugeridos:

- `id`
- `source_company`
- `source_layout`
- `original_file_name`
- `sheet_name`
- `imported_by_user_id`
- `started_at`
- `finished_at`
- `status`
- `summary_json`

#### 3. Tabela de staging ou linha bruta

Sugestao:

- `budget_import_rows_raw`

Campos sugeridos:

- `id`
- `batch_id`
- `row_number`
- `raw_payload_json`
- `normalized_payload_json`
- `status`
- `message`

Essa tabela e muito util para:

- auditoria
- troubleshooting
- reprocessamento
- comparacao entre layouts

#### 4. Campos adicionais no dominio principal

Se os dados da Trox forem relevantes para consulta futura, pode ser util adicionar:

- `customer_code`
- `customer_name`
- `product_line`
- `average_factor`
- `source_status`
- `source_type`

Se esses dados forem apenas informativos ou temporarios, podem ficar so na staging.

### O que eu nao recomendo agora

- criar novas tabelas principais de orcamento por empresa
- duplicar `projects`, `salespeople`, `installers` por origem
- criar APIs distintas de consulta de orcamentos por empresa

## Mapeamento Inicial Sugerido da Trox

### Campos com destino direto

| Coluna Trox | Destino sugerido |
| --- | --- |
| `Orçamento` | `budget_number` |
| `Revisão` | `revision` |
| `Data de Emissão` | `sent_at` |
| `Obra` | `projects.name` / `project_id` |
| `Vendedor` | `salespeople.name` / `salesperson_id` |
| `Instalador` | `installers.name` / `installer_id` |
| `Contato` | `contacts.name` / `contact_id` |
| `Total do orçamento` | `gross_value` |

### Campos que podem ir para staging ou novos campos

| Coluna Trox | Destino sugerido |
| --- | --- |
| `Tipo` | `source_type` ou `current_follow_up`, dependendo da regra |
| `Status` | `source_status` ou `current_follow_up`, dependendo da regra |
| `Linha de produtos` | `product_line` ou staging |
| `Código Cliente` | `customer_code` ou staging |
| `Nome Cliente` | `customer_name` ou staging |
| `Fator Médio` | `average_factor` ou staging |

## Viabilidade de Uso da Mesma Estrutura

### Viavel reutilizar a mesma estrutura principal?

Sim.

Com a leitura atual, e plenamente viavel usar a mesma estrutura principal do sistema para as duas empresas, desde que:

- o importador seja desacoplado do layout Rocktec
- exista uma camada de normalizacao
- a origem da informacao seja registrada

### Viavel reutilizar o mesmo importador atual?

Nao, nao da forma como ele esta hoje.

Motivos:

- aba fixa
- linha fixa de cabecalho
- validacao fixa
- leitura por indice de coluna
- parse de data especifico
- mapeamento sem suporte a variacoes de layout

### Viavel separar apenas a logica de importacao?

Sim, e essa e a melhor opcao.

## Backlog Tecnico Recomendado

### Fase 1 - Fundacao do importador multi-layout

#### TASK-IMP-001 - Desacoplar layout fixo da Rocktec

- `Prioridade`: `P0`
- `Descricao`: extrair do importador atual os pontos fixos de aba, linha de cabecalho, validacao e mapeamento por coluna.
- `Pronto quando`: o importador puder trabalhar com configuracao por layout.

#### TASK-IMP-002 - Criar contrato de layout de importacao

- `Prioridade`: `P0`
- `Descricao`: definir uma interface ou estrategia para layout de importacao, com metodos para identificar aba, cabecalho, conversao e normalizacao.
- `Pronto quando`: for possivel adicionar Rocktec e Trox como implementacoes diferentes.

#### TASK-IMP-003 - Criar DTO intermediario comum

- `Prioridade`: `P0`
- `Descricao`: criar um objeto normalizado unico para receber dados de qualquer planilha antes da persistencia.
- `Pronto quando`: Rocktec e Trox conseguirem gerar a mesma estrutura intermediaria.

### Fase 2 - Rocktec como primeiro layout formal

#### TASK-IMP-004 - Adaptar o layout Rocktec para o novo contrato

- `Prioridade`: `P0`
- `Dependencias`: `TASK-IMP-001`, `TASK-IMP-002`, `TASK-IMP-003`
- `Descricao`: migrar a implementacao atual para um parser especifico Rocktec.
- `Pronto quando`: a importacao Rocktec continuar funcionando dentro da nova arquitetura.

#### TASK-IMP-005 - Garantir testes de regressao da Rocktec

- `Prioridade`: `P0`
- `Descricao`: manter e ampliar testes de preview e execucao para garantir que a refatoracao nao quebre a importacao atual.
- `Pronto quando`: a suite passar cobrindo o fluxo atual da Rocktec.

### Fase 3 - Suporte a Trox

#### TASK-IMP-006 - Criar parser Trox

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-002`, `TASK-IMP-003`
- `Descricao`: implementar parser para a planilha Trox considerando aba `Capa`, cabecalho na linha `1`, data textual e mapeamento proprio.
- `Pronto quando`: a planilha Trox puder ser interpretada e convertida para o DTO intermediario.

#### TASK-IMP-007 - Definir regras de negocio dos campos exclusivos da Trox

- `Prioridade`: `P1`
- `Descricao`: validar com negocio o destino de `Tipo`, `Status`, `Linha de produtos`, `Código Cliente`, `Nome Cliente` e `Fator Médio`.
- `Pronto quando`: todos os campos da Trox tiverem destino definido no dominio ou na staging.

#### TASK-IMP-008 - Adicionar testes automatizados da Trox

- `Prioridade`: `P1`
- `Descricao`: criar testes de preview e importacao para o layout Trox.
- `Pronto quando`: o layout Trox estiver coberto por testes automatizados.

### Fase 4 - Banco e rastreabilidade

#### TASK-IMP-009 - Criar tabela de lotes de importacao

- `Prioridade`: `P1`
- `Descricao`: adicionar `budget_import_batches`.
- `Pronto quando`: cada importacao tiver rastreabilidade completa.

#### TASK-IMP-010 - Criar tabela de linhas brutas de importacao

- `Prioridade`: `P1`
- `Descricao`: adicionar `budget_import_rows_raw`.
- `Pronto quando`: for possivel auditar a origem de cada linha processada.

#### TASK-IMP-011 - Adicionar campos de origem no dominio

- `Prioridade`: `P1`
- `Descricao`: avaliar e implementar `source_company`, `source_layout` e `import_batch_id`.
- `Pronto quando`: cada orçamento importado tiver origem identificavel.

#### TASK-IMP-012 - Avaliar campos adicionais da Trox no dominio principal

- `Prioridade`: `P2`
- `Descricao`: decidir se `customer_code`, `customer_name`, `product_line` e `average_factor` entram em `budgets` ou ficam so em staging.
- `Pronto quando`: o destino desses dados estiver definido e documentado.

### Fase 5 - UX e operacao

#### TASK-IMP-013 - Evoluir tela de importacao para escolha de layout

- `Prioridade`: `P1`
- `Descricao`: permitir que o usuario selecione a origem ou que o sistema detecte automaticamente o layout da planilha.
- `Pronto quando`: a interface suportar multiplas empresas na mesma experiencia.

#### TASK-IMP-014 - Melhorar feedback de preview por layout

- `Prioridade`: `P2`
- `Descricao`: exibir para o usuario qual layout foi identificado, quais campos foram ignorados e quais foram convertidos.
- `Pronto quando`: o preview ficar claro e auditavel para Rocktec e Trox.

## Recomendacao Final

Minha recomendacao final e:

- nao separar tabelas principais por empresa
- manter um unico dominio de orcamentos
- introduzir um importador multi-layout
- adicionar rastreabilidade por lote e por origem
- tratar campos exclusivos da Trox de forma gradual, priorizando staging antes de expandir o dominio

Essa abordagem equilibra:

- menor impacto no sistema atual
- melhor capacidade de manutencao
- melhor escalabilidade para futuras empresas
- menor risco de fragmentar o produto

## Decisao Recomendada

Decisao sugerida neste momento:

- **base unica de dados**
- **layouts separados de importacao**
- **evolucao incremental do modelo**

Essa e, na minha avaliacao, a opcao mais segura tecnicamente e mais inteligente do ponto de vista de produto.
