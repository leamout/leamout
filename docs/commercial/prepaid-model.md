# Prepaid Commercial Model

Leamout commercial state is organized around pay-before-use settlement. PostgreSQL is the monetary source of truth. Redis may cache availability or coordinate realtime authorization, but it never creates, destroys, or settles value.

## Core tables

### Catalog

```text
products
plans
prices
meters
```

`products` describe what Leamout sells. `plans` package product access. `prices` contain one-time, recurring, and metered customer-facing prices. `meters` define measurable consumption units.

### Access

```text
subscriptions
licenses
entitlements
```

`subscriptions` describe recurring software access, `licenses` authorize self-hosted software use, and `entitlements` resolve features and limits. Deployment/runtime inventory is not a Commercial primitive.

### Metering

```text
usage_events
```

Usage events are immutable, idempotent measurements. Observed usage is not automatically billable usage.

### Money

```text
wallets
wallet_ledger_entries
wallet_reservations
```

Wallet ledger entries are immutable posted credits and debits. Reservations temporarily hold spendable balance before the final charge is known or before an asynchronous managed-provider operation becomes irrevocable. Available balance is posted balance minus active reservations.

### Payments

```text
checkouts
payments
payment_provider_events
```

Checkouts represent customer purchase intent. Payments represent reconciled money movement. Provider events are retained for webhook idempotency, verification, and reconciliation.

## Four-quadrant billing rule

| Delivery mode | Software | Cloud consumption | Managed telecom consumption |
| --- | --- | --- | --- |
| Self-Hosted + BYOC | Billable | No | No |
| Self-Hosted + Managed | Billable | No | Billable |
| Leamout Cloud + BYOC | Billable | Billable | No |
| Leamout Cloud + Managed | Billable | Billable | Billable |

Voice and messaging usage may still be observed in every mode. Billability is resolved by Commercial policy; it is not inferred merely from the existence of a usage event.

## Pricing

`prices` is the single customer-facing pricing table. It supports one-time, recurring, and metered prices. A separate `usage_rates` table is not part of the target model.

Examples:

```text
Enterprise software        USD 50,000 / year
Cloud voice processing     USD 0.002 / minute
Managed Ghana voice        USD 0.017 / minute
Managed number purchase    USD 2.00 one-time
```

Provider wholesale cost remains separate from customer-facing prices. Upstream carrier cost is COGS and must never be stored as a customer `price`.

## Prepaid flow

```text
customer pays
    ↓
payment succeeds and is verified
    ↓
wallet ledger credit
    ↓
usage authorization
    ↓
reservation when final cost is not yet known
    ↓
usage occurs
    ↓
final price resolution
    ↓
ledger debit + reservation release/capture
```

A managed-provider obligation must not be created unless sufficient prepaid funds have been authorized. Fixed charges may post an atomic debit directly; asynchronous fixed-cost provider operations may reserve the fixed customer price first so retries and provider reconciliation remain durable.

### Managed number purchase

Managed DID purchasing uses a server-owned Leamout price. DIDWW inventory IDs, SKUs, order amounts, and other provider economics never determine the customer wallet charge.

```text
managed number search
        ↓
resolve active subscription plan
        ↓
resolve Leamout one-time price in the subscription currency
        ↓
return opaque selection + customer quote
        ↓
customer creates managed number
        ↓
revalidate quote
        ↓
atomically reserve wallet funds
        ↓
persist provider operation with durable purchase authorization
        ↓
worker verifies reservation
        ↓
DIDWW order
        ├── pending → keep reservation
        ├── cancelled/failed → release reservation
        └── completed → capture reservation
                              ↓
                       reconcile DID routing
```

Capture happens when the provider order completes because that is the point at which Leamout has incurred the upstream obligation. A later routing or persistence problem does not silently release captured value. Such a correction must be represented by an explicit compensating ledger entry if commercial policy requires a refund.

The purchase authorization and provider operation are independently retry-safe: a captured reservation may be verified again, capture is idempotent for the same amount, and release is idempotent while value has not been captured.

## Deferred concepts

The target prepaid model does not require invoice-centric settlement. `invoices` and `invoice_items` are deferred until Leamout has a concrete need for invoices or postpaid accounts.

The current Commercial migration series is still pre-release and has not been applied, so it is consolidated in place around this model. Once this schema has been applied or released, subsequent schema changes must be introduced through forward migrations rather than rewriting migration history.
