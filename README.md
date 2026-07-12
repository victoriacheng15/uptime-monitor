# Uptime Monitor

Uptime Monitor is a serverless AWS project built with Go, Lambda, EventBridge, S3, OpenTofu (Terraform-compatible), and GitHub Actions.

It monitors multiple websites on an hourly schedule, stores the latest status snapshot and recent history in S3, and exposes lightweight HTTP endpoints through a Lambda Function URL for manual checks and dashboard data.

[Full Documentation](./docs/README.md)

---

## Case Studies

| Case Study | Problem | How it was diagnosed | Result |
| :--- | :--- | :--- | :--- |
| [Why S3 JSON storage](./docs/decisions/002-store-uptime-results-as-s3-json-files.md) | The monitor needed persistent latest status and short per-site history without adding database infrastructure. | S3 fit because the MVP only needed small structured writes, current-state reads, and short history reads; it did not need relational queries, indexes, joins, or high-write throughput. | Results are stored as `latest.json` and `history.json`, keeping storage low-cost, frontend-friendly, and database-free for the MVP. |
| [Why EventBridge scheduling](./docs/decisions/003-use-eventbridge-for-hourly-scheduled-checks.md) | The monitor needed to run automatically without relying on a user or external client to call `/check`. | EventBridge fit because it keeps the scheduler serverless, avoids a VM cron job or always-on worker, and supports the hourly cadence needed for the first version. | EventBridge invokes the Lambda every hour with `rate(1 hour)`, reusing the same monitor path as manual checks. |

---

## Architecture

The system runs as a small serverless monitoring loop: Terraform provisions AWS resources, EventBridge triggers scheduled Lambda checks, and S3 stores the latest and historical status data.

| Path | Use case | Flow |
| :--- | :--- | :--- |
| Scheduled monitoring | Check configured websites every hour | EventBridge -> Lambda -> monitored sites -> S3 |
| Manual operation | Trigger checks or read monitor data on demand | Lambda Function URL -> `/check`, `/latest`, `/history`, `/health` |
| Runtime storage | Keep latest and recent historical status data | Lambda -> S3 `latest.json` and `history.json` |
| Infrastructure deployment | Keep AWS resources managed from source control | GitHub Actions -> Terraform -> AWS resources |

Deployment flow:

```text
                [1] DEPLOYMENT & CI/CD PIPELINE
                ===============================
                       Push to main / PR
                               │
                               ▼
                  ┌───────────────────────┐
                  │  GitHub CI/CD Runner  │
                  └─────┬───────────┬─────┘
                        │           │
      ┌─────────────────┘           └─────────────────┐
      ▼ (Deploy Backend)                              ▼ (Deploy SSG Page)
┌─────────────────────────────┐                 ┌─────────────────────────────┐
│ 1. Run Go Unit/E2E Tests    │                 │ 1. Fetch telemetry logs     │
│ 2. Package Lambda Bootstrap │                 │    from Lambda API Endpoint │
│ 3. Run Terraform Apply (IaC)│                 │ 2. Compile Tailwind CSS     │
└──────────────┬──────────────┘                 │ 3. Generate static HTML     │
               │                                └──────────────┬──────────────┘
    (Reads / Writes tfstate)                                   │
               │                                               │
               ▼                                               │
┌─────────────────────────────┐                                │ (Publishes static
│     S3 Backend Bucket       │                                │  dist assets)
│   (uptime-monitor-tfstate)  │                                │
└──────────────┬──────────────┘                                ▼
               │                                ┌─────────────────────────────┐
               │ (Configures Lambda,            │    GitHub Pages Website     │
               │  S3, & EventBridge)            └─────────────────────────────┘
               ▼
┌─────────────────────────────┐
│   AWS Lambda & S3 Bucket    │
└─────────────────────────────┘
```

Application flow:

```text
              [2] APPLICATION FLOW
              ====================
              ┌───────────────────────┐
              │      EventBridge      │
              └───────────┬───────────┘
                          │ (Hourly Trigger)
                          ▼
              ┌───────────────────────┐
   Client ───►│      AWS Lambda       │
(HTTP Req)    │    (Go API Router)    │
              └────┬──────┬──────┬────┘
                    │      │      │
     ┌──────────────┘      │      └──────────────┐
     ▼ (/latest)           ▼ (/check)            ▼ (/history)
┌──────────┐         ┌───────────┐         ┌──────────┐
│ S3 Read  │         │ Run check │         │ S3 Read  │
│  latest  │         │ against   │         │ history  │
│  .json   │         │ websites  │         │  .json   │
└──────────┘         └─────┬─────┘         └──────────┘
                          │ (Write outputs)
                          ▼
                    ┌───────────┐
                    │ S3 Bucket │
                    │   Store   │
                    └───────────┘
```

---

## Tech Stack

| Layer | Tools |
| :--- | :--- |
| Language | Go |
| Compute | AWS Lambda |
| Scheduling | Amazon EventBridge |
| Storage | Amazon S3 |
| Infrastructure | Terraform |
| Testing | Go `testing` package, table-driven tests |
| CI/CD | GitHub Actions |

---

## API Endpoints

| Endpoint | Method | Purpose |
| :--- | :--- | :--- |
| `/health` | `GET` | Returns Lambda service health |
| `/check` | `POST` | Runs uptime checks and writes results to S3 |
| `/latest` | `GET` | Returns the latest status snapshot |
| `/history` | `GET` | Returns recent status history grouped by URL |

---

## Documentation

- [Documentation Hub](./docs/README.md)
- [Architecture](./docs/architecture/README.md)
- [Decisions](./docs/decisions/README.md)
- [Incidents](./docs/incidents/README.md)

---

## Local Setup

Build and test the Go backend:

```bash
make test
make lambda-package
```

Deploy infrastructure locally:

```bash
cd infra
tofu init
tofu apply
```

Trigger a manual check:

```bash
curl -X POST "$(tofu output -raw lambda_function_url)check"
```

Read monitor data:

```bash
curl "$(tofu output -raw lambda_function_url)latest"
curl "$(tofu output -raw lambda_function_url)history"
```
