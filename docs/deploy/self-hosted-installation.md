# Self-hosted Leamout installation

This document defines the target installation, activation, upgrade, and operating contract for a self-hosted Leamout deployment.

It is intentionally a product and deployment contract before it is an implementation guide. The public installer and `leamout` CLI described here may be introduced incrementally, but their behavior should converge on this document rather than exposing the underlying Docker Compose implementation directly to operators.

## Goals

A new operator should be able to provision a supported Linux host and reach a healthy Leamout installation through one public entry point:

```sh
curl -fsSL https://get.leamout.com/install.sh | sudo sh
```

The bootstrap installer installs the Leamout command-line interface and verifies that the host can support Leamout. The actual deployment lifecycle belongs to the CLI:

```sh
sudo leamout init
sudo leamout up
sudo leamout status
```

The target product experience is:

```text
Download / install Leamout
        ↓
Initialize the host
        ↓
Configure networking, TLS, and storage
        ↓
Start in development/evaluation mode when permitted
        ↓
Activate a Leamout Commercial License
        ↓
Operate a production self-hosted deployment
```

The installer must not require operators to understand OpenSIPS, RTPengine, FreeSWITCH, Coturn, NATS, Redis, PostgreSQL, Atlas, or Docker Compose command details for normal installation and operation.

## Product model

Leamout is proprietary software. Availability of an installer or software artifact does not itself grant production-use rights.

Development, testing, or evaluation use is governed by the Leamout Software License Agreement. Production use, including a self-hosted production deployment, requires a valid Leamout Commercial License issued by Leamout Limited.

Licensing is therefore separate from installation:

```text
software distribution
        ↓
installation
        ↓
local deployment identity
        ↓
license activation
        ↓
production entitlements
```

A customer should not need to contact the licensing service for every telecom operation. A self-hosted deployment should be able to verify locally stored, cryptographically signed license material using a Leamout public verification key.

## Supported deployment tiers

Leamout should expose three deployment tiers over time.

### 1. Developer deployment

The repository remains the primary development workflow:

```sh
make certs
make preflight
make up
make verify
```

This mode is intended for contributors, automated acceptance tests, and local development. It may use self-signed certificates and development defaults that are not appropriate for production.

### 2. Self-hosted production deployment

The first supported customer deployment target should be a single Linux host using the Leamout CLI to manage a pinned Docker Compose stack.

The CLI owns configuration generation, secrets, image versions, migrations, TLS setup, health verification, upgrades, backups, restoration, and license activation.

Docker Compose remains an implementation detail and recovery escape hatch, not the primary product interface.

### 3. Enterprise / managed deployment

A future multi-node deployment may independently scale and operate:

- API nodes
- workers
- OpenSIPS signaling nodes
- RTPengine media nodes
- FreeSWITCH media workers
- Coturn nodes
- PostgreSQL
- Redis
- NATS

This tier is deliberately outside the first self-hosted installer contract. The single-node deployment should become reliable and upgrade-safe before Leamout introduces a second production orchestration model.

## Supported host contract

The first production release should intentionally support a narrow host matrix.

Recommended initial support:

- 64-bit Linux
- Ubuntu LTS and Debian stable
- systemd
- Docker Engine with the Compose plugin
- persistent local storage
- a routable public signaling address
- a routable public media address or correctly configured NAT mapping
- DNS names for the console/API and TLS SIP/WSS/TURN endpoints
- outbound HTTPS access for installation, image retrieval, certificate issuance, updates, and optional license activation

The installer should reject unsupported environments instead of attempting an unknown installation.

Containerized Leamout inside another container, Docker Desktop, arbitrary NAS operating systems, and unsupported distributions may work for development but should not be represented as production-supported until they have explicit acceptance coverage.

## Bootstrap installer

The bootstrap endpoint is:

```sh
curl -fsSL https://get.leamout.com/install.sh | sudo sh
```

Security-conscious operators must also be able to download and inspect the script before running it:

```sh
curl -fsSL https://get.leamout.com/install.sh -o install.sh
less install.sh
sudo sh install.sh
```

