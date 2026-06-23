# Estudo de Data de Entrega em Orcamentos e Avisos ao Vendedor

## Objetivo

Este estudo descreve uma proposta funcional e tecnica para adicionar ao modulo de orcamentos:

- um campo `data de entrega`
- uma regra de uso vinculada ao status `Pedido`
- um mecanismo de monitoramento automatico
- um aviso ao vendedor responsavel quando faltarem `2 dias` para a entrega
- uma tela de consulta para acompanhar os orcamentos com data de entrega preenchida

Tambem considera o reaproveitamento da tela de `Comunicacao` como canal interno de aviso ao vendedor.

## Conclusao Executiva

- A melhor abordagem e tratar `data de entrega` como um campo do proprio orcamento.
- O preenchimento deve ficar disponivel quando o status estiver em `Pedido`, sem obrigatoriedade tecnica de preenchimento.
- O monitoramento nao deve depender de acesso do usuario na tela; deve rodar por processo automatico no backend.
- O canal mais aderente ao sistema atual para aviso interno e a propria `Comunicacao`, via mensagem automatica do sistema para o vendedor vinculado ao orcamento.
- A nova tela de consulta deve funcionar como um painel operacional de acompanhamento de entregas, com foco em prazo, atraso, proximidade e responsavel.

## Problema de Negocio

Hoje o sistema controla o orcamento e seu status, mas nao ha um acompanhamento operacional estruturado da previsao de entrega depois que o orcamento vira `Pedido`.

Isso gera alguns riscos:

- perda de previsibilidade sobre entregas proximas
- falta de aviso proativo ao vendedor
- dificuldade para consultar rapidamente os pedidos com entrega agendada
- risco de atraso sem acompanhamento centralizado

## Regra de Negocio Proposta

### Campo Novo

Adicionar no cadastro de orcamento:

- `delivery_date`

Tipo sugerido:

- `date`

Justificativa:

- a necessidade descrita e de acompanhamento por dia, nao por hora/minuto
- simplifica comparacoes, filtros e exibicao

### Regra de Uso

- o campo deve ficar disponivel para preenchimento no formulario de orcamento
- ele passa a ter relevancia operacional quando o `status` estiver como `Pedido`
- se o status for `Pedido`, o sistema deve permitir o preenchimento da `data de entrega`, mas sem bloquear o salvamento quando ela nao for informada

### Recomendacao

Para manter flexibilidade operacional, a recomendacao e:

- `status = Pedido` => `data de entrega` opcional, porem recomendada

Como apoio de usabilidade, a sugestao e:

1. exibir um destaque visual quando o status for `Pedido` e a data estiver vazia
2. permitir filtros e consultas especificas para identificar pedidos sem data de entrega

## Modelo de Dados Sugerido

### Tabela `budgets`

Adicionar colunas como:

- `delivery_date date null`
- `delivery_notification_sent_at timestamp null`
- `delivery_notification_reference_date date null`

## Finalidade das Colunas Auxiliares

- `delivery_date`
  - armazena a previsao de entrega do pedido
- `delivery_notification_sent_at`
  - registra quando o aviso automatico foi enviado
- `delivery_notification_reference_date`
  - registra para qual data de entrega o aviso foi gerado

Essas colunas ajudam a evitar reenvio duplicado.

## Regra de Monitoramento

### Condicao Principal

O mecanismo deve procurar orcamentos em que:

- `status = Pedido`
- `delivery_date` preenchida
- vendedor vinculado ao orcamento preenchido
- faltam `2 dias` para a `delivery_date`
- ainda nao foi enviado aviso para essa entrega

### Exemplo

Se hoje for `10/07/2026`:

- entrega em `12/07/2026` => deve gerar aviso
- entrega em `11/07/2026` => nao entra na regra de 2 dias, mas pode entrar como `urgente`
- entrega em `10/07/2026` => entrega hoje
- entrega em `09/07/2026` => atrasado

## Tratamento Para Pedido Sem Data de Entrega

Como a `data de entrega` nao sera obrigatoria, o sistema deve tratar explicitamente os pedidos que estiverem com esse campo vazio.

### Comportamento Recomendado

