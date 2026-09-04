# Self-hosted Leamout installation

This document defines the supported product, installation, activation, upgrade, and operating contract for a self-hosted Leamout deployment.

It is a product contract before it is an implementation guide. The public installer and `leamout` CLI may be introduced incrementally, but their behavior must converge on this contract rather than exposing Docker Compose or individual telecom components as the normal customer interface.

## Product boundary

Leamout Self-Hosted means the communications runtime executes on infrastructure controlled by the customer.

It does **not** mean that Leamout's public website or hosted customer dashboard is copied onto the customer's server.

The product surfaces are intentionally separated:

```text
clients/apps/web
    Leamout marketing website
    hosted by Leamout

clients/apps/console
    dashboard at console.leamout.com
    hosted by Leamout

self-hosted release
    Leamout communications runtime
    operated on customer infrastructure
```

The self-hosted runtime includes the Leamout API and workers plus the signaling, media, persistence, and coordination components required to execute communications locally.

The hosted console is the primary commercial and fleet-management experience for both self-hosted and managed customers.

## Customer journey

The intended self-hosted production journey is:

```text
console.leamout.com/sign-up
        ↓
create organization
        ↓
choose a self-hosted plan / subscription
        ↓
create a self-hosted deployment
        ↓
receive a short-lived activation token
        ↓
install the Leamout CLI and runtime
        ↓
initialize a durable deployment identity
        ↓
activate the deployment
        ↓
receive locally verifiable signed license material
        ↓
deployment appears in console.leamout.com
        ↓
operate the communications runtime on customer infrastructure
```

Installation and commercial activation are related but separate operations. Availability of an installer or release artifact does not itself grant production-use rights.

## Authority model

Leamout Cloud and the self-hosted runtime own different classes of state.

### Leamout Cloud

`console.leamout.com` and its backing services are authoritative for commercial and fleet-management state such as:

- organizations and users;
- plans and subscriptions;
- invoices and payments;
- commercial licenses;
- entitlements;
- deployment registrations;
- activation authorization;
- support relationships;
- release/update availability;
- optional deployment health summaries and check-in metadata.

### Self-hosted deployment

The customer deployment remains authoritative for local communications execution and telecom state such as:

- carrier connections and credentials;
- trunks and SIP configuration;
- numbers and routing;
- calls and call state;
- recordings and conferences;
- webhook execution state;
- OpenSIPS signaling;
- RTPengine media relay;
- FreeSWITCH media execution;
- PostgreSQL application state;
- Redis coordination state;
- NATS durable asynchronous workflows.

Sensitive telecom secrets must not be copied to Leamout Cloud merely because the deployment is visible in the hosted console.

## Cloud independence

Leamout Cloud is not part of the media or signaling path for a self-hosted deployment.

A transient outage or loss of connectivity to `console.leamout.com` must not immediately stop an otherwise valid deployment.

In particular, cloud unavailability must not by itself terminate established calls or require every new telecom operation to synchronously contact Leamout Cloud.

The intended model is:

```text
Leamout Cloud unavailable
        │
        ├── existing calls continue
        ├── local API remains available
        ├── workers continue
        ├── SIP routing continues
        ├── media continues
        └── locally signed license remains verifiable
```

Commercial expiration, revocation, grace periods, and enforcement policy are separate documented policies. Runtime enforcement must remain telecom-safe.

## Connectivity to Leamout Cloud

When a self-hosted deployment communicates with Leamout Cloud, prefer customer-initiated outbound HTTPS or another authenticated outbound control channel.

Do not require customers to expose a privileged management API directly to the public internet merely so the hosted console can display the deployment.

Conceptually:

```text
customer deployment
        │
        │ outbound authenticated connection
        ▼
Leamout Cloud
        │
        ▼
console.leamout.com
```

The exact registration/check-in protocol belongs to a later implementation phase.

## Goals

A new operator should eventually be able to provision a supported Linux host and install the self-hosted runtime through one public entry point:

