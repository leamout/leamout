-- name: ListUsageNotificationRecipients :many
SELECT
    o.id AS organization_id,
    o.name AS organization_name,
    u.email AS recipient
FROM organizations AS o
JOIN organization_members AS member
  ON member.organization_id = o.id
 AND member.status = 'active'
 AND member.role = 'owner'
JOIN users AS u
  ON u.id = member.user_id
 AND u.disabled_at IS NULL
 AND u.email_verified = TRUE
WHERE o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY o.id, u.id;

-- name: InsertUsageNotificationEvent :execrows
INSERT INTO usage_notification_events (
    id, organization_id, period_start, threshold_percent
)
VALUES (
    sqlc.arg(id), sqlc.arg(organization_id), sqlc.arg(period_start), sqlc.arg(threshold_percent)
)
ON CONFLICT (organization_id, period_start, threshold_percent) DO NOTHING;