- `status = Pedido` e `delivery_date` vazia => o registro nao entra no job de aviso de entrega
- mesmo assim, o registro deve aparecer em consultas operacionais como pendencia de preenchimento
- essa pendencia deve ficar visivel para vendedor e administracao

### Motivo

Se o sistema apenas ignorar esses casos, o usuario perde visibilidade justamente dos pedidos que mais precisam de acao manual.

### Classificacao Sugerida

Adicionar uma situacao derivada complementar:

- `Pedido sem data de entrega`

Essa situacao nao representa atraso, mas sim ausencia de informacao operacional.

## Melhor Mecanismo Tecnico

### Opcao Recomendada

Implementar um processo agendado no backend, executado por `job` periodico.

Exemplos de execucao:

- a cada 1 hora
- ou 1 vez por dia no inicio da manha

### Recomendacao Pratica

Rodar:

- 1 vez por dia, por exemplo `07:00`

Motivos:

- a regra e diaria, nao em tempo real
- reduz custo operacional
- simplifica observabilidade
- atende bem ao caso de uso

### Como o Job Deve Funcionar

1. consultar orcamentos com `status = Pedido`
2. filtrar os que possuem `delivery_date`
3. identificar os que vencem em `2 dias`
4. verificar se o aviso daquela entrega ja foi enviado
5. gerar mensagem interna para o vendedor
6. gravar os metadados para evitar duplicidade

## Como Evitar Avisos Duplicados

O sistema deve considerar o aviso enviado por combinacao de:

- `budget_id`
- `delivery_date`
- `tipo de aviso = entrega_em_2_dias`

Se a data de entrega mudar, o mecanismo pode enviar novo aviso para a nova data.

## Integracao com a Tela de Comunicacao

### Recomendacao Principal

Usar a tela de `Comunicacao` como canal interno de notificacao.

### Forma Sugerida

Criar uma mensagem automatica do sistema para o vendedor.

Exemplo de conteudo:

`Aviso automatico: o pedido do orcamento RKT5000461 da obra Centro Empresarial Campinas possui entrega prevista para 12/07/2026, faltando 2 dias para a data programada.`

### Vantagens

- reaproveita uma tela que ja existe no produto
- concentra avisos no canal interno do sistema
- evita criar um modulo paralelo de notificacoes logo na primeira fase
- permite historico de avisos enviados

### Recomendacao de Modelagem

Esses avisos devem ser registrados como mensagens de sistema, por exemplo:

- remetente logico: `Sistema`
- destinatario: vendedor do orcamento
- tipo: `delivery_reminder`

Se a arquitetura atual de comunicacao ainda nao tiver suporte a remetente de sistema, a recomendacao e introduzir isso como extensao do modulo.

## Sugestao de Conteudo da Mensagem

Campos uteis no aviso:

- numero do orcamento
- obra
- data de entrega
- status atual
- vendedor responsavel
- link de contexto para abrir o orcamento

### Modelo

`Aviso automatico de entrega`

- Orcamento: `TRX5000201`
- Obra: `Sr. Lucas (Apartamento 171)`
- Status: `Pedido`
- Entrega prevista: `12/07/2026`
- Acao recomendada: acompanhar a entrega do pedido

## Proposta da Tela de Consulta

### Objetivo

Criar uma tela operacional para acompanhar orcamentos com `data de entrega` preenchida.

### Nome Sugerido

- `Acompanhamento de Entregas`
- ou `Entregas de Pedidos`

### Melhor Localizacao

Duas opcoes sao boas:

- nova tela no menu principal
- aba dentro do modulo `Orcamentos`

### Recomendacao

Comecar como nova tela vinculada ao modulo de orcamentos, porque:

- o assunto e operacional
- a consulta tem vida propria
- facilita filtros especializados
- evita poluir a listagem principal de orcamentos

## Alerta Visual Na Tela de Orcamento

### Objetivo

Orientar o vendedor sem bloquear o fluxo de salvamento.

### No Formulario de Orcamento

Quando o status estiver em `Pedido` e a `data de entrega` estiver vazia, a sugestao e exibir:

- um `Alert` de aviso no topo do formulario
- destaque visual no campo `Data de entrega`
- texto de apoio explicando que a data e recomendada para monitoramento e aviso automatico