```sh
curl -fsSL https://get.leamout.com/install.sh | sudo sh
```

The bootstrap installer installs the Leamout CLI. The CLI owns the deployment lifecycle:

```sh
sudo leamout init --activation-token <token>
sudo leamout up
sudo leamout status
```

The operator should not need to understand OpenSIPS, RTPengine, FreeSWITCH, Coturn, NATS, Redis, PostgreSQL, Atlas, or raw Docker Compose commands for normal product operation.

## Product model

Leamout is proprietary software owned by Leamout Limited.

Development, testing, and evaluation use are governed by the Leamout Software License Agreement. Production use of a self-hosted deployment requires a valid Leamout Commercial License issued by Leamout Limited.

The commercial flow is:

```text
subscription / entitlement
        ↓
deployment activation authorization
        ↓
local deployment identity
        ↓
signed commercial license
        ↓
local verification
        ↓
production entitlements
```

For Self-Hosted + BYOC, Leamout charges for the software/platform license while the customer keeps and pays its own carrier relationship unless another commercial agreement says otherwise.

## Supported deployment tiers

### Developer deployment

Repository checkout remains the contributor/development workflow:

```sh
make certs
make preflight
make up
make verify
```

This mode may include the combined marketing/waitlist site, hosted-console application, development image tags, self-signed certificates, and other repository-level conveniences that are not part of the self-hosted production release.

### Self-hosted production deployment

The first customer production target is a single supported Linux host managed through the Leamout CLI and a pinned runtime release.

The self-hosted production release contains runtime components, not `clients/apps/web` or `clients/apps/console`.

The CLI owns configuration generation, release selection, secrets, migrations, TLS setup, health verification, upgrades, backup/restore coordination, and license installation.

Docker Compose may remain the first orchestration implementation, but it is an implementation detail and recovery escape hatch rather than the customer-facing product API.

### Managed deployment

Leamout Managed uses the same product concepts while Leamout operates the infrastructure. Managed deployment architecture and multi-node orchestration are outside the first self-hosted installer milestone.

## Supported host contract

Phase 1 defines the initial production host matrix explicitly:

| Distribution | Minimum release | Architecture | Init system | Minimum kernel |
|---|---:|---|---|---:|
| Ubuntu Server LTS | 24.04 | amd64 | systemd | 6.8 |
| Debian | 13 | amd64 | systemd | 6.12 |

Additional minimums:

- Docker Engine 27.0 or newer;
- Docker Compose plugin 2.30 or newer;
- 64-bit native Linux userspace;
- persistent local storage;
- routable signaling/media addresses or explicitly configured NAT behavior;
- DNS names required by the deployment's SIP, WSS, TURN, and API/TLS configuration;
- outbound HTTPS for release retrieval, certificate operations when applicable, activation, and optional deployment check-in.

Docker Desktop, WSL, NAS operating systems, nested/containerized Leamout hosts, unsupported distributions, and `arm64` are not production-supported by the Phase 1 contract.

A future release may expand this matrix only through an explicit release contract and corresponding acceptance coverage.

## Bootstrap installer

The target bootstrap endpoint is:

```sh
curl -fsSL https://get.leamout.com/install.sh | sudo sh
```

Security-conscious operators must also be able to inspect it first:

```sh
curl -fsSL https://get.leamout.com/install.sh -o install.sh
less install.sh
sudo sh install.sh
```

The bootstrap script should remain intentionally small. Its responsibilities are limited to host detection, prerequisite verification, downloading a pinned CLI artifact, verifying release integrity, installing the CLI, creating minimum base directories, and printing the next action.

It must not contain the complete deployment implementation.

## CLI ownership

The intended operator surface is:

```text
leamout init
leamout up
leamout down
leamout status
leamout logs
leamout doctor
leamout restart
leamout update
leamout backup
leamout restore
leamout license verify --artifact license.json --keyring keyring.json
leamout license install --artifact license.json --keyring keyring.json
```

