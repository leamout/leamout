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

`cmd` and `internal/runtime` mirror each other. A runtime package is the executable-specific composition behind one command.

```text
cmd/cloud                 → internal/runtime/cloud
cmd/cloud-worker          → internal/runtime/cloudworker
cmd/selfhosted            → internal/runtime/selfhosted
cmd/selfhosted-worker     → internal/runtime/selfhostedworker
cmd/backoffice            → internal/runtime/backoffice
cmd/leamout               → internal/runtime/leamout
```

Generic `runtime/server` and `runtime/worker` packages are not used. Shared API and worker implementation lives under `internal/app`, outside the executable runtime namespace.

```text
internal/app/server
internal/app/worker
```

Runtime packages choose the appropriate shared application composition without making shared telecom, identity, tenancy, or platform packages depend on Cloud or Self-Hosted policy.

## Dependency rule

Dependencies flow from runtime composition into shared application and domain packages, never the reverse.

```text
runtime/cloud ───────────────┐
runtime/selfhosted ──────────┼──→ internal/app + shared domains
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
    ├── app/
    │   ├── server/
    │   └── worker/
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

This layout separates executable composition from shared application and domain implementation. Files should move only when the move establishes a real ownership or release boundary.

## Release rule

Monorepo does not mean one release artifact.

Cloud and Self-Hosted may share source packages while producing different binaries, images, manifests, and operational assets.

Self-Hosted release artifacts must contain only what is required to operate Leamout in customer infrastructure. Cloud operational components must not be included merely because they live in the same repository.

## Refactor order

1. Keep command and runtime packages aligned one-to-one.
2. Keep shared API/worker implementation outside `runtime`.
3. Stop constructing the full Commercial module for every runtime.
4. Ensure Self-Hosted BYOC works without wallet, checkout, or payment-provider dependencies.
5. Attach wallet authorization only to managed-provider paths.
6. Split Cloud and Self-Hosted release composition where needed.
7. Keep acceptance coverage proving Cloud, Self-Hosted BYOC, and Self-Hosted Managed independently.
