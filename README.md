# AgentTrace

AgentTrace is a small Go OTLP trace receiver and query service for agent and GenAI telemetry.

It accepts standard OTLP traces, stores projects/traces/spans through GORM, keeps raw span/resource attributes as JSON, and indexes the official OTel GenAI semantic-convention fields.

## Quick Start

Run locally with SQLite:

```sh
make serve
```

Send OTLP/HTTP traces to:

```text
http://localhost:6006/v1/traces
```

Useful query endpoints:

```text
GET /healthz
GET /api/projects
GET /api/traces?project=heyyod&limit=100
GET /api/traces/{trace_id}
GET /api/spans?trace_id={trace_id}
```

Run tests:

```sh
make test
```

Run the optional Postgres ingest integration test against the local deps Postgres:

```sh
AGENTTRACE_POSTGRES_TEST_DSN='postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable' \
go test ./internal/otlp -run TestHTTPReceiverIngestsPostgresWhenConfigured -count=1 -v
```

## Configuration

Environment variables:

| Name | Default | Description |
| --- | --- | --- |
| `AGENTTRACE_HTTP_ADDR` | `:6006` | HTTP receiver and query API address. |
| `AGENTTRACE_GRPC_ADDR` | `:4317` | OTLP/gRPC trace receiver address. Set to `off` to disable. |
| `AGENTTRACE_DATABASE_DRIVER` | inferred | `sqlite` or `postgres`. |
| `AGENTTRACE_DATABASE_DSN` | `data/agenttrace.db` | SQLite path or Postgres DSN. |
| `DATABASE_URL` | unset | Postgres-compatible fallback DSN for production platforms. |
| `AGENTTRACE_DEFAULT_PROJECT` | `default` | Project used when OTLP resource data has no project name. |

Project name resolution uses standard OpenTelemetry resource attributes:

1. `service.name`
2. `service.namespace`
3. `AGENTTRACE_DEFAULT_PROJECT`

## Docker

Default one-command SQLite deployment:

```sh
docker compose up --build
```

Production-style Postgres deployment:

```sh
docker compose -f docker-compose.postgres.yml up --build
```

To reuse the local dependency Postgres from `/Users/ifuryst/projects/deps`, start that compose stack and run:

```sh
AGENTTRACE_DATABASE_DRIVER=postgres \
AGENTTRACE_DATABASE_DSN='postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable' \
go run ./cmd/agenttrace serve
```

## GenAI

AgentTrace preserves all OTLP attributes and indexes the current official GenAI semantic-convention fields:

- `gen_ai.operation.name`
- `gen_ai.provider.name`
- `gen_ai.request.model`
- `gen_ai.response.model`
- `gen_ai.usage.input_tokens`
- `gen_ai.usage.output_tokens`
- `gen_ai.conversation.id`

OpenInference and Phoenix-specific attributes are not part of the supported contract. If a sender includes them, they are preserved only as ordinary raw OTLP attributes.

## License

[MIT](LICENSE)
