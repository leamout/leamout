-- name: GetOrganizationRuntimeUsedSeconds :one
-- Count each persisted Voice Agent session's overlap with the billing window.
-- A session spanning two periods contributes only the overlapping seconds.
SELECT COALESCE(SUM(
    GREATEST(0, CEIL(EXTRACT(EPOCH FROM
        LEAST(COALESCE(ended_at, CURRENT_TIMESTAMP), sqlc.arg(period_end)::timestamptz)
        - GREATEST(started_at, sqlc.arg(period_start)::timestamptz)
    )))::bigint
), 0)::bigint AS used_seconds
FROM voice_agent_sessions
WHERE organization_id = sqlc.arg(organization_id)
  AND started_at < sqlc.arg(period_end)::timestamptz
  AND (ended_at IS NULL OR ended_at > sqlc.arg(period_start)::timestamptz);
