-- name: UpsertIdentityUser :one
INSERT INTO users (identity_issuer, identity_subject, name, email, email_verified)
VALUES (sqlc.arg(identity_issuer), sqlc.arg(identity_subject), sqlc.narg(name), sqlc.arg(email), TRUE)
ON CONFLICT (identity_issuer, identity_subject) DO UPDATE
SET name = EXCLUDED.name, email = EXCLUDED.email, email_verified = TRUE, updated_at = NOW()
WHERE users.disabled_at IS NULL
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = sqlc.arg(id)
  AND disabled_at IS NULL
LIMIT 1;


-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = sqlc.arg(email)
  AND disabled_at IS NULL
LIMIT 1;


-- name: GetUserByEmailIncludingDisabled :one
SELECT *
FROM users
WHERE email = sqlc.arg(email)
LIMIT 1;


-- name: DisableUser :exec
UPDATE users
SET
    disabled_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND disabled_at IS NULL;

-- name: ListBackofficeUsers :many
SELECT
    u.id::TEXT AS id,
    COALESCE(NULLIF(BTRIM(u.name), ''), '—')::TEXT AS name,
    u.email::TEXT AS email,
    u.email_verified,
    u.is_platform_admin,
    CASE WHEN u.disabled_at IS NULL THEN 'active' ELSE 'disabled' END::TEXT AS status,
    COUNT(om.organization_id) FILTER (WHERE om.status = 'active')::BIGINT AS organization_count,
    to_char(u.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI')::TEXT AS created_at
FROM users AS u
LEFT JOIN organization_members AS om ON om.user_id = u.id
WHERE (
    sqlc.arg(search)::text = ''
    OR u.id::text = sqlc.arg(search)::text
    OR strpos(lower(u.email::text), lower(sqlc.arg(search)::text)) > 0
    OR strpos(lower(COALESCE(u.name, '')), lower(sqlc.arg(search)::text)) > 0
)
  AND (sqlc.arg(status)::text = '' OR (u.disabled_at IS NULL) = (sqlc.arg(status)::text = 'active'))
  AND (sqlc.arg(admin)::text = '' OR u.is_platform_admin = (sqlc.arg(admin)::text = 'yes'))
  AND (sqlc.arg(verified)::text = '' OR u.email_verified = (sqlc.arg(verified)::text = 'yes'))
  AND (
      sqlc.narg(organization_id)::uuid IS NULL
      OR EXISTS (
          SELECT 1 FROM organization_members AS member
          WHERE member.user_id = u.id
            AND member.organization_id = sqlc.narg(organization_id)::uuid
      )
  )
GROUP BY
    u.id,
    u.name,
    u.email,
    u.email_verified,
    u.is_platform_admin,
    u.disabled_at,
    u.created_at
ORDER BY u.created_at DESC, u.id DESC
LIMIT sqlc.arg(page_limit)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: GetBackofficeUser :one
SELECT
    u.id::TEXT AS id,
    COALESCE(NULLIF(BTRIM(u.name), ''), '—')::TEXT AS name,
    u.email::TEXT AS email,
    u.email_verified,
    u.is_platform_admin,
    CASE WHEN u.disabled_at IS NULL THEN 'active' ELSE 'disabled' END::TEXT AS status,
    COUNT(DISTINCT om.organization_id) FILTER (WHERE om.status = 'active')::BIGINT AS organization_count,
    COUNT(DISTINCT s.id) FILTER (
        WHERE s.revoked_at IS NULL
          AND s.expires_at > NOW()
    )::BIGINT AS active_session_count,
    COALESCE(
        to_char(MAX(s.last_seen_at) AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI'),
        '—'
    )::TEXT AS last_seen_at,
    to_char(u.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI')::TEXT AS created_at,
    to_char(u.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI')::TEXT AS updated_at,
    COALESCE(
        to_char(u.disabled_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI'),
        '—'
    )::TEXT AS disabled_at
FROM users AS u
LEFT JOIN organization_members AS om ON om.user_id = u.id
LEFT JOIN sessions AS s ON s.user_id = u.id
WHERE u.id = sqlc.arg(id)
GROUP BY
    u.id,
    u.name,
    u.email,
    u.email_verified,
    u.is_platform_admin,
    u.disabled_at,
    u.created_at,
    u.updated_at
LIMIT 1;
