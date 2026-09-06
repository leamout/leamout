package edge

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Resolve(ctx context.Context, req Request) (Route, error) {
	var route Route
	err := r.db.QueryRow(ctx, `
SELECT tc.organization_id, tc.trunk_id, cc.id, te.host, te.port, te.transport
FROM trunk_credentials tc
JOIN trunks customer_trunk ON customer_trunk.id = tc.trunk_id
 AND customer_trunk.organization_id = tc.organization_id
JOIN phone_numbers pn ON pn.organization_id = tc.organization_id
 AND pn.number = $3 AND pn.provisioning_mode = 'managed'
 AND pn.status = 'active' AND pn.voice_enabled = true
JOIN trunks platform_trunk ON platform_trunk.organization_id IS NULL
 AND platform_trunk.provisioning_mode = 'managed'
 AND platform_trunk.managed_default = true AND platform_trunk.status = 'active'
JOIN carrier_connections cc ON cc.id = platform_trunk.carrier_connection_id
 AND cc.scope = 'platform' AND cc.organization_id IS NULL AND cc.status = 'active'
JOIN trunk_endpoints te ON te.trunk_id = platform_trunk.id
 AND te.organization_id IS NULL AND te.enabled = true
 AND te.direction IN ('outbound', 'bidirectional') AND te.health_status <> 'down'
JOIN organizations o ON o.id = tc.organization_id
WHERE tc.username = $1 AND tc.realm = $2
 AND customer_trunk.provisioning_mode = 'managed'
 AND customer_trunk.carrier_connection_id IS NULL
 AND customer_trunk.status = 'active'
 AND customer_trunk.direction IN ('outbound', 'bidirectional')
 AND o.status = 'active' AND o.deleted_at IS NULL
ORDER BY te.priority, te.id
LIMIT 1`, req.Username, req.Realm, req.From).Scan(
		&route.OrganizationID, &route.TrunkID, &route.CarrierConnectionID,
		&route.Host, &route.Port, &route.Transport,
	)
	return route, err
}

func (r *Repository) DailyWholesaleSpend(ctx context.Context, organizationID uuid.UUID, day time.Time) (int64, error) {
	var spent int64
	err := r.db.QueryRow(ctx, `
SELECT COALESCE(sum(amount_micros), 0)::BIGINT
FROM wholesale_charges
WHERE organization_id = $1 AND occurred_at >= $2 AND occurred_at < $2 + interval '1 day'`, organizationID, day).Scan(&spent)
	return spent, err
}
