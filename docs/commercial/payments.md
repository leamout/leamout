# Payments and prepaid checkout

Payments reconcile externally collected money with a server-priced Leamout checkout order. A payment is evidence of collection; it is not itself a subscription, entitlement, license, or wallet balance.

## Collection boundary

```text
card         -> Stripe PaymentIntent + Leamout card checkout
mobile money -> Paystack Charge API + Leamout MoMo checkout
```

Leamout creates every checkout reference and owns the amount, currency, target, and commercial consequence. Stripe is restricted to cards. Paystack is restricted to Mobile Money and exposes only the continuation actions needed by that channel.

Challenge values such as an OTP are relayed to the provider over a protected request and are never persisted. Leamout does not collect a Paystack PIN.

## Checkout order

A checkout order is the durable intent that precedes provider collection. It has exactly one target:

- `subscription` references a server-owned catalog price and may reference its invoice.
- `wallet_topup` references the destination currency wallet.

The provider and payment method are fixed pairs:

| Provider | Method |
| --- | --- |
| Stripe | `card` |
| Paystack | `mobile_money` |

The browser may select a permitted wallet top-up amount, but configured limits and the final amount are enforced by the server. Subscription amounts always come from the selected immutable price and invoice snapshot.

## Verified collection flow

```text
server-priced checkout order
        ↓
provider payment attempt
        ↓
authenticated provider event
        ↓
idempotent payment success
        ↓
subscription transition OR wallet ledger credit
```

Pending or processing payment state never delivers value. Before applying a success, reconciliation compares provider, reference, amount, and currency to the checkout order.

Provider events are stored with a unique `(provider, provider_event_id)` identity. Re-delivery can observe the recorded result but cannot repeat a wallet credit or subscription transition.

## Prepaid wallets

Each organization may have one wallet per ISO currency. Currency is never converted implicitly or mixed within a wallet.

Posted balance is the sum of immutable ledger entries. Spendable balance is:

```text
posted ledger balance - active reservations
```

Positive entries are top-ups, refunds, or credit adjustments. Negative entries are captures, chargebacks, or debit adjustments. Ledger rows cannot be updated or deleted; corrections are compensating entries with their own idempotency key.

## Provider spending

Before Leamout incurs an upstream obligation, it atomically locks the wallet, verifies spendable funds, and creates a reservation. This applies to DIDWW number purchases, CommPeak calls, and every future managed carrier operation.

A successful operation captures no more than the reservation. Failure releases it. Realtime Redis state may accelerate admission and incremental call authorization, but PostgreSQL remains authoritative and Redis cannot mint credit.

The wallet repository acquires a PostgreSQL row lock before reading spendable
balance and inserting a reservation. Capturing a reservation and appending its
immutable debit share one database transaction, so neither half can commit
without the other. Expiration records `expired_at`; an automatic timeout is not
represented as a manual release.

## Provider independence

Provider webhooks must never directly grant entitlements, issue licenses, or mutate wallet balances. They authenticate and record provider facts; commercial services apply the matching Leamout transition in an idempotent database transaction.

## Deferred concerns

The initial pay-before-use model does not implement postpaid credit, customer withdrawals, automatic foreign-exchange conversion, tax calculation, payouts, or Merchant-of-Record infrastructure.
