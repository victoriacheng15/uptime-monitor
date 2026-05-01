# ADR 002: Store Uptime Results as S3 JSON Files

- **Status:** 🔵 Accepted
- **Date:** 2026-05-01
- **Author:** Victoria Cheng

## Context and Problem Statement

Uptime Monitor needs persistent status data for multiple websites. The system has two primary read needs: a current fleet snapshot and a short per-site history that can support simple trend display.

The MVP does not need relational queries, secondary indexes, joins, high-write throughput, or user-managed records. A database would add more operational surface than the project currently needs. The monitor writes a small amount of structured data on each check, and consumers only need to read the latest status and recent history.

The storage design also needs to remain simple for a static or lightweight frontend. JSON files provide a direct handoff between the backend monitor and any dashboard or API consumer.

## Decision Outcome

Store monitor output in Amazon S3 as JSON files.

The runtime data model uses two objects:

- `latest.json`: Current status snapshot for all monitored websites.
- `history.json`: Recent status history grouped by URL and capped at five results per website.

`latest.json` is optimized for the dashboard's current-state view. `history.json` is optimized for small per-site history without requiring a database table. The backend updates both objects after a successful check run.

## Consequences

### Positive

- **Low infrastructure cost**: S3 is sufficient for small JSON status files.
- **Simple read model**: The backend and frontend can consume JSON directly.
- **Clear data separation**: Latest state and historical state are stored separately for different access patterns.
- **No database dependency**: The MVP can persist monitor data without provisioning DynamoDB, RDS, or another database service.
- **Frontend-friendly output**: Stored JSON can map directly to dashboard data without a translation layer.

### Negative

- **Write coordination limits**: This design assumes one monitor writer and does not handle concurrent write contention.
- **Object rewrite behavior**: Updating history requires reading and rewriting the `history.json` object.
- **Growth limit**: If retention grows beyond a short history window, the storage model may need to move to partitioned objects or a database.

## Verification

- [x] Verified `latest.json` and `history.json` were written to S3 after calling `/check`.
- [x] `go test ./...` passed.
