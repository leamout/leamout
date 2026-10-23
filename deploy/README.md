# Deployment

- `compose/docker.yaml` runs Leamout API, worker, media, telephony and data services.
- `compose/observability.yaml` adds Prometheus, Alertmanager, Grafana, Loki,
  Tempo and Alloy to the **same Compose project**.
- `observability/` contains the configs, rules, targets and dashboard provisioning.
- `Caddyfile` is the public proxy configuration. Recording objects live exclusively in customer-owned S3-compatible storage.

From the repository root:

```sh
make deployment-up
GRAFANA_ADMIN_PASSWORD=choose-a-strong-secret docker compose \
  --env-file .env --env-file server/.env \
  -f deploy/compose/docker.yaml -f deploy/compose/observability.yaml up -d
```

For configuration-only validation, provide both example environment files
with `--env-file .env.example --env-file server/.env.example`. No application
configuration file is mounted or injected wholesale into containers.

The overlay uses the existing `control` network. The application stack also
uses a separate public-proxy `edge` network, and isolated `database`, `cache`, and `events` networks. FreeSWITCH and RTPEngine retain their
fixed IP addresses on `control` for SIP/media connectivity.

Grafana is published only on `127.0.0.1:3000`; access it through a secured tunnel or reverse proxy.
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
configuration values consumed only by the individual services that need them.
Compose uses both files for substitution, but each container receives an
explicit environment allowlist. The edge proxy does not receive database,
identity, billing, or encryption credentials.

Use the same PostgreSQL password in both files. Match `DOMAIN` in the root
`.env` and `server/.env`; Caddy uses DOMAIN to configure its public hosts. Default hosted origins are `FRONTEND_URL=https://console.leamout.com`, `BACKEND_URL=https://api.leamout.com`, and `CORS_ORIGINS=https://console.leamout.com`. Set `KINDE_ISSUER`, `KINDE_CLIENT_ID`, and `KINDE_CLIENT_SECRET` from the hosted Kinde application. Set the real `DOMAIN`,
`CORS_ORIGINS`, `TURN_PUBLIC_URLS` and `TURN_REALM` for the deployment;
keep `TRUSTED_PROXY_CIDRS=172.33.0.10/32` so only the static Caddy
address can supply forwarded client IP headers;
replace all example secrets before operating an external service. Keep both
environment files private and out of images and Git.

Validate the Compose model before starting it:

```sh
docker compose \
  --env-file .env.example --env-file server/.env.example \
  -f deploy/compose/docker.yaml config --quiet
```

For additional infrastructure telemetry, use the observability overlay as
described above. Leamout's operator deployment does not constitute a separate
customer self-hosting product.

## First managed deployment: operator checklist

This is Leamout-operated infrastructure, not a customer self-hosting product.
The Compose checks and synthetic acceptance tests do **not** establish that a
live environment is ready for customer traffic.

1. Provision DNS for `api.<DOMAIN>`, `backoffice.<DOMAIN>`, and `turn.<DOMAIN>`. Host the console separately
   at `console.<DOMAIN>`. Confirm public IPv4/NAT, firewalls, SIP signaling,
   RTP, TURN and TLS reachability. Install the certificate/key files required
   by OpenSIPS and Coturn under `/etc/leamout/certs/`, and configure the
   Caddy ACME mount at `/var/lib/leamout/acme` as needed.
2. Copy both example environment files as shown above. Generate **unique**
   secrets for PostgreSQL, FreeSWITCH, media control, TURN, and
   the Kinde client. A URL-safe generator avoids characters that require
   escaping in PostgreSQL URLs or Compose values:

   ```sh
   python3 -c 'import secrets; print(secrets.token_urlsafe(32))'
   python3 -c 'import base64, secrets; print(base64.urlsafe_b64encode(secrets.token_bytes(32)).decode())'
   ```

   Use the second command for `ENCRYPTION_KEY` only. **Back up this key
   securely and never regenerate it for an existing database**; losing it can
   make encrypted organization credentials unrecoverable.
