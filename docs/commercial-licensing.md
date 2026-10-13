# Cloud subscriptions and commercial self-hosted licensing

Leamout Cloud uses Stripe subscriptions. Self-hosted installations use locally
verified commercial licenses. Polar is not integrated. Deployment mode is an
installation setting shared by the API, worker, and media processes.

## Cloud

Set `DEPLOYMENT_MODE=cloud` (the compatibility default) and configure
`STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, and the three `STRIPE_*_PRICE_ID`
variables in `server/.env`. Create monthly USD Stripe prices matching the seeded
Developer ($79), Pro ($299), and Scale ($799) plans. Free needs no Stripe subscription.
Configure Stripe to send checkout.session.completed and
customer.subscription.created/updated/deleted to `/v1/webhooks/stripe`.
Configure the customer portal for allowed upgrades, downgrades and cancellations.

Authenticated organization APIs:

- `GET /v1/plans`: plan catalog.
- `GET /v1/subscription`: current subscription (see subscription route definition).
- `POST /v1/billing/checkout`: plan_code, success_url, cancel_url; requires an Idempotency-Key header.
- `POST /v1/billing/portal`: return_url.
- `GET /v1/limits`: resolved limits and usage period.
- `GET /v1/runtime-usage`: used_seconds, allowance_seconds, unlimited and period boundaries.
- `GET /v1/entitlements`: effective feature access.

Checkout URLs must use the configured frontend origin. Webhooks authenticate the
raw body with Stripe-Signature. Subscription writes reject older provider
observations and conflicting subscription/customer identities. The Cloud worker
reconciles locally known subscriptions with Stripe at startup and every 15 minutes;
individual failures are logged and retried without stopping telephony workloads.
A checkout that never delivers its initial webhook is not discoverable from local
subscription records: replay that event through Stripe.

Only active/trialing subscriptions with a current period receive paid limits and
plan-derived feature access. Missing, past-due, canceled or expired subscriptions
resolve to Free capacity; explicit organization feature overrides retain their
existing semantics. Redis atomically reserves organization concurrency, while
Postgres accounts runtime from session intervals. Runtime minutes are checked at
new-session admission: already admitted calls finish, so usage can exceed the
allowance. This is an admission quota, not a mid-call termination policy.

## Self-hosted issuance

Licenses authorize the core runtime, deployment-wide concurrency, media node count,
and supported feature flags. They do not inherit Cloud minutes, agent/trunk counts,
or retention quotas. Trunk operational limits and local retention policies still apply.
There is no billing API dependency or online activation requirement.

Generate an issuer key offline:

```sh
cd server
go run ./cmd/license -generate-key -output /secure/issuer-private.key
```

The command prints only the base64 public key. Store the private key securely;
never include it in source control, images, customer environments or license files.
Pin public keys in all three distributed binaries using this release build argument:

```sh
docker build --build-arg 'LICENSE_PUBLIC_KEYS=issuer-2026:BASE64_PUBLIC_KEY' --target runtime -t leamout/self-hosted:VERSION server
```

The equivalent Go build flag is
`-ldflags '-X github.com/leamout/leamout/server/internal/commercial/licensing.BuildPublicKeys=issuer-2026:BASE64_PUBLIC_KEY'`.
Multiple keys use comma-separated `key-id:base64-key` entries. Ship a release with
both public keys before rotating the signing key. Runtime environment variables
cannot replace the trusted public keyring.

Create claims JSON:

```json
{
  "version": 1,
  "license_id": "lic_customer_production_2026",
  "customer_id": "customer",
  "deployment_id": "customer-production",
  "valid_from": "2026-10-01T00:00:00Z",
  "expires_at": "2027-10-01T00:00:00Z",
  "grace_until": "2027-10-15T00:00:00Z",
  "max_concurrent_calls": 50,
  "max_nodes": 3,
  "features": ["sso", "advanced_rbac", "retention_policies", "private_networking"]
}
```

Issue the license:

```sh
go run ./cmd/license -claims claims.json -private-key /secure/issuer-private.key -key-id issuer-2026 -output license.json
```

The output cannot overwrite an existing file. Validation requires positive finite
capacities, a supported schema, nonempty identities and a signed grace period of
at most 30 days. No local flag extends license expiry. Verification covers the
exact payload bytes with Ed25519.

## Installation and renewal

Configure `DEPLOYMENT_MODE=self_hosted`, `LICENSE_DEPLOYMENT_ID`, and
`LICENSE_FILE` on the API, worker and media processes. Use the same Redis instance
and deployment ID across the installation. Mount the parent license directory
read-only, with the file readable by the container's application user. Mounting
the directory allows atomic replacement without a stale file bind mount.

`deploy/compose.self-hosted.yaml` provides the overlay. Set LICENSE_PUBLIC_KEYS
for issuer builds, LICENSE_DEPLOYMENT_ID, LICENSE_DIRECTORY, and MEDIA_NODE_ID,
then combine it with `deploy/compose.yaml`. Every scaled media process must have a
unique, stable node ID; the default media-1 is only for a single-node installation.
The overlay uses issuer build inputs, not a public customer-controlled activation
service. Distribute the resulting images with the signed file.

A node means a registered media runtime process, not an API/worker replica or a
physical machine. Each media process reserves a short Redis node lease, renewed
every five seconds and on session creation. Leases expire after 15 seconds so
crashes do not permanently consume node capacity. Duplicate node IDs are unsupported.
Deployment concurrency uses the existing call lease lifecycle, shared across
organizations. Creating more organizations cannot multiply the licensed allowance.

`GET /v1/license` is protected by organization commercial access and returns
cloud, missing, invalid, not_yet_valid, active, grace or expired status with verified
claims. The API remains available for administration when runtime access is blocked.
Replace license.json atomically to renew without restarting. Decisions reread the
bounded file and verify its signature; they never retain a previously valid license
as a fallback when the replacement is invalid.

At expiry, signed grace permits continued usage. After grace ends, new calls and
new media sessions are blocked; existing calls, teardown, data access and locally
configured retention cleanup continue. Missing or tampered licenses block new
runtime admission immediately. Current feature gates (SSO, SCIM, advanced RBAC,
retention policies and private networking) resolve from license features rather
than Cloud records. Audit history access requires the audit_logging license feature, including after
expiry; audit capture continues independently so security records are never lost
because of billing. premium_templates reserves an identifier for a future template
product and does not introduce a template catalog.

Capacity reductions prevent new call admissions until usage is below the new
limit. Existing node leases are not evicted; apply node-count reductions through
a coordinated rollout that stops excess media processes. Offline licensing cannot
reliably prevent host clock rollback, duplicated installations or binary patching,
and supports no immediate remote revocation. Contracts and issuer release controls
remain part of commercial licensing.

## Source license and release policy

This repository currently has an Apache 2.0 LICENSE. These runtime checks do not
change that license or revoke existing recipients' Apache rights. A commercial
binary/source distribution strategy and agreement need to be established separately
before claiming that all use of this public code legally requires payment. No
repository license change is made by this implementation.

## Verification

Run `go test ./...`, `go test -race ./internal/commercial/... ./internal/runtime/...`,
`go vet ./...` and `go build ./...` from server/. Tests cover license identity,
signature, validity/grace boundaries, file renewal, Cloud independence, deployment
concurrency scope and reconciliation retries. Operational release acceptance should
also exercise Stripe test-mode checkout/webhooks and multi-node Redis admission.
