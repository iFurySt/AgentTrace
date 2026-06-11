# AgentTrace

AgentTrace is a small OTLP trace receiver for agent and GenAI telemetry.

It stores projects, traces, and spans in SQLite or Postgres, preserves raw OTLP attributes, and indexes the official OpenTelemetry `gen_ai.*` semantic-convention fields.

## Quick Start

Run with Docker and SQLite:

```sh
docker compose up
```

Run with Docker and Postgres:

```sh
docker compose -f docker-compose.postgres.yml up
```

The compose files pull the published image:

```text
ghcr.io/ifuryst/agenttrace:latest
```

OTLP endpoints:

```text
HTTP: http://localhost:16006/v1/traces
gRPC: localhost:14317
```

Useful API endpoints:

```text
GET http://localhost:16006/healthz
GET http://localhost:16006/api/projects
GET http://localhost:16006/api/traces?project=default&limit=100
```

Local development:

```sh
make serve
make test
```

## Configuration

| Name | Default | Description |
| --- | --- | --- |
| `AGENTTRACE_HTTP_ADDR` | `:6006` | HTTP receiver and query API address. |
| `AGENTTRACE_GRPC_ADDR` | `:4317` | OTLP/gRPC trace receiver address. Set to `off` to disable. |
| `AGENTTRACE_DATABASE_DRIVER` | inferred | `sqlite` or `postgres`. |
| `AGENTTRACE_DATABASE_DSN` | `data/agenttrace.db` | SQLite path or Postgres DSN. |
| `DATABASE_URL` | unset | Postgres-compatible fallback DSN for production platforms. |
| `AGENTTRACE_DEFAULT_PROJECT` | `default` | Project used when OTLP resource data has no project name. |

## GenAI

Supported indexed fields:

- `gen_ai.operation.name`
- `gen_ai.provider.name`
- `gen_ai.request.model`
- `gen_ai.response.model`
- `gen_ai.usage.input_tokens`
- `gen_ai.usage.output_tokens`
- `gen_ai.conversation.id`

OpenInference and Phoenix-specific attributes are not part of the supported contract. If a sender includes them, they are preserved only as ordinary raw OTLP attributes.

## Docs

- [Architecture](docs/ARCHITECTURE.md)
- [CI/CD](docs/CICD.md)
- [Security](docs/SECURITY.md)
- [Reliability](docs/RELIABILITY.md)

## License

[MIT](LICENSE)