### Exemplo de Mensagem

`Este orcamento esta com status Pedido e ainda nao possui data de entrega. Sem essa informacao, o sistema nao conseguira gerar aviso automatico ao vendedor.`

### Na Listagem de Orcamentos

Se fizer sentido expor esse contexto tambem na listagem principal, a recomendacao e:

- exibir uma coluna opcional `Data de entrega`
- exibir um chip ou marcador quando:
  - a entrega estiver proxima
  - a entrega estiver atrasada
  - o pedido estiver sem data de entrega

Isso ajuda a identificar rapidamente os casos pendentes sem obrigar o usuario a abrir cada orcamento.

## Filtros Recomendados para a Tela

- data de entrega de
- data de entrega ate
- vendedor
- empresa
- obra
- numero do orcamento
- status
- situacao da entrega

## Situacoes Derivadas

A tela pode calcular uma coluna ou filtro `situacao da entrega`:

- `Atrasado`
- `Entrega hoje`
- `Entrega em 1 dia`
- `Entrega em 2 dias`
- `Entrega futura`
- `Pedido sem data de entrega`

Essa classificacao melhora muito a leitura operacional.

## Colunas Recomendadas na Grid

- numero do orcamento
- obra
- empresa
- vendedor
- status
- data de entrega
- dias para entrega
- situacao da entrega
- ultima atualizacao
- acoes

## Acoes Recomendadas na Tela

- abrir orcamento
- abrir obra
- abrir conversa/aviso do vendedor
- filtrar apenas atrasados
- filtrar apenas entregas proximas

## Especificacao da Tela Acompanhamento de Entregas

### Objetivo da Tela

- consolidar em uma unica consulta os pedidos com `data de entrega` preenchida e os pedidos sem data
- permitir leitura operacional rapida por vendedor, administracao e coordenacao
- reduzir a necessidade de abrir a listagem geral de orcamentos para localizar pendencias

### Estrutura Recomendada

- faixa superior com titulo, descricao curta e acao de limpar filtros
- bloco de cards resumo com os principais totais operacionais
- bloco de filtros com destaque para `situacao da entrega`, `vendedor` e faixa de data
- grid principal com colunas operacionais e acoes rapidas

### Cards Resumo

- `Total monitorado`
- `Atrasados`
- `Entrega hoje`
- `Entrega em ate 2 dias`
- `Pedidos sem data`

### Filtros da Tela

- `Situacao da entrega`
- `Data de entrega de`
- `Data de entrega ate`
- `Vendedor`
- `Empresa`
- `Obra`
- `Numero do orcamento`
- `Status`
- atalho rapido `Somente pedidos sem data`

### Grid Principal

Colunas recomendadas:

- `Orcamento`
- `Obra`
- `Empresa`
- `Vendedor`
- `Status`
- `Data de entrega`
- `Dias para entrega`
- `Situacao da entrega`
- `Ultima atualizacao`
- `Acoes`

### Acoes por Linha

- `Abrir orcamento`
- `Abrir obra`
- `Abrir comunicacao`

### Regras Visuais

- `Atrasado` em vermelho
- `Entrega hoje` em laranja
- `Entrega em 1 ou 2 dias` em destaque
- `Pedido sem data` com chip de pendencia
- ordenar por prioridade operacional antes da ordenacao alfabetica

### Ordenacao Recomendada

1. atrasados
2. entrega hoje
3. entrega em 1 dia
4. entrega em 2 dias
5. pedido sem data de entrega
6. entregas futuras

### Comportamento dos Filtros

- a tela deve abrir mostrando registros relevantes do contexto operacional
- se nao houver filtro informado, a recomendacao e priorizar pedidos com data preenchida e pedidos sem data
- filtros devem ser cumulativos
- a acao `Limpar` deve restaurar a consulta inicial

### Permissoes na Tela

- vendedor visualiza prioritariamente os proprios registros
- admin visualiza todos e pode consultar por vendedor
- acoes de abrir orcamento e comunicacao devem respeitar as permissoes ja existentes

### Estados da Interface

