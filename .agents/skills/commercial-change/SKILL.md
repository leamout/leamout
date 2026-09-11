---
name: commercial-change
description: Implement or review changes in Leamout Commercial, including Catalog, Billing, Checkout, Payments, Access, Usage, Prepaid wallets, subscriptions, entitlements, and licensing.
---

# Work in Leamout Commercial

Read `AGENTS.md` and `docs/commercial/` before changing the domain. Prefer the current implementation and newest architecture docs when older documents conflict.

## Boundaries

- Catalog defines products, plans, prices, and meters.
- Billing owns Checkout and Payments.
- Checkout owns purchase intent and fulfillment after settlement.
- Payments owns collection state, provider interaction, and provider-event reconciliation; it must not grant access or credit wallets directly.
- Access owns subscriptions, licenses, entitlements, and effective commercial standing.
- Usage records immutable/idempotent observations. Recording usage does not itself make it billable.
- Prepaid wallets own spendable value, immutable ledger entries, and reservations.

## Monetary invariants

- PostgreSQL is the monetary source of truth.
- Redis may coordinate authorization but must never create, destroy, or settle value.
- Managed-provider obligations require prepaid authorization first.
- A wallet must not be overdrawn.
- Ledger history is append-only; corrections/refunds are compensating entries.
- Customer-facing Catalog prices and upstream wholesale carrier costs are separate concepts.
- Provider success is evidence of payment; it is not itself wallet balance or subscription state.
- Idempotency must survive retries, duplicate webhooks, and reconciliation.

All Commercial persistence must use SQLC-generated queries. Keep transaction boundaries explicit when multiple durable monetary records must change atomically.