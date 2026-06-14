# Estudo de Carga de Planilha de Orcamentos

## Objetivo

Definir uma proposta funcional para a tela de carga de planilhas usando somente a aba `ORCAMENTOS` do arquivo `RELATORIO DE ORCAMENTOS-25.xlsx`, considerando:

- layout e comportamento da tela de importacao
- mapeamento da planilha para as tabelas atuais
- regras de saneamento dos dados
- estrategia para preencher campos ausentes com `Nao informado`
- estrategia para tabelas auxiliares com item padrao `Nao informado`

## Origem dos Dados

Arquivo analisado:

- `RELATORIO DE ORCAMENTOS-25.xlsx`

Aba utilizada:

- `ORCAMENTOS`

Observacoes da aba:

- a linha de cabecalho util esta na linha 10
- existem aproximadamente `3156` linhas com conteudo util apos o cabecalho
- existem linhas vazias no final da aba e elas devem ser ignoradas
- alguns textos chegaram com acentuacao corrompida na leitura tecnica do arquivo, entao o processo deve normalizar caracteres antes de persistir

## Estrutura Observada da Aba

Cabecalhos identificados na linha 10:

1. `DATA`
2. `Nº DE ORCA`
3. `REV.`
4. `INSTALADOR`
5. `NOME DA OBRA`
6. `TIPO DE OBRA`
7. `VENDEDOR`
8. `CONTATO`
9. `VALOR BRUTO`
10. `COMISSAO`
11. `M2`
12. `PRIORIDADE`
13. `STATUS`
14. `CONCORRENTE`
15. `MOTIVO`
16. `VALOR CONCORRENTE`
17. `PROJETISTA`
18. `ESPECIFICACOES`
19. `COMISSAO 10mm`
20. coluna sem nome
21. `VALOR`
22. coluna sem nome
23. coluna sem nome

## Leitura Funcional das Colunas

Durante a analise foi observado que o conteudo de algumas colunas nao bate exatamente com o titulo exibido no cabecalho.

Interpretacao sugerida para a primeira versao da carga:

- `DATA`: data de envio do orcamento
- `Nº DE ORCA`: numero do orcamento
- `REV.`: revisao do orcamento
- `INSTALADOR`: nome do instalador
- `NOME DA OBRA`: nome do projeto ou obra
- `TIPO DE OBRA`: tipo do projeto
- `VENDEDOR`: nome do vendedor
- `CONTATO`: nome do contato
- `VALOR BRUTO`: valor bruto do orcamento
- `COMISSAO`: valor ou percentual de comissao, depende de validacao com negocio
- `M2`: area em metros quadrados
- `PRIORIDADE`: aparentemente contem o status macro do orcamento, com valores como `FECHADO`, `CANCELADO`, `PERDIDO`, `COMPRA`, `ORCAMENTO`
- `STATUS`: aparentemente contem acompanhamento textual, com valores como `Negociando`, `Aguardando aprovacao`, `Apenas consulta`
- `CONCORRENTE`: nome do concorrente
- `MOTIVO`: motivo da perda ou observacao de perda
- `VALOR CONCORRENTE`: valor informado do concorrente
- `PROJETISTA`: nome do projetista
- `ESPECIFICACOES`: detalhe tecnico livre

Colunas a principio fora do escopo da carga:

- `COMISSAO 10mm`
- coluna 20 sem nome
- `VALOR`
- colunas 22 e 23 sem nome

Essas colunas podem ficar registradas para estudo futuro, mas nao precisam entrar na primeira versao se nao houver destino claro no banco.

## Tabelas Atuais do Sistema

Tabela principal:

- `budgets`

Campos principais relevantes:

- `budget_number`
- `year_budget`
- `revision`
- `sent_at`
- `gross_value`
- `commission_value`
- `area_m2`
- `status_id`
- `priority_id`
- `installer_id`
- `project_id`
- `salesperson_id`
- `contact_id`
- `loss_reason_id`
- `competitor_name`
- `competitor_price`
- `designer_name`
- `specification_details`
- `current_follow_up`

Tabelas auxiliares relevantes:

- `budget_statuses`
- `priorities`
- `installers`
- `projects`
- `project_types`
- `salespeople`
- `contacts`
- `loss_reasons`

## Mapeamento Sugerido Planilha x Banco

### Orcamento

