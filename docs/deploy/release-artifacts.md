# Self-hosted release artifacts

This document is the machine-facing companion to the [self-hosted installation contract](self-hosted-installation.md). The installation contract is authoritative for the customer experience; this document defines the release artifacts that make that contract implementable.

Phase 1 does **not** publish `get.leamout.com/install.sh` and does not implement the `leamout` operator CLI. Those belong to Phase 2. Phase 1 establishes the compatibility, integrity, and immutability rules that Phase 2 must consume.

## Product boundary

The self-hosted release contains the Leamout communications runtime, not Leamout's hosted customer-facing web applications.

`clients/apps/web` is the Leamout marketing website and `clients/apps/console` is the hosted dashboard at `console.leamout.com`. Neither application is installed on a customer's self-hosted server.

The intended customer flow is:

```text
console.leamout.com/sign-up
        ↓
create organization / subscription
        ↓
obtain a Leamout Commercial License and deployment activation token
        ↓
install the self-hosted Leamout runtime
        ↓
activate/register the deployment
        ↓
see the deployment in console.leamout.com
```

`console.leamout.com` owns the commercial and fleet-management experience: organizations, subscriptions, licenses, entitlements, deployment registrations, billing, release availability, and support workflows.

The self-hosted deployment owns communications execution and local telecom state. Leamout Cloud is not in the call path. A transient loss of connectivity to `console.leamout.com` must not immediately stop an otherwise valid self-hosted communications runtime; locally stored signed license material is intended to remain verifiable without continuous cloud availability.

The exact activation protocol, deployment identity format, cloud check-in mechanism, grace policy, and remote-management capabilities are later implementation phases. Phase 1 only establishes this product boundary so release artifacts do not accidentally package the hosted console.

## Supported host matrix

The first self-hosted production artifact contract intentionally supports a narrow Linux matrix.

| Distribution | Minimum release | Architecture | Init system | Minimum kernel | Status |
|---|---:|---|---|---:|---|
| Ubuntu Server LTS | 24.04 | amd64 | systemd | 6.8 | Supported |
| Debian | 13 | amd64 | systemd | 6.12 | Supported |

Additional constraints:

- Docker Engine 27.0 or newer.
- Docker Compose plugin 2.30 or newer.
- A 64-bit kernel and userspace.
- Native Linux installation. Docker Desktop, WSL, NAS distributions, and a Leamout host nested inside another container are not production-supported by this contract.
- `arm64` is reserved for a future release after it has equivalent telecom and full-stack acceptance coverage.

A future release may expand this matrix, but it must do so explicitly in its release manifest. A CLI must not infer support for an unlisted host.

## Release identity

Leamout release versions use semantic versioning without a leading `v` inside machine-readable metadata.

Examples:

```text
1.0.0
1.1.0-preview.1
```

Git tags may use the conventional `v` prefix:

```text
v1.0.0
v1.1.0-preview.1
```

Every release is bound to one exact Git commit through `source_commit`. Rebuilding a release from a different source commit requires a different release version.

Supported release channels are:

- `stable` — customer production releases;
- `preview` — explicitly opted-in prerelease builds.

There is no production `latest`, `main`, `master`, `edge`, or `dev` channel.

## CLI artifact format

Phase 1 defines the artifact shape even though the CLI implementation arrives in Phase 2.

A Linux CLI artifact is named:

```text
leamout_<version>_linux_<arch>.tar.gz
```

The first supported artifact is therefore shaped like:

```text
leamout_1.0.0_linux_amd64.tar.gz
```

The archive contains exactly the distributable CLI payload required by the installer:

```text
leamout
LICENSE
```

`leamout` must be an executable native binary for the declared OS/architecture. `LICENSE` is the Leamout Software License Agreement shipped from the source tree used to create the release.

Release archives are created deterministically where the release environment supports it: sorted entries, fixed ownership metadata, and a fixed archive timestamp. A repeated build from the same source inputs should therefore be suitable for reproducibility checks.

## Checksums and signatures

Every downloadable release artifact must be covered by SHA-256 before it is eligible for production distribution.

The canonical checksum file is:

```text
checksums.txt
```

Its records use the standard format:

```text
<64 lowercase hex characters>  <filename>
```

Production release publication must additionally provide detached Ed25519 signatures for the release metadata. The canonical signature names are:

```text
checksums.txt.minisig
release-manifest.json.minisig
```

The signature format is Minisign. The public verification key is a Leamout-controlled trust root that will be embedded or otherwise pinned by the Phase 2 installer/CLI. Private signing keys must never be stored in this repository, container images, build output, or customer installations.

The trust chain is:

```text
Leamout offline/restricted signing key
             │
             ├── signs checksums.txt
             └── signs release-manifest.json
                         │
                         ├── binds release to source commit
                         ├── binds compatible CLI artifact checksums
                         ├── binds database migration boundary
                         └── binds every production image by digest
```

A checksum verifies bytes. A signature verifies that the release metadata was authorized by Leamout. Production distribution requires both layers.

The repository's Phase 1 tooling can create deterministic archives and checksums without possessing a production signing key. Final signing is a release-publication responsibility and must use protected key material.

## Release manifest

Every self-hosted production release is described by one `release-manifest.json` document conforming to `release/manifest.schema.json` and the stricter repository validator.

