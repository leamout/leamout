# Subscriptions

Subscriptions connect an organization to a reusable catalog plan and provide the commercial lifecycle used by entitlement resolution, licensing, and managed billing.

## Module boundary

```text
catalog
   ↓
subscription
   ↓
entitlements / state / licensing
```

`commercial/subscriptions` owns the organization's commercial relationship to a plan. It does not own product/plan definitions, entitlement resolution, license signing, usage pricing, invoices, or payment-provider implementations.

## Model

```text
product
   ↓
 plan
   ↓
subscription
   ↓
organization
```

A **product** is a Leamout product family. A **plan** is a reusable commercial offer or edition within that product. A **subscription** is an organization's instance of a plan.

The current plan schema intentionally does not store recurring or fixed price fields. Telecom usage pricing is modeled separately through `carrier_rates`.

## Subscription fields

The current subscription record contains:

```text
organization_id
plan_id
status
starts_at
renews_at
ends_at
billing_provider
provider_subscription_id
created_at
updated_at
```

Provider fields are optional but paired: a provider subscription ID cannot exist without a billing provider and vice versa.

Leamout does not currently store `billing_provider` or `provider_customer_id` on `organizations`.

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

## Catalog eligibility

Creating a subscription or changing its plan requires:

```text
plan.active = true
        +
product.active = true
```

The subscription service checks this through the catalog domain before persistence, while the SQL write path repeats the active plan/product constraint to protect against service or authorization mistakes and races.

A cancelled or expired subscription cannot change plans.

## Current subscription lookup

For commercial decisions, an organization may have historical subscription rows. The current lookup selects the newest subscription in `active` or `past_due` state for an active organization.

A caller must not treat an arbitrary historical subscription as the organization's current commercial state.

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

External providers are reconciliation sources, not the owner of Leamout's domain model.

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

Subscription SQL is organization-scoped for organization-owned reads and mutations. New or changed state requires an active, non-deleted organization. Plan assignment is repeated at SQL level against active plans/products.

This preserves the commercial-domain rule that SQL must enforce ownership and resource validity even when middleware or service authorization fails.

## Deferred lifecycle fields

The current schema does not yet include more detailed cancellation/grace information such as:

```text
past_due_at
cancel_at_period_end
cancelled_at
```

Add these when concrete service behavior requires them rather than encoding a speculative billing lifecycle.
