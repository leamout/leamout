-- name: CreateLicense :one
INSERT INTO licenses (
    organization_id,
    subscription_id,
    status,
    max_deployments,
    signing_key_id,
    issued_at,
    expires_at
) VALUES (
    sqlc.arg(organization_id),
    sqlc.narg(subscription_id),
    COALESCE(sqlc.narg(status), 'pending'),
    COALESCE(sqlc.narg(max_deployments), 1),
    sqlc.narg(signing_key_id),
    COALESCE(sqlc.narg(issued_at), NOW()),
    sqlc.narg(expires_at)
)
RETURNING *;

-- name: GetLicense :one
SELECT *
FROM licenses
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: ListLicensesByOrganization :many
SELECT *
FROM licenses
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY created_at DESC;

-- name: ListLicensesBySubscription :many
SELECT *
FROM licenses
WHERE organization_id = sqlc.arg(organization_id)
  AND subscription_id = sqlc.arg(subscription_id)
ORDER BY created_at DESC;

-- name: UpdateLicenseStatus :one
UPDATE licenses
SET
    status = sqlc.arg(status),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;

-- name: UpdateLicenseExpiration :one
UPDATE licenses
SET
    expires_at = sqlc.narg(expires_at),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;
