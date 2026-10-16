-- name: CreateOIDCLoginAttempt :exec
INSERT INTO oidc_login_attempts (
    state_hash,
    verifier_ciphertext,
    nonce,
    expires_at
)
VALUES (
    sqlc.arg(state_hash),
    sqlc.arg(verifier_ciphertext),
    sqlc.arg(nonce),
    sqlc.arg(expires_at)
);


-- name: ConsumeOIDCLoginAttempt :one
DELETE FROM oidc_login_attempts
WHERE state_hash = sqlc.arg(state_hash)
  AND expires_at > NOW()
RETURNING *;


-- name: DeleteExpiredOIDCLoginAttempts :exec
DELETE FROM oidc_login_attempts
WHERE expires_at <= NOW();


-- name: ApplyOIDCLogout :exec
WITH event AS (
    INSERT INTO oidc_logout_events (
        token_id,
        expires_at
    )
    VALUES (
        sqlc.arg(token_id),
        sqlc.arg(expires_at)
    )
    ON CONFLICT DO NOTHING
    RETURNING token_id
)
UPDATE sessions AS s
SET revoked_at = NOW()
FROM users AS u
WHERE EXISTS (
    SELECT 1
    FROM event
)
  AND u.id = s.user_id
  AND u.identity_issuer = sqlc.arg(issuer)
  AND (
      sqlc.arg(subject)::text = ''
      OR s.identity_subject = sqlc.arg(subject)
  )
  AND (
      sqlc.arg(session_id)::text = ''
      OR s.identity_session_id = sqlc.arg(session_id)
  )
  AND s.revoked_at IS NULL;


-- name: DeleteExpiredOIDCLogoutEvents :exec
DELETE FROM oidc_logout_events
WHERE expires_at <= NOW();
