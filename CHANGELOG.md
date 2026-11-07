# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Resend webhooks (`POST /webhooks/resend`) with Svix signature verification, replay window and deduplication; hard bounces and complaints suppress the address per project and delivery states `delivered`, `bounced` and `complained`.
- RFC 8058 one-click unsubscribe: `List-Unsubscribe` and `List-Unsubscribe-Post` headers on campaign emails (Resend) and `POST /unsubscribe/one-click`.
- CORS allowlist (`NEWSLETTER_ALLOWED_ORIGINS`) for the public `POST /subscribe/{slug}` endpoint, with preflight support, no credentials, and a documented JSON contract (`docs/subscribe-api.md`).
- **security:** Vendor the Cloudflare security-audit skill and run it on pull requests.

### Changed

- Confirmation email shows a copyable fallback link so the plain-text part reads cleanly; an end-to-end test covers subscribe, confirm, campaign, unsubscribe and resubscribe.
- **deps:** Bump vukem to v0.38.0.

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

[Unreleased]: https://github.com/AlexGalhardo/vukem-email/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/AlexGalhardo/vukem-email/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/AlexGalhardo/vukem-email/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/AlexGalhardo/vukem-email/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/AlexGalhardo/vukem-email/releases/tag/v0.1.0
