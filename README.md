<div align="center">

<h1><strong><a href="https://email.vukem.dev" style="text-decoration: none" target="_blank">Vukem Email</a></strong></h1>

[![CI](https://github.com/AlexGalhardo/vukem-email/actions/workflows/ci.yml/badge.svg)](https://github.com/AlexGalhardo/vukem-email/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)
[![Go 1.27.3](https://img.shields.io/badge/go-1.27.3-00ADD8?logo=go&logoColor=white)](go.mod)
<a href="https://github.com/AlexGalhardo/vukem-email/releases"><img src="https://img.shields.io/github/v/release/AlexGalhardo/vukem-email?include_prereleases&label=version" alt="Version"></a>
<a href="https://scorecard.dev/viewer/?uri=github.com/AlexGalhardo/vukem-email"><img src="https://api.scorecard.dev/projects/github.com/AlexGalhardo/vukem-email/badge" alt="OpenSSF Scorecard"></a>

<h3><strong>Open Source Newsletter Management for Multiple Projects</strong></h3>

</div>

## Introduction

Vukem Email is a newsletter dashboard that manages subscribers, campaigns and delivery for separate projects, each with its own public subscription domain and sender address. The dashboard runs at [email.hexascore.app](https://email.hexascore.app) and the documentation at [email.vukem.dev](https://email.vukem.dev).

This application is built using the [Vukem Go framework](https://vukem.dev), with Inertia, React, PostgreSQL, Resend and React Email.

## Summary

- [Introduction](#introduction)
- [Features](#features)
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
- Public JSON subscribe endpoint with a CORS allowlist for external sites ([contract](docs/subscribe-api.md)), capped per address and per project.
- RFC 8058 one-click unsubscribe and Resend webhooks that suppress hard bounces and complaints per project.
- Markdown, MDX and React Email campaigns with a sandboxed preview. Authored JavaScript runs only in QuickJS WebAssembly.
- Persistent delivery queue with per-recipient status checks, stable message IDs, retries and provider idempotency.

## Documentation

- [Online documentation](https://email.vukem.dev/docs/introduction) and its [source](site/).
- [Installation](https://email.vukem.dev/docs/installation), [configuration](https://email.vukem.dev/docs/configuration) and [Railway deployment](https://email.vukem.dev/docs/deployment).
- [Subscribe API](https://email.vukem.dev/docs/guides/subscribe-api) and [Resend webhooks](https://email.vukem.dev/docs/guides/resend-webhooks).
- [Google OIDC](https://email.vukem.dev/docs/guides/google-oidc), [security](https://email.vukem.dev/docs/guides/security) and the [2026-11-07 security audit](docs/security-audit-2026-11-07.md).
- [AGENTS.md](AGENTS.md) for repository instructions.

## Stability

Vukem Email follows [Semantic Versioning](https://semver.org) and tags releases as `vX.Y.Z` on `main`. Each release publishes signed binaries, checksums and SBOMs on [GitHub Releases](https://github.com/AlexGalhardo/vukem-email/releases).

Changes are listed in [CHANGELOG.md](CHANGELOG.md), which follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Contributing

Thank you for considering a contribution to Vukem Email.

- Write the test first. Run the specific test, lint and typecheck that cover your change: `npm run check`, `npm run build`, `go vet ./...` and `go test -race ./...`.
- Keep code, documentation and user-facing copy in English.
- Update `CHANGELOG.md` (under `[Unreleased]`) and the documentation in the same change when behaviour, commands or conventions change.

The full guide (local setup, optional browser and PostgreSQL checks, commit style) is in [CONTRIBUTING.md](CONTRIBUTING.md).

### Pull requests

1. Create a branch from `main`.
2. Make one logical change per commit.
3. Open a pull request against `main`.

`main` accepts changes through a pull request with green checks and linear history.

### Commits

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) in English: `type(scope): description`.

- Types: `feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `build`, `ci`, `chore`, `style`, `revert`.

Run `git config core.hooksPath .githooks` once: the commit-msg hook checks the header locally, and the pre-push hook refreshes the documentation dates. CI checks the commits and the pull request title.

## Code of conduct

To keep the community welcoming to everyone, please review and follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Security vulnerabilities

If you discover a security vulnerability, please follow the [security policy](SECURITY.md) instead of opening a public issue.

## License

Vukem Email is open source software released under the [MIT license](LICENSE.md).
