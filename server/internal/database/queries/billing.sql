-- name: ListBillingReconciliationTargets :many
SELECT
    organization_id,
    COALESCE(provider_subscription_id, '')::TEXT AS subscription_id,
    COALESCE(provider_customer_id, '')::TEXT AS customer_id
FROM subscriptions
WHERE provider = 'polar'
  AND provider_subscription_id IS NOT NULL
  AND provider_customer_id IS NOT NULL
  AND organization_id > sqlc.arg(after_organization_id)
ORDER BY organization_id
LIMIT 100;
