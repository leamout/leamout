-- name: CompareAndSetSubscriptionStatus :one
UPDATE subscriptions AS s
SET
    status = sqlc.arg(status),
    updated_at = NOW()
FROM organizations AS o
WHERE s.organization_id = sqlc.arg(organization_id)
  AND s.id = sqlc.arg(id)
  AND s.status = sqlc.arg(expected_status)
  AND o.id = s.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING s.*;