- estado vazio com mensagem orientando como preencher filtros
- estado sem resultados com indicacao de ajuste de filtros
- loading com skeleton ou indicador consistente com o restante do sistema
- alerta superior quando houver muitos `Pedidos sem data`

### Endpoint Sugerido

- `GET /budgets/delivery-monitor`

### Resposta Sugerida

Campos minimos por item:

- `budget_id`
- `budget_number`
- `project_id`
- `project_name`
- `construction_company`
- `salesperson_id`
- `salesperson_name`
- `status_id`
- `status_name`
- `delivery_date`
- `days_until_delivery`
- `delivery_status`
- `updated_at`

### Filtros de API Sugeridos

- `delivery_date_from`
- `delivery_date_to`
- `salesperson_id`
- `project_name`
- `budget_number`
- `status_id`
- `delivery_status`
- `missing_delivery_date`

### Primeira Versao Recomendada

Para a primeira entrega da tela, o recorte mais seguro e:

1. cards resumo
2. filtros principais
3. grid com destaque visual por situacao
4. acao de abrir o orcamento

Integracao direta com `Comunicacao` pode entrar logo depois, na segunda iteracao.

## Comportamento Visual Recomendado

Para leitura rapida:

- atrasado em vermelho
- entrega hoje em laranja
- entrega em 1 ou 2 dias em amarelo/azul de destaque
- futuras em neutro

## Sugestao de Cards Resumo no Topo

A tela pode ter indicadores como:

- total com data de entrega
- atrasados
- entrega hoje
- entrega em ate 2 dias
- pedidos sem data de entrega

Isso transforma a tela em painel operacional.

## Como A Consulta Deve Exibir Pedidos Sem Data

Para a tela ficar realmente util, a recomendacao e nao mostrar apenas entregas preenchidas.

O ideal e trabalhar com dois grupos operacionais dentro da mesma consulta:

- pedidos com `data de entrega` preenchida
- pedidos com status `Pedido` e `data de entrega` vazia

### Forma Sugerida

A tela pode ter:

- filtro rapido `Somente pedidos sem data de entrega`
- card resumo com contador dessas pendencias
- ordenacao priorizando:
  - atrasados
  - entrega hoje
  - entrega em 2 dias
  - pedido sem data de entrega

### Resultado Esperado

Assim o painel deixa de ser apenas uma agenda de entregas e passa a ser tambem uma ferramenta de saneamento operacional.

## Regras de Permissao

### Vendedor

- visualiza prioritariamente os proprios orcamentos com data de entrega
- recebe os avisos das entregas dos orcamentos pelos quais e responsavel

### Admin

- visualiza todos
- pode consultar por vendedor
- pode acompanhar se os avisos foram enviados

## Comportamentos Importantes

### Quando o Status Deixar de Ser `Pedido`

Sugestao:

- manter historicamente a `data de entrega`
- mas o monitoramento automatico deixa de considerar esse registro

### Quando a Data de Entrega For Alterada

Sugestao:

- limpar o controle de aviso anterior
- permitir novo envio automatico para a nova data

### Quando Nao Houver Vendedor no Orcamento

Sugestao:

- nao gerar mensagem
- registrar esse caso para auditoria
- opcionalmente exibir em consulta administrativa como inconsistencia

## Arquitetura Sugerida

### Backend

Implementacoes futuras sugeridas:

- migration para novas colunas em `budgets`
- ajuste do DTO de detalhe e listagem
- ajuste de create/update de orcamento
- regra de validacao do status `Pedido`
- job agendado de monitoramento
- integracao com modulo de comunicacao
- endpoint da tela de acompanhamento

### Frontend

Implementacoes futuras sugeridas:

- novo campo `Data de entrega` no formulario de orcamento
- exibicao da data na listagem de orcamentos, quando fizer sentido
- nova tela `Acompanhamento de Entregas`
- filtros especializados
- indicacao visual de proximidade e atraso
- acao para abrir a conversa relacionada

## Estrategia de Implementacao

### Fase 1. Base de Dados e Formulario

- adicionar campo `delivery_date`
- exibir no formulario de orcamento
- validar com status `Pedido`

### Fase 2. Tela de Consulta

