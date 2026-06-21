# Guia de Deploy Demo com Vercel, Render e Neon

## Objetivo

Este guia mostra como publicar uma versao de demonstracao do sistema usando:

- `Frontend`: Vercel
- `Backend`: Render
- `Banco PostgreSQL`: Neon

O foco aqui e colocar uma demo funcional no ar para apresentacao ao cliente, com o menor atrito possivel.

## Arquitetura Recomendada

- `Vercel`: hospeda o frontend React/Vite
- `Render`: publica a API em Go
- `Neon`: fornece o PostgreSQL gerenciado
- `Vercel Proxy`: encaminha `/api/*` para a API no Render

Esse proxy no Vercel e importante porque o sistema usa cookie de refresh token. Com frontend e backend em dominios diferentes, voce pode ter problema de autenticacao em navegadores. Com o proxy, o frontend fala com `/api` no mesmo host e a demo fica mais estavel.

## Pre-Requisitos

Antes de comecar, tenha em maos:

- conta no GitHub com o projeto publicado
- conta no Vercel
- conta no Render
- conta no Neon
- Docker funcionando na sua maquina local

## Estrategia De Publicacao

1. Criar o banco no Neon
2. Aplicar as migrations do projeto no banco novo
3. Publicar o backend no Render
4. Publicar o frontend no Vercel
5. Configurar o frontend para chamar a API via `/api`
6. Criar o primeiro usuario admin
7. Validar login, orcamentos, dashboards e comunicacao

## 1. Criar O Banco No Neon

1. Acesse o painel do Neon
2. Crie um novo projeto
3. Crie ou use o database principal com o nome `budget_management`
4. Copie a connection string completa

Exemplo de formato:

```text
postgres://usuario:senha@ep-xxxxxx.us-east-2.aws.neon.tech/budget_management?sslmode=require
```

Anote tambem os valores separados, porque o backend precisa de:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DATABASE_URL`

## 2. Aplicar As Migrations No Neon

Este projeto nao roda migrations automaticamente ao iniciar a API. Entao, antes do deploy da aplicacao, aplique todas as migrations no banco do Neon.

### Opcao recomendada para Windows com PowerShell

Abra um terminal na pasta `backend` do projeto e execute:

```powershell
$env:DATABASE_URL="COLE_AQUI_A_CONNECTION_STRING_DO_NEON"

Get-ChildItem .\db\migrations\*.sql |
  Sort-Object Name |
  ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $up = (($content -split '-- migrate:down')[0] -replace '-- migrate:up', '').Trim()

    if ([string]::IsNullOrWhiteSpace($up)) {
      throw "Migration sem bloco UP: $($_.Name)"
    }

    $up | docker run --rm -i postgres:17 `
      psql -d "$env:DATABASE_URL" -v ON_ERROR_STOP=1
  }
```

Se tudo estiver certo, as tabelas e dados base serao criados no Neon.

### Validacao rapida

Depois, voce pode validar conectando no editor SQL do Neon e executando algo simples:

```sql
select now();
```

## 3. Publicar O Backend No Render

1. Entre no Render
2. Crie um novo `Web Service`
3. Escolha o repositorio GitHub deste projeto
4. Se o `Root Directory` nao funcionar, deixe vazio e use os comandos com `cd backend`

Use:

- `Build Command`:

```bash
cd backend && go build -o app ./cmd
```

- `Start Command`:

```bash
cd backend && ./app
```

### Variaveis de ambiente do backend

Cadastre no Render:

```text
APP_ENV=production
LOG_LEVEL=info
SECRET_JWT=gere-um-valor-bem-forte-aqui
INITIAL_ADMIN_SETUP_TOKEN=gere-um-token-inicial-forte-aqui
REFRESH_COOKIE_NAME=budget_management_refresh
REFRESH_COOKIE_DOMAIN=
REFRESH_COOKIE_SECURE=true
DB_HOST=host-do-neon
DB_PORT=5432
DB_USER=usuario-do-neon
DB_PASSWORD=senha-do-neon
DB_NAME=budget_management
DATABASE_URL=postgres://usuario:senha@host-do-neon/budget_management?sslmode=require&channel_binding=require
ALLOWED_ORIGINS=http://localhost:5173,http://127.0.0.1:5173,https://SEU-FRONTEND.vercel.app
```

Observacoes:

- `REFRESH_COOKIE_SECURE=true` e o ideal em ambiente HTTPS
- `ALLOWED_ORIGINS` deve conter o dominio final do frontend no Vercel
- o projeto valida tanto `DATABASE_URL` quanto `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` e `DB_NAME`, por isso todos precisam ser preenchidos

### Validar o backend

Depois do deploy, abra a URL publica do Render e teste:

```text
https://SEU-BACKEND.onrender.com/check-health
```

O retorno esperado e:

```json
{"message":"service is healthy"}
```

## 4. Configurar O Frontend Para Usar Proxy `/api`

Para evitar problemas de cookie e refresh token em dominios diferentes, o frontend deve chamar a API por `/api`, e o Vercel deve encaminhar isso para o backend do Render.

### 4.1 Criar o arquivo `frontend/vercel.json`

Crie o arquivo `frontend/vercel.json` com este conteudo:

