# Commercial Security

Commercial data is tenant-sensitive and financially sensitive. Leamout therefore uses defense in depth rather than treating HTTP middleware as the only security boundary.

## Defense layers

```text
request authentication
        ↓
HTTP middleware
        ↓
service authorization and business rules
        ↓
sqlc tenant/resource ownership checks
        ↓
foreign keys, uniqueness and CHECK constraints
```

A failure in one layer must not automatically expose or mutate another organization's commercial records.

## Active organization requirement

Organization-owned commercial queries should verify that the organization exists, is active, and has not been soft-deleted.

```sql
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
```

This applies to organization-scoped subscriptions, licenses, organization entitlements, deployments through their license, usage events, invoices, invoice items, and payments.

## Guarded inserts

Do not trust an `organization_id` argument and insert it directly when the record is tenant-owned. Prefer an `INSERT ... SELECT` guarded by the organization and any required parent resources.

```sql
INSERT INTO subscriptions (
    organization_id,
    plan_id,
    status
)
SELECT
    o.id AS organization_id,
    pl.id AS plan_id,
    'pending' AS status
FROM organizations AS o
JOIN plans AS pl ON pl.id = sqlc.arg(plan_id)
JOIN products AS p ON p.id = pl.product_id
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND pl.active = true
  AND p.active = true
RETURNING *;
```

If the organization or parent resource is invalid, the statement inserts no row.

## Tenant-scoped reads

Owned resources should be selected through their tenant relationship, not only by globally unique resource ID.

```sql
SELECT s.*
FROM subscriptions AS s
JOIN organizations AS o ON o.id = s.organization_id
WHERE s.id = sqlc.arg(id)
  AND s.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;
```

This prevents an otherwise valid resource ID from crossing an organization boundary.

## Tenant-scoped writes

Updates and deletes must enforce the same ownership relationship.

```sql
UPDATE licenses AS l
SET status = sqlc.arg(status), updated_at = NOW()
FROM organizations AS o
WHERE l.id = sqlc.arg(id)
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.id = l.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;
```

## Cross-resource ownership

When one commercial resource references another, SQL should verify that the relationship is valid for the same organization.

Examples:

```text
license.subscription_id
    subscription must belong to license.organization_id

deployment.license_id
    license must belong to the requested organization

usage_event.subscription_id
    subscription must belong to usage_event.organization_id

invoice.subscription_id
    subscription must belong to invoice.organization_id

payment.invoice_id
    invoice must belong to payment.organization_id
```

Foreign keys establish existence, but they do not by themselves establish same-tenant ownership when both records independently carry organization identity.

## Catalog resources

Some commercial resources are global catalog/configuration records rather than tenant-owned records:

```text
products
plans
meters
usage_rates
```

They do not use organization guards in the same way. Their mutation must instead be restricted to trusted administrative/service paths, and related active-state checks should prevent inactive catalog records from being used for new commercial state.

## Provider identifiers

Provider identifiers can be globally useful for webhook reconciliation. A provider lookup should still return only a record attached to an active, non-deleted organization unless a deliberately privileged reconciliation path requires otherwise.

Never use an external provider identifier as an authorization credential.

## Idempotency keys

Usage event idempotency keys prevent duplicate accounting under retries and at-least-once event delivery.

Lookup APIs should remain organization-scoped even when the database uniqueness constraint is global.

```text
organization_id + idempotency_key
        ↓
normal application lookup
```

A leaked idempotency key must not become a cross-tenant read primitive.

## Database constraints are still required

Query guards complement, rather than replace:

- foreign keys;
- unique indexes;
- state `CHECK` constraints;
- positive amount/quantity checks;
- JSON object checks;
- temporal checks.

Security rule: middleware improves ergonomics and rejects requests early; SQL and database constraints protect the data when upstream assumptions fail.
