-- name: CreateSession :one
INSERT INTO sessions (
    user_id,
    token_hash,
    identity_subject,
    identity_session_id,
    tokens_ciphertext,
    assurance,
    ip_address,
    user_agent,
    expires_at,
    last_seen_at
)
SELECT
    u.id,
    sqlc.arg(token_hash),
    sqlc.arg(identity_subject),
    sqlc.arg(identity_session_id),
    sqlc.arg(tokens_ciphertext),
    sqlc.arg(assurance),
    sqlc.narg(ip_address)::inet,
    sqlc.narg(user_agent)::text,
    sqlc.arg(expires_at)::timestamptz,
    NOW()
FROM users AS u
WHERE u.id = sqlc.arg(user_id)
  AND u.disabled_at IS NULL
RETURNING *;


-- name: GetSessionByID :one
SELECT s.*
FROM sessions AS s
JOIN users AS u
    ON u.id = s.user_id
WHERE s.id = sqlc.arg(id)
  AND s.expires_at > NOW()
  AND s.revoked_at IS NULL
  AND u.disabled_at IS NULL
LIMIT 1;


-- name: GetSessionByTokenHash :one
SELECT s.*
FROM sessions AS s
JOIN users AS u
    ON u.id = s.user_id
WHERE s.token_hash = sqlc.arg(token_hash)
  AND s.expires_at > NOW()
  AND s.revoked_at IS NULL
  AND u.disabled_at IS NULL
LIMIT 1;


-- name: ListSessionsByUserID :many
SELECT s.*
FROM sessions AS s
JOIN users AS u
    ON u.id = s.user_id
WHERE s.user_id = sqlc.arg(user_id)
  AND s.expires_at > NOW()
  AND s.revoked_at IS NULL
  AND u.disabled_at IS NULL
ORDER BY s.last_seen_at DESC NULLS LAST;


-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = NOW()
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL;


-- name: TouchSession :exec
UPDATE sessions
SET last_seen_at = NOW()
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND expires_at > NOW()
  AND revoked_at IS NULL
  AND (
      last_seen_at IS NULL
      OR last_seen_at < NOW() - INTERVAL '5 minutes'
  );


-- name: RevokeUserSessions :exec
UPDATE sessions
SET revoked_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL;

-- name: LockIdentitySession :one
SELECT s.* FROM sessions AS s
JOIN users AS u ON u.id = s.user_id
WHERE s.id = sqlc.arg(id) AND s.revoked_at IS NULL
  AND s.expires_at > NOW() AND u.disabled_at IS NULL
FOR UPDATE OF s;

-- name: UpdateIdentitySessionTokens :exec
UPDATE sessions SET tokens_ciphertext = sqlc.arg(tokens_ciphertext), last_seen_at = NOW()
WHERE id = sqlc.arg(id) AND revoked_at IS NULL;

-- name: ListBackofficeUserSessions :many
SELECT
    s.id::TEXT AS session_id,
    s.assurance,
    COALESCE(host(s.ip_address), '—')::TEXT AS ip_address,
    COALESCE(NULLIF(BTRIM(s.user_agent), ''), '—')::TEXT AS user_agent,
    CASE
        WHEN s.revoked_at IS NOT NULL THEN 'revoked'
        WHEN s.expires_at <= NOW() THEN 'expired'
        ELSE 'active'
    END::TEXT AS status,
    to_char(s.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI')::TEXT AS created_at,
    COALESCE(
        to_char(s.last_seen_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI'),
        '—'
    )::TEXT AS last_seen_at,
    to_char(s.expires_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI')::TEXT AS expires_at,
    COALESCE(
        to_char(s.revoked_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI'),
        '—'
    )::TEXT AS revoked_at
FROM sessions AS s
WHERE s.user_id = sqlc.arg(user_id)
ORDER BY s.created_at DESC, s.id DESC
LIMIT 20;
