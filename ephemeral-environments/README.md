# ephemeral-env

A CLI tool that spins up ephemeral (preview) environments for services scaffolded by [scaffold](../service-scaffolding). It detects the project stack, generates a `docker-compose.yml` with Traefik routing, and brings the environment up via the Docker daemon.

Designed to run inside CI/CD pipelines as a Docker container with the host Docker socket mounted.

---

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Commands](#commands)
- [Stack Detection](#stack-detection)
- [Generated docker-compose](#generated-docker-compose)
- [CI/CD Integration](#cicd-integration)
- [Project Structure](#project-structure)

---

## Requirements

- Go 1.22 or later (to build from source)
- Docker with Compose v2 plugin

---

## Installation

### Build from source

```bash
git clone https://github.com/alsmanifesto/internal-developer-platform
cd internal-developer-platform/ephemeral-environments
go mod download
go build -o ephemeral-env .
```

### Build the Docker image

```bash
docker build -t ephemeral-env .
```

---

## Commands

### `ephemeral-env create`

Creates an ephemeral environment for a given project.

```bash
ephemeral-env create --path <project_path> --env-id <env_id>
```

**Flags:**

| Flag | Description | Required |
|---|---|---|
| `--path` | Path to a scaffold project folder (must contain a `Dockerfile`) | Yes |
| `--env-id` | Unique environment identifier, e.g. `payments-pr-123` | Yes |

**Example:**

```bash
./ephemeral-env create --path ./my-service --env-id payments-pr-123
```

**Output:**

```
⚡ Creating ephemeral environment payments-pr-123 (stack: go)...
...docker compose output...

🌐 Preview environment ready:
   http://payments-pr-123.local.scaffold.dev

💰 Estimated cost rate: $0.20/hour
```

### Validation

The tool exits with a non-zero status if:

- `--path` does not exist
- `--path` does not contain a `Dockerfile`

```
Dockerfile not found in provided path. A valid project must include a Dockerfile.
```

---

## Stack Detection

The stack is inferred by scanning the content of the project's `Dockerfile` (case-insensitive):

| Dockerfile contains | Detected stack |
|---|---|
| `spark` | `spark` |
| `kafka` | `kafka` |
| `golang` | `go` |
| `python` | `python` |
| _(none match)_ | `unknown` |

> Note: `spark` is checked before `python` because PySpark images contain both keywords.

Stack detection is informational — execution continues regardless of the result.

---

## Generated docker-compose

For each environment, a `docker-compose.yml` is written to:

```
envs/<env-id>/docker-compose.yml
```

It includes two services:

**App service** — built from the project path, with Traefik labels:

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.<env-id>.rule=Host(`<env-id>.local.scaffold.dev`)"
  - "traefik.http.services.<env-id>.loadbalancer.server.port=8080"
```

**Traefik** — reverse proxy that picks up Docker labels automatically:

```yaml
traefik:
  image: traefik:v2.11
  command:
    - "--providers.docker=true"
    - "--entrypoints.web.address=:80"
  ports:
    - "80:80"
  volumes:
    - "/var/run/docker.sock:/var/run/docker.sock:ro"
```

Both services share a `traefik-net` bridge network.

---

## CI/CD Integration

### Run as Docker container (recommended for pipelines)

```bash
docker run \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  ephemeral-env \
  create --path /workspace/my-service --env-id payments-pr-123
```

### GitHub Actions example

```yaml
- name: Deploy preview environment
  run: |
    docker run \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -v ${{ github.workspace }}:/workspace \
      ephemeral-env \
      create \
        --path /workspace/${{ env.SERVICE_NAME }} \
        --env-id ${{ env.SERVICE_NAME }}-pr-${{ github.event.pull_request.number }}
```

### Concourse example

```yaml
- task: create-preview-env
  config:
    platform: linux
    image_resource:
      type: registry-image
      source: { repository: ephemeral-env }
    run:
      path: ephemeral-env
      args:
        - create
        - --path
        - /workspace/my-service
        - --env-id
        - my-service-pr-123
```

---

## Project Structure

```
ephemeral-environments/
├── main.go
├── go.mod
├── Dockerfile                        # Multi-stage build for the tool itself
├── cmd/
│   ├── root.go                       # Cobra root command
│   └── create.go                     # create subcommand + flag definitions
├── internal/
│   ├── cli/cli.go                    # Shared error helpers
│   ├── detector/detector.go          # Stack detection from Dockerfile content
│   ├── compose/compose.go            # docker-compose.yml generation via text/template
│   ├── docker/docker.go              # Runs docker compose up -d --build
│   └── utils/utils.go                # Path and Dockerfile validation
└── envs/
    └── <env-id>/
        └── docker-compose.yml        # Generated per environment
```
