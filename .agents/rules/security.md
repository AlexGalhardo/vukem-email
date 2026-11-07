# Security invariants

- Keep Google OIDC as the only production administrative login, restricted to the verified owner address.
- Preserve double opt-in, single-use hashed confirmation tokens, signed unsubscribe links and project isolation.
- GET requests must never confirm or cancel subscriptions.
- Execute campaign JavaScript only in QuickJS WebAssembly and sanitize rendered HTML.
- Never expose subscriber addresses, OAuth tokens or credentials in public props.
- Recheck subscription state in the queue worker before sending.
- Apply the vendored `security-audit` skill (`.agents/skills/security-audit`, Cloudflare) during development so security issues are audited as part of the work; use guidance mode unless the task is an explicit full audit. CI runs it in `.github/workflows/security-audit.yml`.