The keyring passed to these commands is trust material, not an ordinary input.
Operators must obtain it through the signed release channel or another authenticated
out-of-band channel; accepting a keyring from the same untrusted source as a license
artifact does not establish license authenticity.

The local CLI remains available even when the hosted console is unavailable.

The current release-candidate CLI implements `init`, runtime lifecycle commands,
local diagnostics, staged runtime updates, PostgreSQL/configuration backups and
restores, and offline signed-license installation/verification. Hosted activation,
certificate backup, recording backup, and automatic pre-update backup/drain remain
separate completion work.

### `leamout init`

`init` creates durable local deployment identity and configuration.

For production, the preferred flow is that the operator first creates the deployment in `console.leamout.com`, obtains a short-lived activation token, then supplies it during initialization or activation.

Example target UX:

```sh
sudo leamout init --activation-token lm_act_...
```

The activation token is not the permanent commercial license. It authorizes registration/activation and should be short-lived or one-time use.

The deployment should generate a durable identity that survives container replacement, service restarts, upgrades, and normal host maintenance.

Do not derive identity directly from unstable properties such as MAC address, IP address, hostname, Docker container ID, or disk serial number.

### `leamout up`

`up` converges the local communications runtime to its declared release and configuration.

It should eventually:

1. run host/runtime preflight checks;
2. validate required local configuration and secrets;
3. resolve and verify the signed release manifest;
4. pull immutable runtime images declared by the manifest;
5. start storage and coordination dependencies;
6. apply database migrations;
7. start signaling/media/control-plane services in dependency order;
8. wait for readiness;
9. perform product-level verification;
10. report local runtime and license state.

It does not start a local copy of `console.leamout.com`.

### `leamout status`

`status` reports local product state rather than dumping `docker compose ps`.

Expected information includes:

- installed and desired Leamout version;
- deployment ID;
- self-hosted deployment mode;
- local license state;
- last successful cloud check-in when enabled;
- API and worker health;
- PostgreSQL, Redis, and NATS health;
- OpenSIPS, RTPengine, FreeSWITCH, and Coturn health;
- certificate expiry warnings;
- public signaling/media configuration;
- pending migration/update state.

Cloud connectivity and runtime health are distinct states. A deployment can be temporarily disconnected from Leamout Cloud while its local communications runtime remains healthy.

### `leamout doctor`

`doctor` is the local diagnostic command operators should run before opening a support request.

It should validate host support, disk capacity, Docker/Compose versions, ports, DNS, network configuration, TLS, required secrets, database/coordination connectivity, telecom component readiness, API readiness, worker liveness, local license validity, and deployment registration/check-in diagnostics where applicable.

## Filesystem layout

Suggested stable host locations:

```text
/usr/local/bin/leamout             CLI executable
/etc/leamout/                      operator-controlled configuration
/etc/leamout/leamout.env           generated runtime values
/etc/leamout/certs/                deployment TLS material or references
/etc/leamout/license/              locally stored signed license material
/var/lib/leamout/                  durable Leamout-owned host state
/var/lib/leamout/backups/          local backup staging
/var/log/leamout/                  optional CLI/installer logs
```

Permissions must prevent ordinary users from reading telecom credentials, private keys, activation credentials, encryption keys, or other secrets.

## Secret handling

Production installation must not ship known/default passwords.

`leamout init` should generate cryptographically secure deployment-owned secrets such as FreeSWITCH ESL credentials, carrier credential encryption keys, TURN authentication secrets, and internal service credentials introduced later.

Secrets must not be printed in normal output or uploaded to Leamout Cloud by default.

## TLS and certificates

Development and production certificate behavior remain separate.

Development may use repository-generated self-signed certificates.

Production initialization should validate hostname/DNS prerequisites, key/certificate consistency, validity periods, chains, SIP/WSS/TURN TLS material, and custom carrier trust material when configured.

