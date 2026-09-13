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

## Composition roots

Cloud and Self-Hosted must have explicit runtime composition roots.

Shared packages provide capabilities. Composition roots decide which capabilities are constructed, configured, routed, and shipped.

Conceptually:

```text
shared domains
    │
    ├── Cloud composition
    │     ├── catalog
    │     ├── wallets
    │     ├── checkout
    │     ├── payments
    │     ├── usage
    │     ├── BYOC
    │     └── managed providers
    │
    └── Self-Hosted composition
          ├── licensed runtime
          ├── BYOC
          └── managed-provider extension when configured
```

Self-Hosted-only operational capabilities include the `leamout` operator CLI, deployment identity, offline license verification, backup/restore, host doctor, runtime installation, and update lifecycle.

Cloud-only operational capabilities include Leamout-operated deployment and fleet concerns. Those must not become dependencies of shared product domains.

## Dependency rule

Dependencies flow from distribution composition into shared domains, never the reverse.

```text
Cloud ────────┐
              ├──→ shared product domains
Self-Hosted ──┘
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

The repository should evolve toward this shape without duplicating domains:

```text
server/
├── cmd/
│   ├── cloud/
│   ├── selfhosted/
│   ├── worker/
│   ├── backoffice/
│   └── leamout/
└── internal/
    ├── commercial/
    ├── identity/
    ├── tenancy/
    ├── telecom/
    ├── integrations/
    └── runtime/
        ├── cloud/
        ├── selfhosted/
        └── shared server/worker runtime code

deploy/
├── cloud/
└── self-hosted/
```

This is a target composition boundary, not a requirement to move files solely for appearance. Existing paths should move only when the move establishes a real ownership or release boundary.

## Release rule

Monorepo does not mean one release artifact.

Cloud and Self-Hosted may share source packages while producing different binaries, images, manifests, and operational assets.

Self-Hosted release artifacts must contain only what is required to operate Leamout in customer infrastructure. Cloud operational components must not be included merely because they live in the same repository.

## Refactor order

1. Separate Cloud and Self-Hosted server/worker composition.
2. Stop constructing the full Commercial module for every runtime.
3. Ensure Self-Hosted BYOC works without wallet, checkout, or payment-provider dependencies.
4. Attach wallet authorization only to managed-provider paths.
5. Split Cloud and Self-Hosted executable/release composition where needed.
6. Keep acceptance coverage proving Cloud, Self-Hosted BYOC, and Self-Hosted Managed independently.
