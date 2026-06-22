# Debug Session: login-internal-error

Status: [OPEN]

Symptoms:
- Usuario informa "Erro interno do servidor" ao passar da tela de login.
- Fluxo esperado: autenticar e carregar a tela inicial autenticada sem erro 500.

Hypotheses:
1. A primeira requisicao apos o login para `GET /budgets` ainda falha no backend por incompatibilidade entre schema do banco local e o codigo em execucao.
2. Ha outra migration recente nao aplicada e a query da listagem autenticada esta acessando coluna, indice, tabela ou constraint inexistente.
3. O backend esta retornando 500 em algum endpoint carregado pelo shell autenticado, como `GET /users/me`, `GET /budget-statuses` ou `GET /budgets`.
4. O banco foi parcialmente corrigido, mas a aplicacao backend em execucao usa outra base ou outro container diferente da base que foi validada manualmente.
5. Existe um erro de runtime no frontend disparado por resposta inesperada do backend apos o login, e a mensagem exibida ao usuario mascara o endpoint real que falhou.

Evidence Plan:
- Identificar exatamente qual endpoint retorna 500 apos o login.
- Coletar logs do backend no momento da reproducao.
- Instrumentar pontos minimos de entrada das rotas mais provaveis para correlacionar requisicao, erro e dependencias de banco.
- Confirmar ou refutar as hipoteses antes de aplicar correcao.

Progress Log:
- Sessao criada.
