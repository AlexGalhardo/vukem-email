# Vukem Email documentation

This Next.js site uses the visual shell and navigation architecture of [vukem.dev](https://github.com/AlexGalhardo/vukem/tree/main/site), with content specific to Vukem Email. `/` redirects to `/docs/introduction`.

```bash
cd site
npm ci
npm run build
npm run start
```

Set `NEXT_PUBLIC_SITE_URL=https://email.vukem.dev` in production. Deploy this folder as a separate Railway service; set its root directory to `site` and start with `npm run start`. The web app and queue worker use the root Dockerfile.
