# Observability Stack

Grafana + Prometheus stack completamente definido como código. Sin configuración manual — datasources, dashboards y alertas se provisionan automáticamente al levantar el stack.

---

## Requisitos

- Docker
- Docker Compose v2

---

## Levantar el stack

```bash
cd observability
docker compose up -d
```

Eso es todo. En unos segundos los tres servicios estarán corriendo.

---

## Servicios

| Servicio | URL | Descripción |
|---|---|---|
| Grafana | http://localhost:3000 | Dashboards — sin login requerido |
| Prometheus | http://localhost:9090 | Métricas y alertas |
| mock-app (node-exporter) | http://localhost:9100/metrics | Fuente de métricas simuladas |

---

## Qué viene incluido

### Dashboards

Grafana abre con el dashboard **Service Overview** ya cargado, con tres paneles SLI:

- **p99 Latency** — latencia en el percentil 99
- **Error Rate** — tasa de errores de red (umbral: 5%)
- **Throughput** — bytes recibidos por segundo

El dashboard se refresca automáticamente cada 30 segundos.

### Alertas (Prometheus)

Tres reglas activas definidas en `prometheus/alerts.yml`:

| Alerta | Condición | Severidad | `for` |
|---|---|---|---|
| `HighErrorRate` | error rate > 5% | critical | 5m |
| `HighCpuUsage` | CPU usage > 80% | warning | 5m |
| `InstanceDown` | target sin respuesta | critical | 1m |

Ver el estado de las alertas en: http://localhost:9090/alerts

### Scrape jobs

Prometheus recolecta métricas de tres targets (intervalo: 15s):

| Job | Target |
|---|---|
| `prometheus` | localhost:9090 |
| `app` | mock-app:9100 |
| `grafana` | grafana:3000 |

---

## Comandos útiles

```bash
# Ver logs de todos los servicios
docker compose logs -f

# Ver logs de un servicio específico
docker compose logs -f grafana
docker compose logs -f prometheus

# Ver estado de los contenedores
docker compose ps

# Detener el stack (conserva los volúmenes)
docker compose down

# Detener y eliminar datos persistidos
docker compose down -v
```

---

## Estructura

```
observability/
├── docker-compose.yml
├── prometheus/
│   ├── prometheus.yml        # Configuración de scrape jobs y carga de reglas
│   └── alerts.yml            # Reglas de alertas
└── grafana/
    ├── provisioning/
    │   ├── datasources/
    │   │   └── datasource.yml    # Prometheus como datasource por defecto
    │   └── dashboards/
    │       └── dashboard.yml     # Apunta a la carpeta de dashboards
    └── dashboards/
        └── service-overview.json # Dashboard SLI con 3 paneles
```

---

## Agregar métricas de un servicio real

Para scrapear un servicio propio, agregar un nuevo job en `prometheus/prometheus.yml`:

```yaml
- job_name: mi-servicio
  static_configs:
    - targets: ["host.docker.internal:8080"]
```

Luego recargar la configuración sin reiniciar:

```bash
curl -X POST http://localhost:9090/-/reload
```

> Esto funciona gracias al flag `--web.enable-lifecycle` habilitado en el compose.
