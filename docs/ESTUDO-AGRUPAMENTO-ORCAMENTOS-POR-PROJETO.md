# Estudo De Agrupamento De Orcamentos Por Projeto

## Objetivo

Este documento registra uma proposta funcional e tecnica para suportar o seguinte fluxo:

- um unico projeto pode possuir diversos orcamentos
- deve existir uma tela para associar orcamentos a um projeto
- todos os orcamentos do grupo podem nascer com status `ORCAMENTO`
- quando um dos orcamentos do grupo mudar para `PEDIDO`, os demais do mesmo grupo devem passar automaticamente para `CANCELADO`
- os orcamentos cancelados por essa regra deixam de exigir atencao operacional

O foco aqui e definir:

- a regra de negocio
- o impacto na modelagem atual
- o comportamento de tela
- a estrategia mais segura para implementacao

## Leitura Do Cenario Desejado

O fluxo descrito representa um agrupamento comercial de propostas concorrentes para uma mesma oportunidade.

Interpretacao pratica:

- existe um projeto ou obra
- para esse mesmo projeto podem existir varios orcamentos
- em algum momento um desses orcamentos vira o vencedor do grupo
- ao virar `PEDIDO`, ele define os demais como alternativas descartadas

Na pratica, isso se comporta como:

- um grupo comercial de orcamentos
- com um vencedor
- e com encerramento automatico dos demais

## Estado Atual Do Sistema

Hoje o sistema ja possui:

- cadastro de `projects`
- `project_id` opcional em `budgets`
- relacao tecnica de muitos orcamentos para um projeto, porque nao existe `unique` em `budgets.project_id`

Ou seja:

- tecnicamente, o banco atual ja permite varios orcamentos apontando para o mesmo `project_id`
- porem, isso ainda nao resolve o problema funcional completo

Faltam hoje:

- regra de grupo
- regra de vencedor
- cancelamento automatico dos demais
- tela dedicada para visualizar e gerir os orcamentos vinculados ao mesmo projeto
- indicacao operacional de que os perdedores do grupo nao precisam mais de atencao

## Ponto Importante Sobre O `project_id` Atual

O cadastro atual de `projects` parece representar a obra ou projeto real, com campos como:

- `name`
- `city`
- `state`
- `project_type_id`
- `notes`

Isso sugere que `projects` ja esta semanticamente alinhado com a ideia de obra.

### Conclusao Sobre Reuso

Para este caso, a recomendacao inicial e:

- reutilizar o `project_id` atual como agrupador de orcamentos
- nao criar uma nova entidade de agrupamento nesta primeira versao

Motivos:

- a relacao muitos-para-um ja existe no banco
- a linguagem do usuario e do negocio usa exatamente `projeto`
- a tela e a regra podem ser construidas sobre a estrutura atual com menor custo

### Quando Criar Outra Entidade

So faria sentido criar uma nova entidade, como `budget_groups` ou `commercial_opportunities`, se o projeto perceber que:

- um mesmo orcamento pode pertencer a mais de um agrupamento
- o cadastro `projects` nao representa exatamente a oportunidade comercial
- ha necessidade futura de historico do grupo, vencedor, motivo de encerramento e metadados de pipeline em nivel superior ao orcamento

Para a primeira fase, isso parece desnecessario.

## Regra De Negocio Proposta

### Regra Principal

Um projeto pode ter varios orcamentos associados.

### Regra De Status Do Grupo

Todos os orcamentos vinculados ao mesmo projeto podem existir inicialmente com status `ORCAMENTO`.

Quando um orcamento vinculado ao projeto mudar para `PEDIDO`:

- ele permanece com status `PEDIDO`
- todos os demais orcamentos do mesmo projeto, desde que ainda estejam em status operacionais concorrentes, passam automaticamente para `CANCELADO`

### Recomendacao De Escopo Da Regra

Para evitar efeitos colaterais perigosos, recomendo aplicar o cancelamento automatico apenas sobre orcamentos do mesmo projeto que estejam em status concorrentes ativos, por exemplo:

- `ORCAMENTO`
- `NEGOCIANDO`
- `AGUARDANDO`
- outros status equivalentes definidos pelo negocio

E nao alterar automaticamente itens que ja estejam em estados finais, por exemplo:

- `CANCELADO`
- `PERDIDO`
- `PEDIDO`

### Regra De Unicidade Do Vencedor

Recomendo assumir:

- um projeto nao pode ter mais de um orcamento com status `PEDIDO`

Se existir outro `PEDIDO` no mesmo projeto, a operacao deve ser bloqueada com erro de negocio.

### Regra De Projeto Nulo

Se o orcamento nao estiver associado a projeto:

- a regra de cancelamento automatico nao se aplica

### Regra De Reabertura

Se um projeto ja teve um `PEDIDO` e o usuario tentar reativar outro orcamento do grupo:

- isso deve ser bloqueado ou exigir acao administrativa explicita

Para a primeira versao, recomendo bloquear.

## Comportamento Operacional Esperado

Fluxo sugerido:

1. usuario cria ou edita um orcamento
2. informa o `projeto`
3. o sistema mostra que aquele projeto possui outros orcamentos vinculados
4. todos seguem com status normal enquanto nenhum venceu
5. ao alterar um deles para `PEDIDO`, o sistema:
   - salva o vencedor
   - atualiza os demais para `CANCELADO`
   - registra historico dessas alteracoes
   - remove esses itens da fila de acompanhamento ativo

## Recomendacao De Tela

## Tela Principal Recomendada

A melhor solucao inicial e criar uma tela de detalhe do projeto com foco comercial.

Sugestao de rota:

- `/projects/:projectId/budgets`

Ou incorporar uma aba dentro do detalhe do projeto:

- `Dados do projeto`
- `Orcamentos vinculados`

### Conteudo Da Tela

A tela deveria exibir:

- cabecalho do projeto
- resumo do grupo
- tabela com todos os orcamentos vinculados
- destaque visual do vencedor quando existir
- acao para vincular orcamentos existentes
- acao para abrir criacao de novo orcamento ja pre-preenchido com o projeto

### Resumo No Topo

Indicadores recomendados:

- total de orcamentos vinculados
- quantidade em aberto
- quantidade cancelada
- orcamento vencedor atual
- data da ultima atualizacao

### Tabela De Orcamentos Do Projeto

Colunas recomendadas:

- ID
- numero do orcamento
- ano
- revisao
- data de envio
- vendedor
- contato
- valor bruto
- status
- ultima atualizacao
- acoes

### Acoes Da Tela

Sugestao de botoes:

- `Vincular orcamentos`
- `Novo orcamento para este projeto`
- `Abrir orcamento`
- `Definir como pedido`

### Experiencia Ao Marcar Como `PEDIDO`

Ao clicar para mudar um orcamento do grupo para `PEDIDO`, a interface deve pedir confirmacao:

- informar que os demais orcamentos do projeto serao cancelados automaticamente
- indicar quantos itens serao impactados

Texto sugerido:

- `Ao definir este orcamento como PEDIDO, os demais orcamentos ativos deste projeto serao marcados como CANCELADO. Deseja continuar?`

## Alternativa De Tela Mais Simples

Se a prioridade for velocidade de entrega, existe uma versao mais enxuta:

- manter a lista de orcamentos atual
- adicionar filtro por `projeto`
- adicionar uma coluna `Projeto`
- adicionar um drawer ou modal com os orcamentos do mesmo projeto

Porem, isso resolve menos bem a gestao do grupo.

### Recomendacao

Entre as duas opcoes, recomendo:

- detalhe do projeto com tabela de orcamentos vinculados

Porque deixa a regra mais clara e reduz risco operacional.

## Associacao De Orcamentos Ao Projeto

### Formas Possiveis

Existem tres formas uteis de associacao:

1. associar no create/edit do orcamento
2. associar a partir da tela do projeto
3. criar novo orcamento ja dentro do contexto do projeto

### Recomendacao

Implementar as tres, mas por etapas:

- fase 1: associacao no formulario do orcamento e visualizacao no detalhe do projeto
- fase 2: vincular orcamentos existentes em lote a partir da tela do projeto

## Impacto Na Modelagem

## Minimo Necessario

Se a primeira versao for enxuta, a modelagem pode continuar com a atual relacao:

- `budgets.project_id -> projects.id`

Nesse caso, o essencial e implementar:

- consulta de todos os orcamentos por projeto
- regra de atualizacao em lote de status
- validacao para impedir dois `PEDIDO` no mesmo projeto

## Campos Que Valeria Adicionar

Mesmo sem criar nova tabela, alguns campos extras podem ajudar no futuro:

- em `budgets`:
  - `cancellation_reason`
  - `canceled_by_group_rule` boolean

Esses campos nao sao obrigatorios na primeira entrega, mas melhoram auditoria e UX.

### Recomendacao

Se houver tempo tecnico, recomendo adicionar ao menos:

- `canceled_by_group_rule boolean default false`

Isso permite diferenciar:

- cancelamento manual
- cancelamento automatico por outro orcamento do projeto ter virado `PEDIDO`

## Impacto No Backend

## Novas Regras No Service De Orcamentos

Ao atualizar o status de um orcamento para `PEDIDO`, o backend deve:

1. carregar o orcamento atual
2. validar se ele possui `project_id`
3. verificar se ja existe outro `PEDIDO` no mesmo projeto
4. iniciar transacao
5. atualizar o orcamento atual para `PEDIDO`
6. cancelar os demais orcamentos ativos do mesmo projeto
7. registrar historico das alteracoes
8. concluir a transacao

## Recomendacao De Implementacao

