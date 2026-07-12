# Infrastructure

Uptime Monitor uses AWS infrastructure to run a serverless website monitoring system. Lambda runs the Go backend, EventBridge starts hourly checks, S3 stores monitor output and Terraform state, and IAM connects the services with scoped permissions.

The infrastructure is managed with Terraform-compatible OpenTofu so AWS resources can be reviewed, deployed, and updated from source control.

## AWS Services

| AWS service | How this project uses it |
| :--- | :--- |
| AWS Lambda | Runs the Go uptime monitor backend. |
| Lambda Function URL | Provides HTTP access for health checks, manual checks, latest status, and history. |
| Amazon EventBridge | Triggers the monitor automatically every hour. |
| Amazon S3 | Stores monitor results as JSON files. |
| AWS IAM | Grants Lambda access to S3 and allows EventBridge to invoke the function. |

## Monitoring Data Flow

The monitor stores runtime data as JSON in S3. `latest.json` holds the current status for all monitored websites, while `history.json` keeps recent status history for each website.

```text
EventBridge          Lambda              Websites              S3
    │                  │                    │                  │
    │ (Hourly Trigger) │                    │                  │
    ├─────────────────►│                    │                  │
    │                  │                    │                  │
    │                  │ ── Send request ─► │                  │
    │                  │ ◄── Status/result ─│                  │
    │                  │                    │                  │
    │                  │ ────────────────────────────────────► │
    │                  │   Write latest.json & history.json    │
    │                  │                    │                  │
```

## Scheduled Checks

EventBridge runs the monitor every hour. This makes the project behave like a background uptime service instead of relying only on manual HTTP requests.

Manual requests still exist through the Lambda Function URL, so the same backend can be used for both scheduled checks and direct API access.

## Terraform State

Terraform state is stored in a separate S3 bucket from the monitoring data. This keeps AWS infrastructure state separate from runtime application output.
