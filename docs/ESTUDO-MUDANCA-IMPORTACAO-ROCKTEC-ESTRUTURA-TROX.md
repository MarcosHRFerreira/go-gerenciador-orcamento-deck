# Estudo de Mudanca da Importacao Rocktec para a Estrutura da Trox

## Objetivo

Avaliar a mudanca da importacao da `Rocktec` para passar a usar a mesma estrutura de planilha hoje tratada pelo layout `Trox`, preservando a origem comercial como `Rocktec`.

Neste estudo, a referencia funcional solicitada e a planilha:

- `c:\Users\marco\OneDrive\Projects\go-gerenciador-orcamento-deck\PLANILHA_TESTE_ROCKTEC_100_LINHAS.xlsx`

Diretriz informada para a mudanca:

- a estrutura da planilha da `Rocktec` passara a ser a mesma da `Trox`
- a aba utilizada no arquivo deve passar a se chamar `Rocktec`

## Resumo Executivo

Minha recomendacao e:

- nao reaproveitar o parser legado atual da `Rocktec` como esta
- criar uma nova variante de layout da `Rocktec`, semanticamente alinhada a `Trox`
- manter `source_company = Rocktec`
- manter `source_layout = rocktec`
- trocar a deteccao por aba fixa para suportar a nova aba `Rocktec`
- reutilizar praticamente o mesmo mapeamento funcional hoje usado pela `Trox`

Em termos praticos:

- `Rocktec legado`: continua apontando para aba `ORCAMENTOS`, cabecalho na linha `10` e 18 colunas
- `Rocktec nova`: deve apontar para aba `Rocktec`, cabecalho na linha `1` e 14 colunas
- `Trox`: continua apontando para aba `Capa`, cabecalho na linha `1` e 14 colunas

Conclusao objetiva:

- sim, a mudanca e viavel
- o maior impacto nao esta na persistencia
- o maior impacto esta na identificacao do layout e na compatibilidade de transicao

## Estado Atual do Codigo

Hoje o backend trabalha com layouts registrados em:

- [layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/layout.go)

Pontos importantes do desenho atual:

- a interface de layout expõe apenas `SheetName() string`
- a resolucao do layout tenta carregar o arquivo usando exatamente uma aba por layout
- a validacao do cabecalho ocorre depois de localizar a aba

Trechos relevantes:

