---
applyTo: "server/internal/commercial/**/*.go,server/internal/database/queries/*.sql,server/migrations/*.sql,docs/commercial/**/*.md"
---

# Commercial Instructions

Follow `AGENTS.md` first. These rules apply specifically to Leamout Commercial work.

## Domain ownership

Commercial is organized around:

- Catalog — products, plans, prices, meters.
- Billing — Checkout and Payments.
- Access — subscriptions, licenses, entitlements.
- Usage — authoritative usage observations.
- Prepaid — wallets, immutable ledger entries, reservations.

Do not collapse these responsibilities into one generic billing package.

## Payment and checkout boundary

- Checkout owns purchase intent, commercial target, and fulfillment.
- Payments owns provider-neutral collection state, payment-provider access, webhook parsing/reconciliation, and Settlement results.
- Payments must not credit wallets or grant subscriptions/licenses/entitlements directly.
- Provider success must be authenticated and idempotently reconciled before Checkout fulfills the commercial effect.
- Provider adapters must remain outside Commercial state.

## Prepaid invariants

- Managed-provider obligations require prepaid authorization before Leamout incurs upstream cost.
- Wallet ledger entries are immutable.
- Capture requires a reservation for variable-cost operations.
- Failed/abandoned operations release reservations.
- Successful operations capture no more than the authorized/reserved amount unless a separately defined policy explicitly handles additional authorization.
- PostgreSQL is the monetary source of truth.
- Redis may coordinate authorization but cannot create, settle, or destroy value.
- Currency boundaries are strict.

## Pricing and usage

- Catalog `prices` is the customer-facing pricing primitive.
- Upstream wholesale/provider cost is COGS and must not be stored as a customer price.
- Usage events describe observed consumption; recording usage does not by itself make it billable.
- Do not resurrect removed invoice/postpaid/rate-table concepts unless the active task explicitly changes the product model.

## Persistence

- Commercial persistence uses SQLC.
- Put changed SQL in `server/internal/database/queries/*.sql` and regenerate SQLC.
- Do not embed raw SQL in Commercial repositories.
- Preserve organization ownership and idempotency in SQL where practical.
- Preserve immutable monetary history; corrections are compensating records.

## Tests

For financial flows, cover at least:

- successful path;
- insufficient funds / denied authorization;
- retry/idempotent replay;
- provider mismatch or amount/currency mismatch;
- duplicate webhook/reconciliation;
- transaction rollback/failure where relevant;
- no double credit/debit/capture.
