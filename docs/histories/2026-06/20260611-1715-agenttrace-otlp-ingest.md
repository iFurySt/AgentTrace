## [2026-06-11 17:15] | Task: AgentTrace OTLP ingest bootstrap

### Execution Context

- Agent ID: `codex`
- Base Model: `GPT-5`
- Runtime: `local shell`

### User Query

> Build the first version of a Go/Cobra AgentTrace service that can receive Phoenix-style OTLP trace data, persist it with GORM to SQLite/Postgres, support OTel GenAI semantic-convention fields, and provide Docker/Docker Compose deployment defaults.

### Changes Overview

- Area: Go service, OTLP ingestion, persistence, deployment, documentation.
- Key actions:
  - Added a Cobra `agenttrace` binary with `serve`, `migrate`, and `version` commands.
  - Added OTLP/HTTP `/v1/traces` ingestion compatible with Phoenix's protobuf endpoint, including gzip and deflate bodies.
  - Added OTLP/gRPC TraceService export support.
  - Added GORM-backed `projects`, `traces`, and `spans` persistence with SQLite default and Postgres support.
  - Added GenAI indexing and core OpenInference alias synthesis.
  - Added JSON query endpoints for health, projects, traces, trace detail, and spans.
  - Added Dockerfile plus SQLite and Postgres compose files.
  - Added real protobuf ingest tests for SQLite and optional Postgres persistence.

### Design Intent

The first version intentionally keeps Phoenix compatibility focused on the ingestion contract and trace/span data shape instead of cloning Phoenix's full schema. Raw OTLP data is preserved as JSON while high-value GenAI/OpenInference fields are indexed in columns so the service remains small, legible, and useful for HeyYo's current and future telemetry.

### Files Modified

- `cmd/agenttrace/main.go`
- `internal/cli/root.go`
- `internal/config/config.go`
- `internal/httpapi/api.go`
- `internal/otlp/decode.go`
- `internal/otlp/receiver.go`
- `internal/otlp/receiver_test.go`
- `internal/store/models.go`
- `internal/store/query.go`
- `internal/store/store.go`
- `Dockerfile`
- `docker-compose.yml`
- `docker-compose.postgres.yml`
- `Makefile`
- `README.md`
- `docs/ARCHITECTURE.md`
- `docs/QUALITY_SCORE.md`