- criar a tela de acompanhamento
- criar filtros e grid operacional
- disponibilizar leitura por vendedor e admin

### Fase 3. Aviso Automatico

- criar job diario
- gerar mensagem automatica
- registrar controle de envio

### Fase 4. Refinamentos

- indicadores executivos
- marcacao de leitura
- filtros rapidos por situacao
- historico de avisos

## Especificacao Funcional Inicial

### RF01. Cadastro do Campo

- o sistema deve permitir informar `data de entrega` no cadastro e na edicao do orcamento
- o campo deve ficar visivel independentemente do status, mas com maior destaque quando o status for `Pedido`

### RF02. Regra de Status `Pedido`

- quando o status do orcamento estiver como `Pedido`, a `data de entrega` continua opcional
- o sistema nao deve bloquear o salvamento quando a data nao for preenchida
- o sistema deve orientar o usuario com aviso visual de preenchimento recomendado

### RF03. Consulta Operacional

- o sistema deve disponibilizar uma tela especifica para acompanhamento de entregas
- a consulta deve listar tanto pedidos com `data de entrega` preenchida quanto pedidos sem data de entrega
- a tela deve permitir filtro por vendedor, obra, empresa, numero do orcamento, faixa de data e situacao da entrega

### RF04. Classificacao Operacional

- o sistema deve classificar cada registro em uma `situacao da entrega`
- as situacoes minimas sugeridas sao:
  - `Atrasado`
  - `Entrega hoje`
  - `Entrega em 1 dia`
  - `Entrega em 2 dias`
  - `Entrega futura`
  - `Pedido sem data de entrega`

### RF05. Aviso Automatico

- o sistema deve identificar diariamente os pedidos cuja entrega ocorre em `2 dias`
- para esses casos, deve gerar mensagem automatica para o vendedor responsavel
- o sistema deve evitar aviso duplicado para a mesma combinacao de orcamento e data de entrega

### RF06. Excecoes Operacionais

- se o orcamento estiver em `Pedido` e nao tiver vendedor vinculado, o sistema nao deve tentar enviar a mensagem
- esses casos devem ficar rastreaveis para consulta administrativa ou log operacional

### RF07. Alteracao da Data

- se a `data de entrega` for alterada apos um aviso ja enviado, o sistema deve permitir novo aviso para a nova data
- o controle de duplicidade deve considerar a data de entrega vigente

### RF08. Saida do Status `Pedido`

- se o orcamento deixar de estar em `Pedido`, o monitoramento automatico nao deve mais considerar esse registro
- a data pode ser mantida historicamente no cadastro, sem acionar novos avisos

## Criterios de Aceite Iniciais

### CA01. Formulario

- dado um orcamento em edicao, quando o usuario informar a `data de entrega`, entao o sistema deve salvar a informacao corretamente
- dado um orcamento com status `Pedido` e sem `data de entrega`, quando o usuario salvar, entao o sistema deve permitir a persistencia e exibir alerta visual de recomendacao

### CA02. Monitoramento

- dado um pedido com `data de entrega` para daqui a `2 dias`, quando o job diario executar, entao o sistema deve gerar exatamente um aviso para o vendedor
- dado um pedido ja notificado para a mesma data, quando o job executar novamente, entao o sistema nao deve duplicar o aviso

### CA03. Pendencias

- dado um pedido sem `data de entrega`, quando a tela de acompanhamento for consultada, entao o registro deve aparecer como `Pedido sem data de entrega`
- dado um pedido sem vendedor vinculado, quando atingir a regra de aviso, entao o sistema nao deve quebrar a execucao do job

### CA04. Consulta

- dado um conjunto de pedidos com cenarios mistos, quando o usuario aplicar filtros na tela, entao a consulta deve separar corretamente atrasados, proximos, futuros e sem data

## Backlog Tecnico Inicial

### Banco de Dados

- criar migration para adicionar `delivery_date`
- criar migration para adicionar `delivery_notification_sent_at`
- criar migration para adicionar `delivery_notification_reference_date`
- avaliar indice para consulta por `status` e `delivery_date`, caso o volume cresca

### Backend

