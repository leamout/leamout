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
      BYOC + Managed                  BYOC + optional Managed
```

Cloud and Self-Hosted are not forks and must not duplicate telecom, identity, tenancy, or shared platform code.

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

## Commercial boundaries

Cloud is prepaid PAYG.

Self-Hosted BYOC is governed by the enterprise software license and must not require wallet balance for customer-owned carrier usage.

Self-Hosted Managed combines the enterprise software license with prepaid authorization for Leamout-managed provider obligations.

```text
Cloud
    → prepaid PAYG

Self-Hosted + BYOC
    → enterprise license
    → customer carrier
    → no Leamout managed-carrier wallet charge

Self-Hosted + Managed
    → enterprise license
    + prepaid managed-usage wallet
```

The invariant is:

> PAYG protects Leamout-owned provider exposure. It is not a prerequisite for Self-Hosted BYOC platform operation.

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

deploy/
├── cloud/
└── self-hosted/
```

## Release rule

Monorepo does not mean one release artifact.

Cloud and Self-Hosted may share source packages while producing different binaries, images, manifests, and operational assets.

Self-Hosted release artifacts must contain only what is required to operate Leamout in customer infrastructure. Cloud operational components must not be included merely because they live in the same repository.

## Refactor order

1. Keep command and runtime packages aligned one-to-one.
2. Keep shared behavior in domain and platform packages, not in a generic application layer.
3. Stop constructing the full Commercial module for every runtime.
4. Ensure Self-Hosted BYOC works without wallet, checkout, or payment-provider dependencies.
5. Attach wallet authorization only to managed-provider paths.
6. Split Cloud and Self-Hosted release composition where needed.
7. Keep acceptance coverage proving Cloud, Self-Hosted BYOC, and Self-Hosted Managed independently.
