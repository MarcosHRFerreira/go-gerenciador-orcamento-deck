# Guia Visual do Frontend

## Objetivo

Definir a direcao visual do frontend com foco em um design:

- moderno
- limpo
- profissional
- administrativo
- facil de manter

Este guia complementa:

- `docs/ARQUITETURA-FRONTEND.md`
- `docs/BACKLOG-FRONTEND-IMPLEMENTACAO.md`

## Direcao Visual

O frontend deve transmitir:

- organizacao
- clareza
- confianca
- agilidade operacional

O estilo recomendado nao e chamativo nem institucional.

Ele deve parecer um painel administrativo atual, com:

- muito espaco em branco
- contraste bem controlado
- hierarquia visual clara
- poucos acentos de cor
- superficies discretas
- interacoes suaves

## Principios de Design

### 1. Minimalismo funcional

Cada elemento da tela deve ter uma funcao clara.

Evitar:

- excesso de bordas
- gradientes fortes
- sombras pesadas
- muitos cards competindo entre si
- muitas cores ao mesmo tempo

Preferir:

- composicao limpa
- destaque apenas no que e acionavel ou importante
- componentes com leitura imediata

### 2. Hierarquia visual forte

O usuario deve bater o olho e entender:

- onde esta
- o que pode fazer
- o que e prioritario
- o que e informacao secundaria

Aplicar isso com:

- titulos bem definidos
- subtitulos curtos
- labels consistentes
- espacamento generoso
- tamanhos de fonte previsiveis

### 3. Densidade equilibrada

Como o sistema e administrativo, ele nao deve ser "vazio demais", mas tambem nao deve parecer apertado.

Padrao recomendado:

- densidade media
- tabelas compactas, mas legiveis
- formularios com respiro
- blocos organizados em secoes

### 4. Consistencia acima de criatividade isolada

O frontend deve parecer uma aplicacao unica, nao uma colecao de telas independentes.

Isso significa padronizar:

- botoes
- espacamentos
- titulos
- acoes primarias
- modais
- filtros
- tabelas
- estados vazios

## Referencia de Estilo

O visual ideal se aproxima de:

- dashboards SaaS modernos
- paineis B2B com visual leve
- interfaces administrativas com base neutra e acentos discretos

Em termos praticos:

- base clara
- fundo cinza muito suave
- cards brancos
- bordas sutis
- azul como cor primaria
- tipografia limpa

## Tema Recomendado

### Modo inicial

- iniciar com `light theme`

Motivo:

- combina melhor com leitura operacional
- deixa tabelas e formularios mais limpos
- reduz complexidade no MVP

Modo escuro pode entrar depois, mas nao deve ser prioridade inicial.

### Paleta de cores

Base recomendada:

- `background default`: cinza muito claro
- `background paper`: branco
- `primary`: azul elegante e moderado
- `success`: verde suave
- `warning`: amarelo/ambar controlado
- `error`: vermelho limpo
- `text primary`: quase preto
- `text secondary`: cinza escuro
- `divider`: cinza claro

Sugestao inicial de tons:

```text
primary: #2563EB
primary dark: #1D4ED8
primary light: #DBEAFE

background default: #F5F7FB
background paper: #FFFFFF

text primary: #111827
text secondary: #6B7280

border/divider: #E5E7EB

success: #16A34A
warning: #D97706
error: #DC2626
info: #0284C7
```

## Tipografia

### Familia

Usar uma fonte limpa e neutra.

Sugestao:

- `Inter`

Fallback:

- `system-ui`
- `Segoe UI`
- `Arial`

### Escala tipografica

Padrao sugerido:

- titulo principal da pagina: `28px` semibold
- subtitulo da pagina: `14px` a `16px`
- titulo de secao: `18px` a `20px`
- texto padrao: `14px`
- label de formulario: `13px` a `14px`
- texto auxiliar: `12px`

### Peso visual

Preferir:

- `500` para destaque intermediario
- `600` para titulos e acoes importantes

Evitar:

- muitos textos em `700`
- textos muito pequenos

## Grid e espacamento

### Unidade base

Usar escala multipla de `4px`.

Sugestao pratica:

- `4px`
- `8px`
- `12px`
- `16px`
- `20px`
- `24px`
- `32px`

### Regras gerais

- gap entre campos relacionados: `16px`
- gap entre secoes: `24px` ou `32px`
- padding interno de cards: `20px` ou `24px`
- espacamento vertical entre titulo e conteudo: `16px`

## Layout geral

## Estrutura do AppShell

O layout principal deve ter:

- sidebar fixa ou semi-fixa
- topo discreto
- conteudo central com largura confortavel

### Sidebar

Caracteristicas:

- fundo escuro suave ou branco com borda
- icones simples
- destaque claro da rota ativa
- grupos bem separados

Se optar por sidebar escura, recomendacao:

- fundo `#0F172A`
- texto principal claro
- item ativo com fundo azul discreto

Se optar por sidebar clara, recomendacao:

- fundo branco
- borda lateral suave
- item ativo com `primary.light`

Para manter o visual mais moderno e limpo, minha recomendacao principal e:

- sidebar escura
- area de conteudo clara

Esse contraste ajuda a dar identidade sem poluir o restante da tela.

### Topbar

Caracteristicas:

- altura baixa a media
- sem excesso de elementos
- nome da pagina
- busca futura opcional
- nome do usuario
- menu de perfil e logout

Evitar:

- muitos botoes no topo
- excesso de cor

### Conteudo

Padrao:

- fundo cinza muito claro
- cards brancos
- margens amplas
- largura maxima controlada em algumas telas de formulario

## Componentes-chave

## Cards

Os cards devem ser discretos.

Padrao recomendado:

- fundo branco
- borda de `1px` muito suave
- sombra leve
- raio de borda entre `12px` e `16px`

Evitar:

- cards com sombra pesada
- cards com fundos coloridos sem necessidade

## Botoes

### Botao primario

Uso:

- salvar
- criar
- confirmar
- entrar

Estilo:

- fundo `primary`
- texto branco
- altura consistente
- raio medio

### Botao secundario

Uso:

- cancelar
- voltar
- limpar filtro

Estilo:

- outlined ou tonal
- sem competir com a acao principal

### Botao destrutivo

Uso:

- excluir
- remover

Estilo:

- vermelho apenas no contexto da acao
- nunca como cor dominante da tela

## Campos de formulario

Padrao visual:

- cantos arredondados
- borda discreta
- foco com cor primaria
- labels claras
- mensagens de erro objetivas

Recomendacoes:

- usar largura total por padrao
- agrupar campos em grid quando fizer sentido
- manter labels sempre visiveis

## Tabelas

As tabelas serao um dos elementos mais importantes do sistema.

Padrao recomendado:

- header com fundo muito suave
- linhas brancas
- hover discreto
- tipografia legivel
- acoes no final da linha
- status com chip ou badge

Para manter o visual limpo:

- evitar muitas linhas verticais
- usar divisores suaves
- esconder excesso de informacao secundaria

## Filtros

O bloco de filtros deve ficar em um card proprio acima da tabela.

Padrao:

- layout em grid responsivo
- acao primaria `Filtrar`
- acao secundaria `Limpar`
- resumo opcional dos filtros ativos

## Dialogs e modais

Usar modal apenas quando:

- confirmar exclusao
- executar acao sensivel
- abrir formulario curto

Evitar:

- modais muito grandes para formularios complexos

Para formularios maiores, preferir pagina dedicada ou drawer lateral.

## Status e feedback

### Chips

Usar chips para:

- status de orcamento
- ativo/inativo
- perfil de usuario

Padrao:

- fundo suave
- texto com contraste bom
- sem saturacao excessiva

### Alerts

Usar alertas para:

- erro de salvamento
- sucesso de operacao
- aviso de validacao ou acao irreversivel

### Toasts

Usar toasts para:

- feedback rapido de sucesso
- erro nao bloqueante

Evitar toasts para tudo.

## Paginas principais

## Login

O login deve ser simples e elegante.

Composicao sugerida:

- fundo claro com textura muito sutil ou bloco neutro
- card central
- logo ou nome do sistema
- titulo curto
- dois campos
- botao primario forte

Visual desejado:

- profissional
- silencioso
- sem elementos decorativos excessivos

## Dashboard

Se existir dashboard no MVP, ele deve ser resumido e util.

Sugestao:

- cards de indicadores principais
- tabela curta de ultimos orcamentos
- alertas operacionais

Se nao houver dados relevantes ainda, melhor mostrar uma home simples com atalhos.

## Listagem de orcamentos

Essa deve ser a tela mais madura visualmente.

Estrutura recomendada:

- `PageHeader`
- bloco de filtros
- bloco de tabela
- acoes de criacao
- paginacao clara

## Detalhe do orcamento

Estruturar em blocos:

- resumo principal
- informacoes comerciais
- informacoes financeiras
- follow-ups
- historico de status

Evitar colocar tudo em uma unica coluna longa sem separacao.

## Formularios

Usar secoes com titulos curtos, por exemplo:

- Identificacao
- Relacionamentos
- Valores
- Contexto comercial

## Responsividade

Mesmo sendo um painel administrativo desktop-first, o frontend deve funcionar bem em telas menores.

Prioridade:

- desktop
- notebook
- tablet em paisagem

No mobile:

- sidebar vira drawer
- filtros quebram em mais linhas
- tabelas podem precisar de scroll horizontal

## Motion e microinteracoes

Para manter o aspecto moderno:

- hover suave em botoes e linhas
- transicoes curtas
- feedback visual claro no foco

Evitar:

- animacoes longas
- efeitos chamativos

## O que evitar

Para preservar o visual moderno e limpo, evitar:

- excesso de cores de destaque
- muitos elementos com sombra forte
- telas lotadas de cards
- muitos botoes primarios na mesma area
- labels inconsistentes
- misturar estilos de borda e raio
- muitos tamanhos de fonte sem criterio
- icones decorativos sem funcao

## Componentes base recomendados

No bootstrap do frontend, vale criar estes componentes com cuidado visual:

- `AppShell`
- `AppSidebar`
- `Topbar`
- `PageHeader`
- `SectionCard`
- `DataTable`
- `EmptyState`
- `ErrorState`
- `ConfirmDialog`
- `StatusChip`
- `FormTextField`
- `FormSelectField`
- `PageActions`

## Recomendacao pratica para o MVP

Se for implementar rapido sem perder qualidade visual, recomendo:

1. usar `MUI` com tema customizado leve
2. definir desde o inicio `palette`, `shape`, `typography` e `spacing`
3. criar `AppShell`, `PageHeader`, `SectionCard` e `DataTable` como base
4. manter o resto da UI muito proximo do design system padrao

## Decisao recomendada

Se eu tivesse que resumir a direcao visual ideal para este projeto, seria:

- painel administrativo claro
- sidebar escura elegante
- conteudo claro e espaçado
- azul como cor primaria
- cards brancos com borda suave
- tipografia `Inter`
- formularios limpos
- tabelas organizadas
- feedbacks discretos e profissionais

## Proximo Passo Recomendado

Depois deste guia visual, o proximo passo ideal e:

1. traduzir essas decisoes para o tema do `MUI`
2. criar um arquivo de tokens visuais em `frontend/src/app/theme`
3. implementar `AppShell`, `PageHeader` e `SectionCard` ja seguindo este padrao