3. Set actual public URLs, `PUBLIC_IP`, TURN settings, and service
   credentials. `POSTGRES_PASSWORD` in the root file must match the
   password in `server/.env`'s `DATABASE_URL`; percent-encode URL-reserved
   characters if necessary. Configure the Kinde confidential OIDC app with both console and backoffice callback URLs.
4. Restrict the environment files and run the fail-closed checks **before**
   starting or upgrading services:

   ```sh
   chmod 600 .env server/.env
   make deployment-preflight
   make compose-validate
   python3 tests/ops/validate.py
   docker compose \
     --env-file .env.example --env-file server/.env.example \
     -f deploy/compose/docker.yaml config --format json \
     | python3 tests/ops/assert_isolation.py
   ```

   The preflight checks example credentials, key strength and encoding,
   public IP, DNS/origin consistency, database password consistency, narrow
   forwarded-header trust, and accidental partial Polar configuration. It does not print secrets or
   call external services.
5. If this is an **upgrade**, back up PostgreSQL, NATS JetStream, Redis, and relevant persistent volumes **before**
   running Compose. Verify that backups are restorable. The `migrate` service
   applies migrations automatically on startup; rolling back application
   containers does not roll back database schema changes.
6. Start the stack and inspect service health and logs:

   ```sh
   make deployment-up
   docker compose --env-file .env --env-file server/.env -f deploy/compose/docker.yaml ps
   ```

7. Before enabling customer traffic, verify live HTTPS/TLS and Kinde
   login/logout, API readiness (`https://api.<DOMAIN>/readyz`), private
   metrics access controls, Alertmanager delivery, and a real organization
   onboarding flow. Place an actual BYOC SIP call using a verified BYOA
   integration; confirm two-way audio, interruption, tool execution, call
   teardown, records and usage. Exercise a controlled restart and review
   logs and alerts. Validate Polar checkout and webhooks before enabling
   paid subscriptions.

**Stop the rollout** if any environment preflight, migration, health,
authentication, provider, SIP, or operational check fails. Restore from a
verified backup or roll back only after assessing the migration compatibility;
do not delete named volumes as a troubleshooting shortcut.

## Production reliability gates

Run `python3 tests/ops/validate.py` to check the default origins and Kinde identity configuration, and public metrics boundary. CI also runs `caddy adapt --validate` and `promtool check rules` against the deployment assets. These configuration checks do not replace a live TLS or Kinde login test.

Prometheus alerts cover target availability, API HTTP 5xx spikes, media attachment timeouts, provider/transport failures, mean turn latency, and backpressure. The exporter currently measures aggregate mean latency; no p95/p99 claim is possible without histograms. Suggested initial operational objectives are API success >=99.5% and media attachment success >=99% over rolling 30-day windows. The present short-window alerts are operational signals, not full error-budget burn alerts. Configure a real Alertmanager receiver before production.

The Voice Agent release gate tests continued call, media fork and durable-session identity across a worker restart. An independent media-process restart remains a destructive fault: it is not a call-continuity mechanism. A media **process crash** loses its in-memory provider streams; the system cannot transparently resume an existing call on a replacement media process. Draining and controlled rollover must be preferred to crashing active media nodes.

## BYOS-only recording storage

Leamout does not provide a default object store. Customers configure S3-compatible recording integrations; the runtime does not start or depend on MinIO. The `recordings-data` volume remains temporary FreeSWITCH recording staging, not long-term object storage.

**Existing deployments:** before removing a live MinIO service or volume, back up and migrate every recording where `storage_integration_id IS NULL` into the appropriate customer bucket. Verify object content and transactionally update its integration ID and bucket. Legacy recordings remain intentionally inaccessible to the BYOS-only adapter until migrated; never silently repoint them to the current customer bucket. This PR does not automate migration.
