# Commercial Domain

The commercial domain owns Leamout's software-license relationship today and provides explicit boundaries for managed telecom billing as the product roadmap reaches those phases. External billing/payment providers are adapters; they are never the source of truth for Leamout commercial state.

Leamout uses `organizations` as the canonical tenant and current commercial account identity. There is no separate commercial customer identity in the current model.

## Roadmap alignment

Leamout has two commercial paths that share the same organization and catalog foundations but should not be forced into one billing model.

### Current self-hosted licensing path

```text
catalog
   ↓
subscriptions
   ↓
entitlements
   ↓
state
   ↓
licensing
   ↓
self-hosted deployment(s)
```

This is the Phase 1 commercial core. The subscription describes the software commercial relationship, entitlements describe what the organization is allowed to use, state resolves the effective operational view, and licensing turns those rights into deployable self-hosted authority.

Self-hosted BYOC traffic can use the customer's own carrier and infrastructure. Telecom minutes are therefore not automatically Leamout-billable usage merely because the customer runs licensed Leamout software.

### Future managed / CPaaS money path

```text
telecom domain event
        ↓
      metering
        ↓
       rating
        ↓
 durable monetary charge
        ↓
 accounting / balance
        ↓
     invoicing
        ↓
      payments
```

This path becomes concrete as managed voice, multi-carrier orchestration, number lifecycle, direct carrier connectivity, messaging, media, and hosted carrier products introduce Leamout-owned billable usage and recurring telecom resources.

`metering`, `rating`, `invoicing`, and `payments` are domain boundaries already represented by existing schema/code, but they must not be treated as a complete CPaaS billing engine until real product workflows require them. Future durable `charges`, ledger/accounting, prepaid balance, credit, tax, refund, and billing-account concepts should be introduced with those concrete workflows rather than hidden inside invoices or payment-provider objects.

## Current domain map

```text
organization
    │
    ├── subscription
    │      └── price
    │             └── plan
    │                    └── product
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

The existing commercial tables are:

```text
products
plans
prices
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

Existing tables do not imply that every future commercial workflow is implemented. Package behavior should continue to follow concrete callers and roadmap needs.

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

All commercial application persistence must go through SQLC.

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
3. The current subscription model permits at most one `active`/`past_due` subscription per organization; PostgreSQL enforces that invariant.
4. Payment providers are adapters. Provider state must be reconciled into Leamout state rather than replacing it.
5. Provider webhooks must not directly issue licenses or grant entitlements.
6. Self-hosted license verification must not require a provider to be online for every runtime policy decision.
7. Historical monetary results must be snapshotted. Old invoices must not be re-rated from current rates.
8. Usage ingestion must be idempotent.
9. SQL queries must enforce tenant/resource ownership even when middleware or service authorization fails.
10. Commercial repositories must use SQLC-generated queries rather than raw SQL.

See [security.md](security.md) for the database defense model.

## Commercial domains

- [Catalog](catalog.md) — reusable products, plans, prices, stable offer codes, and availability.
- [Subscriptions](subscriptions.md) — organization-to-price commercial relationships and subscription lifecycle.
- [Entitlements](entitlements.md) — feature and limit grants at plan, organization, and license scope.
- **State** — resolved commercial capabilities and limits consumed by operational code.
- [Licensing](licensing.md) — self-hosted commercial authority and deployment activation.
- [Metering](metering.md) — authoritative usage ingestion and meters for managed/billable services.
- [Rating](rating.md) — telecom usage economic resolution through carrier rates.
- [Invoicing](invoicing.md) — period statements and historical monetary snapshots.
- [Payments](payments.md) — provider-independent payment reconciliation.

## Current boundaries

The current model intentionally does not include a separate commercial customer entity, billing accounts, contracts, durable charges, ledger accounting, prepaid wallets, credit limits, support, notifications, credits, discounts, taxes, refunds, or payout infrastructure.

Those concepts should be added only when concrete product behavior requires them.
