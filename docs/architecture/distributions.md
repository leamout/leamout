# Cloud and Self-Hosted distributions

Leamout is one communications platform maintained in one monorepo and delivered through separate Cloud and Self-Hosted compositions.

The repository must share product behavior while keeping deployment and commercial policy at composition boundaries.

## Canonical model

```text
                         LEAMOUT MONOREPO
                               │
                    shared platform code
                               │
             ┌─────────────────┴─────────────────┐
             │                                   │
        CLOUD COMPOSITION                  SELF-HOSTED COMPOSITION
             │                                   │
      Leamout infrastructure               Customer infrastructure
             │                                   │
       prepaid PAYG                      enterprise license
             │                                   │
      BYOC + Managed                         BYOC only
                                                  │
                                          customer selects
                                          any carrier,
                                          including Leamout
```

Cloud and Self-Hosted are not forks and must not duplicate telecom, identity, tenancy, or shared platform code.

There is no separate Self-Hosted + Managed distribution. If a self-hosted customer chooses Leamout Carrier, the deployment remains Self-Hosted + BYOC and Leamout is acting as the selected telecom provider.

## Shared product core

The following belong to the shared platform and must not depend on Cloud or Self-Hosted policy:

- identity and authentication
- tenancy and credentials
- call control
- SIP routing
- trunks and carriers
- numbers
- recordings
- conferences
- realtime communications
- webhooks and audit
- provider adapters and telecom integrations

Shared telecom code must not require a prepaid wallet merely because it is running in Self-Hosted.

## Connectivity boundary

The runtime composition determines which connectivity policies are available:

```text
Self-Hosted
    → organization-scoped carrier connections
    → BYOC only

Cloud
    → organization-scoped carrier connections
    → BYOC

Cloud
    → platform-scoped carrier connections
    → Managed
```

An organization-scoped carrier connection remains BYOC even when `provider = leamout`.

Platform-scoped carrier connections are Leamout-owned upstream resources used behind Cloud Managed. Provider credentials, physical SIP endpoints, wholesale resources, and provider-specific topology stay inside the Leamout-operated boundary.

Trunks and endpoints inherit connectivity ownership from the carrier connection. Distribution code must not create an independent BYOC/managed type on physical endpoints.

## Commercial boundaries

Cloud is prepaid PAYG.

Self-Hosted is governed by the enterprise software license and must not require a Leamout wallet for the customer's carrier usage.

If a self-hosted customer separately purchases telecom service from Leamout Carrier, that carrier relationship may have its own prepaid telecom balance or carrier billing. It does not create a different Self-Hosted software distribution.

```text
Cloud + BYOC
    → prepaid Cloud/platform charges
    → customer-selected carrier relationship

Cloud + Managed
    → prepaid Cloud/platform charges
    + Leamout-managed telecom usage

Self-Hosted + BYOC
    → enterprise software license
    → customer-selected carrier
       ├── third-party carrier
       └── Leamout Carrier
```

The invariant is:

> Software deployment mode and carrier commercial relationship are separate concerns. Choosing Leamout Carrier must not turn a self-hosted runtime into a managed distribution.

## Runtime ownership

`cmd` and `internal/runtime` mirror each other. A runtime package owns the implementation and composition behind one executable.

```text
cmd/cloud                 → internal/runtime/cloud
cmd/cloud-worker          → internal/runtime/cloudworker
cmd/selfhosted            → internal/runtime/selfhosted
cmd/selfhosted-worker     → internal/runtime/selfhostedworker
cmd/backoffice            → internal/runtime/backoffice
cmd/leamout               → internal/runtime/leamout
```

There is no generic `internal/app`, `runtime/server`, or `runtime/worker` layer.

Shared behavior belongs in the existing reusable domain and platform packages such as `identity`, `tenancy`, `telecom`, `commercial`, `integrations`, and `platform`.

Runtime packages may contain executable-specific wiring, routes, health checks, and worker orchestration. That composition glue may differ between Cloud and Self-Hosted while the actual product capabilities remain shared.

Worker runtime composition uses the same four-file contract in both distributions:

