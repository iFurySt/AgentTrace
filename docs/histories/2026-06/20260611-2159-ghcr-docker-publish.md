## [2026-06-11 21:59] | Task: GHCR Docker image publishing

### Execution Context

- Agent ID: `codex`
- Base Model: `GPT-5`
- Runtime: `local shell`

### User Query

> Make the GitHub repository public and add CI/CD that builds the AgentTrace Docker image to GHCR, then track the run until the image is published.

### Changes Overview

- Area: repository visibility, CI/CD, Docker publishing, documentation.
- Key actions:
  - Changed the GitHub repository visibility from private to public.
  - Added a GitHub Actions workflow that builds and publishes `ghcr.io/ifuryst/agenttrace`.
  - Configured multi-architecture Docker builds for `linux/amd64` and `linux/arm64`.
  - Added GHCR image tags for `latest`, branch refs, git tags, and commit SHA tags.
  - Documented the image, workflow triggers, release artifact, and supply-chain posture.

### Design Intent

The workflow publishes one runtime image that works for both SQLite and Postgres deployments. Runtime mode stays configuration-driven through environment variables and compose files instead of producing separate images.

### Files Modified

- `.github/workflows/docker-image.yml`
- `README.md`
- `docs/CICD.md`
- `docs/SUPPLY_CHAIN_SECURITY.md`
- `docs/histories/2026-06/20260611-2159-ghcr-docker-publish.md`