The bootstrap script should remain intentionally small. It should not contain the complete deployment implementation.

Its responsibilities are limited to:

1. Detect the host operating system and architecture.
2. Refuse unsupported platforms with an actionable error.
3. Verify required base utilities.
4. Verify or install the supported Docker Engine / Compose prerequisites according to the release policy.
5. Download a pinned Leamout CLI artifact over HTTPS.
6. Verify the CLI artifact signature or published checksum before installation.
7. Install the CLI into a stable executable path such as `/usr/local/bin/leamout`.
8. Create only the minimum directories required by the CLI.
9. Print the next command:

```sh
sudo leamout init
```

The script must be safe to run more than once. Re-running it should upgrade or repair the CLI according to explicit rules rather than destroy deployment state.

## CLI ownership

Once installed, operators should interact with Leamout through the `leamout` CLI.

The initial command surface should converge on:

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
leamout license status
leamout license activate
```

Commands should wrap stable product operations rather than expose Docker Compose syntax.

### `leamout init`

`init` creates a deployment identity and writes the local deployment configuration.

Interactive initialization may collect:

```text
Deployment name
Primary application hostname
Public SIP hostname/address
Public media address
TURN hostname/address
Administrator email
Certificate mode
Storage locations
License key or activation token (optional during evaluation)
```

The command should support a non-interactive form for automated deployment later, but interactive initialization is the preferred first user experience.

`init` must be idempotent. Once a deployment has been initialized, running the command again should present the existing configuration and require an explicit reconfiguration operation before replacing security-sensitive state.

### `leamout up`

`up` starts or converges the local installation to its declared version and configuration.

It should:

1. run local preflight checks;
2. validate required secrets and certificates;
3. render deployment configuration;
4. pull or build only explicitly supported artifacts;
5. start storage and coordination dependencies;
6. apply database migrations;
7. start telecom and application services in dependency order;
8. wait for readiness;
9. run post-start verification;
10. report the console/API endpoint and license state.

### `leamout status`

`status` must report product state rather than merely dump `docker compose ps`.

Expected information includes:

- installed Leamout version
- desired version
- deployment ID
- deployment mode
- license state
- API health
- worker health
- PostgreSQL health
- Redis health
- NATS health
- OpenSIPS health
- RTPengine health
- FreeSWITCH health
- Coturn health
- certificate expiry warnings
- public signaling/media configuration
- pending migration or update state

### `leamout doctor`

`doctor` is the diagnostic command operators should run before opening a support request.

Checks should include:

- supported OS and kernel characteristics
- disk capacity and filesystem writability
- Docker daemon and Compose version
- required TCP/UDP port conflicts
- DNS resolution
- public/private network configuration
- TLS certificate readability and expiry
- required secret availability
- PostgreSQL connectivity
- Redis connectivity
- NATS connectivity
- OpenSIPS readiness
- RTPengine readiness
- FreeSWITCH ESL readiness
- Coturn readiness
- API readiness
- worker liveness
- license validity and deployment binding

The command should distinguish warnings from errors and provide machine-readable output later for support automation.

## Filesystem layout

The installation should use conventional host locations and avoid storing durable state inside the downloaded installer.

Suggested layout:

```text
/usr/local/bin/leamout             CLI executable
/etc/leamout/                      operator-controlled configuration
/etc/leamout/leamout.env           generated environment/config values
/etc/leamout/certs/                deployment TLS material or references
/etc/leamout/license/              locally stored signed license material
/var/lib/leamout/                  durable Leamout-owned host state
/var/lib/leamout/backups/          local backup staging
/var/log/leamout/                  optional CLI/installer logs
```

Container volumes remain the durable stores for PostgreSQL, Redis persistence, NATS state, recordings, and other runtime state unless the production storage contract explicitly changes.

Permissions must prevent ordinary users from reading credentials, private keys, commercial license material containing sensitive metadata, or telecom secrets.

## Secret handling

Production installation must not ship known/default passwords.

`leamout init` should generate cryptographically secure values for deployment-owned secrets such as:

- FreeSWITCH ESL credentials
- carrier credential encryption keys
- TURN shared secrets
- internal service authentication secrets introduced later

Secrets must not be printed to normal command output or written to world-readable files.

The first installer can use root-owned files mounted into containers. The configuration interface should leave room for future external secret providers such as Vault, cloud KMS/secret managers, or orchestration-native secrets without changing telecom domain APIs.

## TLS and certificates

Local development and production certificate behavior must stay separate.

### Development

Self-signed certificates may be generated through existing development tooling.

### Production

The CLI should support a production certificate workflow backed by the existing Let's Encrypt/Certbot deployment tooling where applicable.

Production initialization must validate:

- hostname ownership / DNS prerequisites
- certificate/key pair consistency
- certificate validity period
- required certificate chain
- SIP TLS and WSS certificate availability
- TURN TLS certificate availability
- configured carrier CA bundle when mTLS or custom carrier trust is used

Automatic renewal must not silently replace certificates without ensuring the affected runtime reload/restart operation succeeds.

Certificate expiry should become visible through `leamout status` before it becomes an outage.

## Networking

The production installer must preserve Leamout's trust-zone model rather than placing all services onto one shared network.

The intended production topology separates:

```text
public signaling
public media
private control
```

Only components that must bridge a boundary should be attached to more than one trust zone. PostgreSQL, Redis, NATS, the API, workers, and internal FreeSWITCH control interfaces should not be reachable from public telecom networks.

Installation preflight should detect common host conflicts for at least:

- SIP UDP/TCP
- SIP TLS
- SIP WebSocket/WSS
- TURN/STUN
- RTPengine media ranges
- TURN relay ranges
- console/API HTTP(S) listeners

The exact public ports remain versioned deployment configuration rather than assumptions embedded in customer automation.

## Licensing and activation

Installation and commercial activation are separate operations.

The CLI should expose:

```sh
leamout license status
leamout license activate <activation-token>
```

The long-term license artifact should be cryptographically signed by Leamout Limited and locally verifiable by the deployment.

A license should be able to represent at least:

- license identifier
- organization identifier
- plan/edition
- deployment authorization
- issued-at timestamp
- not-before timestamp when needed
- expiration or perpetual status
- maximum deployment count when applicable
- product entitlements
- capacity entitlements
- signature metadata / key identifier

The signed document must not contain secrets required to mint another valid license.

The deployment contains only verification material, never the Leamout private signing key.

### Evaluation state

If Leamout offers an evaluation mode, it should be explicit and visible in the console and CLI.

Evaluation policy is a commercial decision and may include constraints such as:

- expiration date
- limited concurrent calls
- limited projects or users
- test-only destinations
- visible evaluation state

The installer must not invent these policies. They should come from the commercial/license domain.

### Production state

A production self-hosted deployment requires a valid Leamout Commercial License.

Runtime enforcement must be telecom-safe. License transitions must never intentionally terminate an already established call solely because a periodic license check changed state.

Where enforcement is required, prefer boundaries such as:

- admission of new production calls
- creation of new licensed resources
- activation of licensed features
- enforcement of licensed capacity
- administrative configuration changes

The exact expiration and grace-period behavior must be a documented commercial policy rather than hidden behavior in telecom components.

### Offline verification

Normal call handling should not require continuous connectivity to a Leamout licensing service.

After activation, the deployment stores signed license material and verifies it locally. Online communication may be required for activation, renewal, re-hosting, or periodic policy depending on the commercial product, but a transient Leamout control-service outage must not become an immediate customer telecom outage.

## Container image and release policy

Production installation should consume immutable, versioned artifacts.

Avoid floating tags such as `latest` in the production deployment contract.

A release should define a manifest containing the compatible versions of:

- Leamout server
- Leamout worker
- console/web applications
- OpenSIPS image/configuration
- RTPengine image/configuration
- FreeSWITCH image/configuration
- Coturn image/configuration
- migration set
- minimum supported CLI version

The CLI should resolve a Leamout release to this manifest and deploy the exact versions declared by it.

## Upgrades

The public workflow should be:

```sh
sudo leamout update
```

An upgrade must be treated as a controlled operation, not `git pull && docker compose up`.

The command should:

1. resolve the target release;
2. verify signatures/checksums;
3. validate upgrade compatibility;
4. validate free disk space;
5. run a pre-upgrade backup when required;
6. download artifacts before disruption;
7. drain telecom nodes/components when required;
8. apply migrations according to compatibility rules;
9. roll or restart application components;
10. wait for readiness;
11. verify the deployment;
12. record the installed version.

Leamout must document which downgrades are supported. Database migrations that make rollback unsafe should be identified before an upgrade starts.

## Restart behavior

`leamout restart` should preserve the existing graceful-drain contract.

Telecom services must stop accepting new work before they are terminated when the component supports draining. The CLI should surface a timeout or forced-shutdown condition rather than report an unqualified successful restart.

Normal application restarts must not corrupt durable call state. Existing lifecycle reconciliation remains responsible for rebuilding runtime coordination after worker or media-control restarts.

## Backup contract

A customer cannot consider a deployment production-ready without a documented backup and restore path.

`leamout backup` should eventually capture or coordinate:

- PostgreSQL
- persistent deployment configuration
- locally stored signed license material
- certificate configuration/material when appropriate
- NATS durable state when required by recovery semantics
- recordings when the deployment owns their storage and the operator requests them

Secrets should be backed up only through an explicitly protected backup path.

Backups must be versioned with enough metadata to determine which Leamout release and schema created them.

## Restore contract

`leamout restore` should refuse unsafe restoration automatically.

A restore flow should verify:

- backup integrity
- backup format version
- target Leamout compatibility
- database schema compatibility
- deployment/license identity implications
- required encryption material

Re-hosting a licensed deployment onto new infrastructure may require commercial reactivation depending on the license policy. That must be explicit rather than inferred from machine identifiers.

## Deployment identity

Do not bind the commercial license directly to unstable host properties such as a MAC address, Docker container ID, hostname, or disk serial number.

`leamout init` should generate a durable deployment identifier. The deployment identity is persisted with the installation and represented in commercial deployment records.

This supports legitimate host replacement and disaster recovery while still allowing Leamout to enforce purchased deployment counts through activation/re-hosting policy.

## Observability and supportability

A self-hosted product must be supportable without unrestricted access to customer infrastructure.

The CLI should eventually expose a support bundle command such as:

```sh
leamout support-bundle
```

The bundle may include:

- Leamout version and release manifest
- redacted configuration
- service health
- recent service logs
- migration status
- network diagnostics
- certificate metadata without private keys
- license metadata without activation secrets

The command must apply deterministic redaction rules and must never collect carrier passwords, encryption keys, private TLS keys, access tokens, or unredacted customer call content by default.

## Non-interactive and automated installs

After the interactive flow is stable, `leamout init` should support automation through a configuration file and explicit flags.

For example:

```sh
sudo leamout init --config /root/leamout-install.yaml --non-interactive
```

The configuration schema should be versioned. Secrets should be referenceable from protected files rather than required inline in a world-readable YAML document.

This creates a path toward Terraform, Ansible, cloud-init, image baking, and managed deployment automation without making those systems part of the initial product.

## Uninstall behavior

Uninstall must distinguish software removal from data destruction.

A command such as:

```sh
sudo leamout uninstall
```

should stop and remove runtime components while preserving durable data by default.

Destructive deletion must require an explicit operation, for example:

```sh
sudo leamout uninstall --purge-data
```

and should clearly list the data that will be deleted before proceeding.

## Failure principles

The installer and CLI should follow these principles:

1. Fail before mutation when preflight discovers an unsupported host.
2. Never silently regenerate a secret required to decrypt existing state.
3. Never overwrite a valid production certificate without a recoverable path.
4. Never run destructive database migration behavior implicitly.
5. Never report a deployment healthy solely because containers are running.
6. Never expose secret values in normal output or support bundles.
7. Never terminate established telecom sessions merely to enforce a periodic commercial license check.
8. Never require continuous Leamout cloud availability for normal self-hosted call processing.
9. Prefer resumable/idempotent installation steps over one-shot shell behavior.
10. Preserve an operator escape hatch for manual recovery when automation fails.

## Target first-run experience

The desired experience is deliberately simple:

```text
$ curl -fsSL https://get.leamout.com/install.sh | sudo sh

