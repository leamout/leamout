-- name: ListRetentionPolicies :many
SELECT *
FROM retention_policies
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY resource;

-- name: GetRetentionPolicy :one
SELECT *
FROM retention_policies
WHERE organization_id = sqlc.arg(organization_id)
  AND resource = sqlc.arg(resource)
LIMIT 1;

-- name: UpsertRetentionPolicy :one
INSERT INTO retention_policies (
    organization_id,
    resource,
    retention_days,
    enabled
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(resource),
    sqlc.arg(retention_days),
    sqlc.arg(enabled)
)
ON CONFLICT (organization_id, resource)
DO UPDATE SET
    retention_days = EXCLUDED.retention_days,
    enabled = EXCLUDED.enabled
RETURNING *;

-- name: DeleteRetentionPolicy :exec
DELETE FROM retention_policies
WHERE organization_id = sqlc.arg(organization_id)
  AND resource = sqlc.arg(resource);

-- name: ListEnabledRetentionPolicies :many
SELECT *
FROM retention_policies
WHERE enabled
ORDER BY organization_id, resource;

-- name: ListExpiredRecordingsForRetention :many
SELECT id
FROM recordings
WHERE organization_id = sqlc.arg(organization_id)
  AND status = 'completed'
  AND created_at < sqlc.arg(created_before)
ORDER BY created_at
LIMIT sqlc.arg(batch_size);

-- name: ListEffectiveRecordingRetention :many
WITH effective_plans AS (
    SELECT
        o.id AS organization_id,
        COALESCE(active_plan.limits, free_plan.limits) AS limits
    FROM organizations AS o
    LEFT JOIN subscriptions AS s
      ON s.organization_id = o.id
     AND s.status IN ('trialing', 'active')
    LEFT JOIN plans AS active_plan
      ON active_plan.id = s.plan_id
    CROSS JOIN plans AS free_plan
    WHERE free_plan.code = 'free'
      AND o.status = 'active'
      AND o.deleted_at IS NULL
)
SELECT
    ep.organization_id,
    CASE
        WHEN NULLIF(ep.limits ->> 'retention_days', '') IS NULL
            THEN rp.retention_days
        WHEN rp.enabled
            THEN LEAST(
                rp.retention_days,
                (ep.limits ->> 'retention_days')::INTEGER
            )
        ELSE (ep.limits ->> 'retention_days')::INTEGER
    END::INTEGER AS retention_days
FROM effective_plans AS ep
LEFT JOIN retention_policies AS rp
  ON rp.organization_id = ep.organization_id
 AND rp.resource = 'recordings'
WHERE NULLIF(ep.limits ->> 'retention_days', '') IS NOT NULL
   OR (rp.enabled AND rp.retention_days IS NOT NULL)
ORDER BY ep.organization_id;
