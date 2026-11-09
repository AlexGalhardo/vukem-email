# Vukem Email

- Stack: Vukem v0.39, Go, Inertia v3, React, shadcn/ui. PostgreSQL in production; SQLite in tests/local development.
- Keep production authentication exclusively Google OIDC with verified `aleexgvieira@gmail.com`. Never add a development login endpoint or an authentication bypass to production routes. Browser fixtures live only in `_test.go` files.
- Preserve double opt-in, hashed single-use confirmation tokens, signed unsubscribe links, and project isolation. GET requests must not confirm or cancel subscriptions.
- Campaign JS executes only in QuickJS WebAssembly (`scripts/isolated-mdx.ts`). Never evaluate author code in Node, expose host callbacks or expand its imports to arbitrary server modules. Validate the serialized tree and sanitize the rendered HTML.
- Public props must never include subscriber addresses, confirmation hashes, OAuth tokens or credentials.
- Queue delivery must recheck subscription status. Keep stable message IDs, transactional fan-out and provider idempotency time limits.
- Change email wrappers in `resources/js/emails`, then run `npm run build:emails` and commit `resources/emails` too.
- Use `npm run check`, `npm run build`, `go vet ./...`, and `go test -race ./...` for relevant application changes. The README explains optional browser and PostgreSQL integration checks.
- Deploy using the Dockerfile: dynamic campaign rendering needs Node and production dependencies beside the Go executable. Never commit `.env`, SQLite files or credentials.
- Public POST routes (`/subscribe/{slug}`, `/unsubscribe/one-click`, `/webhooks/resend`) bypass CSRF and sessions in `routes/web.go`; keep their own protection (Origin allowlist via `NEWSLETTER_ALLOWED_ORIGINS`, HMAC token, Svix signature with replay window and dedupe) and the 8 KiB body limit. Never widen CORS beyond `/subscribe/`.
- Campaign mail goes through `newsletter.SendWithHeaders` (RFC 8058 headers; direct Resend API call because vukem mail has no header support, see the shared `vukem-feedback.md` at the workspace root, ids `VE-<n>`). Messages carry `project`, `campaign` and `delivery` tags that webhooks use to suppress addresses per project.
- Public subscribe endpoint contract: [docs/subscribe-api.md](docs/subscribe-api.md); change it together with the code. One-click unsubscribe follows RFC 8058; Resend webhooks verify the `svix-*` headers.
- Framework findings go in the single shared `vukem-feedback.md` at the workspace root (the folder that holds `vukem/` and every app; ids `VE-<n>`; fields Severity, Date, Docs, Finding, Minimal repro, Suggested fix, Status). Never create a copy inside this repo. The vukem maintainer fixes entries on one PR, one commit per entry, and updates `Status`.
- Releases follow the vukem pattern: the CHANGELOG section is the release notes (git-cliff), GoReleaser attaches assets with `release.mode: keep-existing`, cosign signatures and SBOMs. See `.agents/skills/release-checklist`.
- The docs site in `site/` is Astro with the same shell and pinned versions as `vukem/site` (not Next.js). Any change under `site/content/docs` needs `npm run last-updated` in `site/`.
- Apply the vendored `security-audit` skill ([.agents/skills/security-audit](.agents/skills/security-audit)) in guidance mode while developing; see [.agents/rules/security.md](.agents/rules/security.md).
