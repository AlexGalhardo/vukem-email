# Contributing

Thank you for helping improve Vukem Email. Open an issue for a substantial change before implementation.

## Local setup

Use Go 1.27.4+, Node 24+ and PostgreSQL for integration tests. Copy `.env.example` to `.env`, install dependencies with `npm ci`, generate `APP_KEY`, run `npm run build`, then `go run ./cmd/app migrate` and `go run ./cmd/app`. Run `git config core.hooksPath .githooks` once: the hooks check commit messages and refresh the doc dates in `site/content/last-updated.json` before a push. A valid Google OAuth client is required for administrative login even locally.

## Verification

Run `npm run check`, `npm run build`, `go vet ./...` and `go test -race ./...`. Changes to `resources/js/emails` require `npm run build:emails` and the generated files in `resources/emails` in the same PR. Run `cd site && npm ci && npm run build` for documentation changes.

## Commits and releases

Use Conventional Commits in English: `feat(scope): description`, `fix(scope): description`, etc. Supported types: `feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `build`, `ci`, `chore`, `style`, `revert`. Use Semantic Versioning tags `vX.Y.Z`; record user-visible changes under `[Unreleased]` in `CHANGELOG.md`. Breaking changes require a `BREAKING CHANGE:` footer.

Before opening a PR, update tests and documentation for changed behavior. Never commit `.env`, credentials, subscriber data, SQLite files or auth bypasses. Security reports follow [SECURITY.md](SECURITY.md).
