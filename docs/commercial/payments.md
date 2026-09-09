# Payments

Payments record provider-independent money movement and reconcile it with Leamout invoices and commercial state.

## Flow

```text
Stripe cards / Paystack Mobile Money / manual
                    ↓
              billing adapter
                    ↓
                 payment
                    ↓
                 invoice
                    ↓
        subscription consequences
                    ↓
              entitlements
                    ↓
                licensing
```

Leamout is not building payment-network or Merchant-of-Record infrastructure. Those capabilities belong behind provider adapters when needed.

## Collection adapters

Payment collection adapters live under `server/internal/integrations/payments`.
They expose one provider-neutral contract for initializing secure checkout,
retrieving authoritative provider state, and authenticating webhook payloads.

The initial channel boundary is explicit:

```text
card         -> Stripe PaymentIntent + Stripe Elements
mobile money -> Paystack transaction + Paystack Popup/redirect
```

Leamout owns the amount, currency, reference, invoice, subscription, prepaid
balance, and all commercial consequences. The browser receives only the Stripe
client secret or Paystack access code/authorization URL needed by the provider's
secure UI. Provider adapters must not activate subscriptions, grant entitlements,
issue licenses, or credit balances.

Both adapters require a Leamout-generated idempotent reference. Stripe receives
that reference as both its idempotency key and PaymentIntent metadata. Paystack
receives it as the transaction reference. Reconciliation must retrieve provider
state and compare the provider, reference, amount, and currency with Leamout's
commercial intent before delivering value.

Webhook parsers authenticate the unmodified request body. Stripe uses the
endpoint-specific webhook secret and a bounded timestamp tolerance. Paystack
uses the account secret key with its HMAC-SHA512 signature. Authenticated events
are normalized, but durable event storage and asynchronous commercial processing
belong to the forthcoming checkout/payment orchestration layer.

## Current payment model

```text
organization_id
invoice_id
provider
provider_payment_id
amount
currency
status
paid_at
metadata
created_at
updated_at
```

Current states are:

```text
pending
succeeded
failed
refunded
partially_refunded
```

The database validates state names and basic amount/currency requirements. Service logic owns valid provider-specific transition handling.

## Provider reconciliation

`provider` identifies the adapter/source. `provider_payment_id` stores the provider's payment identifier when one exists.

The pair is unique when a provider payment ID is present, which prevents the same external payment from being represented by multiple Leamout payment rows.

A typical webhook path is:

```text
provider webhook
    ↓
verify provider signature/authenticity
    ↓
normalize provider event
    ↓
lookup by provider + provider_payment_id
    ↓
idempotently reconcile payment state
    ↓
reconcile invoice/subscription state
```

A provider identifier is a reconciliation key, not an authorization credential.

## Invoice relationship

`invoice_id` is optional because a payment can enter the system before final invoice association or represent a provider flow that is reconciled later.

When an invoice is supplied, SQL should verify:

```text
payment organization == invoice organization
payment currency == invoice currency
```

Service logic should additionally enforce amount/allocation rules appropriate to the payment flow.

## Tenant safety

Payment reads and writes require an active, non-deleted organization. Invoice-scoped payment listing also verifies that the invoice and payment share the same organization.

Provider-ID reconciliation lookups currently filter through active organizations for normal domain access. If a future privileged reconciliation worker must process disabled organizations, create a deliberately named privileged query rather than weakening normal tenant queries.

## Payment does not equal entitlement

Never implement:

```text
payment provider webhook
        ↓
direct entitlement grant or license signing
```

Use Leamout domain transitions:

```text
payment reconciliation
        ↓
invoice/subscription state
        ↓
commercial service decision
        ↓
entitlement resolution
        ↓
license issuance/renewal
```

This keeps business policy provider-independent.

## Manual billing

Manual invoice/payment flows may use the same `payments` domain model without pretending that a card processor exists. Manual billing behavior should be introduced only when an operational workflow requires it.

## Deferred concerns

The current model intentionally avoids premature payment infrastructure such as:

```text
stored cards
payouts
full refund ledger
tax remittance
payment-network settlement
```

Add dedicated models when Leamout owns concrete behavior for them.
