-- name: CreateLicense :one
INSERT INTO licenses (
    organization_id,
    subscription_id,
    status,
    max_deployments,
    signing_key_id,
    issued_at,
    expires_at
)
SELECT
    o.id,
    sqlc.narg(subscription_id)::uuid,
    COALESCE(sqlc.narg(status), 'pending'),
    COALESCE(sqlc.narg(max_deployments), 1),
    sqlc.narg(signing_key_id),
    COALESCE(sqlc.narg(issued_at), NOW()),
    sqlc.narg(expires_at)
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND (
      sqlc.narg(subscription_id)::uuid IS NULL
      OR EXISTS (
          SELECT 1
          FROM subscriptions AS s
          WHERE s.id = sqlc.narg(subscription_id)::uuid
            AND s.organization_id = o.id
      )
  )
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

-- name: ListLicensesBySubscription :many
SELECT l.*
FROM licenses AS l
JOIN subscriptions AS s
  ON s.id = l.subscription_id
 AND s.organization_id = l.organization_id
JOIN organizations AS o ON o.id = l.organization_id
WHERE l.organization_id = sqlc.arg(organization_id)
  AND l.subscription_id = sqlc.arg(subscription_id)
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
RETURNING l.*;

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
RETURNING l.*;
