# Deployment

- `compose/docker.yaml` runs Leamout API, worker, media, telephony and data services.
- `compose/observability.yaml` adds Prometheus, Alertmanager, Grafana, Loki,
  Tempo and Alloy to the **same Compose project**.
- `observability/` contains the configs, rules, targets and dashboard provisioning.
- `Caddyfile` and `minio-policy.json` are deployment assets.

From the repository root:

```sh
docker compose --env-file .env -f deploy/compose/docker.yaml up -d
GRAFANA_ADMIN_PASSWORD=choose-a-strong-secret docker compose --env-file .env \
  -f deploy/compose/docker.yaml -f deploy/compose/observability.yaml up -d
```

For configuration-only validation, set `APP_ENV_FILE=../../server/.env.example`
and use `--env-file .env.example` with `config --quiet`. This relative
`APP_ENV_FILE` path is resolved from `deploy/compose/`.

The overlay uses the existing `control` network. Grafana is published only
on `127.0.0.1:3000`; access it through a secured tunnel or reverse proxy.
All other monitoring services are internal-network-only.

Prometheus scrapes `server:8080/metrics` and `media:8090/metrics`.
Alloy accepts OTLP inside the Docker network, sending traces to Tempo and logs
to Loki. No Docker socket is mounted. Log forwarding requires application or
collector instrumentation; container stdout is **not** automatically ingested.
The Go app must explicitly configure a reachable OTLP endpoint to send traces.
Current tracing initialization should be wired and tested independently.

**Not production-ready by itself:** Configure an Alertmanager receiver,
retention and storage capacities, TLS/authentication for cross-host telemetry,
backup policies and monitoring access controls before running at scale.