Certificate expiry should be visible locally and may be reported to the hosted console as metadata without transferring private keys.

## Networking

The self-hosted production topology should preserve separate trust zones for public signaling, public media, and private control.

PostgreSQL, Redis, NATS, worker control interfaces, and FreeSWITCH ESL must not become publicly reachable merely to support the hosted console.

Installation preflight should detect common conflicts for SIP, SIP TLS, SIP WSS, TURN/STUN, RTPengine media ranges, TURN relay ranges, and the local Leamout API listener.

There is no local hosted-console HTTP listener in the self-hosted production contract.

## Licensing and activation

A production self-hosted deployment requires a valid Leamout Commercial License.

The long-term license artifact should be cryptographically signed by Leamout Limited and locally verifiable by the deployment using public verification material.

A license should be able to represent at least:

- license identifier;
- organization identifier;
- plan/edition;
- deployment authorization;
- issuance and validity timestamps;
- expiration/perpetual state;
- deployment-count policy where applicable;
- product/feature entitlements;
- capacity entitlements;
- signature key identifier/metadata.

The deployment contains verification material, never Leamout's private signing key.

### Activation token

The activation token obtained from `console.leamout.com` should be treated as an enrollment credential, not as the durable license.

A later activation flow should exchange the token and deployment identity for signed license material and deployment-specific cloud credentials.

### Telecom-safe enforcement

License changes must never intentionally terminate an established call solely because a periodic commercial check changed state.

Where enforcement is required, prefer admission boundaries such as accepting new production work, enabling licensed features, creating licensed resources, or enforcing purchased capacity.

### Offline verification

Normal call handling must not require continuous Leamout licensing-service availability.

After successful activation, the deployment stores signed license material and verifies it locally. Online access may be needed for activation, renewal, re-hosting, or commercial-policy checks, but transient Leamout Cloud outages must not become immediate telecom outages.

## Hosted console visibility

After activation, a self-hosted deployment should appear in `console.leamout.com`.

The hosted console may eventually show non-secret fleet information such as:

```text
Deployment: production-ghana
Type: Self-Hosted
Version: 1.4.0
License: Active
Connection: Last seen 20 seconds ago
Runtime health: Healthy
Update: 1.4.1 available
```

`Connection` and `Runtime health` must remain separate concepts. If cloud check-in becomes stale, the console should report that the deployment is disconnected/last seen rather than falsely claiming the customer's telecom runtime is down.

Useful metadata may include version, component health summaries, capacity usage, active-call count, certificate expiry metadata, backup metadata, and update state.

Do not send carrier passwords, SIP digest secrets, encryption keys, TLS private keys, API tokens, webhook secrets, recording contents, or raw call content by default.

## Container image and release policy

Production installation consumes immutable, versioned artifacts.

The self-hosted production release manifest contains the compatible runtime components:

```text
server
worker
opensips
rtpengine
freeswitch
coturn
postgres
redis
nats
atlas
```

It explicitly does **not** contain:

```text
web
console
```

Those applications remain Leamout-hosted web properties. The `web` application owns both the marketing and waitlist surfaces.

Every production runtime image is resolved by OCI digest through the signed release manifest. Repository tags such as `dev`, `latest`, or human-readable semantic-version tags are not the production lockfile.

See [Self-hosted release artifacts](release-artifacts.md) for the machine-readable Phase 1 contract.

## Upgrades

The target public workflow is:

```sh
sudo leamout update
```

An update is a controlled release transition, not `git pull && docker compose up`.

The release-candidate command installs the runtime artifact already staged and
verified by the bootstrap installer for the CLI's exact version, pulls its
digest-pinned images, and converges the Compose deployment. Compatibility and disk
space validation, automatic backup/drain, readiness waiting, and rollback still
need to be added before general availability.

## Restart behavior

`leamout restart` should preserve graceful-drain behavior for telecom components where supported and must not report an unqualified success after a forced or timed-out shutdown.

