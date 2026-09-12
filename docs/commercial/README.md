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

### Managed / CPaaS prepaid money path

```text
telecom domain event
        ↓
      metering
        ↓
       rating
        ↓
 prepaid authorization
        ↓
 funds reservation
        ↓
 provider operation
        ↓
 capture + reconciliation
```

Managed voice, numbers, messaging, and media create an upstream obligation for Leamout. The customer must therefore fund a wallet before Leamout authorizes a managed-provider operation. Usage may be metered and rated, but collection is prepaid; Leamout does not extend postpaid usage credit.

## Pay-before-use invariants

1. Platform access requires an active subscription or enterprise license.
2. A managed-provider operation requires sufficient prepaid funds; BYOC carrier usage does not consume a Leamout carrier wallet.
3. A pending, processing, failed, or unverified provider payment never creates spendable credit.
4. Only an authenticated, idempotently recorded provider success may post a wallet top-up.
5. Spendable balance is posted ledger credits minus posted ledger debits minus active reservations.
6. Leamout must atomically reserve sufficient funds before creating an upstream obligation with DIDWW, CommPeak, or another managed provider.
7. A wallet cannot be overdrawn. Concurrent authorizations serialize on the wallet and fail when funds are insufficient.
8. Successful provider work captures no more than the reserved amount; unused funds are released. Failed or abandoned work releases its reservation.
9. Wallet ledger entries are immutable. Refunds, chargebacks, and corrections are new compensating entries.
10. PostgreSQL is the monetary source of truth. Redis may cache availability and coordinate realtime authorization, but it cannot create or destroy value.
11. Money is never mixed across currencies. Every wallet has exactly one ISO currency.
12. Customer challenge secrets, including OTPs and PINs, are never persisted.

These rules describe subscription/license plus prepaid usage. They prohibit postpaid usage billing, not the metering required to price and reconcile prepaid consumption.

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
    ├── checkout_orders
    │      └── payments
    │             └── payment_provider_events
    └── wallets
           ├── wallet_ledger_entries
           └── wallet_reservations

plan + meter
    └── usage_rates
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
usage_rates
usage_events
invoices
invoice_items
checkout_orders
payments
payment_provider_events
wallets
wallet_ledger_entries
wallet_reservations
```

Managed telecom wholesale cost is a separate concern: provider CDRs reconcile into `wholesale_charges`; they are not usage pricing rules.

The schema establishes financial invariants. The `commercial/wallets` and
`commercial/checkout` repositories implement the four durable prepaid records:
wallets, immutable ledger entries, funds reservations, and checkout orders.
Reservation admission serializes on the wallet row, and capture closes the
reservation and posts its ledger debit in one transaction. Public checkout
routes and provider-side effects remain separate vertical slices.

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
11. Managed-provider spending must be preceded by an atomic prepaid reservation.
12. Provider payment success and wallet credit are separate, idempotent records.
13. Ledger history is append-only; a correction never rewrites a posted entry.
14. Redis is never the monetary system of record.

See [security.md](security.md) for the database defense model.

## Commercial domains

- [Catalog](catalog.md) — reusable products, plans, prices, stable offer codes, and availability.
- [Subscriptions](subscriptions.md) — organization-to-price commercial relationships and subscription lifecycle.
- [Entitlements](entitlements.md) — feature and limit grants at plan, organization, and license scope.
- **State** — resolved commercial capabilities and limits consumed by operational code.
- [Licensing](licensing.md) — self-hosted commercial authority and deployment activation.
- [Metering](metering.md) — authoritative usage ingestion and meters for managed/billable services.
- [Rating](rating.md) — customer-facing telecom usage pricing through usage rates.
- [Invoicing](invoicing.md) — period statements and historical monetary snapshots.
- [Payments](payments.md) — checkout intent and provider-independent payment reconciliation.
- [Wallets](wallets.md) — currency-scoped prepaid value, immutable ledger movements, and provider-operation reservations.

## Current boundaries

The current model intentionally does not include a separate commercial customer entity, contracts, postpaid credit limits, support, discounts, tax calculation, customer withdrawals, or payout infrastructure.

Those concepts should be added only when concrete product behavior requires them.
