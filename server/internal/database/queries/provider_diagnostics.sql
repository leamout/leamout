-- name: GetProviderOperationDiagnosticsSummary :one
SELECT
    COUNT(*) FILTER (WHERE state = 'pending')::BIGINT AS pending_count,
    COUNT(*) FILTER (WHERE state = 'provider_accepted')::BIGINT AS accepted_count,
    COUNT(*) FILTER (WHERE state = 'failed')::BIGINT AS failed_count,
    COUNT(*) FILTER (
        WHERE state IN ('pending', 'provider_accepted')
          AND next_attempt_at IS NOT NULL
          AND next_attempt_at <= now()
    )::BIGINT AS retry_ready_count
FROM provider_operations;

-- name: ListProviderOperationDiagnostics :many
SELECT
    po.id,
    cp.slug AS provider,
    po.operation_type,
    po.state,
    po.attempts,
    po.last_error,
    po.next_attempt_at,
    po.created_at,
    po.updated_at
FROM provider_operations AS po
JOIN carrier_providers AS cp ON cp.id = po.carrier_provider_id
WHERE po.state IN ('pending', 'provider_accepted', 'failed')
ORDER BY
    CASE po.state
        WHEN 'failed' THEN 0
        WHEN 'provider_accepted' THEN 1
        ELSE 2
    END,
    po.next_attempt_at ASC NULLS LAST,
    po.created_at ASC
LIMIT sqlc.arg(limit_count);

-- name: ListProviderCDRPollDiagnostics :many
SELECT
    c.provider,
    c.direction,
    c.window_date,
    c.page,
    c.attempt_count,
    c.next_attempt_at,
    c.last_error,
    c.last_success_at,
    last_page.received_at AS last_page_received_at
FROM provider_cdr_poll_cursors AS c
LEFT JOIN LATERAL (
    SELECT p.received_at
    FROM provider_cdr_pages AS p
    WHERE p.provider = c.provider
      AND p.direction = c.direction
    ORDER BY p.received_at DESC
    LIMIT 1
) AS last_page ON TRUE
ORDER BY c.provider, c.direction;
