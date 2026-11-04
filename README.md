# [Vukem Email](https://email.vukem.dev)

### Open source newsletter management for multiple projects

## Introduction

Vukem Email is a personal newsletter dashboard built with [Vukem](https://vukem.dev), Go, Inertia, React, PostgreSQL, Resend, and React Email. It manages subscribers, campaigns, and delivery for separate projects. The dashboard runs at [email.hexascore.app](https://email.hexascore.app); the technical documentation runs at [email.vukem.dev](https://email.vukem.dev).

## Summary

- [Features](#features)
- [Quick start](#quick-start)
- [Documentation](#documentation)
- [Stability](#stability)
- [Contributing](#contributing)
- [Code of conduct](#code-of-conduct)
- [Security vulnerabilities](#security-vulnerabilities)
- [License](#license)

## Features

- Private dashboard with Google sign-in restricted to one verified owner account.
- Isolated projects with their own public subscription domain and sender address.
- Double opt-in with expiring, single-use confirmation links and signed unsubscribe links.
- Markdown, MDX, and React Email campaigns with a sandboxed preview. Authored JavaScript runs only in QuickJS WebAssembly.
- Persistent delivery queue with per-recipient status checks, stable message IDs, retries, and provider idempotency.

## Quick start

Requires Go 1.27.3+, Node 24+, npm, and a Google OAuth Web application.

```bash
git clone https://github.com/AlexGalhardo/vukem-email.git
cd vukem-email
cp .env.example .env
npm ci
go run ./cmd/app key:generate
# Add the generated key and Google OAuth credentials to .env.
npm run build
go run ./cmd/app migrate
go run ./cmd/app
```

Open [http://localhost:3000](http://localhost:3000). Google sign-in is required locally too. For development with live reload, install Vukem v0.37.0 and run `vukem serve`.

```bash
npm run check
npm run build
go vet ./...
go test -race ./...
```

## Documentation

- [Online documentation](https://email.vukem.dev/docs/introduction) and its [source](site/).
- [Installation](https://email.vukem.dev/docs/installation), [configuration](https://email.vukem.dev/docs/configuration), and [Railway deployment](https://email.vukem.dev/docs/deployment).
- [Google OIDC](https://email.vukem.dev/docs/guides/google-oidc) and [security](https://email.vukem.dev/docs/guides/security).
- [AGENTS.md](AGENTS.md) for repository instructions.

Production runs the root Dockerfile for the web service and worker, plus `site/Dockerfile` for the documentation service. Set `APP_URL=https://email.hexascore.app` on both application services and register `https://email.hexascore.app/auth/google/callback` with Google. Use PostgreSQL, a verified Resend sender, and a separate queue worker. Deploy the documentation at `email.vukem.dev`.

## Stability

Releases follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html), use `vX.Y.Z` tags, and are recorded in [CHANGELOG.md](CHANGELOG.md).

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) for setup, checks, and pull request conventions. Keep code, documentation, and user-facing copy in English.

## Code of conduct

Please follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Security vulnerabilities

Report vulnerabilities through the [security policy](SECURITY.md), not a public issue.

## License

Vukem Email is open source software released under the [MIT license](LICENSE.md).
