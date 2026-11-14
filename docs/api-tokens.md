# API tokens

Server-to-server API for one project (a project is a list). Tokens are created and revoked by the admin in the dashboard under "API tokens", or with `POST /api/projects/{id}/tokens` and `POST /api/projects/{id}/tokens/{token}/revoke` (admin session).

## Tokens

- Format `vke_<id>.<secret>` (vukem `auth.Tokens`). Only `SHA-256(secret)` is stored in `access_tokens`; the plain value is shown once.
- A token belongs to one project (`access_tokens.user_id` holds the project id) and carries explicit abilities. There is no wildcard.
- Tokens do not expire; revoke them. At most 20 active tokens per project.

| Ability | Allows |
| --- | --- |
| `subscribers:write` | `POST /api/v1/subscribers`, `PATCH /api/v1/subscribers/{id}` |
| `campaigns:send` | `POST /api/v1/campaigns/{id}/send` |

Send `Authorization: Bearer <token>`. Missing, wrong or revoked: `401` with `WWW-Authenticate: Bearer realm="api"`. Missing ability: `403`. Another project's subscriber or campaign: `404`.

Rate limits: 300 requests per minute per IP, then 60 per minute per token (`429`). The counters are in process memory.

## Endpoints

All bodies are JSON (8 KiB limit). The routes need no session or CSRF token.

### Create or re-request a subscriber

```http
POST /api/v1/subscribers
{"email": "ana@example.com"}
```

Same double opt-in as the public form: the subscriber is `pending` and receives the confirmation email; consent comes from the link, the API cannot mark an address `confirmed`. Confirmed subscribers are returned unchanged.

| Status | Meaning |
| --- | --- |
| 202 | `{"id", "email", "status"}` (`pending` or `confirmed`). |
| 409 | The address is suppressed for the project (bounce or complaint). |
| 422 | Invalid email. |
| 429 | Per-project confirmation limit (500 an hour) reached. |

### Unsubscribe

```http
PATCH /api/v1/subscribers/{id}
{"status": "unsubscribed"}
```

Only `pending` or `confirmed` subscribers change; a `suppressed` one keeps its status. Other statuses answer `422`.

### Send a campaign

```http
POST /api/v1/campaigns/{id}/send
```

Queues a `draft` campaign of the token's project for its confirmed subscribers, as the dashboard button does. `202` when queued, `409` if it is already queued or sent, `422` without confirmed subscribers.