## Backup and restore

A production-ready self-hosted deployment requires documented backup and restore behavior.

Backups should coordinate PostgreSQL, deployment configuration, signed local license material, certificates where appropriate, NATS durable state when required, and customer-owned recordings when selected.

The release-candidate `backup` command writes a mode-`0600` archive containing a
logical PostgreSQL dump, deployment identity, generated runtime environment, and
installed signed-license/keyring files. `restore --force` validates the archive and
deployment identity, stops the runtime, restores configuration and PostgreSQL, and
then starts the runtime. Certificates, recordings, and NATS stream state are not
yet included and must be preserved separately.

Secrets must only be included through an explicitly protected backup path.

Restore must validate backup integrity, format, release/schema compatibility, required encryption material, and deployment/license identity implications.

Legitimate re-hosting or disaster recovery may require commercial reactivation; it must not depend on unstable machine identifiers.

## Observability and supportability

A self-hosted product must be supportable without unrestricted Leamout access to customer infrastructure.

A future support bundle may include release/version information, redacted configuration, health, recent logs, migration state, network diagnostics, certificate metadata without private keys, and license metadata without activation secrets.

Redaction must exclude carrier passwords, encryption keys, TLS private keys, access tokens, webhook secrets, and customer call content by default.

## Uninstall behavior

Uninstall must distinguish software removal from data destruction.

```sh
sudo leamout uninstall
```

should preserve durable data by default.

Destructive deletion should require an explicit operation such as:

```sh
sudo leamout uninstall --purge-data
```

with clear disclosure of what will be removed.

## Failure principles

The installer and CLI should follow these principles:

1. Fail before mutation when preflight discovers an unsupported host.
2. Never silently regenerate a secret required to decrypt existing state.
3. Never overwrite production TLS material without a recoverable path.
4. Never perform destructive database behavior implicitly.
5. Never report a deployment healthy solely because containers are running.
6. Never expose secrets in normal output or support bundles.
7. Never terminate established telecom sessions merely to enforce a periodic commercial check.
8. Never require continuous Leamout Cloud availability for normal self-hosted call processing.
9. Prefer resumable/idempotent lifecycle operations.
10. Keep a local CLI/API recovery path even when the hosted console is unavailable.
11. Do not require inbound public management access for normal console visibility.
12. Do not package Leamout-hosted frontend applications into the customer runtime.

## Target first-run experience

The target commercial flow starts in the hosted console:

```text
console.leamout.com/sign-up

Create organization
Choose Self-Hosted plan
Create deployment: production-1

License: Active
Activation token: lm_act_...
```

Then on the customer's server:

```text
$ curl -fsSL https://get.leamout.com/install.sh | sudo sh

✓ Supported Linux host
✓ Leamout CLI installed

Run:

    sudo leamout init --activation-token lm_act_...
```

Initialization/activation target:

```text
$ sudo leamout init --activation-token lm_act_...

✓ Deployment identity created
✓ Activation authorized
✓ Signed license installed
✓ Secrets generated
✓ Configuration written
✓ TLS/network prerequisites validated

Deployment: production-1
License: Active
Console: https://console.leamout.com

Run:

    sudo leamout up
```

Startup target:

```text
$ sudo leamout up

✓ PostgreSQL ready
✓ Redis ready
✓ NATS ready
✓ RTPengine ready
✓ FreeSWITCH ready
✓ OpenSIPS ready
✓ Coturn ready
✓ Leamout API ready
✓ Leamout worker ready
✓ Deployment verification passed

Leamout is running.
Deployment: production-1
License: Active
```

The hosted console can then show the registered deployment without becoming part of its call path.

## Implementation phases

### Phase 1 — installation contract and release artifacts

