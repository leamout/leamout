-- name: UpsertRuntimeAttachment :one
INSERT INTO runtime_attachments (
    deployment_id,
    ingress_host,
    ingress_port,
    transport
)
SELECT
    d.id,
    sqlc.arg(ingress_host),
    sqlc.arg(ingress_port),
    sqlc.arg(transport)
FROM deployments AS d
JOIN licenses AS l ON l.id = d.license_id
WHERE d.id = sqlc.arg(deployment_id)
  AND d.status = 'active'
  AND l.status = 'active'
ON CONFLICT (deployment_id) DO UPDATE SET
    ingress_host = EXCLUDED.ingress_host,
    ingress_port = EXCLUDED.ingress_port,
    transport = EXCLUDED.transport,
    verification_status = 'pending',
    health_status = 'unknown',
    verified_at = NULL,
    last_checked_at = NULL,
    updated_at = NOW()
RETURNING *;

-- name: SetRuntimeAttachmentState :one
UPDATE runtime_attachments
SET
    verification_status = sqlc.arg(verification_status),
    health_status = sqlc.arg(health_status),
    verified_at = CASE
        WHEN sqlc.arg(verification_status)::TEXT = 'verified' THEN COALESCE(verified_at, NOW())
        ELSE NULL
    END,
    last_checked_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ResolveManagedInboundRuntimeAttachment :one
WITH candidates AS (
    SELECT
        ra.id AS runtime_attachment_id,
        d.id AS deployment_id,
        d.deployment_id AS deployment_identity,
        ra.ingress_host,
        ra.ingress_port,
        ra.transport
    FROM runtime_attachments AS ra
    JOIN deployments AS d ON d.id = ra.deployment_id
    JOIN licenses AS l ON l.id = d.license_id
    JOIN organizations AS o ON o.id = l.organization_id
    WHERE l.organization_id = sqlc.arg(organization_id)
      AND o.status = 'active'
      AND o.deleted_at IS NULL
      AND l.status = 'active'
      AND d.status = 'active'
      AND ra.verification_status = 'verified'
      AND ra.health_status = 'healthy'
)
SELECT *
FROM candidates
WHERE (SELECT count(*) FROM candidates) = 1
LIMIT 1;