| Coluna da planilha | Campo alvo | Regra sugerida |
| --- | --- | --- |
| `DATA` | `sent_at` | Converter numero serial do Excel para data e salvar com horario padrao `00:00:00` |
| `Nº DE ORCA` | `budget_number` | Salvar como texto normalizado |
| `REV.` | `revision` | Extrair numero; vazio, `-` ou invalido vira `0` |
| `VALOR BRUTO` | `gross_value` | Converter para decimal |
| `COMISSAO` | `commission_value` | Converter para decimal; validar com negocio se o valor representa percentual ou valor monetario |
| `M2` | `area_m2` | Converter para decimal; vazio vira `0` |
| `PRIORIDADE` | `status_id` | Buscar ou criar item em `budget_statuses` pelo nome normalizado |
| `STATUS` | `current_follow_up` | Salvar como texto livre |
| `CONCORRENTE` | `competitor_name` | Salvar texto; se vazio usar `Nao informado` |
| `MOTIVO` | `loss_reason_id` | Buscar ou criar item em `loss_reasons` pelo nome normalizado |
| `VALOR CONCORRENTE` | `competitor_price` | Converter para decimal; se vazio manter `null` ou `0` conforme regra de negocio |
| `PROJETISTA` | `designer_name` | Salvar texto; se vazio usar `Nao informado` |
| `ESPECIFICACOES` | `specification_details` | Salvar texto livre |

### Relacionamentos auxiliares

| Coluna da planilha | Tabela alvo | Regra sugerida |
| --- | --- | --- |
| `INSTALADOR` | `installers` | Buscar ou criar por nome |
| `NOME DA OBRA` | `projects` | Buscar ou criar por nome |
| `TIPO DE OBRA` | `project_types` | Buscar ou criar por nome e relacionar ao projeto |
| `VENDEDOR` | `salespeople` | Buscar ou criar por nome |
| `CONTATO` | `contacts` | Buscar ou criar por nome dentro do instalador |

## Regra de `Nao informado`

Regra solicitada:

- se a informacao existir na tabela de destino mas vier vazia na planilha, a carga deve preencher como `Nao informado`
- para as tabelas auxiliares deve existir um item padrao `Nao informado`

Padronizacao recomendada para considerar um valor como ausente:

- celula vazia
- espacos em branco
- `-`
- `N/E`
- `N/I`
- `null`

### Campos textuais da tabela `budgets`

Quando vierem vazios, preencher com `Nao informado`:

- `competitor_name`
- `designer_name`
- `specification_details`
- `current_follow_up`

Observacao:

- `current_follow_up` tambem pode receber o valor original da coluna `STATUS` quando ele existir

### Campos numericos da tabela `budgets`

Quando vierem vazios, preencher com valor tecnico padrao:

- `gross_value`: `0`
- `commission_value`: `0`
- `area_m2`: `0`
- `revision`: `0`

### Campos de relacionamento

Quando vierem vazios, apontar para o item `Nao informado` da tabela auxiliar correspondente:

- `status_id`
- `priority_id`
- `installer_id`
- `project_id`
- `salesperson_id`
- `contact_id`
- `loss_reason_id`

## Itens Padrao nas Tabelas Auxiliares

Para garantir consistencia da importacao, a aplicacao deve manter um registro padrao `Nao informado` nas tabelas auxiliares usadas pela carga.

Sugestao de seed tecnico:

### `budget_statuses`

- `code`: `nao_informado`
- `name`: `Nao informado`
- `description`: `Status padrao para importacao`
- `is_final`: `false`
- `sort_order`: `999`

### `priorities`

- `code`: `nao_informado`
- `name`: `Nao informado`
- `weight`: `0`

### `loss_reasons`

- `code`: `nao_informado`
- `name`: `Nao informado`
- `description`: `Motivo nao informado na planilha`
- `active`: `true`

### `project_types`

- `code`: `nao_informado`
- `name`: `Nao informado`
- `description`: `Tipo de obra nao informado`

### `projects`

- `name`: `Nao informado`
- `project_type_id`: item `Nao informado` de `project_types`
- `city`: `Nao informado`
- `state`: `NI`
- `notes`: `Projeto padrao para importacao`

### `salespeople`

- `name`: `Nao informado`
- `email`: `nao-informado-salesperson@local`
- `phone`: `00000000000`
- `active`: `true`

### `installers`

- `name`: `Nao informado`
- `email`: `nao-informado-installer@local`
- `phone`: `00000000000`
- `city`: `Nao informado`
- `state`: `NI`
- `notes`: `Instalador padrao para importacao`
- `active`: `true`

### `contacts`

Como `contacts` exige `installer_id`, `email` e `phone`, o contato padrao deve ficar associado ao instalador padrao.

