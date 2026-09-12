-- name: ListBackofficeCommercialAccounts :many
SELECT
    o.id::TEXT AS organization_id,
    o.name AS organization_name,
    'prepaid'::TEXT AS billing_model,
    COUNT(w.id)::TEXT AS wallet_count,
    COALESCE(string_agg(w.currency, ', ' ORDER BY w.currency), '—') AS currencies
FROM organizations AS o
LEFT JOIN wallets AS w
    ON w.organization_id = o.id
   AND w.status <> 'closed'
WHERE o.deleted_at IS NULL
GROUP BY o.id, o.name, o.created_at
ORDER BY o.created_at DESC
LIMIT 100;

-- name: GetBackofficeCommercialAccount :one
SELECT
    o.id::TEXT AS organization_id,
    o.name AS organization_name,
    'prepaid'::TEXT AS billing_model,
    COUNT(w.id)::TEXT AS wallet_count,
    COALESCE(string_agg(w.currency, ', ' ORDER BY w.currency), '—') AS currencies,
    to_char(o.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS organization_created_at
FROM organizations AS o
LEFT JOIN wallets AS w
    ON w.organization_id = o.id
   AND w.status <> 'closed'
WHERE o.id = sqlc.arg(organization_id)
  AND o.deleted_at IS NULL
GROUP BY o.id, o.name, o.created_at
LIMIT 1;
