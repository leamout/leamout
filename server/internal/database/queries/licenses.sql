-- name: CreateLicense :one
INSERT INTO licenses (
    organization_id,
    status,
    signing_key_id,
    issued_at,
    expires_at
)
SELECT
    o.id AS organization_id,
    COALESCE(sqlc.narg(status), 'pending') AS status,
    sqlc.narg(signing_key_id) AS signing_key_id,
    COALESCE(sqlc.narg(issued_at), NOW()) AS issued_at,
    sqlc.narg(expires_at) AS expires_at
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: GetLicense :one
SELECT l.*
FROM licenses AS l
JOIN organizations AS o ON o.id = l.organization_id
WHERE l.organization_id = sqlc.arg(organization_id)
  AND l.id = sqlc.arg(id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListLicensesByOrganization :many
SELECT l.*
FROM licenses AS l
JOIN organizations AS o ON o.id = l.organization_id
WHERE l.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY l.created_at DESC;

-- name: UpdateLicenseStatus :one
UPDATE licenses AS l
SET
    status = sqlc.arg(status),
    updated_at = NOW()
FROM organizations AS o
WHERE l.organization_id = sqlc.arg(organization_id)
  AND l.id = sqlc.arg(id)
  AND o.id = l.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: UpdateLicenseExpiration :one
UPDATE licenses AS l
SET
    expires_at = sqlc.narg(expires_at),
    updated_at = NOW()
FROM organizations AS o
WHERE l.organization_id = sqlc.arg(organization_id)
  AND l.id = sqlc.arg(id)
  AND o.id = l.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;
