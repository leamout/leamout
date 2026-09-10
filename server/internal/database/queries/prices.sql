-- name: GetPriceByID :one
SELECT *
FROM prices
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: ListPricesByPlan :many
SELECT *
FROM prices
WHERE plan_id = sqlc.arg(plan_id)
ORDER BY effective_from DESC, created_at DESC;

-- name: ListActivePricesByPlan :many
SELECT pr.*
FROM prices AS pr
JOIN plans AS pl ON pl.id = pr.plan_id
JOIN products AS p ON p.id = pl.product_id
WHERE pr.plan_id = sqlc.arg(plan_id)
  AND pr.active = true
  AND pr.effective_from <= sqlc.arg(at)
  AND (pr.effective_until IS NULL OR pr.effective_until > sqlc.arg(at))
  AND pl.active = true
  AND p.active = true
ORDER BY pr.effective_from DESC, pr.created_at DESC;

-- name: ListActiveMeteredPricesByPlanAndMeter :many
SELECT pr.*
FROM prices AS pr
JOIN plans AS pl ON pl.id = pr.plan_id
JOIN products AS p ON p.id = pl.product_id
JOIN meters AS m ON m.id = pr.meter_id
WHERE pr.plan_id = sqlc.arg(plan_id)
  AND pr.meter_id = sqlc.arg(meter_id)
  AND pr.pricing_type = 'metered'
  AND pr.active = true
  AND pr.effective_from <= sqlc.arg(at)
  AND (pr.effective_until IS NULL OR pr.effective_until > sqlc.arg(at))
  AND pl.active = true
  AND p.active = true
  AND m.active = true
ORDER BY pr.effective_from DESC, pr.created_at DESC;
