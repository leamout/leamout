# Subscriptions

Subscriptions connect an organization to a commercial plan and provide the commercial lifecycle used by licensing and managed billing.

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

Current states:

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

The database constrains valid state names, while service logic should define which transitions are allowed and why.

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

A provider webhook should normally follow this direction:

```text
provider webhook
    ↓
verify and normalize provider event
    ↓
reconcile payment/subscription state
    ↓
commercial service rules
    ↓
entitlements / licensing consequences
```

It must not directly sign a license.

## Invariants

- The organization must be active and not deleted when new subscription state is created or mutated.
- New subscriptions may use only active plans whose product is active.
- A plan change should resolve to an active plan/product.
- Provider subscription identifiers are unique per provider pair.
- Subscription timestamps must maintain a valid period.

## Deferred lifecycle fields

The current schema does not yet include more detailed cancellation/grace information such as:

```text
past_due_at
cancel_at_period_end
cancelled_at
```

Add these when service behavior requires them rather than encoding a speculative billing lifecycle.
