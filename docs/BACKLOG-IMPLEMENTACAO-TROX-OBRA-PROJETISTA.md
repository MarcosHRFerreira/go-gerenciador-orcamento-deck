# Backlog De Implementacao Trox Obra Projetista

## Objetivo

Converter o estudo `ESTUDO-ALTERACOES-TROX-OBRA-PROJETISTA.md` em um backlog executavel, com ordem de implementacao, dependencias, criterios de aceite e definicao do proximo passo pratico.

## Ordem Recomendada

1. Banco e contratos base
2. Importacao Trox
3. Permissoes de orcamento para `user`
4. Filtros e paginacao
5. Cadastro de `Obra`
6. Renomeacao funcional para `Obra`
7. Renomeacao funcional para `Projetista`
8. Refatoracao estrutural final

## Fase 1. Banco E Contratos Base

### Task 1. Criar tabela de linha de produtos

Descricao:

- criar tabela `product_lines`
- incluir `id`, `code`, `name`, `description`, `created_at`, `updated_at`
- garantir unicidade adequada para `code` e `name`

Dependencias:

- nenhuma

Criterios de aceite:

- migration criada
- tabela criada com rollback
- repositorio e model preparados

### Task 2. Adicionar referencia de linha de produto em budgets

Descricao:

- adicionar `product_line_id` em `budgets`
- criar `foreign key` para `product_lines`
- revisar selects, inserts e updates

Dependencias:

- Task 1

Criterios de aceite:

- migration criada
- `BudgetModel` atualizado
- repositorio de `budgets` persistindo o novo campo

### Task 3. Adicionar campo construtora em budgets

Descricao:

- adicionar `construction_company` em `budgets`
- ajustar model, DTOs, repository, service e responses

Dependencias:

- nenhuma

Criterios de aceite:

- migration criada
- create, update, get e list retornam o campo
- testes atualizados

### Task 4. Garantir status Em Negociacao

Descricao:

- criar migration idempotente para inserir `Em Negociacao` em `budget_statuses`
- definir `sort_order` e demais atributos aprovados

Dependencias:

- nenhuma

Criterios de aceite:

- migration executa sem duplicar registro
- status disponivel para uso na importacao

### Task 5. Preparar compatibilidade para projetista

Descricao:

- definir estrategia de transicao entre `designer` e `projetista`
- revisar nomes de DTO, filtro e resposta

Dependencias:

- nenhuma

Criterios de aceite:

- estrategia documentada
- contratos de API definidos para a fase de implementacao

### Task 6. Adicionar codigo na entidade atual de projeto

Descricao:

- adicionar `code` na tabela atual `projects`
- ajustar DTOs, models, repository e service

Dependencias:

- nenhuma

Criterios de aceite:

- migration criada
- CRUD backend aceita e devolve `code`

## Fase 2. Importacao Trox

### Task 7. Mapear linha de produtos na importacao

Descricao:

- alterar parser Trox para capturar `Linha de produtos`
- localizar ou criar catalogo em `product_lines`
- gravar referencia em `budgets.product_line_id`

Dependencias:

- Task 1
- Task 2

Criterios de aceite:

- importacao cria ou reutiliza a linha de produto
- budget importado fica associado corretamente
- testes de unidade cobrindo fluxo principal e alternativo

### Task 8. Mapear construtora na importacao

Descricao:

- usar `Nome Cliente` para preencher `construction_company`
- normalizar espacos e texto

Dependencias:

- Task 3

Criterios de aceite:

- budget importado recebe `construction_company`
- testes cobrindo o mapeamento

### Task 9. Alterar status inicial da importacao

Descricao:

- trocar status principal de `Nao informado` para `Em Negociacao`
- decidir se o `Status` da planilha segue em `current_follow_up`

Dependencias:

- Task 4

Criterios de aceite:

- orcamentos importados entram com status `Em Negociacao`
- testes atualizados

### Task 10. Revisar testes de importacao

Descricao:

- atualizar testes unitarios de `budgetimport`
- atualizar testes de integracao que dependem do payload importado

Dependencias:

- Tasks 7, 8 e 9

Criterios de aceite:

- suite de testes da importacao verde

## Fase 3. Permissoes De Orcamento

### Task 11. Liberar edicao de orcamento para user

Descricao:

- permitir `PUT /budgets/:id` para `user`
- respeitar escopo por vendedor

Dependencias:

- validacao de regra de escopo atual

Criterios de aceite:

- `user` consegue editar apenas orcamentos do proprio escopo
- `admin` continua podendo editar qualquer orcamento

### Task 12. Manter exclusao apenas para admin

Descricao:

- preservar `DELETE /budgets/:id` como exclusivo de `admin`
- revisar middlewares e agrupamento de rotas

Dependencias:

- nenhuma

Criterios de aceite:

- `user` recebe `403` ao tentar excluir
- testes existentes ou novos cobrindo o bloqueio

### Task 13. Ajustar frontend da listagem de orcamentos

Descricao:

- mover rota de edicao para area autenticada comum
- exibir `Editar` para `user`
- ocultar ou desabilitar `Excluir` para `user`

Dependencias:

- Tasks 11 e 12

Criterios de aceite:

- `user` acessa a tela de edicao
- `user` nao visualiza acao de exclusao
- `admin` mantem as duas acoes

### Task 14. Criar testes de permissao

Descricao:

- teste de integracao para `user` editar o proprio orcamento
- teste de integracao para `user` nao editar fora do escopo
- teste de integracao para `user` nao excluir

