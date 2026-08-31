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

Display names and descriptions may change without changing the code used by provisioning or commercial configuration.

## Active lifecycle

Catalog records are deactivated rather than deleted as part of normal commercial lifecycle.

An inactive product cannot accept new plans. Active-plan discovery also requires the parent product to be active.

```text
product.active = false
        ↓
existing historical references remain valid
        ↓
no new active commercial use should resolve through that product
```

A plan may be deactivated independently. Reactivating a plan requires its parent product to be active.

Catalog activation does not rewrite historical subscriptions, invoices, licenses, or entitlement snapshots.

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
- New plans require an active parent product.
- Activating a plan requires an active parent product.
- Deactivation must not delete historical commercial references.
