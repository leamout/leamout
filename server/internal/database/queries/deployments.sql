-- name: CreateDeployment :one
INSERT INTO deployments (
    license_id,
    deployment_id,
    name
) VALUES (
    sqlc.arg(license_id),
    sqlc.arg(deployment_id),
    sqlc.narg(name)
)
RETURNING *;

-- name: GetDeployment :one
SELECT *
FROM deployments
WHERE license_id = sqlc.arg(license_id)
  AND deployment_id = sqlc.arg(deployment_id)
LIMIT 1;

-- name: ListDeploymentsByLicense :many
SELECT *
FROM deployments
WHERE license_id = sqlc.arg(license_id)
ORDER BY created_at DESC;

-- name: CountActiveDeploymentsByLicense :one
SELECT COUNT(*)
FROM deployments
WHERE license_id = sqlc.arg(license_id)
  AND status = 'active';

-- name: TouchDeployment :one
UPDATE deployments
SET
    last_seen_at = COALESCE(sqlc.narg(last_seen_at), NOW()),
    updated_at = NOW()
WHERE license_id = sqlc.arg(license_id)
  AND deployment_id = sqlc.arg(deployment_id)
  AND status = 'active'
RETURNING *;

-- name: DeactivateDeployment :one
UPDATE deployments
SET
    status = 'deactivated',
    deactivated_at = COALESCE(deactivated_at, NOW()),
    updated_at = NOW()
WHERE license_id = sqlc.arg(license_id)
  AND deployment_id = sqlc.arg(deployment_id)
  AND status = 'active'
RETURNING *;
