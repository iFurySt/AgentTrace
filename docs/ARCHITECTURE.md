# Architecture

AgentTrace is a single Go binary that receives OTLP trace data and stores it for lightweight agent and GenAI trace querying.

## Runtime Topology

- `agenttrace serve` starts one HTTP server on `AGENTTRACE_HTTP_ADDR`.
- The HTTP server exposes Phoenix-compatible OTLP ingestion at `/v1/traces`.
- The same HTTP server exposes JSON query APIs under `/api/*`.
- `agenttrace serve` also starts an OTLP/gRPC trace receiver on `AGENTTRACE_GRPC_ADDR` unless that value is `off`.
- SQLite is the default local persistence layer; Postgres is supported through GORM for production.

## Package Layout

- `cmd/agenttrace`: binary entry point.
- `internal/cli`: Cobra command wiring and process lifecycle.
- `internal/config`: environment and default configuration.
- `internal/otlp`: OTLP protobuf decoding, HTTP receiver, and gRPC TraceService receiver.
- `internal/store`: GORM models, migrations, ingestion, and query methods.
- `internal/httpapi`: health and JSON query routes.
- `docs/`: repository knowledge base and change history.

## Data Model

AgentTrace uses three core tables:

- `projects`: named project buckets.
- `traces`: trace-level summary rows keyed by project and OTLP trace ID.
- `spans`: span rows keyed by OTLP trace ID and span ID.

Span resource attributes, span attributes, and events are stored as JSON. High-value GenAI/OpenInference fields are duplicated into columns for filtering and summaries:

- GenAI operation, provider, request model, response model.
- OpenInference span kind.
- Input and output token counts.
- Input and output values when present as strings.

## OTLP Compatibility

The first ingestion target is Phoenix's practical OTLP/HTTP behavior:

- `POST /v1/traces`
- `Content-Type: application/x-protobuf`
- optional `Content-Encoding: gzip` or `deflate`
- `x-project-name` header override

OTLP/gRPC TraceService export is also supported for standard collectors and SDK exporters.

## GenAI/OpenInference Mapping

The receiver keeps all original attributes. When a span only provides OTel `gen_ai.*` semantic-convention fields, AgentTrace synthesizes the core OpenInference aliases needed by Phoenix-style workflows:

- span kind from `gen_ai.operation.name`
- `llm.provider` from `gen_ai.provider.name` or `gen_ai.system`
- `llm.model_name` from `gen_ai.request.model`
- `llm.token_count.*` from `gen_ai.usage.*`
- `openinference.session.id` from `gen_ai.conversation.id`
