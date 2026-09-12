# Commercial Domain

The Commercial domain owns Leamout customer pricing, prepaid money, payment reconciliation, self-hosted licensing, and durable usage observations. External payment and telecom providers are adapters; they are never the source of truth for Leamout commercial state.

`organizations` is the canonical tenant and commercial account identity. There is no separate commercial customer entity.

## Canonical commercial rule

Everything in Leamout is prepaid pay-as-you-go except Self-Hosted software licenses.

Self-Hosted software licenses are enterprise software agreements. They are sold separately from prepaid usage, normally through a negotiated contract, invoice, and bank transfer. They are not Stripe/Paystack checkout products and do not consume a Leamout wallet.

Every Cloud or managed communications charge is prepaid. Leamout does not extend postpaid telecom or platform usage credit.

| Delivery mode | Software / platform | Managed Leamout usage |
| --- | --- | --- |
| Self-Hosted + BYOC | Enterprise Self-Hosted license | None unless the customer uses a Leamout managed service |
| Self-Hosted + Managed | Enterprise Self-Hosted license | Prepaid PAYG |
| Leamout Cloud + BYOC | Prepaid PAYG | Customer carrier cost remains outside Leamout |
| Leamout Cloud + Managed | Prepaid PAYG | Prepaid PAYG |

BYOC does not mean postpaid. It means the customer owns the carrier relationship, so that carrier cost is not a Leamout managed-provider charge. Any Leamout Cloud/platform consumption that is monetized remains prepaid PAYG.

## Product model

```text
Leamout Cloud
    ↓
prepaid PAYG
    ↓
wallet-funded Leamout consumption
```

```text
Self-Hosted
    ↓
enterprise software agreement
    ↓
Leamout license
    ↓
customer deployment(s)
```

Self-Hosted + Managed combines both commercial paths:

```text
enterprise Self-Hosted license
            +
prepaid PAYG managed usage
```

The license is not funded from the usage wallet. Managed usage still requires prepaid wallet authorization before Leamout creates managed-provider obligations.

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

- [PAYG](payg.md) — the default commercial model for all Cloud and managed usage.
- [Catalog](catalog.md) — Leamout-owned products, plans, prices, meters, and stable offer configuration.
- [Payments](payments.md) — prepaid wallet funding and payment-provider reconciliation.
- [Wallets](wallets.md) — currency-scoped prepaid value, immutable ledger movements, and reservations.
- [Usage](usage.md) — immutable organization-scoped usage observations.
- [Licensing](licensing.md) — enterprise Self-Hosted license lifecycle and signed deployment authority.
- [Deployments](deployments.md) — activated self-hosted installations under a license.
- [Security](security.md) — tenant and financial defense rules.

## Pay-before-use invariants

1. All Cloud and managed communications charges are prepaid PAYG.
2. A managed-provider obligation requires sufficient prepaid wallet authorization first.
3. BYOC carrier usage does not consume a Leamout managed-carrier wallet merely because it passes through Leamout.
4. Pending, processing, failed, or unverified provider payments never create spendable credit.
5. Only authenticated, idempotently reconciled payment success may create a wallet top-up credit.
6. Spendable balance is posted ledger credits minus posted ledger debits minus active reservations.
7. Reservation admission serializes on the wallet so concurrent operations cannot overdraw it.
8. Successful provider work captures authorized value; failed or abandoned work releases its reservation according to operation policy.
9. Ledger history is append-only. Refunds, chargebacks, and corrections are compensating entries.
10. PostgreSQL is the monetary source of truth. Redis may coordinate realtime authorization but cannot create or destroy value.
11. Money is never mixed across currencies.
12. Provider wholesale cost remains separate from customer-facing Catalog prices.
13. Self-Hosted software licensing is independent of prepaid usage wallets and does not require a customer subscription.
14. A Self-Hosted + Managed customer still prepays managed usage; the enterprise license does not create usage credit.

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
11. Enterprise Self-Hosted license settlement does not create prepaid wallet balance.

## Explicit non-goals

The current Commercial model does not include:

- customer subscriptions;
- commercial access or entitlement services;
- postpaid telecom or platform usage credit;
- invoice-centric usage settlement;
- a generalized telecom rating engine;
- provider-cost pass-through pricing;
- customer withdrawals or payouts;
- tax calculation/remittance infrastructure.

Enterprise contracts and invoices for Self-Hosted software licenses are a sales/procurement process, not a postpaid usage model.

Add new commercial concepts only when concrete product behavior requires them.
