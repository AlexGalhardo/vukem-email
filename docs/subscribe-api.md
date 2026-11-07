# Subscribe API

Public, unauthenticated endpoint that external sites (for example the HexaScore footer) call from the browser.

```http
POST https://email.hexascore.app/subscribe/{slug}
Content-Type: application/json
Accept: application/json

{"email": "ana@example.com"}
```

`{slug}` is the project slug (HexaScore: `hexascore`).

## Request

| Field | Type | Rules |
| --- | --- | --- |
| `email` | string | Required, a valid address of up to 254 characters. Lowercased and trimmed by the server. |
| `website` | string | Honeypot. Leave empty or omit. A non-empty value is acknowledged with 202 but no email is sent. |

Form-encoded bodies (`application/x-www-form-urlencoded`) are accepted too. The body is limited by the framework (1 MiB by default). Do not send cookies: use `fetch(url, { method: "POST", headers: { "Content-Type": "application/json", Accept: "application/json" }, body })` with no `credentials`.

## Responses

Every response is JSON.

| Status | When | Body |
| --- | --- | --- |
| 202 | Accepted. Always the same body, whether or not an email was sent (new, pending, already confirmed, suppressed, rate limited per address, honeypot), so addresses cannot be enumerated. | `{"message": "If this address can be subscribed, you will receive a confirmation link. Check your spam folder too."}` |
| 403 | `Origin` is not allowlisted. | problem+json |
| 404 | Unknown `{slug}`. | problem+json |
| 422 | Missing or invalid `email`. | problem+json with `detail: "Enter a valid email address."` |
| 429 | Per-IP rate limit (10 requests per minute). Honor `Retry-After`. | problem+json |
| 400 | Malformed body. | problem+json |

Errors use `application/problem+json` (RFC 9457): `{"type","title","status","detail"?,"instance","request_id"}`. Show `detail` when present, otherwise a generic message.

A 202 does not mean subscribed: the person must open the confirmation email and confirm (double opt-in). Tell them to check spam.

## CORS

- The server answers preflight (`OPTIONS /subscribe/{slug}`) with 204, `Access-Control-Allow-Methods: POST, OPTIONS`, `Access-Control-Allow-Headers: Content-Type, Accept` and `Access-Control-Max-Age: 3600`.
- `Access-Control-Allow-Origin` echoes the request origin only when it is in the allowlist, with `Vary: Origin`. Credentials are never allowed (`Access-Control-Allow-Credentials` is never sent).
- The allowlist is the comma-separated env var `NEWSLETTER_ALLOWED_ORIGINS`, exact origins only (scheme, host and optional port; no path, no `*`). Production value: `https://hexascore.app`. Add `https://www.hexascore.app` if the site is served there too.
- A request from any other origin gets 403 without CORS headers, and no email is sent. Only `/subscribe/{slug}` is CORS-enabled; the dashboard and every other route stay same-origin.

## Example

```ts
const res = await fetch("https://email.hexascore.app/subscribe/hexascore", {
  method: "POST",
  headers: { "Content-Type": "application/json", Accept: "application/json" },
  body: JSON.stringify({ email }),
});
const body = await res.json();
if (res.status === 202) show(body.message);
else show(body.detail ?? "Something went wrong. Try again later.");
```
