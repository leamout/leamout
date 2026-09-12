# Prepaid PAYG commercial model

Everything in Leamout is prepaid pay-as-you-go except Self-Hosted software licenses.

Self-Hosted software licenses are enterprise software agreements and are settled separately, normally through contract, invoice, and bank transfer. They are not funded from a Leamout usage wallet.

All Cloud and managed communications charges are prepaid.

## Delivery modes

| Delivery mode | Commercial treatment |
| --- | --- |
| Self-Hosted + BYOC | Enterprise Self-Hosted software license |
| Self-Hosted + Managed | Enterprise Self-Hosted software license + prepaid PAYG managed usage |
| Leamout Cloud + BYOC | Prepaid PAYG for Leamout Cloud/platform consumption; customer carrier cost remains outside Leamout |
| Leamout Cloud + Managed | Prepaid PAYG |

BYOC means the customer owns the carrier relationship. It does not create a postpaid Leamout billing model.

## Prepaid customer flow

```text
create organization
    ↓
create wallet
    ↓
fund wallet
    ↓
use Leamout Cloud or managed services
    ↓
Leamout authorizes prepaid value before charge/provider exposure
    ↓
service completes → capture/debit
service does not create the obligation → release
```

Stripe and Paystack are automated wallet-funding rails. Enterprise customers may also fund usage wallets through a verified bank-transfer workflow when that capability exists. The payment rail does not change the prepaid requirement.

## Self-Hosted software licensing

Self-Hosted software is the sole exception to the PAYG rule.

```text
enterprise agreement
        ↓
contract / PO / invoice
        ↓
verified settlement
        ↓
Self-Hosted license
        ↓
deployment(s)
```

The enterprise license and prepaid usage balance are separate commercial facts.

For Self-Hosted + Managed:

```text
Self-Hosted license
        +
prepaid managed-usage wallet
```

The license never acts as managed-usage credit.

## Commercial responsibilities

- Catalog owns Leamout customer-facing prices and commercial offer configuration.
- Checkout + Payments fund prepaid wallets.
- Wallets own balances, immutable ledger entries, reservations, capture and release.
- Licensing owns enterprise Self-Hosted license/deployment lifecycle.
- Usage records durable observations; it does not itself charge a wallet.
- Telecom owns service fulfillment and provider orchestration.
- DIDWW, CommPeak and other provider economics remain Leamout COGS and never determine customer wallet prices.

## Pay-before-use invariant

A chargeable Leamout Cloud or managed operation must not proceed beyond its authorized prepaid value.

For a managed-provider operation, Leamout must have enough committed customer funds before the operation can create an upstream provider obligation.

Examples:

- managed DID: reserve fixed customer price before DIDWW order;
- managed voice: reserve an initial authorization window before carrier originate, then increase the target reservation before extending authorized call time;
- managed messaging: reserve customer price before provider submission;
- Cloud platform consumption: authorize prepaid customer value before allowing chargeable consumption beyond the funded amount.

## Explicit non-goals

- customer subscription lifecycle;
- postpaid telecom or platform usage credit;
- Net-30/Net-60 usage settlement;
- invoice-centric usage collection;
- provider-cost pass-through pricing;
- generalized telecom billing or rating platform.

Enterprise contracting and invoicing for Self-Hosted software licenses are not exceptions to these usage rules because the license is a separate enterprise software sale, not metered Leamout consumption.
