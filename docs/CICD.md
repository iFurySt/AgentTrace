# CI/CD Guide

AgentTrace publishes its runtime container image to GitHub Container Registry.

## Workflows

### Docker Image

File: `.github/workflows/docker-image.yml`

Triggers:

- `push` to `main`: builds and pushes `ghcr.io/ifuryst/agenttrace:latest`, `:main`, and `:sha-<short-sha>`.
- `push` of tags matching `v*`: builds and pushes the matching version tag plus `:sha-<short-sha>`.
- `pull_request` to `main`: builds the image without pushing it.
- `workflow_dispatch`: allows a manual build and publish run.

The workflow currently builds a single `linux/amd64` image. Multi-architecture publishing can be added later once the CGO SQLite build is optimized for CI runtime.

## Release Artifact

Published image:

```text
ghcr.io/ifuryst/agenttrace
```

The same image supports both runtime modes:

- SQLite: run the container directly or use `docker-compose.yml`.
- Postgres: provide `AGENTTRACE_DATABASE_DRIVER=postgres` and `AGENTTRACE_DATABASE_DSN`, or use `docker-compose.postgres.yml`.

Both compose files use `ghcr.io/ifuryst/agenttrace:latest` with `pull_policy: always`, so `docker compose up` checks GHCR for the current published image instead of building a local image. The service is pinned to `platform: linux/amd64` until the workflow publishes a native arm64 image.

## Local Validation

Before changing CI/CD, run:

```sh
go test ./...
docker compose config
docker compose -f docker-compose.postgres.yml config
```

## Maintenance Notes

- GitHub Actions are pinned to immutable commit SHAs with comments recording the source major tag.
- The workflow grants only `contents: read`, `packages: write`, and `id-token: write`.
- The Docker build emits BuildKit SBOM and provenance attestations.
