# Backlog Tecnico de Implementacao da Importacao Rocktec x Trox

## Objetivo

Este backlog detalha a implementacao tecnica para evoluir o importador atual, hoje orientado ao layout da Rocktec, para um modelo multi-layout capaz de suportar tambem a planilha da Trox.

Este documento complementa o estudo:

- [ESTUDO-IMPORTACAO-ROCKTEC-TROX.md](file:///c:/Users/marco/OneDrive/Projects/go-gerenciador-orcamento-deck/docs/ESTUDO-IMPORTACAO-ROCKTEC-TROX.md)

## Diretriz Geral

A diretriz recomendada para esta evolucao e:

- manter uma unica base principal de dominio
- nao duplicar tabelas principais de negocio por empresa
- separar a logica de importacao por layout
- normalizar os dados em um modelo intermediario comum
- registrar a origem e o lote de importacao

## Convencoes

- codigo, structs, tabelas e campos em `ingles`
- documentacao em `portugues`
- banco `PostgreSQL`
- backend em `Go`
- frontend em `TypeScript`

## Escala de Prioridade

- `P0`: fundacional ou bloqueante
- `P1`: muito importante para colocar em uso
- `P2`: importante para consolidacao
- `P3`: melhoria futura

## Ordem Macro Recomendada

1. estabilizar a fundacao do importador multi-layout
2. migrar o importador Rocktec para a nova arquitetura
3. mapear formalmente o layout Trox
4. implementar parser e preview Trox
5. adicionar rastreabilidade de importacao
6. evoluir UX e operacao

## Fase 1 - Fundacao do Importador Multi-Layout

### TASK-IMP-001 - Mapear acoplamentos do importador atual

- `Prioridade`: `P0`
- `Dependencias`: nenhuma
- `Descricao`: identificar todos os pontos do importador atual que assumem regras fixas da Rocktec, como nome da aba, linha do cabecalho, validacao, parse de data e posicao das colunas.
- `Pronto quando`: existir um inventario tecnico dos acoplamentos no backend e um plano claro de extracao.

### TASK-IMP-002 - Definir contrato de layout de importacao

- `Prioridade`: `P0`
- `Dependencias`: `TASK-IMP-001`
- `Descricao`: criar uma interface ou estrategia de layout com responsabilidades claras para detectar a planilha, identificar a aba, validar cabecalho, extrair colunas e normalizar valores.
- `Pronto quando`: a aplicacao puder instanciar um layout especifico sem depender de `if` espalhado no service principal.

### TASK-IMP-003 - Criar DTO intermediario comum

- `Prioridade`: `P0`
- `Dependencias`: `TASK-IMP-002`
- `Descricao`: definir um modelo intermediario normalizado, por exemplo `NormalizedBudgetImportRow`, para servir de ponte entre qualquer layout e o dominio atual.
- `Pronto quando`: existir um unico payload intermediario capaz de representar Rocktec e Trox sem persistencia acoplada ao formato de origem.

### TASK-IMP-004 - Separar pipeline em etapas claras

- `Prioridade`: `P0`
- `Dependencias`: `TASK-IMP-002`, `TASK-IMP-003`
- `Descricao`: reorganizar o fluxo em etapas como leitura do arquivo, identificacao do layout, preview, normalizacao, validacao e persistencia.
- `Pronto quando`: o service principal estiver orquestrando etapas e nao mais executando tudo de forma monolitica.

## Fase 2 - Migracao da Rocktec para a Nova Arquitetura

### TASK-IMP-005 - Implementar layout Rocktec no novo contrato

- `Prioridade`: `P0`
- `Dependencias`: `TASK-IMP-003`, `TASK-IMP-004`
- `Descricao`: encapsular as regras da Rocktec em uma implementacao propria, preservando aba `ORCAMENTOS`, linha `10`, parse atual e mapeamento ja conhecido.
- `Pronto quando`: a Rocktec deixar de depender de regras fixas espalhadas no service principal.

### TASK-IMP-006 - Migrar preview da Rocktec

- `Prioridade`: `P0`
- `Dependencias`: `TASK-IMP-005`
- `Descricao`: fazer o preview da Rocktec passar pelo novo parser e pelo DTO intermediario comum.
- `Pronto quando`: a funcionalidade de preview continuar funcionando com o mesmo comportamento funcional atual.

### TASK-IMP-007 - Migrar execucao da Rocktec

- `Prioridade`: `P0`
- `Dependencias`: `TASK-IMP-006`
- `Descricao`: adaptar a execucao/importacao da Rocktec para usar o novo fluxo normalizado.
- `Pronto quando`: a importacao efetiva da Rocktec continuar funcionando apos a refatoracao.

### TASK-IMP-008 - Garantir regressao automatizada da Rocktec

- `Prioridade`: `P0`
- `Dependencias`: `TASK-IMP-006`, `TASK-IMP-007`
- `Descricao`: revisar e ampliar testes unitarios e de integracao para garantir que a refatoracao nao quebre a importacao atual.
- `Pronto quando`: a suite automatizada cobrir preview e execucao da Rocktec dentro da nova arquitetura.

## Fase 3 - Mapeamento Formal da Trox

### TASK-IMP-009 - Catalogar colunas da Trox

- `Prioridade`: `P1`
- `Dependencias`: nenhuma
- `Descricao`: consolidar o mapeamento das colunas observadas da Trox, incluindo semantica funcional e destino esperado.
- `Pronto quando`: existir tabela de mapeamento Trox x sistema aprovada para implementacao tecnica.

### TASK-IMP-010 - Validar regras de negocio da Trox

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-009`
- `Descricao`: definir com clareza o destino dos campos `Tipo`, `Status`, `Linha de produtos`, `Código Cliente`, `Nome Cliente` e `Fator Médio`.
- `Pronto quando`: cada campo da Trox estiver classificado como dominio principal, staging ou ignorado.

### TASK-IMP-011 - Definir politica de datas e numeros da Trox

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-009`
- `Descricao`: especificar parse de `Data de Emissão`, `Total do orçamento` e `Fator Médio`, considerando formatos textuais, monetarios e separadores locais.
- `Pronto quando`: as regras de conversao da Trox estiverem documentadas e testaveis.

## Fase 4 - Implementacao do Layout Trox

### TASK-IMP-012 - Implementar parser Trox

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-003`, `TASK-IMP-010`, `TASK-IMP-011`
- `Descricao`: criar uma implementacao de layout Trox considerando aba `Capa`, cabecalho na linha `1`, colunas textuais e destino dos campos normalizados.
- `Pronto quando`: o backend conseguir gerar `NormalizedBudgetImportRow` a partir da planilha Trox.

### TASK-IMP-013 - Implementar preview Trox

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-012`
- `Descricao`: habilitar preview da Trox com resumo, avisos e erros no mesmo padrao do preview atual.
- `Pronto quando`: a Trox puder passar pela etapa de validacao antes da importacao.

### TASK-IMP-014 - Implementar execucao Trox

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-013`
- `Descricao`: permitir a importacao efetiva da Trox usando o pipeline comum e persistindo no modelo principal.
- `Pronto quando`: o fluxo Trox conseguir gravar orcamentos e relacionamentos com sucesso.

### TASK-IMP-015 - Criar testes automatizados da Trox

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-013`, `TASK-IMP-014`
- `Descricao`: adicionar testes unitarios e de integracao cobrindo o fluxo principal e cenarios alternativos da Trox.
- `Pronto quando`: a suite automatizada garantir a estabilidade do novo layout.

## Fase 5 - Banco e Rastreabilidade

### TASK-IMP-016 - Criar tabela de lotes de importacao

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-004`
- `Descricao`: adicionar uma tabela como `budget_import_batches` para rastrear origem, arquivo, usuario, status e resumo de cada importacao.
- `Pronto quando`: cada importacao puder ser auditada como um lote identificavel.

### TASK-IMP-017 - Criar tabela de staging de linhas importadas

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-016`
- `Descricao`: adicionar uma tabela como `budget_import_rows_raw` para armazenar linha bruta, linha normalizada, status e mensagens.
- `Pronto quando`: for possivel rastrear tecnicamente o que foi lido e transformado em cada linha.

### TASK-IMP-018 - Adicionar origem no dominio principal

- `Prioridade`: `P1`
- `Dependencias`: `TASK-IMP-016`
- `Descricao`: avaliar e implementar campos como `source_company`, `source_layout` e `import_batch_id` no dominio principal ou em relacao vinculada.
- `Pronto quando`: cada registro importado tiver origem claramente identificavel.

### TASK-IMP-019 - Avaliar extensao do modelo principal

- `Prioridade`: `P2`
- `Dependencias`: `TASK-IMP-010`
- `Descricao`: decidir se dados como `customer_code`, `customer_name`, `product_line` e `average_factor` entram em `budgets` ou permanecem apenas em staging.
- `Pronto quando`: o destino desses dados estiver decidido, documentado e refletido no modelo.

## Fase 6 - UX da Tela de Importacao

### TASK-IMP-020 - Exibir layout identificado no preview

- `Prioridade`: `P2`
- `Dependencias`: `TASK-IMP-013`
- `Descricao`: mostrar na tela qual layout foi identificado, qual empresa de origem foi reconhecida e qual estrategia sera usada.
- `Pronto quando`: o usuario entender claramente qual parser esta sendo aplicado.

### TASK-IMP-021 - Permitir escolha ou deteccao do layout

- `Prioridade`: `P2`
- `Dependencias`: `TASK-IMP-012`
- `Descricao`: definir se o usuario escolhe manualmente a empresa/layout ou se o sistema detecta automaticamente com possibilidade de confirmacao.
- `Pronto quando`: a UX do upload suportar multiplas empresas com baixa ambiguidade.

### TASK-IMP-022 - Melhorar exibicao de colunas ignoradas

- `Prioridade`: `P2`
- `Dependencias`: `TASK-IMP-013`
- `Descricao`: mostrar de forma clara quais campos foram mapeados, quais foram enviados para staging e quais foram ignorados.
- `Pronto quando`: o preview ficar transparente o suficiente para uso operacional.

## Fase 7 - Operacao e Governanca

### TASK-IMP-023 - Definir estrategia de seeds e defaults por layout

- `Prioridade`: `P2`
- `Dependencias`: `TASK-IMP-005`, `TASK-IMP-012`
- `Descricao`: revisar `Nao informado`, catálogos auxiliares e comportamento padrao para Rocktec e Trox.
- `Pronto quando`: os dois layouts tiverem estrategia clara e consistente para valores ausentes.

### TASK-IMP-024 - Definir politica de duplicidade por origem

- `Prioridade`: `P2`
- `Dependencias`: `TASK-IMP-014`
- `Descricao`: revisar se a chave de duplicidade continua sendo apenas `budget_number + year_budget` ou se exige considerar a origem da empresa.
- `Pronto quando`: a regra de deduplicacao estiver formalizada para importacoes multiempresa.

### TASK-IMP-025 - Definir observabilidade do processo de importacao

- `Prioridade`: `P3`
- `Dependencias`: `TASK-IMP-016`, `TASK-IMP-017`
- `Descricao`: adicionar logs estruturados, contadores e rastros suficientes para monitorar importacoes multi-layout em producao.
- `Pronto quando`: o processo puder ser auditado e investigado com baixo esforço.

## Sequencia Recomendada de Execucao

Ordem sugerida para comecar:

1. `TASK-IMP-001`
2. `TASK-IMP-002`
3. `TASK-IMP-003`
4. `TASK-IMP-004`
5. `TASK-IMP-005`
6. `TASK-IMP-006`
7. `TASK-IMP-007`
8. `TASK-IMP-008`
9. `TASK-IMP-009`
10. `TASK-IMP-010`
11. `TASK-IMP-011`
12. `TASK-IMP-012`
13. `TASK-IMP-013`
14. `TASK-IMP-014`
15. `TASK-IMP-015`

Depois disso:

- avaliar banco e rastreabilidade
- evoluir UX
- reforcar observabilidade

## Recomendacao de Inicio Imediato

Se a implementacao for comecar agora, o melhor proximo passo tecnico e:

- iniciar pela `Fase 1`
- nao mexer ainda no banco
- refatorar primeiro a fundacao do importador atual
- preservar a Rocktec funcionando antes de ligar a Trox

Essa ordem reduz risco porque:

- evita quebrar o fluxo em producao
- cria base boa para multiplos layouts
- diminui retrabalho futuro

## Entregavel Minimo da Primeira Onda

Uma primeira onda bem sucedida de implementacao deve entregar:

- importador multi-layout fundacional
- Rocktec migrada sem regressao
- parser Trox com preview funcional
- backlog de campos exclusivos decidido

Somente depois disso vale acoplar com mais profundidade:

- staging persistente
- analytics de importacao
- expansao de dominio com novos campos

## Conclusao

Este backlog organiza a evolucao do importador para suportar Rocktec e Trox sem fragmentar o produto.

O direcionamento recomendado continua sendo:

- um dominio principal unico
- layouts de importacao separados
- normalizacao intermediaria
- rastreabilidade progressiva

Essa abordagem e a mais segura tecnicamente e a mais sustentavel para a evolucao comercial do sistema.
