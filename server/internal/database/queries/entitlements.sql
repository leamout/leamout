-- name: CreateEntitlement :one
INSERT INTO entitlements (
    plan_id,
    organization_id,
    license_id,
    entitlement_key,
    kind,
    enabled,
    limit_value,
    starts_at,
    expires_at
) VALUES (
    sqlc.narg(plan_id),
    sqlc.narg(organization_id),
    sqlc.narg(license_id),
    sqlc.arg(entitlement_key),
    sqlc.arg(kind),
    sqlc.narg(enabled),
    sqlc.narg(limit_value),
    sqlc.narg(starts_at),
    sqlc.narg(expires_at)
)
RETURNING *;

-- name: GetEntitlement :one
SELECT *
FROM entitlements
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: ListPlanEntitlements :many
SELECT *
FROM entitlements
WHERE plan_id = sqlc.arg(plan_id)
ORDER BY entitlement_key;

-- name: ListOrganizationEntitlements :many
SELECT *
FROM entitlements
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY entitlement_key;

-- name: ListLicenseEntitlements :many
SELECT *
FROM entitlements
WHERE license_id = sqlc.arg(license_id)
ORDER BY entitlement_key;

-- name: DeleteEntitlement :exec
DELETE FROM entitlements
WHERE id = sqlc.arg(id);
