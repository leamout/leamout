-- name: AddOrganizationMember :one
INSERT INTO organization_members (
    organization_id,
    user_id,
    role
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(user_id),
    sqlc.arg(role)
FROM users AS u
WHERE u.id = sqlc.arg(user_id)
AND u.disabled_at IS NULL
RETURNING *;

-- name: GetOrganizationMember :one
SELECT tm.*
FROM organization_members AS tm
JOIN users AS u ON u.id = tm.user_id
WHERE tm.organization_id = sqlc.arg(organization_id)
AND tm.user_id = sqlc.arg(user_id)
AND tm.status = 'active'
AND u.disabled_at IS NULL
LIMIT 1;

-- name: ListMembersByOrganizationID :many
SELECT tm.*
FROM organization_members AS tm
JOIN users AS u ON u.id = tm.user_id
WHERE tm.organization_id = sqlc.arg(organization_id)
AND tm.status = 'active'
AND u.disabled_at IS NULL
ORDER BY tm.created_at ASC;

-- name: ListOrganizationMembershipsByUserID :many
SELECT tm.*
FROM organization_members AS tm
JOIN organizations AS t ON t.id = tm.organization_id
WHERE tm.user_id = sqlc.arg(user_id)
AND tm.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
ORDER BY tm.created_at ASC;

-- name: UpdateMemberRole :one
UPDATE organization_members
SET
    role = sqlc.arg(role),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
AND user_id = sqlc.arg(user_id)
AND status = 'active'
RETURNING *;

-- name: DisableOrganizationMember :exec
UPDATE organization_members
SET
    status = 'disabled',
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
AND user_id = sqlc.arg(user_id)
AND status = 'active';

-- name: EnableOrganizationMember :exec
UPDATE organization_members
SET
    status = 'active',
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
AND user_id = sqlc.arg(user_id)
AND status = 'disabled';

-- name: RemoveOrganizationMember :exec
DELETE FROM organization_members
WHERE organization_id = sqlc.arg(organization_id)
AND user_id = sqlc.arg(user_id);
