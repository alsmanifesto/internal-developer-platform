# Internal Developer Platform

A set of self-contained tools for bootstrapping and running services locally. Each module works independently — but `ephemeral-env` pairs naturally with projects created by `scaffold`.

---

## Modules

### `service-scaffolding` — scaffold

Creates a new service from scratch: source code, Dockerfile, Terraform, CI/CD pipeline, and README — all generated in one command.

```bash
cd service-scaffolding

# Interactive (guided TUI)
./scaffold create

# Non-interactive (for scripts and pipelines)
./scaffold create --name payments-api --service-type api --workload app --stack go --pipeline gh-actions
```

To build the binary:
```bash
go build -o scaffold .
```

---

### `ephemeral-environments` — ephemeral-env

Spins up a preview environment for any project that has a `Dockerfile`. Works great with services created by `scaffold`. Generates a `docker-compose.yml` with Traefik routing and brings it up.

```bash
cd ephemeral-environments

./ephemeral-env create --path ../service-scaffolding/payments-api --env-id payments-pr-123
```

Use `--dry-run` to generate the compose file without starting anything — useful for jobs, functions, or work in progress:
```bash
./ephemeral-env create --path ../service-scaffolding/my-fn --env-id my-fn-pr-42 --dry-run
```

Preview available at: `http://payments-pr-123.local.ravon.dev`

To build the binary:
```bash
go build -o ephemeral-env .
```

---

### `observability` — Grafana + Prometheus

A ready-to-use monitoring stack. No manual setup — dashboards, datasources, and alert rules are all provisioned as code.

```bash
cd observability
docker compose up -d
```

| Service | URL |
|---|---|
| Grafana | http://localhost:3000 |
| Prometheus | http://localhost:9090 |

Grafana opens with anonymous access enabled, the Prometheus datasource pre-configured, and the **Service Overview** dashboard (p99 latency, error rate, throughput) already loaded.

---

## How they fit together

```
scaffold create → generates service with Dockerfile
      ↓
ephemeral-env create --path ./my-service --env-id my-service-pr-1
      ↓
Preview environment running behind Traefik
      ↓
observability stack scrapes metrics from running services
```
