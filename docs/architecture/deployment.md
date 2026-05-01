# Deployment

Uptime Monitor deploys to AWS through GitHub Actions and OpenTofu.

## Workflow

The repository uses GitHub Actions for both CI and CD:

| Workflow | Purpose |
| :--- | :--- |
| `.github/workflows/ci.yaml` | Runs Go checks and tests for backend changes. |
| `.github/workflows/infra.yaml` | Runs Terraform formatting and validation for infrastructure changes. |
| `.github/workflows/deploy.yaml` | Builds the Lambda package and applies AWS infrastructure changes. |

## CI

CI protects changes before deployment.

| Area | Check | Purpose |
| :--- | :--- | :--- |
| Go | `test -z "$(gofmt -l ./cmd ./internal)"` | Verifies Go formatting. |
| Go | `make test` | Runs the Go test suite. |
| Infrastructure | `tofu fmt -check -recursive infra` | Verifies Terraform formatting. |
| Infrastructure | `tofu init -backend=false` | Initializes Terraform without touching remote state. |
| Infrastructure | `tofu validate` | Validates Terraform configuration. |

## CD

CD deploys the Lambda and AWS infrastructure from `main`.

| Step | Command | Purpose |
| :--- | :--- | :--- |
| Test | `make test` | Runs the Go test suite before deployment. |
| Package Lambda | `make lambda-package` | Builds the Go Lambda binary and creates `bin/lambda.zip`. |
| Initialize Terraform | `tofu init` | Initializes the remote S3 backend and providers. |
| Apply Infrastructure | `tofu apply -auto-approve` | Deploys AWS infrastructure and updates the Lambda package. |

The Lambda package is generated during the workflow run. It is not committed to the repository.

## AWS Authentication

GitHub Actions authenticates to AWS using OIDC.

Required repository secret:

| Secret | Purpose |
| :--- | :--- |
| `AWS_ROLE_TO_ASSUME` | IAM role ARN assumed by GitHub Actions. |

The role must trust:

```text
repo:victoriacheng15/uptime-monitor:*
```

## Runtime Configuration

Required repository secret:

| Secret | Purpose |
| :--- | :--- |
| `MONITOR_TARGETS` | Comma-separated list of URLs checked by the monitor. |

The workflow maps this secret into OpenTofu:

```yaml
TF_VAR_monitor_targets: ${{ secrets.MONITOR_TARGETS }}
```

OpenTofu then sets the Lambda environment variable:

```text
MONITOR_TARGETS
```
