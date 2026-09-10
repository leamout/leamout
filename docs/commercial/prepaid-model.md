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

`products` describe what Leamout sells. `plans` package product access. `prices` contain recurring and metered customer-facing prices. `meters` define measurable consumption units.

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

Wallet ledger entries are immutable posted credits and debits. Reservations temporarily hold spendable balance before the final charge is known. Available balance is posted balance minus active reservations.

### Payments

```text
checkout_orders
payments
payment_provider_events
```

Checkout orders represent customer payment intent. Payments represent reconciled money movement. Provider events are retained for webhook idempotency, verification, and reconciliation.

## Four-quadrant billing rule

| Delivery mode | Software | Cloud consumption | Managed telecom consumption |
| --- | --- | --- | --- |
| Self-Hosted + BYOC | Billable | No | No |
| Self-Hosted + Managed | Billable | No | Billable |
| Leamout Cloud + BYOC | Billable | Billable | No |
| Leamout Cloud + Managed | Billable | Billable | Billable |

Voice and messaging usage may still be observed in every mode. Billability is resolved by Commercial policy; it is not inferred merely from the existence of a usage event.

## Pricing

`prices` is the single customer-facing pricing table. It must support both recurring software prices and metered prices. A separate `usage_rates` table is not part of the target model.

Examples:

```text
Enterprise software        USD 50,000 / year
Cloud voice processing     USD 0.002 / minute
Managed Ghana voice        USD 0.017 / minute
Managed phone number       USD 2.00 / month
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

A managed-provider obligation must not be created unless sufficient prepaid funds have been authorized. Fixed charges may post an atomic debit directly; variable-cost operations should reserve funds first.

## Deferred concepts

The target prepaid model does not require invoice-centric settlement. `invoices` and `invoice_items` are deferred until Leamout has a concrete need for invoices or postpaid accounts.

Historical migrations are append-only. Existing deployed databases must be moved toward this model using forward migrations rather than rewriting already-applied migration files.
