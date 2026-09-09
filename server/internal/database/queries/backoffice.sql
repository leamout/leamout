-- name: ListBackofficeCalls :many
SELECT
    c.id::TEXT AS id,
    o.name AS organization_name,
    c.from_uri,
    c.to_uri,
    c.direction,
    c.state,
    GREATEST(
        0::BIGINT,
        COALESCE(
            EXTRACT(EPOCH FROM (COALESCE(c.ended_at, NOW()) - c.answered_at))::BIGINT,
            0::BIGINT
        )
    ) AS duration_seconds,
    to_char(c.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS created_at
FROM calls AS c
JOIN organizations AS o ON o.id = c.organization_id
ORDER BY c.created_at DESC
LIMIT 100;

-- name: ListBackofficeCarrierConnections :many
SELECT
    cc.id::TEXT AS id,
    COALESCE(o.name, 'Platform') AS organization_name,
    cc.name,
    cp.name AS provider_name,
    cc.scope,
    cc.status,
    cc.inbound_enabled,
    cc.max_cps,
    cc.max_concurrent_calls,
    COUNT(t.id)::BIGINT AS trunk_count
FROM carrier_connections AS cc
JOIN carrier_providers AS cp ON cp.id = cc.provider_id
LEFT JOIN organizations AS o ON o.id = cc.organization_id
LEFT JOIN trunks AS t ON t.carrier_connection_id = cc.id
GROUP BY
    cc.id,
    o.name,
    cc.name,
    cp.name,
    cc.scope,
    cc.status,
    cc.inbound_enabled,
    cc.max_cps,
    cc.max_concurrent_calls,
    cc.created_at
ORDER BY cc.created_at DESC
LIMIT 100;

-- name: ListBackofficeCommercialAccounts :many
SELECT
    o.id::TEXT AS organization_id,
    o.name AS organization_name,
    COALESCE(p.name, '—') AS plan_name,
    COALESCE(s.status, 'none') AS subscription_status,
    COALESCE(s.billing_provider, '—') AS billing_provider,
    COALESCE(
        to_char(s.renews_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI'),
        '—'
    ) AS renews_at
FROM organizations AS o
LEFT JOIN subscriptions AS s
    ON s.organization_id = o.id
   AND s.status IN ('active', 'past_due')
LEFT JOIN plans AS p ON p.id = s.plan_id
WHERE o.deleted_at IS NULL
ORDER BY o.created_at DESC
LIMIT 100;

-- name: ListBackofficeOrganizations :many
SELECT
    o.id::TEXT AS id,
    o.name,
    o.status,
    COUNT(om.user_id) FILTER (WHERE om.status = 'active')::BIGINT AS member_count,
    COALESCE(p.name, '—') AS plan_name,
    to_char(o.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS created_at
FROM organizations AS o
LEFT JOIN organization_members AS om ON om.organization_id = o.id
LEFT JOIN subscriptions AS s
    ON s.organization_id = o.id
   AND s.status IN ('active', 'past_due')
LEFT JOIN plans AS p ON p.id = s.plan_id
WHERE o.deleted_at IS NULL
GROUP BY o.id, o.name, o.status, p.name, o.created_at
ORDER BY o.created_at DESC
LIMIT 100;

-- name: ListBackofficePhoneNumbers :many
SELECT
    pn.id::TEXT AS id,
    o.name AS organization_name,
    pn.number,
    pn.country_code::TEXT AS country_code,
    pn.provisioning_mode,
    COALESCE(cp.name, 'BYOC') AS provider_name,
    pn.voice_enabled,
    pn.sms_enabled,
    pn.status,
    to_char(pn.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS created_at
FROM phone_numbers AS pn
JOIN organizations AS o ON o.id = pn.organization_id
LEFT JOIN carrier_providers AS cp ON cp.id = pn.provider_id
ORDER BY pn.created_at DESC
LIMIT 100;

-- name: ListBackofficeProviders :many
SELECT
    cp.id::TEXT AS id,
    cp.slug,
    cp.name,
    cp.adapter,
    cp.status,
    COUNT(DISTINCT cc.id)::BIGINT AS connection_count,
    COUNT(DISTINCT po.id) FILTER (
        WHERE po.state IN ('pending', 'provider_accepted')
    )::BIGINT AS pending_operation_count,
    COUNT(DISTINCT po.id) FILTER (
        WHERE po.state = 'failed'
    )::BIGINT AS failed_operation_count
FROM carrier_providers AS cp
LEFT JOIN carrier_connections AS cc ON cc.provider_id = cp.id
LEFT JOIN provider_operations AS po ON po.carrier_provider_id = cp.id
GROUP BY cp.id, cp.slug, cp.name, cp.adapter, cp.status, cp.created_at
ORDER BY cp.created_at DESC;

-- name: ListBackofficeTrunks :many
SELECT
    t.id::TEXT AS id,
    COALESCE(o.name, 'Platform') AS organization_name,
    t.name,
    t.provisioning_mode,
    COALESCE(cp.name, '—') AS provider_name,
    t.direction,
    t.status,
    COUNT(te.id)::BIGINT AS endpoint_count
FROM trunks AS t
LEFT JOIN organizations AS o ON o.id = t.organization_id
LEFT JOIN carrier_connections AS cc ON cc.id = t.carrier_connection_id
LEFT JOIN carrier_providers AS cp ON cp.id = cc.provider_id
LEFT JOIN trunk_endpoints AS te ON te.trunk_id = t.id
GROUP BY
    t.id,
    o.name,
    t.name,
    t.provisioning_mode,
    cp.name,
    t.direction,
    t.status,
    t.created_at
ORDER BY t.created_at DESC
LIMIT 100;
