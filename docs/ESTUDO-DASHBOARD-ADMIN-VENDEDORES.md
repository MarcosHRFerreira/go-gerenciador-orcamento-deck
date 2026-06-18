# Estudo Dashboard Administrativo De Vendedores

## Objetivo

Definir uma proposta funcional e tecnica para evoluir o dashboard do perfil `admin` com foco em acompanhamento de vendedores, produtividade comercial, conversao e atividade recente.

## Contexto Atual

- O dashboard atual do frontend usa os dados de orcamentos para montar indicadores em memoria no navegador.
- A pagina atual esta em `frontend/src/features/dashboard/pages/DashboardPage.tsx`.
- A consulta atual reutiliza a listagem de orcamentos com filtros e pagina grandes, em vez de consumir endpoints agregados especificos de dashboard.
- O backend ja expoe informacoes suficientes para listar orcamentos com `salesperson_id`, `salesperson_name`, `status`, `gross_value`, `sent_at`, `updated_at`, `current_follow_up` e `project_id`.
- O sistema ja possui historico de status e follow-up, o que abre espaco para indicadores de atividade mais confiaveis em etapas futuras.

## Objetivo De Negocio

O dashboard do `admin` deve responder rapidamente perguntas como:

- quem esta vendendo mais
- quem esta gerando mais orcamentos
- quem esta com mais valor em negociacao
- quem esta sem atividade recente
- quais oportunidades estao paradas
- quais vendedores convertem melhor

## Proposta Funcional

## 1. Cards Executivos

Primeira linha do dashboard com resumo geral do periodo selecionado:

- `Total de vendedores ativos`
- `Total de orcamentos no periodo`
- `Valor bruto total`
- `Ticket medio`
- `Total em negociacao`
- `Taxa de conversao`

### Definicoes

- `vendedores ativos`: vendedores com ao menos um orcamento no periodo
- `ticket medio`: `valor bruto total / quantidade de orcamentos`
- `total em negociacao`: soma do valor bruto dos orcamentos com status de negociacao
- `taxa de conversao`: percentual de orcamentos com status final vencedor sobre o total do periodo

## 2. Rankings Principais

Blocos visuais com `Top 10`:

- `Top 10 vendedores por valor orcado`
- `Top 10 vendedores por quantidade de orcamentos`
- `Top 10 vendedores por valor em negociacao`
- `Top 10 vendedores por conversao`
- `Top 10 vendedores por ticket medio`

### Recomendacao

Para evitar distorcao em rankings de conversao e ticket medio:

- aplicar quantidade minima de orcamentos no periodo, por exemplo `>= 5`
- permitir alternar entre `valor` e `quantidade`

## 3. Ultima Atividade

Tabela ou lista lateral com:

- `vendedor`
- `ultimo orcamento movimentado`
- `ultima acao`
- `data/hora`
- `status atual`

### O Que Considerar Como Atividade

Ordem recomendada de prioridade:

- alteracao de status do orcamento
- inclusao de follow-up
- atualizacao do orcamento
- criacao/importacao do orcamento

### Beneficio

Permite ao `admin` identificar rapidamente quem esta atuando no funil e quem esta parado.

## 4. Carteira Em Aberto

Bloco para acompanhar pipeline por vendedor:

- `quantidade em negociacao`
- `valor em negociacao`
- `quantidade sem follow-up recente`
- `quantidade sem atualizacao ha X dias`

### Faixas Recomendadas

- `sem atividade ha 7 dias`
- `sem atividade ha 15 dias`
- `sem atividade ha 30 dias`

## 5. Orçamentos Parados

Lista priorizada para acao do `admin`:

- vendedor
- numero do orcamento
- obra
- construtora
- status
- valor bruto
- ultima atualizacao
- dias sem atividade

### Regra Sugerida

Considerar parado quando:

- `updated_at` ou ultima atividade for anterior a `7` dias
- e o status ainda nao for finalizador como `Pedido` ou `Cancelado`

## 6. Funil Comercial Por Vendedor

Grafico comparativo por vendedor com etapas como:

- `Orcamento`
- `Em negociacao`
- `Pedido`
- `Cancelado`

### Objetivo

Permitir leitura rapida de onde cada vendedor concentra o volume e em que etapa perde oportunidades.

## 7. Evolucao Mensal

Grafico de linha ou barras por vendedor com:

- quantidade de orcamentos por mes
- valor bruto por mes
- valor convertido por mes

### Utilidade

Permite comparar consistencia, sazonalidade e tendencia.

## Filtros Recomendados

