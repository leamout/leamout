-- name: InsertAuditEvent :exec
INSERT INTO audit_events (
    organization_id,
    actor_type,
    actor_id,
    action,
    target_type,
    target_id,
    metadata
) VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(actor_type),
    sqlc.arg(actor_id),
    sqlc.arg(action),
    sqlc.arg(target_type),
    sqlc.arg(target_id),
    sqlc.arg(metadata)::jsonb
);

-- name: ListAuditEventsByOrganizationID :many
SELECT
    id,
    organization_id,
    actor_type,
    actor_id,
    action,
    target_type,
    target_id,
    metadata,
    occurred_at
FROM audit_events
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(action)::text IS NULL OR action = sqlc.narg(action)::text)
  AND (sqlc.narg(actor_type)::text IS NULL OR actor_type = sqlc.narg(actor_type)::text)
  AND (sqlc.narg(actor_id)::uuid IS NULL OR actor_id = sqlc.narg(actor_id)::uuid)
  AND (sqlc.narg(target_type)::text IS NULL OR target_type = sqlc.narg(target_type)::text)
  AND (sqlc.narg(target_id)::uuid IS NULL OR target_id = sqlc.narg(target_id)::uuid)
  AND (sqlc.narg(occurred_from)::timestamptz IS NULL OR occurred_at >= sqlc.narg(occurred_from)::timestamptz)
  AND (sqlc.narg(occurred_before)::timestamptz IS NULL OR occurred_at < sqlc.narg(occurred_before)::timestamptz)
ORDER BY occurred_at DESC, id DESC
LIMIT sqlc.arg(limit_count)::integer
OFFSET sqlc.arg(offset_count)::integer;
