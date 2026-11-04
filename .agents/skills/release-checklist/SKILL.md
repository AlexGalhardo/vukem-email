---
name: release-checklist
description: Verify Vukem Email releases before tagging.
---

# Release checklist

1. Confirm SemVer impact and update `CHANGELOG.md` from `[Unreleased]`.
2. Run `npm run check`, `npm run build`, `go vet ./...`, `go test -race ./...` and the documentation build.
3. Confirm generated email assets are committed and the Docker image builds.
4. Tag `vX.Y.Z` only after checks pass. The release workflow attaches binaries, checksums, SBOMs and a keyless signature.