The manifest carries:

- manifest schema version;
- Leamout release version and channel;
- exact source commit;
- minimum compatible CLI version;
- explicit supported host tuples;
- database migration boundary;
- CLI artifact filenames and SHA-256 values;
- immutable image references for all services required by the self-hosted runtime.

Conceptually:

```json
{
  "schema_version": 1,
  "release_version": "1.0.0",
  "channel": "stable",
  "source_commit": "0123456789abcdef0123456789abcdef01234567",
  "minimum_cli_version": "1.0.0",
  "supported_hosts": [
    {"os": "ubuntu", "version": "24.04", "arch": "amd64"},
    {"os": "debian", "version": "13", "arch": "amd64"}
  ],
  "database": {
    "migration": "039_create_idempotency.sql"
  },
  "cli_artifacts": [
    {
      "os": "linux",
      "arch": "amd64",
      "filename": "leamout_1.0.0_linux_amd64.tar.gz",
      "sha256": "..."
    }
  ],
  "images": {
    "server": "ghcr.io/leamout/server@sha256:...",
    "worker": "ghcr.io/leamout/worker@sha256:..."
  }
}
```

The fragment above is illustrative only; an actual production manifest must include every required image and full non-placeholder hashes.

## Required image set

A Phase 1 self-hosted production manifest must bind all of these runtime dependencies:

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

The frontend applications are intentionally excluded:

```text
web       -> Leamout-hosted marketing site
console   -> Leamout-hosted dashboard at console.leamout.com
waitlist  -> Leamout-hosted website application
```

The hosted console may show self-hosted deployment registration, license, version, connectivity/check-in, health summaries, usage/capacity metadata, and update availability in later phases, but its container image is never a dependency of the customer's self-hosted runtime.

A release validator may add mandatory runtime components in a future manifest schema version, but it must never silently remove a component from an existing schema version.

## Image immutability policy

Production releases must use OCI digest references:

```text
registry/repository@sha256:<64 lowercase hex characters>
```

Examples of references that are **not** production release pins:

```text
leamout/rtpengine:dev
leamout/server:latest
leamout/server:main
postgres:18-alpine
redis:8-alpine
```

Even a semantic-version tag such as `leamout/server:1.0.0` is not sufficient for the release manifest because a tag can be moved. Tags remain useful for humans and development, but the production manifest records the resolved digest.

The repository's `deploy/compose.yaml` remains a development/CI stack. It may contain build directives and human-readable tags, including the hosted web applications needed for repository-level full-stack development. It is **not** a production release lockfile. Phase 2 must render or otherwise consume a self-hosted production deployment from a validated release manifest rather than treating repository Compose services as the customer runtime definition.

## Database migration boundary

The manifest records the highest migration included in the release, for example:

```text
039_create_idempotency.sql
```

The migration value must match a migration present in `server/migrations` at the manifest's source commit. A future upgrade planner can use this boundary to reason about compatibility and rollback safety.

The manifest does not replace `atlas.sum`; Atlas remains the integrity source for the migration directory itself.

## Manifest schema compatibility

`schema_version` starts at `1`.

Consumers must fail closed on a schema version they do not understand. They must not guess field meanings or silently ignore an unsupported future schema.

Changes that alter required fields, validation semantics, or compatibility behavior require a new manifest schema version. Optional backward-compatible metadata may be added only when older consumers can safely ignore it.

## Repository layout

Phase 1 owns the following release paths:

```text
release/
  manifest.schema.json
  fixtures/
    valid-manifest.json

scripts/release/
  package-cli.sh
  validate-manifest.py

.github/workflows/
  release-artifacts.yml
```

`release/fixtures/valid-manifest.json` is synthetic CI data. It is not a Leamout release and must never be published as one.

Real release manifests should be generated from an actual release build and attached to the corresponding release publication rather than hand-authored with guessed image digests.

## CI contract

Changes to release policy must pass the Release Artifacts workflow.

CI verifies at least:

1. the manifest validator accepts the known-good runtime-only fixture;
2. the validator rejects mutable image references;
3. the validator rejects placeholder/all-zero digests;
4. the validator rejects hosted frontend images added to a self-hosted manifest;
5. the CLI packager creates the canonical archive shape;
6. the emitted SHA-256 checksum verifies the produced archive;
7. shell and Python release tooling is syntactically valid.

The same validator is intended to run in a future release-publishing workflow before any stable or preview release is published.

## Phase 1 completion criteria

Phase 1 is complete when:

- the self-hosted installation document is treated as the supported installation contract;
- the hosted console versus self-hosted runtime boundary is explicit;
- the supported Linux distributions, versions, architecture, Docker, and kernel minimums are explicit;
- the CLI release archive, checksum, and detached-signature formats are defined;
- the release manifest has a versioned machine-readable schema and validation rules;
- production image references are required to be immutable OCI digests;
- `web`, `console`, and `waitlist` are excluded from the customer self-hosted runtime;
- development Compose tags/services are explicitly separated from production release pins;
- CI enforces the artifact and manifest invariants.

Publishing `get.leamout.com/install.sh`, implementing the real CLI, implementing the license activation/register/check-in protocol, installing prerequisites, and adding operator commands remain later-phase work.
