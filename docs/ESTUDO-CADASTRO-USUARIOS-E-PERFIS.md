# Estudo: Cadastro de Usuarios e Controle de Perfil Administrativo

## Objetivo

Mapear como o sistema pode:

- cadastrar um novo usuario
- definir se o usuario sera `admin` ou `user`
- permitir futura administracao desses perfis a partir da tela de login ou de uma area administrativa

Este estudo foi feito com base no comportamento atual do backend e do frontend existentes no projeto.

## Estado Atual

Hoje o sistema ja possui suporte de backend para autenticacao e para criacao de usuarios, mas esse fluxo ainda nao possui tela administrativa no frontend.

### Backend ja existente

- tabela `users` com os campos `name`, `email`, `username`, `password_hash`, `role` e `active`
- restricao de `role` apenas para `admin` ou `user`
- endpoint publico `POST /auth/register`
- endpoint publico `POST /auth/login`
- endpoint protegido `GET /users/me`
- endpoint administrativo `POST /users`
- endpoint administrativo `GET /users`

### Frontend ja existente

- tela de login
- contexto de autenticacao
- leitura do usuario autenticado em `/users/me`
- uso do campo `role` apenas para refletir sessao e permissoes basicas recebidas do backend

### O que ainda nao existe

- tela para cadastrar usuarios
- tela para listar usuarios
- tela para editar perfil de usuario
- acao para promover ou rebaixar usuario entre `admin` e `user`
- acao para ativar ou desativar usuario

## Regras Atuais de Negocio

### 1. Primeiro usuario do sistema

O endpoint `POST /auth/register` existe somente para criar o primeiro usuario do sistema.

Comportamento atual:

- se ainda nao existe nenhum usuario na base, o cadastro publico e permitido
- o primeiro usuario sempre nasce com perfil `admin`
- depois que existe pelo menos um usuario, o endpoint passa a responder `403` com a mensagem `public registration is no longer available`

Na pratica:

- o cadastro publico nao serve para operacao cotidiana
- ele serve apenas para bootstrap inicial do sistema

### 2. Cadastro de novos usuarios apos o primeiro acesso

Depois que o primeiro admin existe, novos usuarios so podem ser criados por um administrador autenticado, usando:

- `POST /users`

Payload atual:

```json
{
  "name": "Usuario Comum",
  "email": "user@local.dev",
  "username": "user",
  "password": "123456",
  "password_confirm": "123456",
  "role": "user"
}
```

Observacoes:

- o campo `role` ja define se o usuario sera `admin` ou `user`
- o backend ja valida o valor com `oneof=admin user`
- o usuario novo nasce com `active = true`

### 3. Perfil administrativo

Hoje o perfil administrativo e determinado por:

- coluna `role` na tabela `users`
- claim `role` emitida no JWT
- middleware `RequireRoles(model.RoleAdmin)` nas rotas administrativas

Isso significa que:

- um usuario `admin` acessa criacao e listagem de usuarios
- um usuario `user` nao consegue acessar essas rotas

### 4. Ativacao e desativacao

O modelo ja possui:

- campo `active` na tabela `users`

Mas hoje nao existe endpoint para:

- desativar usuario
- reativar usuario
- editar esse campo pela aplicacao

O efeito atual e:

- se `active = false`, o login falha
- porem a mudanca desse campo ainda depende de banco ou de nova API

## Como Cadastrar Um Novo Usuario Hoje

## Opcao 1: Fluxo suportado pelo sistema

Pre-requisito:

- existir um usuario `admin`
- estar autenticado com token valido

Fluxo:

1. fazer login com um usuario administrador
2. chamar `POST /users`
3. enviar o `role` desejado no payload
4. o backend cria o usuario com senha criptografada por `bcrypt`

Resultado:

- se `role = "admin"`, o usuario ja nasce administrador
- se `role = "user"`, o usuario nasce como usuario comum

## Opcao 2: Bootstrap inicial

Se a base estiver vazia:

1. chamar `POST /auth/register`
2. informar nome, email, username, password e password_confirm

Resultado:

- o primeiro usuario sempre sera `admin`

## Como Tornar Administrador Ou Nao Hoje

Hoje existem duas formas reais, sendo uma suportada e outra manual.

### Forma suportada pelo backend atual

No momento da criacao, basta definir:

- `role = "admin"` para criar administrador
- `role = "user"` para criar usuario comum

### Forma manual apos o usuario ja existir

Atualmente nao existe endpoint para alterar o `role` de um usuario existente.

Entao, para promover ou rebaixar um usuario ja cadastrado, hoje seria necessario:

1. atualizar a coluna `role` diretamente no banco
2. usar `admin` ou `user`
3. pedir novo login ao usuario para que um novo JWT seja emitido com o papel correto

Exemplo conceitual:

```sql
UPDATE users
SET role = 'admin',
    updated_at = NOW()
WHERE email = 'usuario@empresa.com';
```

Ou para remover privilegios:

```sql
UPDATE users
SET role = 'user',
    updated_at = NOW()
WHERE email = 'usuario@empresa.com';
```

## Limitacoes Identificadas

### 1. Nao existe tela administrativa de usuarios

O frontend atual nao consome:

- `POST /users`
- `GET /users`

Logo, o cadastro adicional de usuarios ja existe no backend, mas nao esta disponivel na interface.

### 2. Nao existe edicao de usuario

Faltam endpoints para:

- alterar `role`
- alterar `active`
- redefinir senha
- alterar nome, email e username

### 3. Nao existe protecao de seguranca para evitar erros operacionais

Exemplos de regras que ainda nao estao implementadas:

- impedir que o ultimo admin seja rebaixado para `user`
- impedir que um admin desative a si proprio sem outro admin ativo
- exigir confirmacao para promover usuario a `admin`

### 4. Login ainda e apenas de autenticacao

A tela de login hoje nao deve concentrar cadastro administrativo de usuarios.

Ela pode, no maximo:

- ter um atalho para `Solicitar acesso`
- ter um link para `Primeiro acesso`

Mas o gerenciamento de usuarios deve ficar em area protegida por admin.

## Recomendacao de Produto

### Recomendacao principal

Separar em dois fluxos:

### Fluxo A: Primeiro acesso

Manter:

- `POST /auth/register` somente para a primeira conta do sistema

Sugestao de UX:

- na tela de login, exibir o link `Primeiro acesso`
- esse link so deve funcionar se a base ainda nao tiver usuario
- caso contrario, mostrar mensagem informando que o cadastro inicial ja foi concluido

### Fluxo B: Gestao de usuarios por admin

Criar uma nova area administrativa:

- menu `Usuarios`
- listagem com nome, email, username, perfil e status
- botao `Novo usuario`
- acao `Tornar administrador`
- acao `Tornar usuario comum`
- acao `Ativar`
- acao `Desativar`
- acao `Redefinir senha`

Essa e a abordagem mais segura e coerente com o sistema atual.

## Recomendacao Tecnica

### Fase 1: Expor gestao basica de usuarios

Implementar no backend:

- `POST /users` reutilizando o que ja existe
- `GET /users` reutilizando o que ja existe
- `PATCH /users/:user_id/role`
- `PATCH /users/:user_id/active`

Payload sugerido para perfil:

```json
{
  "role": "admin"
}
```

Payload sugerido para status:

```json
{
  "active": false
}
```

### Fase 2: Tela administrativa

Implementar no frontend:

- pagina `Usuarios`
- formulario `Novo usuario`
- dialogos de confirmacao para alteracao de perfil
- filtros por perfil e status

### Fase 3: Regras de seguranca

Adicionar regras como:

- nao permitir remover o perfil do ultimo admin
- nao permitir desativar o unico admin ativo
- registrar auditoria de alteracoes de perfil

## Proposta de Experiencia na Tela de Login

Se a intencao for atuar primeiro na experiencia da tela de login, a melhor evolucao e:

### Manter a tela de login simples

Campos:

- email
- senha

Acoes:

- entrar
- primeiro acesso
- opcionalmente solicitar acesso

### Nao colocar gestao de perfil na tela de login

Motivo:

- promover usuario para admin e uma acao administrativa
- isso deve ficar atras de autenticacao e autorizacao
- misturar essa regra na tela de login tende a gerar risco operacional

## Proposta Objetiva Para o Proximo Passo

Se a prioridade for preparar a gestao de usuarios, a sequencia recomendada e:

1. criar documentacao funcional da area `Usuarios`
2. implementar backend para alterar `role` e `active`
3. criar tela administrativa de usuarios no frontend
4. manter na tela de login apenas `Primeiro acesso`, se fizer sentido para o produto

## Conclusao

O sistema ja permite:

- criar o primeiro admin via `POST /auth/register`
- criar novos usuarios via `POST /users`
- definir no cadastro se o novo usuario sera `admin` ou `user`

O sistema ainda nao permite, pela interface ou por API dedicada:

- promover um usuario existente para admin
- rebaixar um admin para user
- ativar ou desativar um usuario existente

Portanto, o melhor proximo investimento nao e ampliar a tela de login para gerir perfis, e sim criar uma area administrativa de usuarios com controle de perfil e status.