✓ Supported Linux host
✓ Docker available
✓ Leamout CLI installed

Run:

    sudo leamout init
```

Initialization:

```text
$ sudo leamout init

Deployment name: production-1
Application hostname: voice.example.com
Public SIP hostname: sip.example.com
Public media address: 203.0.113.10
TURN hostname: turn.example.com
Administrator email: ops@example.com
License activation token (optional):

✓ Deployment identity created
✓ Secrets generated
✓ Configuration written
✓ TLS prerequisites validated

Run:

    sudo leamout up
```

Startup:

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

Console: https://voice.example.com
License: Evaluation
```

Production activation:

```text
$ sudo leamout license activate <activation-token>

✓ License signature verified
✓ Organization authorized
✓ Deployment activated
✓ Production entitlements installed

License: Active
```

## Implementation phases

### Phase 1 — installation contract and release artifacts

- [ ] Treat this document as the supported installation contract.
- [ ] Define supported Linux distributions and minimum versions.
- [ ] Define a signed/checksummed CLI release artifact format.
- [ ] Define a versioned Leamout release manifest.
- [ ] Remove floating production image assumptions.

### Phase 2 — bootstrap installer and CLI foundation

- [ ] Publish `https://get.leamout.com/install.sh`.
- [ ] Implement OS/architecture detection and prerequisite validation.
- [ ] Install and verify a pinned `leamout` CLI artifact.
- [ ] Add `leamout init`, `up`, `down`, `status`, `logs`, and `doctor`.
- [ ] Wrap existing deployment scripts rather than duplicating their logic.

