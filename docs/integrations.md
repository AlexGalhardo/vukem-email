# Integrations

How HexaScore and EletroRadar feed subscribers into vukem-email, and how to run the whole thing locally with `compose.integration.yaml`.

Both integrations keep double opt-in: nobody becomes `confirmed` without opening the confirmation email and confirming with a POST.

| App | Path | Credential |
| --- | --- | --- |
| HexaScore (server) | `POST /api/v1/subscribers` ([api-tokens.md](api-tokens.md)) | Project token with `subscribers:write`, server side only |
| EletroRadar (static site) | `POST /subscribe/{slug}` ([subscribe-api.md](subscribe-api.md)) | None; Origin allowlist, honeypot and a per-IP rate limit (10 a minute) |

The public endpoint already exists and needs no token: it is the same one HexaScore's footer used. A token must never ship in a static site or browser bundle.

## Local stack

```sh
docker compose -f compose.integration.yaml up -d --build --wait
```

| Service | URL |
| --- | --- |
| vukem-email | `http://localhost:3010` (`VUKEM_PORT`) |
| Mailpit (mail catcher) | `http://localhost:8026` (`MAILPIT_UI_PORT`), SMTP `mailpit:1025` inside the network |
| PostgreSQL | internal only (`postgres:5432`, user, password and database `vukem`) |

`migrate` runs once before the app starts. The stack runs with `APP_ENV=development`, a public development `APP_KEY` and unencrypted SMTP; never reuse it outside a dev machine. Tear down with `docker compose -f compose.integration.yaml down -v`.

Browser origins allowed to call `/subscribe/{slug}` come from `NEWSLETTER_ALLOWED_ORIGINS` (default `http://localhost:8080,http://localhost:3000`, the EletroRadar static server and HexaScore dev server). Override it in the shell when your sites use other ports.

### Projects and tokens

The dashboard needs Google sign-in, which a local stack does not have. `integration:seed` creates the project if missing and prints a new token (it refuses to run with `APP_ENV=production`):

```sh
docker compose -f compose.integration.yaml run --rm --no-deps migrate integration:seed \
  --name HexaScore --slug hexascore --domain hexascore.test --from news@hexascore.test
docker compose -f compose.integration.yaml run --rm --no-deps migrate integration:seed \
  --name EletroRadar --slug eletroradar --domain eletroradar.test --from news@eletroradar.test
```

Each call prints one token (`vke_...`, shown once). EletroRadar needs no token; its project only has to exist.

### Confirmation links

Confirmation emails link to `https://<project domain>/confirm?token=...` (ROADMAP: audit item L4). Locally, open the message in Mailpit, copy the token and open `http://localhost:3010/confirm?token=<token>`, then press the confirm button (the page confirms with a POST, a GET never confirms).

## HexaScore: sign-up calls the API

Server side, after the account is created. Keep the token in an environment variable (`VUKEM_EMAIL_URL`, `VUKEM_EMAIL_TOKEN`) and never log it.

```ts
export async function subscribeToNewsletter(email: string): Promise<void> {
  const res = await fetch(`${process.env.VUKEM_EMAIL_URL}/api/v1/subscribers`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${process.env.VUKEM_EMAIL_TOKEN}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email }),
    signal: AbortSignal.timeout(5000),
  });
  // 202: pending (or already confirmed). 409: suppressed, 422: invalid address.
  if (!res.ok && res.status !== 409) {
    console.warn("newsletter subscribe failed", res.status);
  }
}
```

Do not fail the sign-up when this call fails; log and move on. Inside the compose network the base URL is `http://vukem-email:3000`; from the host, `http://localhost:3010`.

## EletroRadar: static form on the public endpoint

Plain HTML plus a few lines of JS. The browser sends `Origin`, so the site origin must be listed in `NEWSLETTER_ALLOWED_ORIGINS`. The `website` field is a honeypot: hidden from people, filled by bots, silently ignored.

```html
<form id="newsletter" action="https://email.example.com/subscribe/eletroradar" method="post">
  <label for="nl-email">Email</label>
  <input id="nl-email" name="email" type="email" required autocomplete="email" />
  <input name="website" type="text" tabindex="-1" autocomplete="off" aria-hidden="true"
    style="position:absolute;left:-9999px" />
  <button type="submit">Subscribe</button>
  <p id="nl-status" role="status"></p>
</form>
<script>
  const form = document.getElementById("newsletter");
  const status = document.getElementById("nl-status");
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const res = await fetch(form.action, {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({
        email: form.email.value,
        website: form.website.value,
      }),
    });
    const body = await res.json().catch(() => ({}));
    status.textContent =
      res.status === 202
        ? body.message
        : (body.detail ?? "Something went wrong. Try again later.");
  });
</script>
```

Without JavaScript the form still posts, but the browser then shows the JSON answer; keep the script. Locally, set `action` to `http://localhost:3010/subscribe/eletroradar` and serve the page from an origin in the allowlist (for example `python -m http.server 8080`).

## Smoke checklist

1. `up -d --build --wait`, then the two `integration:seed` commands.
2. `curl -X POST localhost:3010/api/v1/subscribers -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"email":"api@example.com"}'` returns 202.
3. `curl -X POST localhost:3010/subscribe/eletroradar -H "Origin: http://localhost:8080" -H "Content-Type: application/json" -d '{"email":"site@example.com"}'` returns 202; an Origin outside the allowlist returns 403.
4. Both confirmation emails show up in Mailpit (`http://localhost:8026`, or `GET /api/v1/messages`).
5. `POST /confirm` with the token confirms the subscriber.
6. `down -v` leaves no containers or volumes.