```text
consumers.go  asynchronous inputs and consumer construction
health.go     liveness, readiness, and component health state
modules.go    runtime dependency and background-module composition
worker.go     worker lifecycle, component registry, and shutdown
```

Do not split individual jobs into one-file wrappers inside a runtime package. Job behavior remains in the domain package that owns it; runtime worker files only compose and run those jobs.

## Dependency rule

Dependencies flow from runtime composition into shared domains and platform packages, never the reverse.

```text
runtime/cloud ───────────────┐
runtime/selfhosted ──────────┼──→ shared domains + platform
runtime/cloudworker ─────────┤
runtime/selfhostedworker ────┘
```

Do not introduce dependencies such as:

```text
telecom → cloud
telecom → self-hosted
identity → wallet
tenancy → licensing
Self-Hosted BYOC → wallet
```

## Repository target

```text
server/
├── cmd/
│   ├── cloud/
│   ├── cloud-worker/
│   ├── selfhosted/
│   ├── selfhosted-worker/
│   ├── backoffice/
│   └── leamout/
└── internal/
    ├── commercial/
    ├── identity/
    ├── integrations/
    ├── platform/
    │   └── middleware/
    ├── tenancy/
    ├── telecom/
    └── runtime/
        ├── cloud/
        ├── cloudworker/
        ├── selfhosted/
        ├── selfhostedworker/
        ├── backoffice/
        └── leamout/

containers/
├── opensips/
├── freeswitch/
├── rtpengine/
├── coturn/
└── nats/

deploy/
├── cloud/
│   ├── compose.yaml
│   └── README.md
└── self-hosted/
    ├── compose.yaml
    ├── .env.example
    ├── install.sh
    ├── update.sh
    ├── uninstall.sh
    └── README.md
```

Each deployment directory owns its composition, environment contract, and operator documentation.
Reusable container image sources live under `containers/`; neither distribution reaches into the
other distribution's tree. Distribution-specific policy remains explicit—for example, OpenSIPS has
separate Cloud and Self-Hosted configurations—while common image mechanics are maintained once.
Shared product behavior continues to belong in the server domain and platform packages.

## Release rule

Monorepo does not mean one release artifact.

Cloud and Self-Hosted may share source packages while producing different binaries, images, manifests, and operational assets.

Self-Hosted release artifacts must contain only what is required to operate Leamout in customer infrastructure. Cloud operational components must not be included merely because they live in the same repository.

## Refactor order

1. Keep command and runtime packages aligned one-to-one.
2. Keep shared behavior in domain and platform packages, not in a generic application layer.
3. Keep Self-Hosted free of Commercial, wallet, payment-provider, and managed-provider runtime dependencies.
4. Make every organization-scoped carrier connection BYOC regardless of provider slug.
5. Keep Cloud Managed on platform-scoped carrier resources and Cloud-only policy paths.
6. Remove customer-facing self-hosted managed trunk/endpoint states and stale naming.
7. Split Cloud and Self-Hosted release composition where needed.
8. Keep acceptance coverage proving Self-Hosted BYOC with third-party carriers, Self-Hosted BYOC with Leamout Carrier, Cloud BYOC, and Cloud Managed independently.

## Deployment compositions

The source tree exposes independent Compose entry points:

```text
deploy/cloud/compose.yaml         Cloud API, Cloud worker, Backoffice, managed providers, and PAYG
deploy/self-hosted/compose.yaml   Self-Hosted API and worker with customer-selected BYOC connectivity
```

`make` defaults to the Self-Hosted composition. Operators and CI can select Cloud explicitly with
`COMPOSE_FILE=deploy/cloud/compose.yaml`. The Self-Hosted image contains the runtime server, worker,
and `leamout` lifecycle CLI; it does not contain Backoffice. Its API and worker dependency graphs
exclude Commercial, payment adapters, managed-provider jobs, provider diagnostics, and managed SIP/CDR
internal handlers. Offline software-license validation remains in the `leamout` lifecycle boundary and
does not introduce wallet authorization into the communications runtime.
