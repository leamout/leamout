-- name: DisableActiveDeploymentCredentials :exec
UPDATE deployment_credentials
SET disabled_at = COALESCE(disabled_at, now())
WHERE deployment_id = sqlc.arg(deployment_id)
  AND purpose = sqlc.arg(purpose)
  AND disabled_at IS NULL;

-- name: CreateDeploymentCredential :one
INSERT INTO deployment_credentials (
    deployment_id,
    purpose,
    token_hash,
    token_prefix,
    scopes,
    expires_at
)
SELECT
    d.id,
    sqlc.arg(purpose),
    sqlc.arg(token_hash),
    sqlc.arg(token_prefix),
    sqlc.arg(scopes)::jsonb,
    sqlc.narg(expires_at)
FROM deployments AS d
JOIN licenses AS l ON l.id = d.license_id
JOIN organizations AS o ON o.id = l.organization_id
WHERE d.id = sqlc.arg(deployment_id)
  AND d.status = 'active'
  AND l.status = 'active'
  AND (l.expires_at IS NULL OR l.expires_at > now())
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING deployment_credentials.*;

-- name: GetDeploymentCredentialByTokenHash :one
SELECT
    dc.id AS credential_id,
    dc.deployment_id AS deployment_row_id,
    dc.purpose,
    dc.token_hash,
    dc.token_prefix,
    dc.scopes,
    dc.expires_at AS credential_expires_at,
    dc.last_used_at,
    dc.disabled_at,
    dc.created_at,
    dc.updated_at,
    d.deployment_id,
    d.status AS deployment_status,
    d.license_id,
    l.organization_id,
    l.status AS license_status,
    l.expires_at AS license_expires_at
FROM deployment_credentials AS dc
JOIN deployments AS d ON d.id = dc.deployment_id
JOIN licenses AS l ON l.id = d.license_id
JOIN organizations AS o ON o.id = l.organization_id
WHERE dc.token_hash = sqlc.arg(token_hash)
  AND dc.disabled_at IS NULL
  AND (dc.expires_at IS NULL OR dc.expires_at > now())
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: TouchDeploymentCredential :one
UPDATE deployment_credentials
SET last_used_at = now()
WHERE id = sqlc.arg(id)
  AND disabled_at IS NULL
  AND (expires_at IS NULL OR expires_at > now())
RETURNING id;
