# Security audit, 2026-11-07

Target: `vukem-email` at branch `feat/todo-2026-11` (PR #6), commit `b250ad5` plus a clean worktree.
Method: the vendored Cloudflare `security-audit` skill (`.agents/skills/security-audit`), full audit mode, profile `quick`.

Machine-readable records: [findings](security-audit-2026-11-07.findings.json), [coverage ledger](security-audit-2026-11-07.coverage.json).

## Scope and coverage statement

This run is partial coverage by design. Profile `quick`: one engineer, coarse ledger units (surface by boundary by attack class, 13 units), one hunting pass, no follow-up waves. No prior ledger existed. Source review only; no live endpoints were probed and no new dependency was installed. Two scanners ran locally: `govulncheck ./...` (no reachable vulnerabilities) and `npm audit --omit=dev` (4 low advisories, one root cause, not reachable; see Rejected).

In scope: the Go app (admin dashboard with Google login, public subscribe, confirm, unsubscribe and one-click endpoints, CORS allowlist, Resend webhook and suppression, campaign rendering and sanitizer, direct Resend sends, migrations), `resources/js`, `scripts/` runtime renderer, the Astro docs site in `site/`, CI workflows, Dockerfile, GoReleaser and the Railway IaC. Skipped: vendored skills, `node_modules`, generated output, and the `vukem` framework itself (consumed as a dependency; its IP, CSRF and session behaviour was only read where the app relies on it).

## Summary

| Severity | Confirmed | Needs validation | Rejected |
| --- | --- | --- | --- |
| Critical | 0 | - | - |
| High | 0 | - | - |
| Medium | 0 | - | - |
| Low | 1 | - | - |
| Informational | 2 | - | - |
| Unrated | - | 1 | 3 |

The application's stated invariants hold: Google OIDC is the only login and requires the verified owner address, confirmation tokens are 128-bit, hashed, single use and expiring, unsubscribe is an HMAC bound to the subscriber id and never changes state on GET, the queue worker rechecks status and suppression, public props carry no subscriber data, and campaign HTML is rendered from an isolated QuickJS tree and sanitized.

## Findings

### SA-001 Public subscribe lets anyone repeat confirmation emails (low)

- Fingerprint: `vukem-email:subscribe:unbounded-confirmation-mail`
- CWE: CWE-770 (allocation of resources without limits)
- Location: `app/controllers/newsletter.go:134`, `app/controllers/newsletter.go:152`, `routes/web.go:91`, `routes/web.go:126`
- Impact: any unauthenticated client can keep triggering project-branded confirmation emails to a third-party address, including one that already unsubscribed. Limits are only 10 requests per minute per IP and a 60 second cooldown per address, with no per-address or per-project cap over a longer window. Abuse costs Resend quota and sender reputation; there is no data exposure.
- Reproduction (code path): POST `{"email":"victim@example.com"}` to `/subscribe/{slug}` every 61 seconds. The row is not confirmed and `LastRequestedAt` is older than 60 seconds, so `Subscribe` rotates the token and `MailLaterTx` queues another email, answering the same 202.
- Fix: confirmations are capped at 3 per address per rolling 24 hours and 500 per project per hour, with the same 202 body (`subscribers.confirmations` and `confirmations_since`, migration `20261107000002_add_confirmation_caps`). Regression test: `TestConfirmationsCappedPerAddress`. Status: fixed.

### SA-002 Renderer subprocess inherits the full server environment (informational)

- Fingerprint: `vukem-email:renderer:env-inherited-by-node`
- CWE: CWE-526 (sensitive information in environment variables)
- Location: `app/newsletter/renderer.go:21`
- Impact: Node receives `APP_KEY`, `RESEND_API_KEY`, the Google client secret, the webhook secret and `DATABASE_URL`. QuickJS keeps authored code away from them today; a future sandbox or host-library escape would expose all of them.
- Reproduction: start the server with a secret in the environment, render with `node` replaced by a stub that prints its environment; the secret is present.
- Fix: set `cmd.Env` to an allowlist (`PATH`, `NODE_ENV`, platform essentials).

### SA-003 CodeQL checkout persists the GitHub token (informational)

- Fingerprint: `vukem-email:ci:codeql-checkout-persists-credentials`
- CWE: CWE-522 (insufficiently protected credentials)
- Location: `.github/workflows/codeql.yml:28`
- Impact: the job token stays in `.git/config` while autobuild runs repository code. Token scope is `contents: read` and `security-events: write`.
- Reproduction: inspect the workflow; the checkout step has no `persist-credentials: false`, unlike the other workflows.
- Fix: add `with: persist-credentials: false`.

## Needs validation

### NV-001 Railway IaC may drop `RESEND_WEBHOOK_SECRET` and `NEWSLETTER_ALLOWED_ORIGINS`

- Fingerprint: `vukem-email:deploy:railway-iac-omits-webhook-and-origins`
- Location: `.railway/railway.ts:8`, `routes/web.go:53`, `routes/web.go:132`
- Blocker: whether a Railway IaC apply deletes variables missing from the declaration is platform behaviour outside the repository. If it does, the webhook route disappears (bounces and complaints stop suppressing) and the CORS allowlist empties.
- Owner check: compare live service variables before and after an apply in a staging environment, or add both with `preserve()` once the semantics are confirmed.
- Status: resolved defensively. Both variables are declared with `preserve()` in `.railway/railway.ts`, which keeps the live value whatever the apply semantics are.

## Rejected after validation

- `vukem-email:renderer:tailwind-config-injection`: author JSON config reaches the React Email Tailwind compile step, but JSON cannot carry functions, output is clipped by the sanitizer style allowlist, and the author is the sole admin.
- `vukem-email:renderer:quickjs-escape`: no path from authored MDX to host objects; import allowlist, hydrate checks, memory and time limits hold, and tests cover process, require, fetch and constructor escapes.
- `vukem-email:deps:deepmerge-prototype-pollution`: GHSA-gcqx-w48m-w2h4 in the nested `html-to-text`; only static options reach it.

Dropped before records (could not be validated or are intended behaviour): timing differences revealing whether an address is already confirmed (no measurable reproduction; the 202 body is uniform); CRLF in `Project.Name` reaching the confirmation subject (admin-only input); `release.yml` dispatch input newline handling (needs write access already); `LIKE` wildcards in the admin search; missing CSP on the static docs site (hardening only, no dynamic content); stateless cookie sessions not revocable on logout (framework design, owner-only account).

Hardening notes, not findings: `docs/subscribe-api.md` states a 1 MiB body limit while the code enforces 8 KiB; container image tags are pinned by version rather than digest.

## Verification

Findings and ledger were validated with the skill's `validate-findings.cjs` and `validate-coverage-ledger.cjs` (run on Windows Node with a preload shim that defines the missing O_NOFOLLOW and O_NONBLOCK constants, since the validators require POSIX no-follow opens; both passed: 7 findings, 13 coverage units). Records were then re-read against source by an independent agent.
