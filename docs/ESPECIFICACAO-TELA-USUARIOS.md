# Especificacao Funcional: Area Administrativa de Usuarios

## Objetivo

Definir como deve funcionar a area administrativa de `Usuarios` do sistema, considerando:

- o backend atual
- as limitacoes existentes
- a necessidade de criar novos usuarios
- a necessidade de controlar se um usuario e `admin` ou `user`
- a futura necessidade de ativar ou desativar acessos

Este documento complementa:

- `docs/ESTUDO-CADASTRO-USUARIOS-E-PERFIS.md`

## Resultado Esperado

Ao final da implementacao, um usuario administrador deve conseguir:

- listar usuarios
- cadastrar novo usuario
- definir se o novo usuario sera `admin` ou `user`
- visualizar status de acesso
- promover ou rebaixar um usuario existente
- ativar ou desativar um usuario
- opcionalmente redefinir senha

Usuarios comuns nao devem acessar essa area.

## Escopo Recomendado

## Fase 1

Entregar o minimo funcional com maior valor:

- listagem de usuarios
- criacao de usuario
- visualizacao do perfil atual

## Fase 2

Completar o controle operacional:

- alteracao de perfil `admin/user`
- ativacao e desativacao

## Fase 3

Melhorias administrativas:

- redefinicao de senha
- auditoria de alteracoes
- filtros mais completos

## Regras de Acesso

### Quem acessa

- apenas usuarios com `role = admin`

### Quem nao acessa

- usuarios com `role = user`

### Comportamento esperado no frontend

- esconder o item `Usuarios` do menu para quem nao for admin
- proteger a rota `/users` com guard de permissao
- se um usuario comum tentar acessar a rota diretamente, redirecionar ou mostrar `403`

## Estrutura de Rotas Recomendada

### Rotas protegidas para admin

- `/users`
- `/users/new`
- `/users/:userId/edit`

### Rotas opcionais futuras

- `/me`

Observacao:

- o fluxo de criacao pode ser pagina separada ou dialog
- para a primeira versao, pagina separada tende a ser mais simples

## Estrutura da Tela de Listagem

## Cabecalho

Deve conter:

- titulo `Usuarios`
- subtitulo curto explicando que a tela controla acesso e perfil
- botao `Novo usuario`

## Bloco de filtros

Filtros recomendados:

- nome
- email
- username
- perfil
- status

Para a primeira entrega, os filtros minimos recomendados sao:

- nome
- email
- perfil
- status

## Tabela principal

Colunas recomendadas:

- nome
- email
- username
- perfil
- status
- criado em
- atualizado em
- acoes

### Como exibir cada coluna

- `perfil`: usar `Chip` com `Administrador` ou `Usuario`
- `status`: usar `Chip` com `Ativo` ou `Inativo`
- datas: mostrar formato local amigavel

### Acoes por linha

Versao minima:

- `Editar`

Versao recomendada:

- `Editar`
- `Tornar administrador` ou `Tornar usuario`
- `Ativar` ou `Desativar`
- `Redefinir senha`

## Estado vazio

Quando nao houver usuarios alem do admin inicial, a tela deve exibir:

- mensagem clara
- incentivo para cadastrar o primeiro usuario adicional

Exemplo:

- `Nenhum usuario adicional encontrado. Cadastre um novo usuario para iniciar o controle de acessos.`

## Estrutura da Tela de Novo Usuario

## Campos recomendados

- nome
- email
- username
- senha
- confirmar senha
- perfil

## Tipo de campo

- `nome`: texto
- `email`: email
- `username`: texto curto
- `senha`: password
- `confirmar senha`: password
- `perfil`: select com `Administrador` e `Usuario`

## Valor padrao recomendado

- `perfil = user`

Motivo:

- reduz risco operacional
- privilegio administrativo deve ser concedido conscientemente

## Validacoes

- nome obrigatorio
- email obrigatorio e valido
- username obrigatorio
- senha obrigatoria com minimo de 6 caracteres
- confirmacao igual a senha
- perfil obrigatorio

## Comportamento ao salvar

1. enviar `POST /users`
2. exibir mensagem de sucesso
3. voltar para a listagem
4. atualizar a tabela automaticamente

## Payload atual suportado

```json
{
  "name": "Maria Souza",
  "email": "maria@empresa.com",
  "username": "maria.souza",
  "password": "123456",
  "password_confirm": "123456",
  "role": "user"
}
```

## Estrutura da Tela de Edicao

Esta tela ainda depende de novos endpoints no backend.

Campos recomendados:

- nome
- email
- username
- perfil
- status

Campos que podem ficar separados:

- senha

## Acoes recomendadas

- salvar dados cadastrais
- alterar perfil
- ativar ou desativar
- redefinir senha

## Regras Operacionais Importantes

