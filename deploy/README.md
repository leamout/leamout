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

## Environment configuration

Copy the deployment examples from the repository root:

```sh
cp .env.example .env
cp server/.env.example server/.env
```

The root `.env` supplies Compose infrastructure settings including
`POSTGRES_PASSWORD` and `PUBLIC_IP`. `server/.env` supplies application
configuration to server, worker, media, Caddy, FreeSWITCH, and Coturn.

Use the same PostgreSQL password in both files. Set the real `DOMAIN`,
`CORS_ORIGINS`, `TURN_PUBLIC_URLS` and `TURN_REALM` for the deployment;
replace all example secrets before operating an external service. Keep both
environment files private and out of images and Git.

Validate the Compose model before starting it:

```sh
APP_ENV_FILE=../../server/.env.example docker compose \
  --env-file .env.example \
  -f deploy/compose/docker.yaml config --quiet
```

For additional infrastructure telemetry, use the observability overlay as
described above. Leamout's operator deployment does not constitute a separate
customer self-hosting product.
