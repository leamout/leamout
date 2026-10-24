-- name: ApplyKindeIdentityEvent :one
-- Signed events are applied atomically. Event IDs prevent replay, and newer
-- events prevent stale profile writes. Disabling is deliberately irreversible
-- through webhooks, including suspension and deletion events.
WITH recorded AS (
    INSERT INTO kinde_identity_events (
        issuer,
        event_id,
        subject,
        event_type,
        occurred_at
    )
    VALUES (
        sqlc.arg(issuer),
        sqlc.arg(event_id),
        sqlc.arg(subject),
        sqlc.arg(event_type),
        sqlc.arg(occurred_at)
    )
    ON CONFLICT (issuer, event_id) DO NOTHING
    RETURNING issuer, subject, event_type, occurred_at
),
current_event AS (
    SELECT r.*
    FROM recorded AS r
    WHERE sqlc.arg(disable_user)::boolean
       OR NOT EXISTS (
           SELECT 1
           FROM kinde_identity_events AS later
           WHERE later.issuer = r.issuer
             AND later.subject = r.subject
             AND later.occurred_at > r.occurred_at
       )
),
changed_users AS (
    UPDATE users AS u
    SET
        name = CASE
            WHEN NOT sqlc.arg(disable_user)::boolean
                AND e.event_type = 'user.updated'
            THEN COALESCE(
                NULLIF(BTRIM(sqlc.arg(profile_name)::text), ''),
                u.name
            )
            ELSE u.name
        END,
        disabled_at = CASE
            WHEN sqlc.arg(disable_user)::boolean
            THEN COALESCE(u.disabled_at, NOW())
            ELSE u.disabled_at
        END,
        updated_at = NOW()
    FROM current_event AS e
    WHERE u.identity_issuer = e.issuer
      AND u.identity_subject = e.subject
      AND (
          sqlc.arg(disable_user)::boolean
          OR (
              e.event_type = 'user.updated'
              AND u.disabled_at IS NULL
              AND NULLIF(BTRIM(sqlc.arg(profile_name)::text), '') IS NOT NULL
          )
      )
    RETURNING u.id, u.disabled_at
),
revoked_sessions AS (
    UPDATE sessions AS s
    SET revoked_at = NOW()
    FROM changed_users AS u
    WHERE s.user_id = u.id
      AND u.disabled_at IS NOT NULL
      AND s.revoked_at IS NULL
    RETURNING s.id
)
SELECT EXISTS(SELECT 1 FROM recorded)::boolean AS recorded;
