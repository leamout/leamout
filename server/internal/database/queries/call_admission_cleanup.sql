-- name: CountPendingAdmissionCleanup :one
SELECT count(*) FROM call_admission_cleanup;

-- name: EnqueueCallAdmissionCleanup :exec
INSERT INTO call_admission_cleanup (call_id, organization_id)
VALUES (sqlc.arg(call_id), sqlc.arg(organization_id))
ON CONFLICT (call_id) DO NOTHING;

-- name: ListDueCallAdmissionCleanup :many
SELECT call_id, organization_id
FROM call_admission_cleanup
WHERE next_attempt_at <= NOW()
ORDER BY next_attempt_at ASC, call_id ASC
LIMIT 100;

-- name: DeferCallAdmissionCleanup :exec
UPDATE call_admission_cleanup
SET next_attempt_at = NOW() + INTERVAL '1 minute'
WHERE call_id = sqlc.arg(call_id);

-- name: DeleteCallAdmissionCleanup :exec
DELETE FROM call_admission_cleanup
WHERE call_id = sqlc.arg(call_id);

-- name: ListCallCleanupTrunks :many
SELECT trunk_id FROM calls
WHERE id = sqlc.arg(call_id) AND trunk_id IS NOT NULL
UNION
SELECT trunk_id FROM call_origination_attempts
WHERE call_id = sqlc.arg(call_id);