- `installer_id`: instalador `Nao informado`
- `name`: `Nao informado`
- `email`: `nao-informado-contact@local`
- `phone`: `00000000000`
- `role`: `Nao informado`
- `is_primary`: `false`

## Proposta de Tela de Carga

## Objetivo da tela

Permitir que o usuario envie um arquivo Excel, valide a aba `ORCAMENTOS`, visualize um resumo do que sera importado e execute a carga com seguranca.

## Estrutura sugerida

### 1. Cabecalho

- titulo: `Carga de planilha de orcamentos`
- subtitulo: `Importe dados da aba ORCAMENTOS e converta para a base do sistema`

### 2. Card de upload

Campos e acoes:

- campo para selecionar arquivo `.xlsx`
- informacao fixa: `A aba utilizada sera ORCAMENTOS`
- botao `Validar arquivo`
- botao `Importar planilha`

### 3. Resumo da leitura

Exibir apos validacao:

- nome do arquivo
- aba encontrada
- quantidade de linhas validas
- quantidade de linhas vazias ignoradas
- quantidade de colunas reconhecidas
- quantidade de colunas ignoradas

### 4. Preview do mapeamento

Tabela de conferência com colunas:

- coluna da planilha
- destino no sistema
- tipo de tratamento
- regra de `Nao informado`

### 5. Resultado da validacao

Blocos separados:

- `Pronto para importar`
- `Avisos`
- `Erros impeditivos`

Exemplos de avisos:

- tipo de obra vazio, sera usado `Nao informado`
- instalador nao encontrado, sera criado
- contato vazio, sera usado contato padrao
- revisao invalida, sera usada revisao `0`

Exemplos de erros impeditivos:

- aba `ORCAMENTOS` nao encontrada
- arquivo invalido
- `Nº DE ORCA` vazio
- data impossivel de converter

### 6. Resultado da importacao

Ao finalizar, apresentar:

- total lido
- total importado
- total atualizado
- total ignorado
- total com erro

E tambem uma grade final por linha:

- numero da linha
- numero do orcamento
- status da importacao
- mensagem

## Fluxo Recomendado

1. Usuario seleciona o arquivo
2. Sistema abre somente a aba `ORCAMENTOS`
3. Sistema identifica a linha 10 como cabecalho
4. Sistema ignora linhas totalmente vazias
5. Sistema normaliza textos e acentos
6. Sistema converte tipos numericos e datas
7. Sistema resolve referencias nas tabelas auxiliares
8. Sistema aplica `Nao informado` quando necessario
9. Sistema mostra preview e avisos
10. Usuario confirma a importacao
11. Sistema grava e apresenta relatorio final

## Regras Tecnicas Recomendadas

- usar importacao em lote por paginas, evitando processar tudo em memoria em uma unica operacao
- manter transacao por bloco para facilitar rollback controlado
- gerar log por linha importada
- normalizar comparacoes por texto sem diferenciar maiusculas, minusculas e acentos
- evitar duplicidade usando chave composta `budget_number + year_budget`

## Regra de Duplicidade

Como a tabela `budgets` possui restricao unica em `budget_number` e `year_budget`, a importacao deve seguir uma destas opcoes:

- `modo atualizar`: se ja existir, atualiza os dados
- `modo ignorar`: se ja existir, apenas registra aviso

Para a primeira versao, a recomendacao e:

- validar duplicidade antes da gravacao
- exibir contagem de registros novos e existentes
- permitir ao usuario escolher entre `Atualizar existentes` ou `Ignorar existentes`

## Pontos de Atencao

- a coluna `PRIORIDADE` da planilha parece representar o `status` do orcamento, nao a prioridade operacional do sistema
- a coluna `STATUS` da planilha parece representar acompanhamento textual, nao o status catalogado
- o campo `COMISSAO` precisa de confirmacao com negocio para saber se representa percentual ou valor monetario
- a coluna `DATA` vem como numero serial do Excel e precisa de conversao correta
- a tabela `contacts` exige `installer_id`, `email` e `phone`, entao o item padrao deve ser criado com dados tecnicos fixos

## Recomendacao Final

A primeira versao da tela deve:

- aceitar somente `.xlsx`
- processar somente a aba `ORCAMENTOS`
- trabalhar com preview antes da importacao
- criar ou reutilizar itens auxiliares normalizados
- usar `Nao informado` como padrao para valores ausentes
- ignorar inicialmente as colunas sem destino claro no banco

Essa abordagem permite colocar a carga em producao com menor risco e deixa preparado o caminho para evoluir o importador depois, caso as colunas extras passem a ter destino definido no modelo de dados.
