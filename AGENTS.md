# Vukem Email

- Stack: Vukem v0.37, Go, Inertia v3, React, shadcn/ui. PostgreSQL in production; SQLite in tests/local development.
- Keep production authentication exclusively Google OIDC with verified `aleexgvieira@gmail.com`. Never add a development login endpoint or an authentication bypass to production routes. Browser fixtures live only in `_test.go` files.
- Preserve double opt-in, hashed single-use confirmation tokens, signed unsubscribe links, and project isolation. GET requests must not confirm or cancel subscriptions.
- Campaign JS executes only in QuickJS WebAssembly (`scripts/isolated-mdx.ts`). Never evaluate author code in Node, expose host callbacks or expand its imports to arbitrary server modules. Validate the serialized tree and sanitize the rendered HTML.
- Public props must never include subscriber addresses, confirmation hashes, OAuth tokens or credentials.
- Queue delivery must recheck subscription status. Keep stable message IDs, transactional fan-out and provider idempotency time limits.
- Change email wrappers in `resources/js/emails`, then run `npm run build:emails` and commit `resources/emails` too.
- Use `npm run check`, `npm run build`, `go vet ./...`, and `go test -race ./...` for relevant application changes. The README explains optional browser and PostgreSQL integration checks.
- Deploy using the Dockerfile: dynamic campaign rendering needs Node and production dependencies beside the Go executable. Never commit `.env`, SQLite files or credentials.
- Public POST routes (`/subscribe/{slug}`, `/unsubscribe/one-click`, `/webhooks/resend`) bypass CSRF and sessions in `routes/web.go`; keep their own protection (Origin allowlist via `NEWSLETTER_ALLOWED_ORIGINS`, HMAC token, Svix signature with replay window and dedupe) and the 8 KiB body limit. Never widen CORS beyond `/subscribe/`.
- Campaign mail goes through `newsletter.SendWithHeaders` (RFC 8058 headers; direct Resend API call because vukem mail has no header support, see `vukem-feedback.md`). Messages carry `project`, `campaign` and `delivery` tags that webhooks use to suppress addresses per project.
