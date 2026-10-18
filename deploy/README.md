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

Prometheus scrapes `server:8080/metrics` and `media:8090/metrics` directly on the private `control` network. Caddy returns HTTP 404 at the public `api.<DOMAIN>/metrics` endpoint; do not publish these internal ports.
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

Use the same PostgreSQL password in both files. Match `DOMAIN` in the root `.env` and `server/.env`; the Compose value sets Keycloak's hostname and the application value configures Caddy. Default hosted origins are `FRONTEND_URL=https://console.leamout.com`, `BACKEND_URL=https://api.leamout.com`, `KEYCLOAK_ISSUER=https://auth.leamout.com/realms/leamout` and `CORS_ORIGINS=https://console.leamout.com`. Set the real `DOMAIN`,
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

## Production reliability gates

Run `python3 tests/ops/validate.py` to check the default origins, Keycloak realm placeholders, and public metrics boundary. CI also runs `caddy adapt --validate` and `promtool check rules` against the deployment assets. These configuration checks do not replace a live TLS or Keycloak login test.

Prometheus alerts cover target availability, API HTTP 5xx spikes, media attachment timeouts, provider/transport failures, mean turn latency, and backpressure. The exporter currently measures aggregate mean latency; no p95/p99 claim is possible without histograms. Suggested initial operational objectives are API success >=99.5% and media attachment success >=99% over rolling 30-day windows. The present short-window alerts are operational signals, not full error-budget burn alerts. Configure a real Alertmanager receiver before production.

The Voice Agent release gate tests recovery of worker control for an ongoing audio fork and a media restart between calls. A media **process crash** loses its in-memory provider streams; the system cannot transparently resume an existing call on a replacement media process. Draining and controlled rollover must be preferred to crashing active media nodes.
