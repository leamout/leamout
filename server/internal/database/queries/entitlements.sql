-- name: CreatePlanEntitlement :one
INSERT INTO entitlements (
    plan_id,
    entitlement_key,
    kind,
    enabled,
    limit_value,
    starts_at,
    expires_at
)
SELECT
    pl.id AS plan_id,
    sqlc.arg(entitlement_key) AS entitlement_key,
    sqlc.arg(kind) AS kind,
    sqlc.narg(enabled) AS enabled,
    sqlc.narg(limit_value) AS limit_value,
    sqlc.narg(starts_at) AS starts_at,
    sqlc.narg(expires_at) AS expires_at
FROM plans AS pl
JOIN products AS p ON p.id = pl.product_id
WHERE pl.id = sqlc.arg(plan_id)
  AND pl.active = true
  AND p.active = true
RETURNING *;

-- name: CreateOrganizationEntitlement :one
INSERT INTO entitlements (
    organization_id,
    entitlement_key,
    kind,
    enabled,
    limit_value,
    starts_at,
    expires_at
)
SELECT
    o.id AS organization_id,
    sqlc.arg(entitlement_key) AS entitlement_key,
    sqlc.arg(kind) AS kind,
    sqlc.narg(enabled) AS enabled,
    sqlc.narg(limit_value) AS limit_value,
    sqlc.narg(starts_at) AS starts_at,
    sqlc.narg(expires_at) AS expires_at
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: CreateLicenseEntitlement :one
INSERT INTO entitlements (
    license_id,
    entitlement_key,
    kind,
    enabled,
    limit_value,
    starts_at,
    expires_at
)
SELECT
    l.id AS license_id,
    sqlc.arg(entitlement_key) AS entitlement_key,
    sqlc.arg(kind) AS kind,
    sqlc.narg(enabled) AS enabled,
    sqlc.narg(limit_value) AS limit_value,
    sqlc.narg(starts_at) AS starts_at,
    sqlc.narg(expires_at) AS expires_at
FROM licenses AS l
JOIN organizations AS o ON o.id = l.organization_id
WHERE l.id = sqlc.arg(license_id)
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: ListPlanEntitlements :many
SELECT *
FROM entitlements
WHERE plan_id = sqlc.arg(plan_id)
ORDER BY entitlement_key;

-- name: ListOrganizationEntitlements :many
SELECT e.*
FROM entitlements AS e
JOIN organizations AS o ON o.id = e.organization_id
WHERE e.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY e.entitlement_key;

-- name: ListLicenseEntitlements :many
SELECT e.*
FROM entitlements AS e
JOIN licenses AS l ON l.id = e.license_id
JOIN organizations AS o ON o.id = l.organization_id
WHERE e.license_id = sqlc.arg(license_id)
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY e.entitlement_key;

-- name: DeletePlanEntitlement :exec
DELETE FROM entitlements
WHERE id = sqlc.arg(id)
  AND plan_id = sqlc.arg(plan_id);

-- name: DeleteOrganizationEntitlement :exec
DELETE FROM entitlements AS e
USING organizations AS o
WHERE e.id = sqlc.arg(id)
  AND e.organization_id = sqlc.arg(organization_id)
  AND o.id = e.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL;

-- name: DeleteLicenseEntitlement :exec
DELETE FROM entitlements AS e
USING licenses AS l, organizations AS o
WHERE e.id = sqlc.arg(id)
  AND e.license_id = sqlc.arg(license_id)
  AND l.id = e.license_id
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.id = l.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL;
