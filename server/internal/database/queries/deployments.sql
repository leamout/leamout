-- name: CreateDeployment :one
INSERT INTO deployments (
    license_id,
    deployment_id,
    name
)
SELECT
    l.id AS license_id,
    sqlc.arg(deployment_id) AS deployment_id,
    sqlc.narg(name) AS name
FROM licenses AS l
JOIN organizations AS o ON o.id = l.organization_id
WHERE l.id = sqlc.arg(license_id)
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: GetDeployment :one
SELECT d.*
FROM deployments AS d
JOIN licenses AS l ON l.id = d.license_id
JOIN organizations AS o ON o.id = l.organization_id
WHERE d.license_id = sqlc.arg(license_id)
  AND d.deployment_id = sqlc.arg(deployment_id)
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListDeploymentsByLicense :many
SELECT d.*
FROM deployments AS d
JOIN licenses AS l ON l.id = d.license_id
JOIN organizations AS o ON o.id = l.organization_id
WHERE d.license_id = sqlc.arg(license_id)
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY d.created_at DESC;

-- name: CountActiveDeploymentsByLicense :one
SELECT COUNT(*)
FROM deployments AS d
JOIN licenses AS l ON l.id = d.license_id
JOIN organizations AS o ON o.id = l.organization_id
WHERE d.license_id = sqlc.arg(license_id)
  AND d.status = 'active'
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL;

-- name: TouchDeployment :one
UPDATE deployments AS d
SET
    last_seen_at = COALESCE(sqlc.narg(last_seen_at), NOW()),
    updated_at = NOW()
FROM licenses AS l
JOIN organizations AS o ON o.id = l.organization_id
WHERE d.license_id = l.id
  AND d.license_id = sqlc.arg(license_id)
  AND d.deployment_id = sqlc.arg(deployment_id)
  AND d.status = 'active'
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: DeactivateDeployment :one
UPDATE deployments AS d
SET
    status = 'deactivated',
    deactivated_at = COALESCE(d.deactivated_at, NOW()),
    updated_at = NOW()
FROM licenses AS l
JOIN organizations AS o ON o.id = l.organization_id
WHERE d.license_id = l.id
  AND d.license_id = sqlc.arg(license_id)
  AND d.deployment_id = sqlc.arg(deployment_id)
  AND d.status = 'active'
  AND l.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;
