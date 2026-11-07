# Vukem Email documentation

The public docs at <https://email.vukem.dev>: Astro 7.4 (static HTML), Tailwind CSS v4, Base UI and MDX. The shell, content pipeline, search, last-updated tracking and scripts mirror [vukem.dev](https://github.com/AlexGalhardo/vukem/tree/main/site); the content is specific to Vukem Email. `/` and `/docs` open `/docs/introduction`.

## Run

```bash
cd site
npm ci
npm run dev       # http://localhost:3000
npm run build     # static output in out/
npm run preview   # serve out/ locally
npm run lint
npm run check:mdx            # compile MDX, check components and /docs links
npm run last-updated         # refresh content/last-updated.json (needs full git history)
npm run check:last-updated
npm run test:scripts
```

## Add a page

1. Create `content/docs/<category>/<slug>.mdx` with `title` and `description` front matter. Do not repeat the title as a heading: the page template renders it.
2. Add `<slug>` to the category's `pages` in `content/navigation.ts`.
3. Run `npm run last-updated` and commit `content/last-updated.json` with the page.

Search, `llms.txt`, `llms-full.txt`, `sitemap.xml` and `search-index.json` pick the page up automatically.

## Deploy (Railway `docs` service)

The service builds the `Dockerfile` in this folder (root directory `site`): `npm ci` and `npm run build` in a Node 24 stage, then the `out/` directory is served by `nginx:1.31.0-alpine` on `$PORT` (default 3000, set by Railway). There is no Node process at runtime and no start command to configure. To run without Docker, build with `npm ci && npm run build` and serve `out/` with any static server that maps `/docs/x/` to `x/index.html`. Set `SITE_URL=https://email.vukem.dev` at build time if the canonical host differs.

The web app and queue worker use the root Dockerfile.
