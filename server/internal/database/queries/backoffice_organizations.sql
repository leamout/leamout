-- name: ListBackofficeOrganizations :many
SELECT
    o.id,
    o.name,
    CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END::TEXT AS status,
    o.created_at
FROM organizations AS o
WHERE (
    sqlc.arg(search)::TEXT = ''
    OR strpos(lower(o.name), lower(sqlc.arg(search)::TEXT)) > 0
    OR o.id::TEXT = sqlc.arg(search)::TEXT
)
AND (
    sqlc.arg(status)::TEXT = ''
    OR (CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END) = sqlc.arg(status)::TEXT
)
ORDER BY o.created_at DESC, o.id DESC
LIMIT sqlc.arg(page_limit)::INT
OFFSET sqlc.arg(page_offset)::INT;

-- name: GetBackofficeOrganization :one
SELECT
    o.id,
    o.name,
    CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END::TEXT AS status,
    o.created_at,
    o.updated_at,
    o.deleted_at,
    (SELECT count(*) FROM organization_members AS m WHERE m.organization_id = o.id AND m.status = 'active')::BIGINT AS active_members,
    (SELECT count(*) FROM voice_agents AS a WHERE a.organization_id = o.id)::BIGINT AS agents,
    (SELECT count(*) FROM trunks AS t WHERE t.organization_id = o.id)::BIGINT AS trunks,
    COALESCE(p.name, 'No subscription')::TEXT AS plan_name,
    COALESCE(s.status, 'none')::TEXT AS subscription_status,
    s.current_period_end
FROM organizations AS o
LEFT JOIN subscriptions AS s ON s.organization_id = o.id
LEFT JOIN plans AS p ON p.id = s.plan_id
WHERE o.id = sqlc.arg(id);