- [layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/layout.go#L10-L28)
- [layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/layout.go#L30-L64)
- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/service.go#L839-L950)

Isso significa que hoje a identificacao depende de dois fatores combinados:

- nome da aba
- cabecalho esperado

## Layout Atual da Rocktec

O layout atual da `Rocktec` esta implementado em:

- [rocktec_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/rocktec_layout.go)

Caracteristicas atuais:

- aba fixa: `ORCAMENTOS`
- cabecalho na linha `10`
- 18 colunas uteis
- semantica original Rocktec

Campos principais hoje esperados:

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

Pontos de codigo:

- [rocktec_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/rocktec_layout.go#L10-L14)
- [rocktec_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/rocktec_layout.go#L61-L71)
- [rocktec_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/rocktec_layout.go#L149-L222)

## Layout Atual da Trox

O layout atual da `Trox` esta implementado em:

- [trox_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/trox_layout.go)

Caracteristicas atuais:

- aba fixa: `Capa`
- cabecalho na linha `1`
- 14 colunas uteis
- parser mais enxuto
- semantica comercial normalizada para o dominio atual

Cabecalhos tratados pela `Trox`:

1. `Orcamento`
2. `Revisao`
3. `Data de Emissao`
4. `Tipo`
5. `Status`
6. `Contato`
7. `Linha de produtos`
8. `Codigo Cliente`
9. `Nome Cliente`
10. `Obra`
11. `Vendedor`
12. `Instalador`
13. `Total do orcamento`
14. `Fator Medio`

Pontos de codigo:

- [trox_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/trox_layout.go#L11-L15)
- [trox_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/trox_layout.go#L122-L171)
- [trox_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/trox_layout.go#L174-L196)

## Estrutura da Planilha de Referencia

Com base no arquivo de referencia informado para a mudanca, a estrutura observada corresponde ao mesmo desenho ja usado pela `Trox`:

- 14 colunas
- cabecalho na primeira linha
- dados comerciais resumidos
- mesma semantica do layout atual da `Trox`

Colunas observadas:

1. `Orcamento`
2. `Revisao`
3. `Data de Emissao`
4. `Tipo`
5. `Status`
6. `Contato`
7. `Linha de produtos`
8. `Codigo Cliente`
9. `Nome Cliente`
10. `Obra`
11. `Vendedor`
12. `Instalador`
13. `Total do orcamento`
14. `Fator Medio`

## Leitura Tecnica da Mudanca

Funcionalmente, o pedido nao e "adaptar a Rocktec antiga".

Na pratica, o pedido e:

- substituir a `Rocktec` legada por uma `Rocktec v2`
- esta `Rocktec v2` deve usar o mesmo formato estrutural da `Trox`
- o diferencial entre `Rocktec v2` e `Trox` deixa de ser a estrutura de colunas
- o diferencial passa a ser principalmente a origem do arquivo e o nome da aba

Isso muda o problema tecnico de:

- parser diferente por estrutura

para:

- parser quase igual, com identidade de layout/origem diferente

## Mapeamento Recomendado da Nova Rocktec

Se a nova `Rocktec` usar exatamente a estrutura da `Trox`, o mapeamento recomendado e o mesmo:

| Coluna da nova Rocktec | Destino sugerido |
| --- | --- |
| `Orcamento` | `budget_number` |
| `Revisao` | `revision` |
| `Data de Emissao` | `sent_at` |
| `Tipo` | manter apenas para rastreabilidade neste momento |
| `Status` | `current_follow_up` |
| `Contato` | `contacts.name` / `contact_id` |
| `Linha de produtos` | `product_line` |
| `Codigo Cliente` | staging / rastreabilidade |
| `Nome Cliente` | `construction_company` |
| `Obra` | `projects.name` / `project_id` |
| `Vendedor` | `salespeople.name` / `salesperson_id` |
| `Instalador` | `installers.name` / `installer_id` |
| `Total do orcamento` | `gross_value` |
| `Fator Medio` | staging / rastreabilidade |

Regras recomendadas, iguais a `Trox`:

- `statusName = Em Negociacao`
- `currentFollowUp = coluna Status`
- remover prefixo `DECK - ` do vendedor
- `Linha de produtos` continua como catalogo auxiliar
- `Nome Cliente` continua alimentando `construction_company`
- `Tipo` e `Fator Medio` permanecem fora do dominio principal por enquanto

## Impactos Necessarios no Backend

### 1. Deteccao da aba

Hoje cada layout aceita apenas um nome de aba por meio de `SheetName() string`.

Se a nova `Rocktec` deve usar aba `Rocktec`, existem tres opcoes:

1. trocar `ORCAMENTOS` por `Rocktec` diretamente no layout atual
2. criar uma nova variante `rocktec_v2`
3. permitir multiplos nomes de aba no mesmo layout

Minha recomendacao:

- nao trocar direto a `Rocktec` antiga se ainda existirem arquivos legados
- preferir compatibilidade progressiva

Melhor abordagem tecnica:

- evoluir a interface para algo como `SheetNames() []string`
- permitir que a `Rocktec` reconheca `Rocktec` e, se necessario na transicao, tambem `ORCAMENTOS`

Arquivos afetados:

- [layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/layout.go)
- [service.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/service.go#L839-L950)

### 2. Parser Rocktec

O parser atual da `Rocktec` nao deve ser reaproveitado como base principal da nova regra porque ele pressupoe:

- data serial do Excel
- header na linha `10`
- colunas legadas

Minha recomendacao:

- criar um novo parser `Rocktec` baseado na mesma semantica do parser `Trox`
- reaproveitar funcoes e regras onde fizer sentido
- manter identidade propria de `source_company` e `source_layout`

Arquivos diretamente impactados:

- [rocktec_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/rocktec_layout.go)
- [trox_layout.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/trox_layout.go)

### 3. Testes automatizados

Hoje existem testes separados por layout:

- [rocktec_layout_test.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/rocktec_layout_test.go)
- [trox_layout_test.go](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/backend/internal/service/budgetimport/trox_layout_test.go)

Para a mudanca, recomendo:

- atualizar os testes do layout `Rocktec`
- cobrir a nova aba `Rocktec`
- validar header na linha `1`
- validar parse da data textual `dd/mm/yyyy`
- validar o mesmo mapeamento funcional da `Trox`

## Opcoes de Implementacao

### Opcao A - Substituir a Rocktec legada diretamente

Descricao:

- o layout `rocktec` deixa de usar `ORCAMENTOS`
- passa a usar aba `Rocktec`
- passa a usar o parser estrutural da `Trox`

Vantagens:

- implementacao mais simples
- menos codigo para manter

Riscos:

- quebra imediata de qualquer arquivo legado Rocktec
- risco operacional alto se ainda existirem planilhas antigas em uso

### Opcao B - Criar Rocktec v2 com convivio temporario

Descricao:

- manter `Rocktec legado`
- criar `Rocktec v2`
- detectar pela aba e pelo header

Vantagens:

- menor risco de transicao
- facilita homologacao gradual

Riscos:

- mais codigo temporario
- exige clareza na selecao do layout

### Opcao C - Tornar a Rocktec tolerante a mais de um formato

Descricao:

- um unico layout `rocktec`
- aceita mais de um nome de aba
- aceita dois formatos de cabecalho

Vantagens:

- menor impacto aparente na API interna

Riscos:

- parser mais complexo
- testes mais dificeis
- maior risco de ambiguidade

## Recomendacao Tecnica Final

Minha recomendacao e a `Opcao B`.

Implementacao sugerida:

1. manter compatibilidade do legado por um periodo
2. criar a nova `Rocktec` com o mesmo desenho funcional da `Trox`
3. aceitar a aba `Rocktec`
4. isolar a regra de escolha do layout
5. so depois remover o parser antigo

## Backlog Sugerido

### Fase 1 - Preparacao

- revisar a interface de layout para suportar alias de aba ou selecao mais flexivel
- decidir se a transicao sera com um layout novo ou substituicao direta

### Fase 2 - Parser da nova Rocktec

- criar parser `Rocktec` com cabecalho na linha `1`
- usar as mesmas 14 colunas da `Trox`
- manter `SourceCompany() = Rocktec`

### Fase 3 - Testes

- teste de validacao de aba `Rocktec`
- teste de cabecalho com 14 colunas
- teste de parse com data textual
- teste de mapeamento comercial equivalente ao da `Trox`

### Fase 4 - Transicao

- homologar com planilhas reais da nova origem
- validar preview e importacao completa
- decidir data de desativacao do layout legado

## Riscos e Cuidados

- risco de quebra de arquivos antigos se a troca for direta
- risco de ambiguidade se dois layouts aceitarem headers muito parecidos
- risco de dependencia operacional no nome da aba
- risco de confusao de negocio se `Rocktec` e `Trox` passarem a ter mesma estrutura, mas regras diferentes no futuro

## Conclusao

Sim, e totalmente viavel mudar a importacao da `Rocktec` para usar a mesma estrutura da `Trox`.

O ponto central desta mudanca nao e o banco nem o dominio principal.

O ponto central e:

- trocar a identificacao da aba
- trocar o parser estrutural da `Rocktec`
- preservar a identidade comercial da origem `Rocktec`
- tratar com cuidado a compatibilidade com o layout legado

Se a decisao for seguir com a implementacao, o melhor caminho e:

- criar uma `Rocktec` nova baseada no parser da `Trox`
- usar a aba `Rocktec`
- homologar primeiro com arquivos novos
- apos estabilizar, descontinuar o layout `ORCAMENTOS`
