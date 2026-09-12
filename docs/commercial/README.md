# Commercial Domain

The Commercial domain owns Leamout customer pricing, prepaid money, payment reconciliation, self-hosted licensing, and durable usage observations. External payment and telecom providers are adapters; they are never the source of truth for Leamout commercial state.

`organizations` is the canonical tenant and commercial account identity. There is no separate commercial customer entity.

## Product model

Leamout uses two commercial paths:

```text
Leamout Cloud
    ↓
prepaid PAYG
    ↓
wallet-funded managed services
```

```text
Self-Hosted
    ↓
Leamout license
    ↓
customer deployment(s)
```

Self-Hosted + Managed combines both: the customer needs a self-hosted license and prepaid wallet authorization before Leamout creates managed-provider obligations.

There is no customer subscription lifecycle in the current model.

## Domain map

```text
organization
    │
    ├── wallets
    │      ├── immutable ledger entries
    │      └── reservations
    │
    ├── checkouts
    │      └── payments
    │             └── provider events
    │
    ├── licenses
    │      └── deployments
    │
    └── usage events

catalog
    ├── products
    ├── plans
    ├── prices
    └── meters
```

Managed telecom wholesale cost is separate COGS. DIDWW, CommPeak, and other provider costs do not determine customer wallet prices.

## Responsibilities

- [PAYG](payg.md) — the overall prepaid commercial model.
- [Catalog](catalog.md) — Leamout-owned products, plans, prices, meters, and stable offer configuration.
- [Payments](payments.md) — wallet top-up checkout and payment-provider reconciliation.
- [Wallets](wallets.md) — currency-scoped prepaid value, immutable ledger movements, and reservations.
- [Usage](usage.md) — immutable organization-scoped usage observations.
- [Licensing](licensing.md) — self-hosted license lifecycle and signed deployment authority.
- [Deployments](deployments.md) — activated self-hosted installations under a license.
- [Security](security.md) — tenant and financial defense rules.

## Pay-before-use invariants

1. A managed-provider obligation requires sufficient prepaid wallet authorization first.
2. BYOC carrier usage does not consume a Leamout managed-carrier wallet merely because it passes through Leamout.
3. Pending, processing, failed, or unverified provider payments never create spendable credit.
4. Only authenticated, idempotently reconciled payment success may create a wallet top-up credit.
5. Spendable balance is posted ledger credits minus posted ledger debits minus active reservations.
6. Reservation admission serializes on the wallet so concurrent operations cannot overdraw it.
7. Successful provider work captures authorized value; failed or abandoned work releases its reservation according to operation policy.
8. Ledger history is append-only. Refunds, chargebacks, and corrections are compensating entries.
9. PostgreSQL is the monetary source of truth. Redis may coordinate realtime authorization but cannot create or destroy value.
10. Money is never mixed across currencies.
11. Provider wholesale cost remains separate from customer-facing Catalog prices.
12. Self-hosted licensing is independent of Cloud wallet state and does not require a customer subscription.

## Module structure

Commercial modules use files only when they own the corresponding responsibility:

| File | Responsibility |
| --- | --- |
| `model.go` | Domain models, inputs, outputs, states, and errors. |
| `repository.go` | Durable persistence through SQLC. |
| `service.go` | Business rules and transaction/orchestration boundaries. |
| `validation.go` | Reusable domain/input validation. |
| `handler.go` | HTTP handlers. |
| `routes.go` | HTTP route registration. |
| `consumer.go` | Asynchronous events entering the module. |
| `publisher.go` | Asynchronous events leaving the module. |
| `jobs.go` | Scheduled or recurring work. |

Do not create empty scaffold files for speculative future behavior.

## Persistence rule

Commercial application persistence must use SQLC-generated queries.

```text
commercial/<module>/repository.go
        ↓
internal/database/sqlc
        ↓
internal/database/queries/*.sql
        ↓
PostgreSQL
```

Repository code may construct SQLC params, call generated queries, convert rows into domain models, and map database errors. Application repositories must not embed raw persistence SQL.

## Source-of-truth rules

1. PostgreSQL commercial state is authoritative.
2. Organization-owned records must remain organization-scoped in SQL as well as middleware.
3. Payment providers report external collection facts; they do not own Leamout wallet state.
4. Telecom providers report fulfillment/wholesale facts; they do not own Leamout customer pricing.
5. Wallet ledger history is immutable.
6. Managed-provider spending must be preceded by committed prepaid authorization.
7. Usage ingestion is idempotent and does not itself debit a wallet.
8. Self-hosted license verification must not require a payment provider to be online.
9. Redis is never the monetary system of record.
10. Historical customer pricing used for an authorized operation must not silently change because Catalog configuration changes later.

## Explicit non-goals

The current Commercial model does not include:

- customer subscriptions;
- commercial access or entitlement services;
- postpaid credit;
- invoice-centric settlement;
- a generalized telecom rating engine;
- provider-cost pass-through pricing;
- customer withdrawals or payouts;
- tax calculation/remittance infrastructure.

Add new commercial concepts only when concrete product behavior requires them.
