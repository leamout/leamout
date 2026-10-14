# Leamout observability (Compose bootstrap)

Optional, standalone metrics/logs/traces stack for Leamout deployments. It runs independently from the Community runtime and Enterprise extension process; losing the stack must never interrupt calls. **This is a single-host bootstrap, not a supported HA, multi-tenant or air-gapped production deployment.**

## Components

- Prometheus (metrics) and Alertmanager (alert routing)
- Loki (logs) and Tempo (traces)
- Grafana (datasources and the **stack-health** dashboard)
- Grafana Alloy (OTLP logs/traces ingest and fan-out)

Versions are pinned in `../compose/observability.yaml`. Grafana requires an administrator password. No bundled default password, external telemetry upload, or anonymous Grafana access.

## Start

From the Enterprise repository root, create a locally ignored environment file, for example `deploy/compose/.env`, containing `GRAFANA_ADMIN_PASSWORD=<a-long-random-secret>`. Run:

```sh
docker compose --env-file deploy/compose/.env -f deploy/compose/observability.yaml config --quiet
docker compose --env-file deploy/compose/.env -f deploy/compose/observability.yaml up -d
```

Visit http://127.0.0.1:3000 and sign in as `admin` (or the configured `GRAFANA_ADMIN_USER`). Grafana has provisioned Prometheus, Loki and Tempo data sources and a stack-health dashboard. OTLP gRPC/HTTP ports are bound only to localhost at 4317/4318 by default (override with `ALLOY_OTLP_GRPC_PORT` / `ALLOY_OTLP_HTTP_PORT`). The collection backends are **not** exposed on host ports.

Stop with `docker compose --env-file deploy/compose/.env -f deploy/compose/observability.yaml down`. Never use `down -v` unless permanent destruction of retained monitoring data is intentional.

## Connect Leamout Core

The public Core `deploy/compose.yaml` runs in another Compose project, so container names are **not** automatically reachable across projects. The monitoring stack creates the shared Docker network `leamout-observability`. Attach only explicitly instrumented Core services to that network using a separately reviewed Core Compose override, for example:

```yaml
services:
  server:
    networks:
      - control
      - leamout-observability

networks:
  leamout-observability:
    external: true
    name: leamout-observability
```

Run this override alongside Core's normal Compose file (from the **Core repository**), after the monitoring network is created. Preserve any existing Core networks; do not replace the deployment's SIP/media connectivity. The sketch above is for the `server` service only: attach other services as needed.

**Metrics**: verify the actual Core or exporter `/metrics` address and port, then copy `prometheus/targets/leamout.json.example` to `prometheus/targets/leamout.json` and replace its placeholder. The sample is deliberately ignored until copied. No invented Leamout metric names or ports are shipped. Prometheus checks `/metrics` by default. A private metrics endpoint is not automatically present merely because the service is reachable.

**Logs/traces**: configure instrumented Core processes or an approved log-forwarder to send OTLP logs/traces to `alloy:4317` (gRPC) or `http://alloy:4318` (HTTP) from the shared network. From the Docker host, use localhost ports instead. Existing stdout logs and call traces are **not** ingested automatically: instrumentation/exporters or a securely scoped agent must be added separately. Do not mount the Docker daemon socket just to enable log collection.

## Alerting, retention and security

The included `ObservabilityTargetDown` and `PrometheusScrapeSamplesMissing` alerts are evaluated by Prometheus and routed to Alertmanager. **Alertmanager's default receiver has no notification destination.** Configure a real receiver and verify a test notification before using it for incident response.

Prometheus retains metrics for 15 days. Loki and Tempo use single-instance local filesystem storage without guaranteed automatic retention or object-storage redundancy. Do not treat this bootstrap as a production long-term log/trace archive. Define retention, backups, quotas, HA architecture and external storage separately for production.

Ports published to localhost are not automatically secure when tunneled/proxied. Restrict ingress, authenticate users, enforce TLS for remote ingestion, avoid logging secrets/PII, and do not use call IDs, phone numbers or tenant IDs as high-cardinality Prometheus/Loki index labels. Loki is single-tenant (`auth_enabled: false`); keep access restricted. Use OpenTelemetry trace context to correlate telemetry, without sending raw conversation audio or customer data by default.

## CI / limits

`.github/workflows/observability.yml` validates Compose, Prometheus/Alertmanager rules and Alloy syntax and parses Grafana JSON. It does not establish integration tests, guaranteed readiness of containers, verified Core instrumentation, data retention, Enterprise license behavior, production HA, or Helm packaging. Helm remains a separate milestone.
