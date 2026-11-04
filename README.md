# Vukem Email

Gerenciador de newsletters para os seus projetos, feito com **Vukem + Go, Inertia + React, shadcn/ui, Resend e React Email**. Preparado para PostgreSQL e deploy no Railway.

## O que está implementado

- Dashboard em `/`, com sidebar, métricas reais por projeto e layout responsivo.
- Login exclusivo com Google: somente **`aleexgvieira@gmail.com`**, com endereço verificado pelo provedor. Sem cadastro público, senha ou acesso administrativo alternativo.
- Vários projetos, cada um com nome, subdomínio, descrição e remetente próprios.
- Landing pública em `/subscribe`, identificada pelo hostname, ou `/subscribe/{slug}`.
- Double opt-in: inscrição pendente → e-mail → confirmação explícita → inscrição confirmada. Tokens aleatórios, armazenados como hash, expiram em 24 horas e funcionam uma única vez.
- Busca, filtros, paginação, descadastro administrativo e exportação CSV por projeto.
- Campanhas em Markdown GFM, MDX e componentes React Email, com modelos de boas-vindas, lançamento e resumo.
- Rascunhos editáveis, prévia desktop/mobile e revisão antes do envio.
- Fila persistente e transacional. Cada destinatário tem uma tarefa própria; o worker verifica novamente o status antes de enviar. Os destinatários nunca são expostos uns aos outros.
- Tentativas automáticas, IDs estáveis de mensagem e estados `queued`, `accepted`, `skipped` e `failed` por destinatário.
- Descadastro assinado, sem necessidade de login. Visitar um link via GET não confirma nem cancela a inscrição: é necessário clicar no botão, para evitar ações de scanners de e-mail.

`accepted` significa **aceito pelo provedor**, e não entrega confirmada na caixa de entrada. Não há métricas inventadas de abertura, cliques ou entrega. Webhooks de entrega, bounce e reclamações do Resend ainda não estão implementados; acompanhe esses eventos no painel do Resend.

## Desenvolvimento

Requisitos: Go **1.27.3+**, Node **24+**, npm e uma aplicação Google OAuth do tipo Web.

```bash
cp .env.example .env
npm ci
go run ./cmd/app key:generate
# Coloque a chave gerada em APP_KEY no .env.
npm run build
go run ./cmd/app migrate
go run ./cmd/app
```

Abra `http://localhost:3000`. O login exige credenciais Google válidas, inclusive localmente. Para desenvolvimento com recarga automática:

```bash
go install github.com/AlexGalhardo/vukem/cmd/vukem@v0.37.0
vukem serve
```

SQLite é usado localmente. `MAIL_MAILER=array` não envia mensagens externas; os testes verificam o conteúdo das mensagens em memória. Para enviar de verdade, use `MAIL_MAILER=resend`, uma chave Resend e um remetente verificado. Os links de confirmação e descadastro usam o domínio configurado no projeto e HTTPS; configure esse domínio antes de fazer testes reais com e-mail.

O processo Go inicia um subprocesso Node apenas ao compilar ou pré-visualizar conteúdo de campanha. **O binário Go isolado não substitui o runtime Docker**: Node, as dependências de produção e os scripts de renderização precisam estar disponíveis. O envio dos e-mails já compilados é feito pelo Go.

## Google OAuth

No Google Cloud Console:

1. Configure a tela de consentimento e, se o app estiver em modo de teste, adicione `aleexgvieira@gmail.com` como test user.
2. Crie um OAuth Client ID do tipo **Web application**.
3. Cadastre as URIs de redirecionamento exatas:

```text
http://localhost:3000/auth/google/callback
https://news.hexascore.app/auth/google/callback
```

4. Preencha `SOCIAL_GOOGLE_CLIENT_ID`, `SOCIAL_GOOGLE_CLIENT_SECRET` e `APP_URL`.

O Vukem valida state de uso único, PKCE, nonce, assinatura do ID token, issuer, audience e expiração. O callback só cria uma sessão administrativa se o e-mail verificado for o autorizado. Tokens Google não são enviados para o frontend nem persistidos pela aplicação.

Para administrar vários domínios, use o domínio canônico de `APP_URL` para o dashboard e login. Os outros domínios podem servir suas páginas públicas; as sessões são restritas ao hostname.

## Criando um projeto

Após o login, abra **Projetos → Novo projeto**. Exemplo:

| Campo | Valor |
| --- | --- |
| Nome | Hexascore |
| Identificador | hexascore |
| Domínio | news.hexascore.app |
| Remetente | newsletter@hexascore.app |

Verifique o domínio do remetente no Resend, com os registros DNS que ele fornecer. O domínio da página pública e o domínio remetente podem ser diferentes.

A página pública estará em `https://news.hexascore.app/subscribe`. Adicione um link para ela na landing page do projeto. A tela **Configurações** fornece o HTML do link. Essa implementação usa uma página hospedada; não expõe uma API de inscrição com CORS aberto para formulários de outros sites.

## Markdown, MDX e React Email

Markdown usa GFM: títulos, listas, links, tabelas, código e texto riscado. O modo **React Email** oferece modelos e botões para inserir componentes. O modo **MDX** combina Markdown, JSX, variáveis, expressões, funções e componentes definidos no próprio conteúdo:

```mdx
import { Button, Section, Text } from '@react-email/components'

export const features = ['Mais simples', 'Mais rápido']
export const Highlight = ({ children }) => <Text>{children}</Text>

# Temos novidades!

<Section>
  {features.map(feature => <Highlight key={feature}>{feature}</Highlight>)}
  <Button href="https://hexascore.app">Conhecer as novidades</Button>
</Section>
```

