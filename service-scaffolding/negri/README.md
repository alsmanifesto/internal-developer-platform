# negri

> Scaffolded by [Scaffold](https://github.com/ravon/scaffold) on 2026-03-25

## Service Configuration

| Field        | Value         |
|--------------|---------------|
| Name         | negri         |
| Service Type | job           |
| Workload     | app           |
| Stack        | go            |
| Pipeline     | gh-actions    |

## Description

This service was bootstrapped using Scaffold, the internal developer platform CLI.
It is a **app** workload running as a **job** using **go**.

## Running the Service

```bash
go run cmd/main.go
```

## Docker

```bash
docker build -t negri .
docker run --rm negri
```

## Infrastructure

Terraform configuration is available in the `terraform/` directory.

```bash
cd terraform
terraform init
terraform apply
```

> Note: Configured for LocalStack compatibility.