```json
{
  "routes": [
    {
      "src": "/api/(.*)",
      "dest": "https://SEU-BACKEND.onrender.com/$1"
    },
    {
      "handle": "filesystem"
    },
    {
      "src": "/(.*)",
      "dest": "/index.html"
    }
  ]
}
```

Troque `SEU-BACKEND.onrender.com` pela URL real do seu backend.

### 4.2 Ajustar a variavel do frontend

No Vercel, cadastre a variavel:

```text
VITE_API_URL=/api
```

Voce tambem pode definir:

```text
VITE_APP_NAME=Gestao de Orcamentos
```

## 5. Publicar O Frontend No Vercel

1. Entre no Vercel
2. Importe o mesmo repositorio GitHub
3. Defina o `Root Directory` como `frontend`
4. Confirme os comandos:

- `Install Command`:

```bash
yarn
```

- `Build Command`:

```bash
yarn build
```

- `Output Directory`:

```text
dist
```

### Variaveis de ambiente do frontend

Cadastre no Vercel:

```text
VITE_API_URL=/api
VITE_APP_NAME=Gestao de Orcamentos
```

Depois disso, publique o projeto.

## 6. Criar O Primeiro Usuario Admin

O primeiro usuario nao e criado pela interface. Ele precisa ser criado via API, usando o header `X-Setup-Token`.

Esse token precisa ser exatamente o mesmo valor configurado em:

```text
INITIAL_ADMIN_SETUP_TOKEN
```

### Exemplo em PowerShell

Substitua os valores abaixo e execute:

```powershell
$headers = @{
  "Content-Type" = "application/json"
  "X-Setup-Token" = "SEU_TOKEN_INICIAL"
}

$body = @{
  name = "Administrador"
  email = "admin@suaempresa.com"
  username = "admin"
  password = "SenhaForte#123"
} | ConvertTo-Json

Invoke-WebRequest `
  -Uri "https://SEU-BACKEND.onrender.com/auth/register" `
  -Method POST `
  -Headers $headers `
  -Body $body
```

Se tudo estiver certo, a API deve retornar `201 Created`.

Depois disso:

- o primeiro usuario admin passa a existir
- a rota de bootstrap inicial deixa de aceitar novos cadastros publicos

## 7. Validacao Final Da Demo

Depois dos deploys e do primeiro usuario criado, valide este fluxo:

1. abrir o frontend no Vercel
2. fazer login com o usuario admin
3. acessar a tela de orcamentos
4. acessar a tela de obras
5. acessar a tela de comunicacao
6. acessar o dashboard
7. validar os catalogos administrativos
8. testar logout e novo login

## 8. Checklist De Apresentacao

Antes de apresentar ao cliente:

- confirme se o backend responde `check-health`
- confirme se o frontend esta chamando `/api`
- confirme se as migrations foram aplicadas no Neon
- confirme se o primeiro usuario admin foi criado
- confira se o login esta funcionando
- confira se as rotas administrativas estao acessiveis para admin
- carregue alguns dados reais ou semi-reais para a demonstracao

## 9. Problemas Comuns

### Frontend sobe, mas login ou refresh falham

Causa comum:

- frontend e backend em dominios diferentes, sem proxy

Correcao:

- usar `VITE_API_URL=/api`
- configurar `frontend/vercel.json` com proxy para o Render

### CORS bloqueado no navegador

Causa comum:

- `ALLOWED_ORIGINS` no backend nao contem a URL final do Vercel

Correcao:

- adicionar `https://SEU-FRONTEND.vercel.app` em `ALLOWED_ORIGINS`
- redeploy do backend

### API sobe, mas faltam tabelas ou catalogos

Causa comum:

- migrations nao aplicadas no Neon

Correcao:

- executar novamente o passo de migrations

### Erro ao criar o primeiro admin

Causas comuns:

- `X-Setup-Token` diferente do configurado
- ja existe um usuario no banco

Correcao:

- revisar `INITIAL_ADMIN_SETUP_TOKEN`
- revisar se o banco ja possui usuarios

### Rotas como `/login` ou `/dashboard` retornam `404` no Vercel

Causa comum:

- falta de fallback SPA no `frontend/vercel.json`

Correcao:

- usar a configuracao com `routes`, `handle: filesystem` e fallback para `/index.html`

## 10. Limites Do Plano Gratis

Para demo e apresentacao, esse stack atende bem. Mas considere:

- pode haver `cold start`
- a API pode demorar mais na primeira requisicao
- o banco gratis tem limites
- esse ambiente nao deve ser tratado como producao

## 11. Proximo Passo Quando O Cliente Aprovar

Depois da aprovacao, o ideal e montar a esteira correta:

- ambiente `dev`
- ambiente `homolog`
- ambiente `prod`
- dominio oficial da empresa
- CI/CD
- backup
- monitoramento
- estrategia de seguranca e acesso interno

## Resumo Rapido

Se quiser a versao curta:

1. criar banco no Neon
2. aplicar migrations
3. publicar backend no Render
4. configurar `ALLOWED_ORIGINS`
5. criar `frontend/vercel.json`
6. publicar frontend no Vercel com `VITE_API_URL=/api`
7. criar primeiro admin via `POST /auth/register` com `X-Setup-Token`
8. validar a demo