- [x] Treat this document as the supported self-hosted installation contract.
- [x] Define the hosted-console versus self-hosted-runtime boundary.
- [x] Define supported Linux distributions and minimum versions.
- [x] Define a signed/checksummed CLI release artifact format.
- [x] Define a versioned Leamout release manifest.
- [x] Remove floating production image assumptions from the production release contract.
- [x] Exclude `web` and `console` from the self-hosted runtime manifest.
- [x] Add CI enforcement for release-artifact and runtime-boundary invariants.

### Phase 2 — bootstrap installer and CLI foundation

- [ ] Publish `https://get.leamout.com/install.sh`.
- [ ] Implement OS/architecture detection and prerequisite validation.
- [ ] Install and verify a pinned `leamout` CLI artifact.
- [x] Add `leamout init`, `up`, `down`, `status`, `logs`, and `doctor`.
- [ ] Wrap existing deployment primitives rather than duplicating their logic.

### Phase 3 — production configuration

- [x] Generate deployment-owned secrets securely.
- [x] Adopt separated signaling/media/control network topology.
- [ ] Integrate production TLS provisioning and renewal.
- [ ] Define persistent filesystem/volume ownership.
- [ ] Add host/network/port preflight checks.

### Phase 4 — commercial activation and cloud registration

- [ ] Generate durable deployment IDs and deployment key material.
- [ ] Create deployment/activation-token flow in `console.leamout.com`.
- [ ] Define the signed license document format.
- [x] Add Leamout public-key verification to the self-hosted runtime.
- [ ] Exchange short-lived activation tokens for signed local licenses.
- [ ] Register deployment ownership with Leamout Cloud.
- [ ] Add outbound authenticated deployment check-in.
- [ ] Surface self-hosted deployments in `console.leamout.com`.
- [ ] Define evaluation, expiration, grace, revocation, and re-hosting policy.
- [ ] Ensure license enforcement is safe for active telecom sessions.

### Phase 5 — lifecycle operations

- [ ] Add `leamout update` with release compatibility checks.
- [ ] Integrate graceful drain into restart/update paths.
- [x] Add `leamout backup` and `leamout restore`.
- [ ] Add support-bundle generation and deterministic redaction.
- [ ] Add non-interactive initialization.
- [ ] Add uninstall/purge behavior.

### Phase 6 — acceptance

- [ ] Create acceptance from a clean supported VM/image.
- [ ] Create a self-hosted deployment/license through the hosted console fixture.
- [ ] Install through the public bootstrap URL or equivalent release fixture.
- [ ] Initialize without repository checkout.
- [ ] Activate/register using a short-lived activation credential.
- [ ] Start the complete runtime-only production stack.
- [x] Probe API and worker readiness from the deployed runtime.
- [ ] Verify the deployment becomes visible in the hosted-console fixture.
- [ ] Verify cloud disconnection does not immediately interrupt valid call processing.
- [ ] Complete inbound and outbound BYOC calls.
- [ ] Exercise API-controlled media.
- [x] Restart the control-plane API and worker while preserving an active SIP/media session.
- [ ] Upgrade without losing authoritative control-plane state.
- [ ] Backup and restore the deployment.
- [ ] Verify diagnostics and support data do not leak secrets.

## Definition of done

The first self-hosted installer milestone is complete when a customer can:

1. sign up at `console.leamout.com`;
2. obtain a valid self-hosted subscription/license and activation credential;
3. install the Leamout CLI on a clean supported Linux host without cloning the repository;
4. initialize a durable local deployment identity;
5. activate and install locally verifiable signed license material;
6. start and verify the runtime-only Leamout stack;
7. see that deployment in `console.leamout.com`;
8. configure BYOC through supported product APIs/interfaces;
9. complete real inbound/outbound communications locally;
10. continue normal licensed telecom processing through transient Leamout Cloud unavailability;
11. inspect/recover the deployment through the local CLI; and
12. safely update, back up, and restore authoritative state.

The product succeeds when the customer experiences one Leamout account and console while retaining an operationally independent communications runtime on their own infrastructure.
