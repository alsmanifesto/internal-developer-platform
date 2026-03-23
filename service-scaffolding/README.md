# Scaffold

Scaffold is a developer platform CLI that bootstraps services across application, data, and ML workloads. It scaffolds a fully functional project structure — including source code, Dockerfile, Terraform, CI/CD pipeline, and README — through either a guided TUI experience or standard CLI flags.

---

## Table of Contents

- [Requirements](#requirements)
- [Dependencies](#dependencies)
- [Installation](#installation)
- [Commands](#commands)
  - [create](#ravon-create)
  - [delete](#ravon-delete)
- [Available Values](#available-values)
- [Generated Structure](#generated-structure)
- [Metadata](#metadata)

---

## Requirements

- Go 1.21 or later
- Terraform CLI (optional — required only for infrastructure operations)
- Docker + LocalStack (optional — required to simulate AWS infrastructure locally)

---

## Dependencies

| Package | Version | Purpose |
|---|---|---|
| `github.com/spf13/cobra` | v1.8.0 | CLI framework |
| `github.com/charmbracelet/bubbletea` | v0.25.0 | TUI framework |
| `github.com/charmbracelet/lipgloss` | v0.9.1 | TUI styling |

---

## Installation

### Build from source

```bash
git clone https://github.com/ravon/scaffold.git
cd service-scaffolding
go mod download
go build -o scaffold .
```

### Run without installing

```bash
go run . create
```

### Install to PATH

```bash
go install .
```

---

## Commands

### `scaffold create`

Scaffolds a new service. Supports two execution modes.

#### Interactive (TUI)

Launches a step-by-step guided terminal interface. Navigate with arrow keys, confirm with Enter.

```bash
./scaffold create
```

#### Non-interactive (flags)

All five flags are required when using this mode. Useful for scripting, CI/CD pipelines, or automation.

```bash
./scaffold create \
  --name <service-name> \
  --service-type <type> \
  --workload <workload> \
  --stack <stack> \
  --pipeline <pipeline>
```

**Example:**

```bash
./scaffold create \
  --name payment-api \
  --service-type api \
  --workload app \
  --stack go \
  --pipeline gh-actions
```

**Flags:**

| Flag | Description | Required |
|---|---|---|
| `--name` | Unique service name | Yes |
| `--service-type` | How the service runs | Yes |
| `--workload` | Primary workload type | Yes |
| `--stack` | Technology stack | Yes |
| `--pipeline` | CI/CD pipeline tool | Yes |

---

### `scaffold delete`

Deletes an existing service via an interactive TUI list. Executes in order:

1. Runs `terraform destroy` if a `terraform/` directory exists
2. Deletes the service folder from disk
3. Removes the entry from `.scaffold/services.json`

```bash
./scaffold delete
```

---

## Available Values

### `--service-type`

| Value | Description |
|---|---|
| `api` | HTTP/gRPC service that handles requests |
| `worker` | Long-running background process |
| `job` | One-off or scheduled batch execution |

### `--workload`

| Value | Description |
|---|---|
| `app` | General application workload |
| `data` | Data processing or pipeline workload |
| `ml` | Machine learning or model serving workload |

### `--stack`

| Value | Scaffolded files | Base image |
|---|---|---|
| `go` | `cmd/main.go` — HTTP hello world server | `golang:1.21-alpine` (multi-stage) |
| `python` | `main.py` — runnable script | `python:3.11-slim` |
| `spark` | `main.py` — PySpark transformation job | `python:3.11-slim` + pyspark |
| `kafka` | `producer.py` + `consumer.py` | `python:3.11-slim` + kafka-python |

### `--pipeline`

| Value | Generated file | Description |
|---|---|---|
| `gh-actions` | `.github/workflows/ci.yml` | GitHub Actions build & test workflow |
| `concourse` | `pipeline.yml` | Concourse CI pipeline |
| `airflow` | `dags/dag.py` | Apache Airflow DAG (extract → transform → load) |
| `mlflow` | `mlflow_pipeline.py` | MLflow experiment pipeline placeholder |

---

## Generated Structure

For a service named `payment-api` with stack `go` and pipeline `gh-actions`:

```
payment-api/
├── cmd/
│   └── main.go              # HTTP hello world server
├── terraform/
│   └── main.tf              # S3 bucket (LocalStack-compatible)
├── .github/
│   └── workflows/
│       └── ci.yml           # CI pipeline
├── Dockerfile               # Multi-stage build
└── README.md                # Service documentation
```

For `spark` or `kafka` stacks, `cmd/` is replaced by Python source files at the root level.

---

## Metadata

Ravon tracks all created services in `.scaffold/services.json`, created automatically on first use.

```json
{
  "services": [
    {
      "name": "payment-api",
      "service_type": "api",
      "workload": "app",
      "stack": "go",
      "pipeline": "gh-actions",
      "path": "payment-api",
      "created_at": "2026-03-21T10:00:00Z"
    }
  ]
}
```

This file is used by `scaffold delete` to list and manage existing services. Services are listed sorted by creation date, newest first.

---

## Infrastructure (LocalStack)

All generated Terraform configurations target LocalStack. The provider is pre-configured with path-style S3 URLs (`s3_use_path_style = true`) to avoid virtual-hosted routing issues.

### Start LocalStack

```bash
docker run --rm -d \
  -p 4566:4566 \
  --name localstack \
  localstack/localstack
```

Or with Docker Compose:

```yaml
services:
  localstack:
    image: localstack/localstack
    ports:
      - "4566:4566"
```

```bash
docker compose up -d
```

### Apply Terraform

```bash
cd <service-name>/terraform
terraform init
terraform apply
```

### Verify the bucket was created

```bash
aws --endpoint-url=http://localhost:4566 s3 ls
```

> The AWS CLI will accept any value for `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` when pointing to LocalStack.

```bash
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1
```
