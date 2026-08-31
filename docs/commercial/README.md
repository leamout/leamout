# Commercial Domain

The commercial domain turns Leamout usage and self-hosted deployments into durable commercial state without making an external billing provider the source of truth.

Leamout uses `organizations` as the canonical tenant and commercial account. There is no separate commercial customer identity in the current model.

## Domain map

```text
organization
    │
    ├── subscription
    │      └── plan
    │            └── product
    │
    ├── entitlements
    │
    ├── licenses
    │      └── deployments
    │
    ├── usage_events
    │      └── meters
    │
    ├── invoices
    │      └── invoice_items
    │
    └── payments

plan + meter
    └── carrier_rates
```

The current commercial tables are:

```text
products
plans
subscriptions
licenses
entitlements
deployments
meters
carrier_rates
usage_events
invoices
invoice_items
payments
```

## Self-hosted commercial flow

```text
organization
    ↓
subscription
    ↓
plan
    ↓
effective entitlements
    ↓
license
    ↓
deployment(s)
```

The subscription describes the commercial relationship. Entitlements describe what is allowed. Commercial state resolves the effective operational view. A license carries the deployable commercial state for self-hosted installations, including activated deployments under that license.

## Managed / CPaaS commercial flow

```text
telecom domain event
        ↓
      meter
        ↓
   usage_event
        ↓
 carrier rate resolution
        ↓
   invoice_item
        ↓
      invoice
        ↓
      payment
```

Telecom domains produce authoritative usage. The commercial domain meters and rates it, snapshots charges into invoice items, and reconciles payments.

## Strict module structure

Commercial modules use the following file convention. A file exists only when the module owns that responsibility.

| File | Add it when |
| --- | --- |
| `model.go` | The module defines domain models, states, errors, commands, inputs, or outputs. |
| `repository.go` | The module owns durable persistence or database queries. |
| `service.go` | The module contains business rules, use cases, orchestration, or transaction boundaries. |
| `validation.go` | The module has reusable domain or input validation. |
| `handler.go` | The module exposes HTTP endpoints. |
| `routes.go` | The module registers HTTP routes. It normally exists together with `handler.go`. |
| `consumer.go` | The module consumes asynchronous events or messages. Events entering the module. |
| `publisher.go` | The module publishes asynchronous events or messages. Events leaving the module. |
| `jobs.go` | The module owns scheduled or recurring background work. |

Do not create empty scaffold files for possible future behavior. Add the file when the responsibility is implemented.

## SQLC-only persistence rule

All commercial persistence must go through SQLC.

```text
commercial/<module>/repository.go
        ↓
internal/database/sqlc
        ↓
internal/database/queries/*.sql
        ↓
PostgreSQL
```

Repository files may construct `sqlc.*Params`, call generated `*sqlc.Queries` methods, convert generated rows into domain models, and map PostgreSQL/pgx errors into domain errors. They must not embed SQL strings or call `Query`, `QueryRow`, or `Exec` directly for application persistence.

New or changed SQL belongs in `server/internal/database/queries/*.sql`. Generated bindings belong in `server/internal/database/sqlc` and must remain reproducible by `sqlc generate`. The Server workflow verifies that regeneration produces no diff.

## Source-of-truth rules

1. PostgreSQL commercial state is authoritative.
2. `organizations` is the tenant boundary for organization-owned commercial records.
3. Payment providers are adapters. Provider state must be reconciled into Leamout state rather than replacing it.
4. Provider webhooks must not directly issue licenses or grant entitlements.
5. Historical invoice pricing must be snapshotted. Old invoices must not be re-rated from current rates.
6. Usage ingestion must be idempotent.
7. SQL queries must enforce tenant/resource ownership even when middleware or service authorization fails.
8. Commercial repositories must use SQLC-generated queries rather than raw SQL.

See [security.md](security.md) for the database defense model.

## Commercial domains

- [Catalog](catalog.md) — reusable products, plans, stable offer codes, and activation lifecycle.
- [Subscriptions](subscriptions.md) — organization-to-plan commercial relationships and subscription lifecycle.
- [Entitlements](entitlements.md) — feature and limit grants at plan, organization, and license scope.
- **State** — resolved commercial capabilities and limits consumed by operational code.
- [Licensing](licensing.md) — signed self-hosted commercial state and deployment activation.
- [Metering](metering.md) — authoritative usage ingestion and meters.
- [Rating](rating.md) — telecom usage price resolution through carrier rates.
- [Invoicing](invoicing.md) — period charges and historical price snapshots.
- [Payments](payments.md) — provider-independent payment reconciliation.

## Current boundaries

The current model intentionally does not include separate commercial customers, contracts, support, notifications, credits, discounts, taxes, refunds, or payout infrastructure.

Those concepts should be added only when concrete product behavior requires them.
