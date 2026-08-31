# Catalog

The catalog defines reusable Leamout commercial products, plans, and prices. It is global commercial configuration: catalog records do not belong to an organization and do not contain subscription, entitlement, usage, rating, invoice, or payment state.

## Model

```text
product
   ↓
 plan
   ↓
prices
```

A **product** is a commercial product family. A **plan** is a reusable offer or edition within that product. A **price** is one immutable set of recurring acquisition terms for a plan.

Examples:

```text
product: self-hosted
  plan: community
  plan: enterprise
    price: USD / month
    price: USD / year

product: managed-voice
  plan: developer
  plan: production
```

The names above are illustrative only; the catalog does not hard-code editions.

## Product fields

```text
id
code
name
description
active
created_at
updated_at
```

## Plan fields

```text
id
product_id
code
name
description
active
created_at
updated_at
```

## Price fields

```text
id
plan_id
currency
amount_minor
billing_interval
active
effective_from
effective_until
created_at
```

`amount_minor` is stored in the currency's minor unit. Currency uses a three-letter uppercase code. The first supported recurring intervals are `month` and `year`.

Commercial price terms are immutable after creation:

```text
plan_id
currency
amount_minor
billing_interval
effective_from
```

Changing a plan's price means creating another price record. Existing subscriptions continue to reference the price they acquired. A price can be retired from future acquisition by changing its availability (`active` / `effective_until`) without rewriting historical terms.

Only one open-ended active price for a given `(plan_id, currency, billing_interval)` may exist at a time.

## Codes

Product and plan codes are stable machine identifiers. They must be non-empty and cannot contain whitespace. The database keeps product codes unique and plan codes unique.

## Current application boundary

`commercial/catalog` is currently read-only application behavior.

It supports commercial workflows that need to resolve or discover existing products, plans, and prices:

```text
GetProduct
GetProductByCode
ListProducts
GetPlan
GetPlanByCode
ListPlans
GetPrice
ListPrices
```

Active price discovery evaluates the price's effective window at the service's current time and also requires the parent plan and product to be active.

The application module does not currently own product, plan, or price creation/update workflows because Leamout has no concrete operator/admin surface for those mutations yet.

Catalog records are populated outside this application module through bootstrap/seed/configuration mechanisms. If a real operator workflow is introduced later, catalog mutations should be added together with that concrete caller rather than kept speculatively in the domain service.

## Active discovery

Active-plan discovery requires both the plan and its parent product to be active. Active-price discovery additionally requires the price to be active and effective at the evaluation time.

```text
product/plan/price no longer available
        ↓
existing historical references remain valid
        ↓
record is excluded from new commercial acquisition
```

Catalog state changes must not rewrite historical subscriptions, charges, invoices, licenses, or entitlement snapshots.

## Catalog prices are not telecom rates

A catalog price answers what it costs to acquire a plan under recurring commercial terms. It does not price individual telecom consumption.

```text
catalog price
= plan acquisition terms

carrier/rating rate
= economic value of telecom usage
```

Destination, carrier, connection, billing increment, buy rate, sell rate, and other telecom-specific economics remain the responsibility of the rating side of the commercial domain.

## Boundaries

The catalog answers:

```text
What products exist?
What plans belong to a product?
What recurring prices can acquire a plan?
Is this product, plan, or price available?
What stable code identifies this offer?
```

It does not answer:

```text
Which plan/price did an organization acquire?
What features does the organization receive?
How much did telecom usage cost?
Was an invoice paid?
Is a self-hosted deployment licensed?
```

Those responsibilities belong to subscriptions, entitlements/state, rating/charges/invoicing/payments, and licensing respectively.

## Persistence rules

- PostgreSQL is authoritative for catalog state.
- Product and plan codes are unique.
- A plan always references a product.
- A price always references a plan.
- Acquired price terms are immutable; new terms require a new price record.
- Application-level catalog access is read-only until a concrete operator mutation workflow exists.
- Historical commercial references must remain valid when catalog availability changes outside this module.
