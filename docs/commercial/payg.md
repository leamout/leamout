# Prepaid PAYG commercial model

Leamout Cloud uses prepaid pay-as-you-go rather than customer subscriptions.

## Customer flow

```text
create organization
    ↓
create wallet
    ↓
top up wallet through Stripe or Paystack
    ↓
buy managed numbers / consume managed telecom services
    ↓
Leamout reserves wallet funds before provider exposure
    ↓
provider succeeds → capture
provider fails before obligation → release
```

## Self-hosted

Self-hosted deployments use Leamout licenses. License lifecycle is independent of Cloud wallet usage and does not require a customer subscription.

## Commercial responsibilities

- Catalog owns Leamout customer-facing prices.
- Checkout + Payments are used to fund wallets.
- Wallets own balances, immutable ledger entries, reservations, capture and release.
- Licensing owns self-hosted license/deployment lifecycle.
- Telecom owns service fulfillment and provider orchestration.
- DIDWW, CommPeak and other provider economics remain Leamout COGS and never determine customer wallet prices.

## Pay-before-use invariant

Leamout must have enough customer funds authorized before an operation can create an upstream managed-provider obligation.

Examples:

- managed DID: reserve fixed customer price before DIDWW order
- managed voice: reserve an initial authorization window before carrier originate, then increase the target reservation before extending authorized call time
- managed messaging: reserve customer price before provider submission

## Explicit non-goals

- customer subscription lifecycle
- postpaid credit
- invoicing as the primary customer payment model
- provider-cost pass-through pricing
- generalized telecom billing or rating platform
