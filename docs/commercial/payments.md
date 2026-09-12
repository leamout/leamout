# Payments and prepaid checkout

Payments reconcile externally collected money into a Leamout wallet top-up. A payment is evidence of collection; it is not itself wallet balance, a license, or telecom authorization.

## Collection boundary

```text
card         → Stripe
mobile money → Paystack
```

Leamout owns every checkout reference, amount, currency, destination wallet, and resulting wallet credit. Payment providers execute collection; they do not own Commercial state.

Challenge values such as an OTP are relayed to the provider over a protected request and are never persisted as Commercial state.

## Checkout purpose

The current checkout type is:

```text
wallet_topup
```

Buying a managed number or consuming managed voice is not a Stripe/Paystack checkout. Those operations spend already-funded wallet value.

```text
Stripe / Paystack
        ↓
wallet top-up checkout
        ↓
verified payment success
        ↓
wallet ledger credit
```

## Supported provider/method pairs

| Provider | Method |
| --- | --- |
| Stripe | `card` |
| Paystack | `mobile_money` |

The server validates permitted amount and currency against the destination wallet before collection proceeds.

## Verified collection flow

```text
Leamout checkout
      ↓
provider payment attempt
      ↓
authenticated provider event / verified status
      ↓
reconcile provider + reference + amount + currency
      ↓
idempotent payment success
      ↓
wallet ledger credit
```

Pending or processing payment state never creates spendable value.

Provider events are recorded with stable provider/event identity so retries cannot create duplicate credits.

## Wallet credit boundary

A successful payment and a wallet credit are separate durable facts.

The payment service records provider collection state. The checkout/wallet workflow applies the matching top-up idempotently.

Provider webhooks must never mutate wallet balance directly without Leamout reconciliation.

## Provider independence

Never implement:

```text
provider webhook
    ↓
direct telecom fulfillment
```

Use:

```text
provider collection
    ↓
verified Leamout checkout success
    ↓
wallet credit
    ↓
customer later authorizes managed service from wallet
```

## Deferred concerns

The current model does not implement:

- customer subscriptions;
- postpaid credit;
- invoice collection as the primary payment path;
- customer withdrawals;
- foreign-exchange conversion;
- tax calculation/remittance;
- payouts or Merchant-of-Record infrastructure.