Dependencias:

- Tasks 11, 12 e 13

Criterios de aceite:

- cenarios principais e alternativos cobertos

## Fase 4. Filtros E Paginacao

### Task 15. Expor filtro por periodo no frontend

Descricao:

- adicionar `sent_at_from` e `sent_at_to` na tela de orcamentos
- persistir os filtros em query string

Dependencias:

- nenhuma, pois backend ja suporta

Criterios de aceite:

- usuario consegue filtrar por periodo
- recarregamento da pagina preserva o filtro

### Task 16. Enviar filtros de periodo para a API

Descricao:

- ajustar types e `buildBudgetListParams`
- revisar leitura de query params

Dependencias:

- Task 15

Criterios de aceite:

- requisicao enviada com `sent_at_from` e `sent_at_to`
- listagem retorna resultados coerentes

### Task 17. Revisar paginacao padrao

Descricao:

- alterar pagina padrao de `20` para valor aprovado
- incluir seletor de quantidade por pagina

Dependencias:

- definicao do novo valor padrao

Criterios de aceite:

- listagem usa novo padrao
- usuario pode mudar a quantidade por pagina

### Task 18. Revisar limite maximo do backend

Descricao:

- aumentar limite maximo de `100` para valor aprovado, se necessario
- revisar testes da validacao de `page_size`

Dependencias:

- definicao do novo maximo

Criterios de aceite:

- validacao atualizada
- testes refletindo o novo limite

## Fase 5. Cadastro De Obra

### Task 19. Evoluir backend da entidade atual de projeto para suportar obra

Descricao:

- reaproveitar CRUD atual de `projects`
- incluir `code`
- preparar contratos para apresentacao como `Obra`

Dependencias:

- Task 6

Criterios de aceite:

- backend pronto para cadastro/manutencao com codigo e descricao

### Task 20. Criar tela de listagem de obra

Descricao:

- criar pagina com listagem
- adicionar busca por codigo e descricao

Dependencias:

- Task 19

Criterios de aceite:

- admin consegue consultar obras cadastradas

### Task 21. Criar tela de cadastro e edicao de obra

Descricao:

- formulario de criacao
- formulario de edicao
- validacoes basicas

Dependencias:

- Task 19

Criterios de aceite:

- admin consegue criar e editar obras

### Task 22. Incluir menu e navegacao de obra

Descricao:

- adicionar item de menu
- ajustar rotas e links da aplicacao

Dependencias:

- Tasks 20 e 21

Criterios de aceite:

- fluxo completo acessivel a partir do menu

## Fase 6. Renomeacao Funcional Para Obra

### Task 23. Trocar labels de Projeto para Obra

Descricao:

- atualizar telas, breadcrumbs, titulos, textos e mensagens

Dependencias:

- nenhuma

Criterios de aceite:

- frontend apresenta `Obra` ao usuario final

### Task 24. Revisar filtros e agrupamentos por obra

Descricao:

- revisar a listagem agrupada hoje tratada como projeto
- ajustar textos e descricoes para `Obra`

Dependencias:

- Task 23

Criterios de aceite:

- agrupamentos e detalhes com nomenclatura coerente

## Fase 7. Renomeacao Funcional Para Projetista

### Task 25. Trocar labels de Designer para Projetista

Descricao:

- atualizar formularios, colunas, filtros e detalhes

Dependencias:

- definicao da estrategia de compatibilidade da Task 5

Criterios de aceite:

- frontend usa `Projetista`

### Task 26. Ajustar API para compatibilidade de projetista

Descricao:

- aceitar e/ou expor nomenclatura de `projetista`
- minimizar risco de quebra durante a transicao

Dependencias:

- Task 5

Criterios de aceite:

- backend e frontend funcionando durante a transicao

## Fase 8. Refatoracao Estrutural Final

### Task 27. Renomear estrutura tecnica de projeto para obra

Descricao:

- avaliar rename fisico de tabela, rotas, DTOs e services

Dependencias:

- fases anteriores concluidas

Criterios de aceite:

- codigo alinhado ao dominio final
- testes atualizados

### Task 28. Renomear coluna tecnica de designer

Descricao:

- renomear fisicamente `designer_name`
- ajustar todos os acessos e filtros

Dependencias:

- Task 26

Criterios de aceite:

- coluna renomeada com migration segura
- suite verde

## Definicoes Pendentes

- confirmar se `Nome Cliente` representa `Construtora`
- confirmar valor padrao da pagina: `50` ou `100`
- confirmar valor maximo de `page_size`: `100` ou `200`
- confirmar nome fisico final da coluna de projetista
- confirmar se o status da planilha continua sendo usado como `current_follow_up`
- confirmar regra de geracao do codigo da obra
- confirmar regra de geracao do codigo da linha de produto

## Proximo Passo Pratico

Executar primeiro a `Fase 1. Banco E Contratos Base`, nesta ordem:

1. criar `product_lines`
2. adicionar `product_line_id` em `budgets`
3. adicionar `construction_company` em `budgets`
4. garantir `Em Negociacao`
5. adicionar `code` em `projects`
6. definir estrategia de compatibilidade para `projetista`

## Entrega Sugerida Em Desenvolvimento

Se a implementacao for iniciada agora, a menor entrega de risco e maior valor e:

- migrations de banco
- ajuste da importacao Trox
- liberacao de edicao para `user` sem exclusao

Isso entrega rapidamente:

- novas informacoes persistidas
- regra comercial da importacao
- melhoria operacional para usuarios comuns