Filtros globais no topo do dashboard:

- periodo
- empresa origem
- vendedor
- status
- obra
- construtora
- faixa de valor

### Recomendacao De UX

- manter filtros simples na primeira fase
- iniciar com `periodo`, `empresa origem` e `vendedor`
- evoluir filtros adicionais em fase seguinte

## Estrutura Recomendada De Tela

## Linha 1

- cards executivos

## Linha 2

- `Top 10 vendedores por valor`
- `Top 10 vendedores por quantidade`

## Linha 3

- `Carteira em aberto por vendedor`
- `Funil por vendedor`

## Linha 4

- `Ultima atividade`
- `Orcamentos parados`

## Linha 5

- `Evolucao mensal`

## Viabilidade Tecnica

## O Que Ja Da Para Fazer Rapido

Com os dados atuais de listagem de orcamentos, ja e possivel montar no frontend:

- total de orcamentos por vendedor
- valor bruto total por vendedor
- ticket medio
- rankings por quantidade e valor
- distribuicao por status
- carteira em aberto baseada em `status_name`
- itens parados usando `updated_at`

### Vantagem

- baixo custo inicial
- reaproveita API existente
- entrega rapida de valor

### Limitacoes

- calculo pesado no navegador conforme a base crescer
- leituras agregadas limitadas pelo total de registros carregados
- `ultima atividade` pode ficar imprecisa se usar apenas `updated_at`
- filtros mais complexos aumentam custo de processamento no cliente

## O Que Vale Evoluir No Backend

Para uma solucao mais robusta, o ideal e criar endpoints especificos de dashboard.

### Endpoints Sugeridos

```text
GET /dashboard/admin/salespeople/summary
GET /dashboard/admin/salespeople/rankings
GET /dashboard/admin/salespeople/pipeline
GET /dashboard/admin/salespeople/recent-activities
GET /dashboard/admin/salespeople/stale-budgets
GET /dashboard/admin/salespeople/monthly-evolution
```

### Beneficios

- consultas agregadas direto no banco
- menor trafego entre backend e frontend
- melhor tempo de resposta
- regras centralizadas no backend
- mais facilidade para paginacao de listas analiticas

## Fontes De Dados Recomendadas

## Fonte Imediata

- `budgets`
- `salespeople`
- `budget_statuses`
- `projects`
- `budget_follow_ups`
- `budget_status_history`

## Regras Relevantes

- `valor total`: `SUM(budgets.gross_value)`
- `quantidade`: `COUNT(*)`
- `ultima atividade`: `MAX` entre follow-up, historico de status e `updated_at`
- `conversao`: status final `Pedido` sobre total do periodo
- `em negociacao`: status em aberto configurados por nome

## Risco Tecnico

Hoje os status parecem ser dirigidos por nome em alguns pontos do frontend. Para o dashboard administrativo, o ideal e padronizar no backend uma classificacao de status por categoria:

- `open`
- `won`
- `lost`
- `cancelled`

### Beneficio

Evita regras espalhadas como comparacao textual de `Pedido`, `Cancelado` ou `Em negociacao`.

## Proposta De Fases

## Fase 1

Entrega rapida com reaproveitamento da API atual:

- cards executivos
- top 10 por valor
- top 10 por quantidade
- carteira em aberto por vendedor
- orcamentos parados por vendedor usando `updated_at`

## Fase 2

Melhoria analitica:

- ultima atividade por vendedor
- funil comercial por vendedor
- evolucao mensal
- filtros por periodo e vendedor mais refinados

## Fase 3

Escalabilidade:

- endpoints agregados especificos de dashboard
- consultas SQL otimizadas
- classificacao de status por categoria
- paginacao e ordenacao para listas analiticas

## Recomendacao Final

A melhor estrategia e iniciar com uma `Fase 1` enxuta, focada em valor gerencial imediato:

- `Top 10 vendedores por valor`
- `Top 10 vendedores por quantidade`
- `Carteira em negociacao por vendedor`
- `Orcamentos parados`
- `Resumo executivo do periodo`

Isso entrega um dashboard util para a administracao sem depender de uma reestruturacao completa do backend. Em seguida, a `Fase 2` pode evoluir para atividade real e funil comercial, e a `Fase 3` consolida tudo em endpoints agregados.

## Sugestao De Nome Do Bloco No Sistema

- `Performance de vendedores`
- `Painel comercial`
- `Resumo comercial por vendedor`

### Nome Recomendado

`Performance de vendedores`

Porque comunica claramente o objetivo do bloco para o perfil `admin`.
