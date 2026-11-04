# Security policy

Vukem Email receives private vulnerability reports through [GitHub Security Advisories](https://github.com/AlexGalhardo/vukem-email/security/advisories/new). Do not open a public issue or send subscriber data in a report.

## Supported versions

The latest tagged release and the current main branch receive security fixes. Older pre-1.0 releases should be upgraded to the latest release.

## Reporting

Include the affected version or commit, impact, reproduction steps and a suggested fix if available. The maintainer aims to acknowledge a report within three business days and provide an initial assessment within seven days. Critical issues take priority.

Disclosure is coordinated with the reporter. Please keep details private until an advisory and patched release are published.

## Security boundaries

Production administration uses Google OIDC only and requires the verified owner email. Confirmation tokens are hashed, expire and work once. Unsubscribe links are signed. GET requests do not change subscription state. Campaign JavaScript runs only in QuickJS WebAssembly; the serialized tree is validated and rendered HTML is sanitized. Public props never include subscriber addresses or credentials. Queue workers recheck subscription state before delivery.

Release archives include checksums, SPDX SBOMs, a keyless cosign bundle and GitHub provenance from the release workflow. The Go binary needs Node 24 and production npm dependencies for dynamic campaign rendering; use the Dockerfile for production deployment.
