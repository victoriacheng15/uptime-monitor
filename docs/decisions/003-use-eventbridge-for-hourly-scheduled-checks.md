# ADR 003: Use EventBridge for Hourly Scheduled Checks

- **Status:** 🔵 Accepted
- **Date:** 2026-05-01
- **Author:** Victoria Cheng

## Context and Problem Statement

Uptime Monitor needs to run checks automatically without relying on a user or external client to call the API. A monitoring service should continue collecting status data even when nobody manually triggers `/check`.

The project also needs to stay serverless. Running a cron process on a VM, container host, or always-on worker would add infrastructure that does not match the rest of the design. The scheduled trigger should be managed by AWS and should invoke the existing Lambda backend directly.

The first version only needs hourly checks. Exact alignment to the top of the hour is not critical for the MVP because the goal is periodic uptime visibility, not second-level scheduling precision.

## Decision Outcome

Use Amazon EventBridge to invoke the Lambda every hour.

The schedule uses:

```hcl
rate(1 hour)
```

EventBridge invokes the same Lambda backend used by the HTTP API. The Lambda detects scheduled invocations and runs the check path internally.

This keeps scheduled checks and manual checks aligned around the same monitor behavior. The HTTP API remains useful for operator-triggered checks and reads, while EventBridge owns the normal background cadence.

## Consequences

### Positive

- **Serverless scheduling**: No worker server or cron host is required.
- **Consistent check behavior**: Scheduled checks reuse the same monitor path as manual checks.
- **MVP-friendly timing**: Hourly checks are enough for the current personal-site monitoring scope.
- **AWS-native trigger**: EventBridge integrates directly with Lambda permissions and keeps scheduling inside the AWS infrastructure definition.
- **Low maintenance**: There is no separate scheduler process to patch, monitor, or restart.

### Negative

- **Not aligned to the top of the hour**: `rate(1 hour)` runs hourly from when the rule becomes active, not necessarily at `xx:00`.
- **Limited schedule precision**: If exact wall-clock timing becomes important, the schedule should move to a cron expression.
- **Lambda event handling complexity**: The Lambda entrypoint must distinguish scheduled events from HTTP Function URL requests.
- **No missed-run recovery**: The MVP does not backfill missed checks if an invocation fails.

## Verification

- [x] Verified the EventBridge rule exists with `rate(1 hour)` and permission to invoke Lambda.
- [x] `go test ./...` passed.