### Phase 3 — production configuration

- [ ] Generate deployment-owned secrets securely.
- [ ] Adopt the separated signaling/media/control network topology.
- [ ] Integrate production TLS provisioning and renewal.
- [ ] Define persistent filesystem and volume ownership.
- [ ] Add host/network/port preflight checks.

### Phase 4 — commercial activation

- [ ] Generate durable deployment IDs.
- [ ] Define the signed license document format.
- [ ] Add Leamout public-key verification to self-hosted runtime code.
- [ ] Add `leamout license status` and `leamout license activate`.
- [ ] Define evaluation, expiration, grace-period, and re-hosting policy.
- [ ] Ensure license enforcement is safe for active telecom sessions.

### Phase 5 — lifecycle operations

- [ ] Add `leamout update` with release compatibility checks.
- [ ] Integrate graceful drain into restart/update paths.
- [ ] Add `leamout backup` and `leamout restore`.
- [ ] Add support-bundle generation and deterministic redaction.
- [ ] Add non-interactive initialization.
- [ ] Add uninstall/purge behavior.

### Phase 6 — acceptance

- [ ] Create an acceptance test from a clean supported VM/image.
- [ ] Install through the public bootstrap URL or equivalent fixture.
- [ ] Initialize without repository checkout.
- [ ] Start the complete production stack.
- [ ] Activate a test license through the public flow.
- [ ] Complete inbound and outbound BYOC calls.
- [ ] Exercise API-controlled media.
- [ ] Upgrade to the next fixture release without losing control-plane state.
- [ ] Backup and restore the deployment.
- [ ] Verify failure output and support diagnostics do not leak secrets.

## Definition of done

The first self-hosted installer milestone is complete when a customer can take a clean supported Linux host and, without cloning the Leamout repository or manually operating Docker Compose:

1. install the Leamout CLI through the documented bootstrap command;
2. initialize a deployment;
3. configure production TLS and networking;
4. start and verify the complete Leamout stack;
5. activate a Leamout Commercial License;
6. configure a BYOC carrier through public product interfaces;
7. complete real inbound and outbound calls;
8. inspect deployment and license health through the CLI;
9. safely restart and update the deployment; and
10. recover the authoritative state through the documented backup/restore workflow.

The installer is successful when Leamout feels like one product to the operator even though it is implemented by multiple control-plane and telecom components.
