# Commercial Security

Commercial data is tenant-sensitive and financially sensitive. Leamout uses defense in depth rather than treating HTTP middleware as the only security boundary.

Everything in Leamout is prepaid pay-as-you-go except Self-Hosted software licenses. Security boundaries must preserve that distinction: enterprise license settlement must not create wallet value, and chargeable Cloud or managed usage must not exceed authorized prepaid value.

## Defense layers

```text
request authentication
        ↓
HTTP middleware
        ↓
service authorization and business rules
        ↓
SQLC tenant/resource ownership checks
        ↓
foreign keys, uniqueness and CHECK constraints
```

A failure in one layer must not expose or mutate another organization's commercial records.

## Organization ownership

Organization-owned commercial queries should verify that the organization exists, is active, and has not been soft-deleted.

This applies to wallets, wallet reservations, wallet ledger entries, checkouts, payments, licenses, deployments through their license, and usage events.

Owned resources should be read and mutated through their tenant relationship rather than only by globally unique resource ID.

A UUID is an identifier, not authorization.

## Cross-resource ownership

When one Commercial resource references another, SQL and service logic should ensure the relationship remains within the same organization.

Examples:

```text
checkout.wallet_id
    wallet must belong to checkout.organization_id

deployment.license_id
    license must belong to the requested organization

payment.checkout_id
    checkout must belong to payment.organization_id

wallet reservation
    reservation must belong to the requested wallet and organization
```

Foreign keys establish existence. They do not replace tenant ownership checks.

## Monetary authority

PostgreSQL is authoritative for wallet value.

Redis may cache availability or coordinate realtime admission, but it must never create, destroy, capture, release, or otherwise settle money independently of committed PostgreSQL state.

A chargeable Cloud or managed operation must not consume beyond committed prepaid authorization.

A managed-provider operation must not create upstream exposure before prepaid wallet authorization is committed.

## Reservation concurrency

Wallet admission must serialize against the wallet before deciding whether enough spendable balance exists.

```text
lock wallet
    ↓
read posted balance + active reservations
    ↓
verify sufficient spendable balance
    ↓
create / increase reservation
    ↓
commit
    ↓
charge/provider exposure allowed
```

Concurrent requests must not be able to authorize the same value twice.

## Immutable ledger

Wallet ledger history is append-only.

Never repair money by updating or deleting a posted ledger row. Refunds, chargebacks, and corrections are new compensating entries with their own idempotency identity.

Enterprise Self-Hosted license payments must never be posted as wallet credits unless they are separately identified and reconciled as actual prepaid wallet funding.

## Payment providers

Stripe and Paystack are external collection providers for prepaid wallet funding, not authorization systems for Leamout resources.

Provider events must be authenticated and reconciled against a Leamout checkout before a wallet credit can be posted.

A provider identifier or provider status is never sufficient authorization by itself.

Self-Hosted software license settlement is a separate enterprise procurement path and must not be inferred from Stripe/Paystack checkout state.

## Provider spending

DIDWW, CommPeak, and other managed telecom providers can create real upstream cost.

Provider fulfillment must sit behind committed Leamout authorization. Provider-reported wholesale amounts remain COGS and must not become customer debit authority.

An active Self-Hosted software license does not authorize managed-provider spending. Self-Hosted + Managed still requires prepaid wallet authorization.

## Catalog resources

Catalog products, plans, prices, and meters are global configuration rather than organization-owned records.

Their mutation belongs to trusted operator/configuration paths. Customer-facing workflows consume active/effective catalog terms but must not be able to create arbitrary pricing records.

Recurring catalog terms do not imply postpaid usage or a subscription lifecycle.

## Usage events

Usage event idempotency keys prevent duplicate observations under retries and at-least-once delivery.

Usage remains organization-scoped even when an idempotency key is globally unique. A leaked key must not become a cross-tenant lookup primitive.

Recording usage does not authorize provider spending or debit a wallet.

## Licensing

Self-hosted licenses are organization-owned and independent of prepaid wallet state.

Deployment IDs do not authorize themselves. Deployment operations must resolve the organization → license → deployment ownership chain.

Private license-signing keys must remain on trusted authority infrastructure. Self-hosted runtimes receive public verification material only.

License activation and wallet funding are separate authority paths. Neither may be treated as proof of the other.

## Database constraints remain required

Query guards complement, rather than replace:

- foreign keys;
- unique indexes;
- lifecycle `CHECK` constraints;
- positive amount/quantity checks;
- currency constraints;
- temporal checks;
- reservation lifecycle constraints.

Security rule: middleware rejects invalid requests early; SQL and database constraints protect the data when upstream assumptions fail.
