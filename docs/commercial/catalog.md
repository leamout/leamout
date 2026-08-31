# Catalog

The catalog defines reusable Leamout commercial products and plans. It is global commercial configuration: catalog records do not belong to an organization and do not contain subscription, entitlement, usage, rating, invoice, or payment state.

## Model

```text
product
   ↓
 plans
```

A **product** is a commercial product family. A **plan** is a reusable offer or edition within that product.

Examples:

```text
product: self-hosted
  plan: community
  plan: enterprise

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

## Codes

Product and plan codes are stable machine identifiers. They must be non-empty and cannot contain whitespace. The database keeps product codes unique and plan codes unique.

## Current application boundary

`commercial/catalog` is currently read-only application behavior.

It supports commercial workflows that need to resolve or discover existing products and plans:

```text
GetProduct
GetProductByCode
ListProducts
GetPlan
GetPlanByCode
ListPlans
```

It does not currently own product or plan creation/update workflows because Leamout has no concrete operator/admin surface for those mutations yet.

Catalog records are populated outside this application module through bootstrap/seed/configuration mechanisms. If a real operator workflow is introduced later, catalog mutations should be added together with that concrete caller rather than kept speculatively in the domain service.

## Active discovery

Active-plan discovery requires both the plan and its parent product to be active.

```text
product.active = false
        ↓
existing historical references remain valid
        ↓
plan is excluded from active commercial discovery
```

Catalog state changes must not rewrite historical subscriptions, invoices, licenses, or entitlement snapshots.

## Boundaries

The catalog answers:

```text
What products exist?
What plans belong to a product?
Is this product or plan active?
What stable code identifies this offer?
```

It does not answer:

```text
Which plan does an organization have?
What features does the organization receive?
How much did telecom usage cost?
Was an invoice paid?
Is a self-hosted deployment licensed?
```

Those responsibilities belong to subscriptions, entitlements/state, rating/invoicing/payments, and licensing respectively.

## Persistence rules

- PostgreSQL is authoritative for catalog state.
- Product and plan codes are unique.
- A plan always references a product.
- Application-level catalog access is read-only until a concrete operator mutation workflow exists.
- Historical commercial references must remain valid when catalog state changes outside this module.
