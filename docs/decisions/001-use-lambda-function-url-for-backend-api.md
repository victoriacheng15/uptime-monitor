# ADR 001: Use Lambda Function URL for Backend API

- **Status:** 🔵 Accepted
- **Date:** 2026-04-30
- **Author:** Victoria Cheng

## Context and Problem Statement

Uptime Monitor needs a small HTTP surface for health checks, manual uptime checks, and reading stored monitor results. The backend is intentionally narrow: it has a health endpoint, a check trigger, and read endpoints for the latest and historical status data.

The project needed an AWS-native way to expose the Go Lambda without introducing more infrastructure than the MVP required. API Gateway was considered unnecessary for the first version because the service does not need advanced routing, request transformation, usage plans, custom authorizers, or stage-level API management.

The backend still needs to be easy to test and operate. Keeping routing inside the Go Lambda makes the Lambda the owner of the API contract and avoids splitting simple endpoint behavior across multiple AWS services.

## Decision Outcome

Use AWS Lambda with a Lambda Function URL as the backend API entrypoint.

The Go Lambda serves the HTTP routes directly:

- `GET /health`
- `POST /check`
- `GET /latest`
- `GET /history`

The Lambda Function URL provides the public HTTP access point, while the Go router handles method validation and response formatting inside the Lambda process. This keeps the deployment model small and makes the HTTP surface easy to reason about during the MVP stage.

## Consequences

### Positive

- **Simple deployment**: The backend can expose HTTP endpoints without adding API Gateway resources.
- **Lower operational overhead**: Fewer AWS components are needed for the MVP.
- **Clear backend ownership**: The Lambda owns routing, monitor execution, and S3 read/write behavior.
- **Fast iteration**: Endpoint behavior can be changed and tested in Go without coordinating API Gateway configuration.

### Negative

- **Limited edge controls**: Function URLs do not provide built-in usage plans, request throttling, stage management, or request and response transformation.
- **Simpler request handling**: The Lambda must handle method checks, route behavior, and response formatting in application code.

## Verification

- [x] Verified Lambda Function URL endpoints with `curl`.
- [x] `go test ./...` passed.
