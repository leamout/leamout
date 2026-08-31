# Subscriptions

Subscriptions connect an organization to acquired catalog terms and provide the commercial lifecycle used by entitlement resolution, licensing, and managed billing.

## Module boundary

```text
catalog price
     ↓
subscription
     ↓
entitlements / state / licensing
```

`commercial/subscriptions` owns the organization's commercial relationship to an acquired catalog price and its derived plan. It does not own product/plan/price definitions, entitlement resolution, license signing, telecom usage pricing, invoices, or payment-provider implementations.

## Model

```text
product
   ↓
 plan
   ↓
 price
   ↓
subscription
   ↓
organization
```

A **product** is a Leamout product family. A **plan** is a reusable commercial offer or edition within that product. A **price** captures immutable recurring acquisition terms. A **subscription** records which price an organization acquired and preserves the corresponding plan identity used by entitlements.

Telecom usage pricing remains separate from catalog plan prices and is resolved by the rating side of the commercial domain.

## Subscription fields

The subscription record contains:

```text
organization_id
plan_id
price_id
status
starts_at
renews_at
ends_at
billing_provider
provider_subscription_id
created_at
updated_at
```

`price_id` is nullable only to preserve compatibility with rows created before price-backed subscriptions were introduced. New subscription creation requires a price.

Provider fields are optional but paired: a provider subscription ID cannot exist without a billing provider and vice versa.

Leamout does not currently store `billing_provider` or `provider_customer_id` on `organizations`.

## Acquiring commercial terms

A caller selects a price rather than independently supplying a plan and price.

```text
price_id
   ↓
catalog price
   ↓
price.plan_id
   ↓
subscription {
  price_id,
  plan_id
}
```

This prevents callers from constructing mismatched commercial terms. The service resolves and validates the price, plan, and product before persistence. PostgreSQL repeats the relationship through the `(price_id, plan_id)` foreign key and SQL write constraints.

A new subscription requires:

```text
price.active = true
price effective at starts_at
plan.active = true
product.active = true
```

Retiring a catalog price does not rewrite an existing subscription. The acquired `price_id` remains the historical identity of the terms that subscription received.

## Changing commercial terms

`ChangePrice` is the current operation for changing a subscription's catalog terms.

The selected price determines the new plan atomically:

```text
new price
   ↓
new price.plan_id
   ↓
UPDATE price_id + plan_id together
```

The service validates availability first, and SQL repeats the price/plan/product checks. Selecting the already-acquired price is an idempotent no-op.

A cancelled or expired subscription cannot change commercial terms.

## Lifecycle

Current states and service-owned transitions:

```text
pending
   ↓
active
   ├──→ past_due ──→ active
   │        ├──→ cancelled
   │        └──→ expired
   ├──→ cancelled
   └──→ expired
```

`cancelled` and `expired` are terminal states. Re-applying the current state is treated as an idempotent no-op so repeated reconciliation/provider events do not create artificial transition failures.

New subscriptions may start only as `pending` or `active`. The default is `pending`.

## Current subscription lookup

For commercial decisions, an organization may have historical subscription rows. The current lookup selects the newest subscription in `active` or `past_due` state for an active organization.

A caller must not treat an arbitrary historical subscription as the organization's current commercial state.

The singular current-subscription model is the current implemented constraint. Supporting independent concurrent commercial relationships per product should be introduced only together with the corresponding resolution semantics rather than inferred from historical rows.

## Provider boundary

```text
external billing provider
        ↓
provider subscription identifier
        ↓
Leamout subscription
        ↓
Leamout commercial state
```

External providers are reconciliation sources, not the owner of Leamout's domain model or catalog price identity.

Provider references are normalized and persisted as paired values. The `(billing_provider, provider_subscription_id)` pair is globally unique when present.

A provider webhook should normally follow this direction:

```text
provider webhook
    ↓
verify and normalize provider event
    ↓
lookup/reconcile Leamout subscription
    ↓
subscription service transition
    ↓
entitlements / state / licensing consequences
```

It must not directly sign a license or grant entitlements.

## Period rules

`starts_at` is immutable after creation in the current model. `renews_at` and `ends_at` may be advanced through subscription period updates.

The following must always hold:

```text
renews_at >= starts_at
ends_at >= starts_at
renews_at <= ends_at   (when both exist)
```

The current update query treats omitted renewal/end values as "leave unchanged". Explicit clearing of an existing timestamp is intentionally not modeled yet.

## Database defenses

Subscription SQL is organization-scoped for organization-owned reads and mutations. New or changed state requires an active, non-deleted organization.

For acquisition and price changes, SQL also repeats:

```text
price exists and matches plan
price active/effective
plan active
product active
```

The database foreign key additionally prevents a persisted `price_id` from referring to a different `plan_id`.

This preserves the commercial-domain rule that SQL must enforce ownership and resource validity even when middleware or service authorization fails.

## Deferred lifecycle fields

The current schema does not yet include more detailed cancellation/grace information such as:

```text
past_due_at
cancel_at_period_end
cancelled_at
```

Add these when concrete service behavior requires them rather than encoding a speculative billing lifecycle.
