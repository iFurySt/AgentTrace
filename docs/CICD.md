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

The workflow builds a single multi-architecture image for:

- `linux/amd64`
- `linux/arm64`

## Release Artifact

Published image:

```text
ghcr.io/ifuryst/agenttrace
```

The same image supports both runtime modes:

- SQLite: run the container directly or use `docker-compose.yml`.
- Postgres: provide `AGENTTRACE_DATABASE_DRIVER=postgres` and `AGENTTRACE_DATABASE_DSN`, or use `docker-compose.postgres.yml`.

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
