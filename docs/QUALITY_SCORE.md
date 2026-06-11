# Quality Score

Track quality by product area and architectural layer so agents can prioritize the weakest parts of the system.

## Suggested Scale

- `A`: strong coverage, stable behavior, clear docs, low operational risk.
- `B`: acceptable but still has known gaps.
- `C`: works but needs targeted hardening.
- `D`: fragile or underspecified.

## Initial Template

| Area | Score | Why | Next Step |
| --- | --- | --- | --- |
| Product surface | C | The first OTLP ingest and JSON query surface exists, but filtering and UI are still minimal. | Add richer trace/span filters and a compact web explorer. |
| Architecture docs | B | Runtime topology, package boundaries, storage model, and OTLP compatibility are documented. | Keep docs aligned as query APIs evolve. |
| Testing | B | Real OTLP/HTTP protobuf ingest tests cover gzip, official GenAI attributes, SQLite persistence, and optional Postgres persistence. | Add gRPC receiver tests and broader query API coverage. |
| Observability | C | The service logs receiver startup and ingest counts and exposes `/healthz`. | Add structured request logs and basic Prometheus metrics. |
| Security | C | No auth is enabled; defaults are intended for local/private deployment. | Add deployment guidance for network boundaries and optional API auth before public exposure. |