### 1. Nao rebaixar o ultimo admin

O sistema deve impedir:

- trocar o ultimo `admin` para `user`

Motivo:

- evita travar a administracao do sistema

### 2. Nao desativar o unico admin ativo

O sistema deve impedir:

- desativar a ultima conta administrativa ativa

### 3. Confirmacao para acoes sensiveis

Deve haver confirmacao antes de:

- tornar usuario administrador
- remover privilegio administrativo
- desativar usuario

### 4. Autoedicao com cuidado

E recomendavel impedir que um admin:

- desative a si proprio
- remova o proprio perfil de admin sem existir outro admin ativo

## Contratos de API Ja Existentes

## Ja disponiveis

- `GET /users/me`
- `GET /users`
- `POST /users`

## Resposta atual de usuario

```json
{
  "id": 1,
  "name": "Admin Local",
  "email": "admin@local.dev",
  "username": "admin",
  "role": "admin",
  "active": true,
  "created_at": "2026-06-13T10:00:00Z",
  "updated_at": "2026-06-13T10:00:00Z"
}
```

## Lacunas de API

Para a area ficar completa, faltam endpoints para:

- editar usuario existente
- alterar `role`
- alterar `active`
- redefinir senha

## Recomendacao de Novos Endpoints

### 1. Atualizar dados basicos

- `PUT /users/:user_id`

Payload sugerido:

```json
{
  "name": "Maria Souza",
  "email": "maria@empresa.com",
  "username": "maria.souza"
}
```

### 2. Alterar perfil

- `PATCH /users/:user_id/role`

Payload sugerido:

```json
{
  "role": "admin"
}
```

### 3. Alterar status

- `PATCH /users/:user_id/active`

Payload sugerido:

```json
{
  "active": false
}
```

### 4. Redefinir senha

- `PATCH /users/:user_id/password`

Payload sugerido:

```json
{
  "password": "novaSenha123",
  "password_confirm": "novaSenha123"
}
```

## Recomendacoes de UX

### Recomendacao 1

Nao misturar a gestao de usuarios com a tela de login.

Motivo:

- login deve focar em autenticacao
- gestao de usuario e acao administrativa autenticada

### Recomendacao 2

Usar linguagem clara nos perfis:

- `Administrador`
- `Usuario`

Em vez de expor apenas:

- `admin`
- `user`

### Recomendacao 3

Usar `Chip` colorido para perfil e status:

- administrador: destaque mais forte
- usuario: destaque neutro
- ativo: verde
- inativo: cinza ou vermelho suave

### Recomendacao 4

Privilegiar seguranca no valor padrao:

- todo novo usuario nasce como `Usuario`

### Recomendacao 5

Manter a criacao simples na primeira versao:

- formulario curto
- sem excesso de filtros
- sem auditoria visual complexa no inicio

## Recomendacoes Tecnicas

### Recomendacao 1

Implementar primeiro o backend que falta:

1. `PATCH /users/:user_id/role`
2. `PATCH /users/:user_id/active`

Esses dois endpoints liberam o nucleo da administracao de acesso.

### Recomendacao 2

Depois implementar o frontend nesta ordem:

1. `UsersListPage`
2. `UserCreatePage`
3. acoes de alterar perfil
4. acoes de ativar ou desativar

### Recomendacao 3

Criar uma feature dedicada:

```text
frontend/src/features/users/
  api/
  components/
  pages/
  schemas/
  types/
```

### Recomendacao 4

Separar claramente:

- tela de `Usuarios`
- tela `Meu perfil`

Motivo:

- `Usuarios` e administrativo
- `Meu perfil` e individual

## Sequencia Recomendada de Implementacao

### Etapa 1

- criar os endpoints faltantes de role e active

### Etapa 2

- adicionar item `Usuarios` no menu apenas para admin
- criar rota protegida `/users`

### Etapa 3

- implementar listagem
- implementar criacao

### Etapa 4

- implementar alteracao de perfil
- implementar ativacao e desativacao

### Etapa 5

- implementar redefinicao de senha
- implementar auditoria

## Criterio de Pronto da Primeira Entrega

A primeira entrega da area `Usuarios` pode ser considerada pronta quando:

- admin consegue acessar `/users`
- admin consegue listar usuarios
- admin consegue cadastrar novo usuario com `role`
- usuario comum nao acessa essa area
- a interface deixa claro quem e admin e quem esta ativo

## Conclusao

A melhor evolucao para o tema de login e usuarios nao e expandir a tela de login, e sim criar uma area administrativa dedicada.

O caminho mais recomendado e:

1. completar a API de usuarios
2. criar a tela `Usuarios`
3. depois evoluir perfil, status e senha

Esse caminho reduz risco, organiza melhor a operacao e deixa a manutencao do sistema mais previsivel.
