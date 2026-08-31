package carrier_tests

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/modules/audit"
	"github.com/leamout/leamout/internal/telecom/routing"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, organizationID, connectionID uuid.UUID, actor audit.Actor, req Request) (Result, error) {
	var result Result
	err := r.db.QueryRow(ctx, `INSERT INTO carrier_test_calls (organization_id, carrier_connection_id, trunk_id, actor_type, actor_id, from_number, to_number) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, organization_id, carrier_connection_id, trunk_id, trunk_endpoint_id, actor_type, actor_id, from_number, to_number, status, sip_call_id, response_code, error_message, started_at, answered_at, completed_at`, organizationID, connectionID, req.TrunkID, actor.Type, actor.ID, req.From, req.To).Scan(resultFields(&result)...)
	return result, err
}

func (r *Repository) AttributeRoute(ctx context.Context, id uuid.UUID, route routing.OutboundDecision) error {
	_, err := r.db.Exec(ctx, `UPDATE carrier_test_calls SET trunk_endpoint_id=$2 WHERE id=$1`, id, route.EndpointID)
	return err
}

func (r *Repository) Finish(ctx context.Context, id uuid.UUID, status string, sipCallID, responseCode, message *string, answered bool) (Result, error) {
	var result Result
	err := r.db.QueryRow(ctx, `UPDATE carrier_test_calls SET status=$2, sip_call_id=COALESCE($3,sip_call_id), response_code=$4, error_message=$5, answered_at=CASE WHEN $6 THEN COALESCE(answered_at,NOW()) ELSE answered_at END, completed_at=NOW() WHERE id=$1 RETURNING id, organization_id, carrier_connection_id, trunk_id, trunk_endpoint_id, actor_type, actor_id, from_number, to_number, status, sip_call_id, response_code, error_message, started_at, answered_at, completed_at`, id, status, sipCallID, responseCode, message, answered).Scan(resultFields(&result)...)
	return result, err
}

func (r *Repository) List(ctx context.Context, organizationID, connectionID uuid.UUID, limit, offset int32) ([]Result, error) {
	rows, err := r.db.Query(ctx, `SELECT id, organization_id, carrier_connection_id, trunk_id, trunk_endpoint_id, actor_type, actor_id, from_number, to_number, status, sip_call_id, response_code, error_message, started_at, answered_at, completed_at FROM carrier_test_calls WHERE organization_id=$1 AND carrier_connection_id=$2 ORDER BY started_at DESC, id DESC LIMIT $3 OFFSET $4`, organizationID, connectionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Result, 0)
	for rows.Next() {
		var item Result
		if err := rows.Scan(resultFields(&item)...); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func resultFields(result *Result) []any {
	return []any{&result.ID, &result.OrganizationID, &result.CarrierConnectionID, &result.TrunkID, &result.TrunkEndpointID, &result.ActorType, &result.ActorID, &result.From, &result.To, &result.Status, &result.SIPCallID, &result.ResponseCode, &result.ErrorMessage, &result.StartedAt, &result.AnsweredAt, &result.CompletedAt}
}
