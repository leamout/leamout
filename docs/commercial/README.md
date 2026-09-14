# Commercial Domain

The Commercial domain owns Leamout customer pricing, prepaid money, payment reconciliation, self-hosted licensing, and durable usage observations. External payment and telecom providers are adapters; they are never the source of truth for Leamout commercial state.

`organizations` is the canonical tenant and commercial account identity. There is no separate commercial customer entity.

## Canonical commercial rule

Everything in Leamout is prepaid pay-as-you-go except Self-Hosted software licenses.

Self-Hosted software licenses are enterprise software agreements. They are sold separately from prepaid usage, normally through a negotiated contract, invoice, and bank transfer. They are not Stripe/Paystack checkout products and do not consume a Leamout wallet.

Every Cloud or Leamout-provided communications charge is prepaid. Leamout does not extend postpaid telecom or platform usage credit.

| Delivery mode | Software / platform | Telecom relationship |
| --- | --- | --- |
| Self-Hosted + BYOC | Enterprise Self-Hosted license | Customer-selected carrier; may be a third party or Leamout Carrier |
| Leamout Cloud + BYOC | Prepaid PAYG | Customer-selected carrier cost remains outside Cloud-managed usage |
| Leamout Cloud + Managed | Prepaid PAYG | Leamout-managed telecom usage is prepaid PAYG |

There is no separate Self-Hosted + Managed commercial mode. If a self-hosted customer chooses Leamout Carrier, the software relationship remains an enterprise Self-Hosted license and the telecom relationship is billed separately as a carrier service.

BYOC means the customer selects and configures the carrier relationship. It does not mean the carrier must be a third party and it does not create a postpaid Leamout billing model.

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
    ↓
customer-selected carrier
    ├── third-party carrier
    └── Leamout Carrier
```

When a self-hosted customer also buys Leamout Carrier service, the two commercial relationships remain separate:

```text
enterprise Self-Hosted license
            +
prepaid Leamout Carrier telecom service
```

The software license is not funded from the telecom wallet and never acts as telecom usage credit.

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

- [PAYG](payg.md) — the default commercial model for all Cloud and Leamout-provided usage.
- [Catalog](catalog.md) — Leamout-owned products, plans, prices, meters, and stable offer configuration.
- [Payments](payments.md) — prepaid wallet funding and payment-provider reconciliation.
- [Wallets](wallets.md) — currency-scoped prepaid value, immutable ledger movements, and reservations.
- [Usage](usage.md) — immutable organization-scoped usage observations.
- [Licensing](licensing.md) — enterprise Self-Hosted license lifecycle and signed deployment authority.
- [Deployments](deployments.md) — activated self-hosted installations under a license.
- [Security](security.md) — tenant and financial defense rules.

## Pay-before-use invariants

1. All Cloud and Leamout-provided communications charges are prepaid PAYG.
2. A managed-provider obligation requires sufficient prepaid wallet authorization first.
3. Customer-selected third-party carrier usage does not consume a Leamout managed-carrier wallet merely because it passes through Leamout.
4. A self-hosted customer using Leamout Carrier may have a separate prepaid carrier balance; this does not change the Self-Hosted + BYOC deployment mode.
5. Pending, processing, failed, or unverified provider payments never create spendable credit.
6. Only authenticated, idempotently reconciled payment success may create a wallet top-up credit.
7. Spendable balance is posted ledger credits minus posted ledger debits minus active reservations.
8. Reservation admission serializes on the wallet so concurrent operations cannot overdraw it.
9. Successful provider work captures authorized value; failed or abandoned work releases its reservation according to operation policy.
10. Ledger history is append-only. Refunds, chargebacks, and corrections are compensating entries.
11. PostgreSQL is the monetary source of truth. Redis may coordinate realtime authorization but cannot create or destroy value.
12. Money is never mixed across currencies.
13. Provider wholesale cost remains separate from customer-facing Catalog prices.
14. Self-Hosted software licensing is independent of prepaid usage wallets and does not require a customer subscription.

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
| `consumer.go` | Asynchronous events entering a module. |
| `publisher.go` | Asynchronous events leaving a module. |
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
