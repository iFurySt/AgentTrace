# Supply Chain Security

This document records AgentTrace's current supply-chain posture and the controls to add next.

## Current State

AgentTrace publishes a Docker image to GitHub Container Registry through `.github/workflows/docker-image.yml`.

Current controls:

- Do not commit secrets, tokens, or local private configuration.
- Keep Go dependency manifests committed through `go.mod` and `go.sum`.
- Pin GitHub Actions to immutable commit SHAs instead of floating tags.
- Build and push a single `linux/amd64` image.
- Emit BuildKit SBOM and provenance attestations with the pushed image.

## Tooling To Add Later

- `actions/dependency-review-action`: reviews pull-request dependency changes.
- `google/osv-scanner-action`: scans for known open source vulnerabilities.
- `actions/attest-build-provenance`: generates signed build provenance for release artifacts.

## Limits And Assumptions

- Dependency Review is available for public repositories and private repositories with GitHub Advanced Security.
- There is no automated vulnerability scan or dependency-review gate right now.
- OpenSSF Scorecard is intentionally not enabled by default because a new repository may not yet have real branch protection, release history, or SAST posture to score. Add it after repository rules are configured.

## What To Do Next

- Add ecosystem-specific vulnerability scanning for Go modules and container images.
- Add `linux/arm64` publishing after optimizing the CGO SQLite build for CI runtime.
- Make release tags intentional and document versioning rules before the first public release.
- Gate production deployment on release artifact provenance verification when possible.
- Consider verifying attestations in the deployment environment or cluster admission layer.