Essa regra nao deve ficar espalhada no handler.

Ela deve ficar concentrada no `service` de `budgets`, porque ali esta a orquestracao correta da regra de negocio.

## Novas Consultas Necessarias

Provavelmente sera necessario adicionar ao repository:

- listar orcamentos por `project_id`
- localizar orcamento `PEDIDO` do projeto
- atualizar status em lote dos demais orcamentos do grupo

## Impacto Em Historico

Como o sistema ja possui `budget_status_history`, a recomendacao e registrar:

- mudanca do orcamento vencedor para `PEDIDO`
- mudanca dos demais para `CANCELADO`
- preferencialmente com observacao indicando que o cancelamento foi automatico por decisao do grupo

## Impacto No Frontend

## Lista De Orcamentos

A lista atual pode ser enriquecida com:

- coluna `Projeto`
- indicador de grupo
- atalho para ver os demais orcamentos do mesmo projeto

## Formulario Do Orcamento

O formulario deve:

- permitir selecionar projeto
- mostrar aviso quando o projeto ja possuir outros orcamentos

Mensagem sugerida:

- `Este projeto ja possui outros orcamentos vinculados. Ao definir este item como PEDIDO, os demais podem ser cancelados automaticamente.`

## Tela Do Projeto

No frontend, essa nova tela pode ser desenhada como:

- cabecalho com dados do projeto
- cards-resumo
- tabela de orcamentos
- destaque visual do vencedor
- badge para cancelados automaticamente

## Impacto Em Permissoes

Esse fluxo mexe com decisao comercial e fechamento do grupo.

Por isso, recomendo que a acao de definir `PEDIDO` em um projeto:

- respeite as mesmas permissoes de alteracao de orcamento ja existentes
- e, se necessario, seja restrita a `admin` ou ao dono comercial do item

## Riscos E Cuidados

### 1. Sobrescrever Status Indevidamente

Risco:

- cancelar automaticamente item que ja estava finalizado por outro motivo

Mitigacao:

- aplicar a regra apenas sobre statuses ativos e concorrentes

### 2. Dois Pedidos No Mesmo Projeto

Risco:

- o grupo ficar incoerente

Mitigacao:

- bloquear no backend
- validar em transacao

### 3. Uso Ambiguo Do Cadastro De Projeto

Risco:

- o cadastro `projects` ser usado como obra fisica, enquanto o negocio pensa em projeto como oportunidade comercial

Mitigacao:

- alinhar com usuarios-chave antes da implementacao
- se houver ambiguidade forte, criar entidade separada em fase futura

### 4. Falta De Auditoria

Risco:

- usuario nao entender porque outro orcamento virou `CANCELADO`

Mitigacao:

- registrar historico
- exibir badge ou observacao `Cancelado automaticamente por pedido do projeto`

## Recomendacao De UX

Quando houver um vencedor no grupo, a tela deve deixar isso muito evidente.

Sugestoes:

- card `Orcamento vencedor`
- destaque visual em verde para `PEDIDO`
- etiqueta nos outros: `Cancelado automaticamente`
- esconder esses itens das filas principais por padrao, mas permitir consulta

## Estrategia De Entrega

## Fase 1 - Regra E Visualizacao

- manter `project_id` como agrupador
- criar consulta de orcamentos do projeto
- criar tela de detalhe do projeto com lista de orcamentos
- implementar regra de `PEDIDO` cancelando os demais
- registrar historico das alteracoes

## Fase 2 - Melhorias Operacionais

- vincular orcamentos existentes em lote
- destacar cancelamento automatico
- adicionar filtros e indicadores de grupo na listagem geral

## Fase 3 - Evolucao Se Necessario

- avaliar criacao de entidade especifica de agrupamento comercial
- adicionar metadados do grupo
- adicionar politicas mais refinadas de reabertura e troca de vencedor

## Recomendacao Final

Para a primeira versao, a melhor estrategia e:

1. reutilizar o `project_id` atual como agrupador
2. criar uma tela de detalhe do projeto com a lista de orcamentos vinculados
3. implementar a regra de que, ao marcar um orcamento do projeto como `PEDIDO`, os demais orcamentos ativos do mesmo projeto passam automaticamente para `CANCELADO`
4. registrar historico dessas alteracoes para auditoria
5. destacar visualmente que os demais itens nao precisam mais de atencao

Essa abordagem entrega o comportamento desejado com baixo retrabalho e aproveita bem a modelagem que o sistema ja possui hoje.

## Checklist De Implementacao

- validar com negocio se `projects` representa corretamente a obra/oportunidade
- definir quais statuses entram como concorrentes ativos
- bloquear mais de um `PEDIDO` por projeto
- criar operacao transacional no backend
- atualizar historico de status
- criar tela de detalhe do projeto com orcamentos vinculados
- permitir associacao de orcamentos ao projeto
- destacar cancelamento automatico no frontend
