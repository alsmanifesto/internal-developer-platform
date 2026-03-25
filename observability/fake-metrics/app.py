"""
Fake metrics server — generates synthetic HTTP histogram data for Prometheus.
Exposes metrics on :8081/metrics.
"""

import random
import time
import threading
from prometheus_client import start_http_server, Histogram, Counter, REGISTRY

REQUEST_LATENCY = Histogram(
    "http_request_duration_seconds",
    "Simulated HTTP request latency",
    ["method", "endpoint", "status"],
    buckets=[0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0],
)

REQUEST_COUNT = Counter(
    "http_requests_total",
    "Total simulated HTTP requests",
    ["method", "endpoint", "status"],
)

ENDPOINTS = ["/api/payments", "/api/orders", "/api/users", "/health"]
METHODS = ["GET", "POST"]


def simulate_traffic():
    """Continuously simulate request latencies with realistic distribution."""
    while True:
        endpoint = random.choice(ENDPOINTS)
        method = random.choices(METHODS, weights=[70, 30])[0]

        # 95% fast requests (10–150ms), 4% slow (150–800ms), 1% very slow (800ms–2s)
        roll = random.random()
        if roll < 0.95:
            latency = random.uniform(0.010, 0.150)
            status = "200"
        elif roll < 0.99:
            latency = random.uniform(0.150, 0.800)
            status = random.choices(["200", "500"], weights=[80, 20])[0]
        else:
            latency = random.uniform(0.800, 2.0)
            status = random.choices(["200", "503"], weights=[60, 40])[0]

        REQUEST_LATENCY.labels(method=method, endpoint=endpoint, status=status).observe(latency)
        REQUEST_COUNT.labels(method=method, endpoint=endpoint, status=status).inc()

        time.sleep(random.uniform(0.05, 0.2))


if __name__ == "__main__":
    start_http_server(8081)
    print("Fake metrics server running on :8081/metrics")

    thread = threading.Thread(target=simulate_traffic, daemon=True)
    thread.start()

    while True:
        time.sleep(1)
