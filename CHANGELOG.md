# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **docs:** Add `docs/audit-2026-11-14.md`, the gap analysis against a minimal listmonk.

### Changed

- **ci:** Dependabot no longer proposes TypeScript 7 for `site/`: typescript-eslint 8.x peers `typescript <6.1.0`, so `npm ci` failed with ERESOLVE.
- Require vukem v0.39.1, which sends Resend addresses as `email` or `Name <email>` (campaign and confirmation mail failed with "Invalid from" on v0.39.0).

## [0.4.0] - 2026-11-07

### Added

- **subscribe:** Add CORS allowlist and documented JSON contract for the public subscribe endpoint ([7b7db17](https://github.com/AlexGalhardo/vukem-email/commit/7b7db171df79fd36822b675d421d5909dd178f0a))
- **unsubscribe:** Add RFC 8058 one-click unsubscribe headers and endpoint ([f4699bc](https://github.com/AlexGalhardo/vukem-email/commit/f4699bcb479a07b524096bf2b8a04afd6597c2f3))
- **webhooks:** Verify Resend events and suppress hard bounces and complaints per project ([a185460](https://github.com/AlexGalhardo/vukem-email/commit/a185460222efc5a9dab68d38e98983be2d23301b))
- **emails:** Add confirmation fallback link and end-to-end subscriber flow test ([2a63706](https://github.com/AlexGalhardo/vukem-email/commit/2a6370661562a54707d956cf580ef5f6d17d9595))

### Changed

- Group dashboard status counts and add delivery and campaign indexes ([da41753](https://github.com/AlexGalhardo/vukem-email/commit/da41753c5fb15943372d824d91fcfb2fd6e34531))
- **site:** Migrate the documentation site from Next.js to Astro with the vukem.dev shell ([8daa20b](https://github.com/AlexGalhardo/vukem-email/commit/8daa20bb4f14dee15bcb4b11ab6547b3c776c544))

### Fixed

- **delivery:** Retry transient Resend 408/409 responses ([437d2c3](https://github.com/AlexGalhardo/vukem-email/commit/437d2c35d42f93fbf3c065d1ae2ae0d461a942b5))
- **webhooks:** Keep provider delivery events from being overwritten or dropped ([0c0dbad](https://github.com/AlexGalhardo/vukem-email/commit/0c0dbad2710a90217ecd2e238276bd2cd49c7b73))
- **security:** Run the campaign renderer with an allowlisted environment ([260dbb8](https://github.com/AlexGalhardo/vukem-email/commit/260dbb8504430489fde97853bbaf7accbcad3313))
- **security:** Cap confirmation emails per address and per project ([63fda1d](https://github.com/AlexGalhardo/vukem-email/commit/63fda1d6f043315cab2f8a30fa2a04bd183b21e8))
- **deploy:** Preserve the webhook secret and allowed origins in the Railway declaration ([e21bc8f](https://github.com/AlexGalhardo/vukem-email/commit/e21bc8fcee06e5b9beb944f3c95cfaa6ff2fd24f))
- **site:** Date doc pages by author date so rebase merges keep them valid ([991f2fc](https://github.com/AlexGalhardo/vukem-email/commit/991f2fcba54bfe55c3ebf27c6f7a2f5329e5512a))

## [0.3.0] - 2026-11-04

### Added

- Use English copy and canonical dashboard domain ([f310209](https://github.com/AlexGalhardo/vukem-email/commit/f3102094e37ca16c76d45ebdcf72625f2efd5117))

## [0.2.1] - 2026-11-04

### Fixed

- **ci:** Use supported CodeQL build modes ([1c00e13](https://github.com/AlexGalhardo/vukem-email/commit/1c00e130e5e0bafca4e76e77300297a3b6f80cd1))

## [0.2.0] - 2026-11-04

### Added

- Add documentation site and release infrastructure ([31de857](https://github.com/AlexGalhardo/vukem-email/commit/31de85726d7b2572b1142e5e85bfa6f02294a4ca))

## [0.1.0] - 2026-11-04

### Added

- Build multi-project newsletter manager with Vukem and React ([086a101](https://github.com/AlexGalhardo/vukem-email/commit/086a10189335995a726dc024ecddf26fd5fc11f8))

[Unreleased]: https://github.com/AlexGalhardo/vukem-email/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/AlexGalhardo/vukem-email/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/AlexGalhardo/vukem-email/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/AlexGalhardo/vukem-email/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/AlexGalhardo/vukem-email/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/AlexGalhardo/vukem-email/releases/tag/v0.1.0