- ajustar model de orcamento para incluir os novos campos
- ajustar DTOs de create, update, detail e listagem
- ajustar validacoes do servico de orcamento para aceitar a data opcional em `Pedido`
- preparar regra derivada de `situacao da entrega`
- criar consulta especifica para a tela de acompanhamento
- criar job agendado para localizar entregas em `2 dias`
- integrar o job com o modulo de conversa/comunicacao para gerar mensagem automatica
- registrar logs operacionais para casos sem vendedor ou com erro de envio

### Frontend

- incluir campo `Data de entrega` no formulario de orcamento
- exibir `Alert` contextual quando `status = Pedido` e a data estiver vazia
- decidir se a listagem principal de orcamentos mostrara a coluna `Data de entrega`
- criar a tela `Acompanhamento de Entregas`
- implementar filtros, chips e destaque visual por situacao
- adicionar atalho para abrir o orcamento e a conversa relacionada

### Testes

- criar testes de backend para persistencia e leitura da `data de entrega`
- criar testes da regra de classificacao da `situacao da entrega`
- criar testes do job para evitar duplicidade de aviso
- criar testes para pedidos sem data e pedidos sem vendedor
- criar testes de frontend para renderizacao do alerta visual e comportamento do formulario

## Mapa Inicial de Impacto no Projeto

### Backend Ja Identificado

- `backend/internal/model/budget_model.go`
- `backend/internal/dto/budget_dto.go`
- `backend/internal/repository/budget/repository.go`
- `backend/internal/service/budget/service.go`
- `backend/internal/handler/budget/handler.go`
- `backend/internal/service/conversation/service.go`
- `backend/internal/repository/conversation/repository.go`
- `backend/internal/server/router.go`

### Frontend Ja Identificado

- `frontend/src/features/budgets/components/BudgetForm.tsx`
- `frontend/src/features/budgets/pages/BudgetCreatePage.tsx`
- `frontend/src/features/budgets/pages/BudgetEditPage.tsx`
- `frontend/src/features/budgets/pages/BudgetListPage.tsx`
- `frontend/src/features/communication/pages/CommunicationPage.tsx`

### Novos Artefatos Provaveis

- arquivo de migration para a tabela `budgets`
- endpoint dedicado para a consulta de acompanhamento de entregas
- servico ou job especifico para monitoramento de entregas
- tipos e funcoes de API no frontend para a nova tela operacional

## Ordem Recomendada de Execucao

1. criar a migration e propagar os campos no backend de orcamentos
2. expor `delivery_date` nos endpoints existentes de cadastro, edicao e detalhe
3. incluir o campo no formulario de orcamento com alerta visual para `Pedido`
4. implementar a regra derivada de `situacao da entrega`
5. criar o endpoint da tela `Acompanhamento de Entregas`
6. construir a tela de consulta com filtros e destaques visuais
7. implementar o job diario de aviso automatico
8. integrar o job com a `Comunicacao` e registrar o controle de duplicidade

## Corte Recomendado para Primeira Entrega

Se for necessario reduzir escopo inicial, o menor recorte com valor operacional e:

1. salvar `data de entrega` no orcamento
2. exibir alerta visual para `Pedido` sem data
3. criar consulta com filtros e coluna `situacao da entrega`

Com isso, o time ganha visibilidade operacional antes mesmo da automacao de aviso.

## Melhor Recomendacao Final

A melhor solucao para esse caso e:

1. adicionar `data de entrega` na tabela `budgets`
2. manter o campo opcional quando o status estiver em `Pedido`, sem bloqueio de salvamento
3. criar uma tela propria de `Acompanhamento de Entregas`
4. usar a `Comunicacao` como canal interno de aviso automatico ao vendedor
5. executar um `job` diario no backend para identificar entregas em `2 dias`

Essa abordagem mantem a regra clara, preserva flexibilidade para o vendedor, aproveita o que o sistema ja possui e cria um fluxo operacional consistente sem depender de processos manuais.

## Proximo Passo Recomendado

Se esta direcao for aprovada, a proxima etapa mais segura e elaborar:

1. especificacao funcional do campo `data de entrega`
2. backlog tecnico da migration, backend e frontend
3. especificacao da tela `Acompanhamento de Entregas`
4. desenho do evento automatico na `Comunicacao`
