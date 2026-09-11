# Cloud and self-hosted security roadmap

This roadmap deliberately separates Leamout Cloud security from self-hosted
security. The two products may reuse small security primitives, but neither
deployment mode may require the other's control plane, credentials, identity
provider, telemetry, or network access.

## Boundary rule

The runtime has one explicit deployment profile selected at installation:

- `cloud`: Leamout operates the runtime and the database contains multiple
  organizations;
- `self_hosted`: the customer operates the runtime in its VPC and no Leamout
  Cloud dependency is required for normal communications traffic.

The profile is configuration, not a request parameter. It is loaded at startup,
cannot change during the process lifetime, and must not be inferred from a host
name, organization, license response, or HTTP header.

Cloud-only and self-hosted-only dependencies belong behind separate composition
roots. A shared package may define interfaces and local cryptographic or policy
primitives, but it must not silently select a remote implementation. Unsupported
configuration combinations fail at startup.

| Concern | Leamout Cloud | Self-hosted | Shared primitive only |
| --- | --- | --- | --- |
| Tenant boundary | Organization on every row and operation | Deployment-local organizations | Scoped authorization types |
| Human identity | Cloud identity and organization membership | Local identity; optional customer SSO | Password, token, session primitives |
| Carrier ownership | BYOC per tenant; Leamout provider accounts platform-only | Customer BYOC credentials only by default | Scoped envelope encryption |
| Key custody | Leamout KMS/HSM and workload identity | Customer-owned key or local secret file | AEAD envelope format |
| Licensing | Cloud service is authoritative | Locally cached signed entitlement | Offline signature verification |
| Telemetry | Leamout-operated private observability | Local-only by default; explicit opt-in export | Redaction rules |
| Updates | Leamout deployment pipeline | Customer-triggered signed artifacts | Manifest verification |
| Backups | Leamout-operated encrypted storage | Customer-selected local destination | Restore verification |

## Next: establish deployment composition roots

Before adding more controls, make the deployment boundary executable:

1. Add a typed `DeploymentMode` with only `cloud` and `self_hosted` values.
2. Validate the mode at startup and reject unknown or empty values outside
   development.
3. Create separate Cloud and self-hosted constructors that receive only their
   permitted dependencies.
4. Make provider-master integrations, Cloud enrollment clients, and Cloud
   telemetry impossible to construct from the self-hosted composition root.
5. Add startup tests proving self-hosted mode works without DNS or Internet
   access and rejects Cloud-only credentials.

This is the immediate implementation milestone because it prevents later work
from accidentally coupling the two trust models.

Current status: the typed mode, startup validation, API and worker composition
roots, Cloud-only provider wiring, and rejection of Cloud provider credentials
in self-hosted mode are implemented. The remaining gate for this milestone is
the network-isolated self-hosted startup and acceptance harness.

## Cloud track: multi-tenant isolation

After composition is separated, implement the Cloud track independently:

1. **Database enforcement.** Add PostgreSQL row-level security for every
   tenant-owned table. Set the organization in each transaction with a
   transaction-local database setting, use a non-owner/non-`BYPASSRLS` runtime
   role, and force RLS on protected tables.
2. **Repository contracts.** Require organization identity in tenant repository
   methods and retain ownership joins for defense in depth. Ban unscoped lookup
   helpers from customer-facing modules.
3. **Isolation tests.** For every resource, seed two organizations and verify
   list, get, create-through-parent, update, delete, pagination, idempotency,
   jobs, exports, and webhooks cannot cross the boundary. Run these tests using
   the production database role.
4. **Provider secret vault.** Move Leamout master carrier credentials out of
   general application configuration into a Cloud-only secret provider backed
   by KMS/HSM and workload identity. Customer handlers never receive that
   interface. Separate tenant BYOC keys from platform-provider keys.
5. **Operator plane.** Put cross-tenant support access on a separately deployed
   service and identity, require phishing-resistant MFA and just-in-time grants,
   and emit immutable reason-coded audit events. Do not add a support bypass to
   tenant middleware.
6. **Cloud egress policy.** Allow only configured carrier, payment, email, and
   observability destinations from the workloads that require them. Database,
   Redis, NATS, metrics, media control, and operator endpoints stay private.

Cloud completion gate: automated tests demonstrate that compromise of a normal
API credential, guessed resource UUID, forged organization header, background
job payload, or copied ciphertext cannot expose or mutate another organization
or retrieve a platform provider credential.

## Self-hosted track: sovereignty and instance integrity

The self-hosted track does not wait for, or call into, Cloud tenant controls:

1. **No-egress baseline.** Start successfully and complete local API, SIP, and
   media acceptance tests with external DNS and Internet routes blocked. External
   telemetry is absent unless the operator supplies a destination and opts in.
2. **Local bootstrap.** Generate deployment-unique secrets locally, create the
   first administrator through a single-use bootstrap flow, expire the bootstrap
   credential, and never send it to Leamout Cloud.
3. **Local key custody.** Support a customer-provided encryption key or a
   permission-restricted local key file. Backups include encrypted data but not
   the key by default; restore tooling explicitly verifies key availability.
4. **Network hardening.** Publish only HTTPS, SIP, RTP, and configured TURN
   ports. Bind databases, Redis, NATS, FreeSWITCH ESL, internal APIs, and metrics
   to private networks. Run containers as non-root with read-only filesystems,
   dropped capabilities, and explicit writable volumes.
5. **Signed offline lifecycle.** Verify release manifests and artifacts before
   install or upgrade. Entitlements are signed and verified locally. Enrollment
   is optional after installation and never carries tenant or communications
   data.
6. **Local audit and recovery.** Keep security audit data and support bundles
   local, redact secrets automatically, document key and administrator recovery,
   and test an encrypted restore without contacting Leamout.

Self-hosted completion gate: the production acceptance suite passes in a VPC
with deny-all Internet egress, no Leamout credentials, and no Leamout-controlled
services, while signed offline upgrades, backup restore, administrator recovery,
and local audit export remain functional.

## What remains shared

Only controls intrinsic to communications execution remain shared: constant-time
credential verification, scoped authenticated encryption, request size limits,
SIP parsing and normalization, carrier admission leases, destination policy,
caller-ID ownership checks, audit event schemas, and secret-redaction helpers.

Shared code accepts its dependencies explicitly and fails closed. In particular,
it must not contain a global Cloud client, default external endpoint, provider
master credential, or background telemetry exporter. Fraud limits may have
different policy values in each product, but calls cannot bypass the shared
admission interface in either mode.

## Delivery order

1. Deployment mode and separate composition roots.
2. In parallel, Cloud RLS/isolation harness and self-hosted no-egress harness.
3. Cloud provider vault and operator-plane separation.
4. Self-hosted local bootstrap, key custody, and container hardening.
5. Run the shared telecom fraud suite against both assembled products.

Each change belongs to exactly one track or to the narrowly defined shared
primitive layer. Pull requests that introduce a cross-track runtime dependency
must include an architecture decision explaining why the boundary cannot be
preserved.
