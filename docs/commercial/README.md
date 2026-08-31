# Commercial Domain

The commercial domain turns Leamout usage and self-hosted deployments into durable commercial state without making an external billing provider the source of truth.

Leamout uses `organizations` as the canonical tenant and commercial account. There is no separate commercial customer identity in the current model.

## Domain map

```text
organization
    │
    ├── subscription
    │      └── plan
    │            └── product
    │
    ├── entitlements
    │
    ├── licenses
    │      └── deployments
    │
    ├── usage_events
    │      └── meters
    │
    ├── invoices
    │      └── invoice_items
    │
    └── payments

plan + meter
    └── carrier_rates
```

The current commercial tables are:

```text
products
plans
subscriptions
licenses
entitlements
deployments
meters
carrier_rates
usage_events
invoices
invoice_items
payments
```

## Self-hosted commercial flow

```text
organization
    ↓
subscription
    ↓
plan
    ↓
effective entitlements
    ↓
license
    ↓
deployment(s)
```

The subscription describes the commercial relationship. Entitlements describe what is allowed. Commercial state resolves the effective operational view. A license carries the deployable commercial state for self-hosted installations, including activated deployments under that license.

## Managed / CPaaS commercial flow

```text
telecom domain event
        ↓
      meter
        ↓
   usage_event
        ↓
 carrier rate resolution
        ↓
   invoice_item
        ↓
      invoice
        ↓
      payment
```

Telecom domains produce authoritative usage. The commercial domain meters and rates it, snapshots charges into invoice items, and reconciles payments.

## Source-of-truth rules

1. PostgreSQL commercial state is authoritative.
2. `organizations` is the tenant boundary for organization-owned commercial records.
3. Payment providers are adapters. Provider state must be reconciled into Leamout state rather than replacing it.
4. Provider webhooks must not directly issue licenses or grant entitlements.
5. Historical invoice pricing must be snapshotted. Old invoices must not be re-rated from current rates.
6. Usage ingestion must be idempotent.
7. SQL queries must enforce tenant/resource ownership even when middleware or service authorization fails.

See [security.md](security.md) for the database defense model.

## Commercial domains

- [Catalog](catalog.md) — reusable products, plans, stable offer codes, and activation lifecycle.
- [Subscriptions](subscriptions.md) — organization-to-plan commercial relationships and subscription lifecycle.
- [Entitlements](entitlements.md) — feature and limit grants at plan, organization, and license scope.
- **State** — resolved commercial capabilities and limits consumed by operational code.
- [Licensing](licensing.md) — signed self-hosted commercial state and deployment activation.
- [Metering](metering.md) — authoritative usage ingestion and meters.
- [Rating](rating.md) — telecom usage price resolution through carrier rates.
- [Invoicing](invoicing.md) — period charges and historical price snapshots.
- [Payments](payments.md) — provider-independent payment reconciliation.

## Current boundaries

The current model intentionally does not include separate commercial customers, contracts, support, notifications, credits, discounts, taxes, refunds, or payout infrastructure.

Those concepts should be added only when concrete product behavior requires them.