São permitidos imports de `@react-email/components` para os componentes disponíveis e de `react` para renderização estática (`createElement` e `Fragment`). Não há acesso a módulos arbitrários do servidor, hooks ou APIs de navegador. Componentes disponíveis: `Button`, `Column`, `Container`, `Heading`, `Hr`, `Img`, `Link`, `Row`, `Section`, `Text`, `CodeInline`, `CodeBlock`, `Markdown` e `Tailwind`. O documento externo, cabeçalho, marca e rodapé com descadastro são controlados pela aplicação.

O JavaScript do MDX executa em **QuickJS dentro de WebAssembly**, sem expor Node, arquivos, rede, variáveis de ambiente ou callbacks do host. Há limites de memória, CPU, tamanho e profundidade. Apenas uma árvore de elementos serializada sai do interpretador; ela é validada, renderizada com React Email e sanitizada. A prévia usa uma rota autenticada temporária e um iframe com sandbox; scripts e formulários ficam desativados. A CSP global continua restrita.

Os templates de confirmação e envelope da campanha vivem em `resources/js/emails/`. `npm run build:emails` gera os templates Go em `resources/emails/`, que são incorporados ao binário. Alterou um template? Gere e versione novamente os arquivos HTML.

## Railway

O repositório inclui um `Dockerfile` de produção e a configuração declarativa `.railway/railway.ts`. Você também pode criar os serviços pelo painel do Railway:

1. Crie um projeto e adicione **PostgreSQL**.
2. Adicione um serviço web apontando para este repositório, usando o Dockerfile.
3. Configure **Pre-deploy command**: `/app/app migrate`; **Healthcheck path**: `/up`.
4. Configure as variáveis abaixo no serviço web.
5. Crie um segundo serviço usando o mesmo repositório, com **Start command** `/app/app queue:work`, as mesmas variáveis e sem domínio público.
6. Configure `QUEUE_IN_SERVE=false` no web e mantenha um worker com `QUEUE_CONCURRENCY=1` inicialmente. Ajuste capacidade conforme os limites de envio da sua conta Resend.

| Variável | Valor |
| --- | --- |
| `APP_ENV` | `production` |
| `APP_DEBUG` | `false` |
| `APP_KEY` | Chave gerada; igual no web e no worker |
| `APP_URL` | `https://news.hexascore.app` |
| `DATABASE_URL` | Referência ao PostgreSQL do Railway |
| `SOCIAL_GOOGLE_CLIENT_ID` | ID do cliente Google |
| `SOCIAL_GOOGLE_CLIENT_SECRET` | Segredo do cliente Google |
| `RESEND_API_KEY` | Chave Resend com permissão de envio |
| `MAIL_MAILER` | `resend` |
| `MAIL_FROM` | Remetente padrão verificado |
| `CACHE_STORE` | `database` |
| `SESSION_DRIVER` | `cookie` |
| `QUEUE_CONNECTION` | `database` |
| `QUEUE_IN_SERVE` | `false`, com worker separado |
| `QUEUE_CONCURRENCY` | `1` inicialmente |
| `TRUSTED_PROXIES` | `100.64.0.0/24`, confirme os peers nos logs do Railway |
| `CLIENT_IP_HEADER` | `X-Real-IP` |
| `TRUST_FORWARDED_PROTO` | `true` |

O Railway injeta `PORT`. A imagem já define `HOST=0.0.0.0`. Não inclua segredos no Git ou em build arguments. Para apenas um serviço, use `QUEUE_IN_SERVE=true`: o servidor também processa a fila. **Não use SQLite em produção sem volume persistente**; a configuração recomendada usa PostgreSQL.

Adicione `news.hexascore.app` em **Networking → Custom Domain** do serviço web, e crie o CNAME/registro de verificação exatamente como o Railway informar. Para outro projeto, adicione outro subdomínio ao mesmo serviço e crie o projeto correspondente no dashboard. As migrações devem terminar antes do novo web e worker receberem tráfego.

Fila e falhas podem ser inspecionadas pelo CLI da aplicação:

```bash
/app/app queue:failed
/app/app queue:retry
```

Envios usam uma chave de idempotência estável por campanha/inscrito. As tarefas de entrega recusam novas tentativas após 23 horas da criação, para não ultrapassar a janela de 24 horas do Resend. Rotacionar `APP_KEY` sem manter a chave anterior interrompe sessões e tarefas criptografadas; mantenha `APP_PREVIOUS_KEYS` durante a rotação. Links de descadastro anteriores continuam válidos enquanto a chave anterior estiver em `APP_PREVIOUS_KEYS`.

## Verificação

```bash
npm run check
npm run build
go vet ./...
go test -race ./...
```

O teste de navegador usa fixtures e cookies apenas dentro de `go test`; o servidor de produção não possui bypass de autenticação:

```bash
# Chromium local ou outro executável indicado por BROWSER_EXECUTABLE
RUN_BROWSER_TESTS=1 BROWSER_EXECUTABLE=/usr/bin/chromium go test ./tests -run TestBrowserSmoke -v
```

Para executar a integração com um PostgreSQL descartável:

```bash
TEST_DATABASE_URL=postgres://newsletter:test@localhost:5432/newsletter_test?sslmode=disable go test ./tests -run TestPostgresWorkflow -v
```

Os testes cobrem autorização, dados públicos, double opt-in, tokens expirados/consumidos, descadastro, isolamento entre projetos, envio único, fila persistente, renderização, isolamento do MDX e fluxos no navegador. Os testes locais não fazem login no Google nem enviam e-mails reais pelo Resend.

Framework: [Vukem](https://vukem.dev/docs/introduction). Interface: [shadcn/ui](https://ui.shadcn.com/). Deploy: [Railway](https://railway.com/).
