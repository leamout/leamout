package calls

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leamout/leamout/internal/database/sqlc"
)

func (r *Repository) createOutbound(
	ctx context.Context,
	organizationID uuid.UUID,
	req CreateCallRequest,
	route RouteAttribution,
	sipCallID string,
) (sqlc.Call, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.Call{}, fmt.Errorf("begin outbound call transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := r.queries.WithTx(tx)
	call, err := queries.CreateCall(ctx, sqlc.CreateCallParams{
		OrganizationID: organizationID,
		ApplicationID:  req.ApplicationID,
		Direction:      string(DirectionOutbound),
		State:          "initiating",
		FromUri:        req.From,
		ToUri:          req.To,
		SipCallID:      &sipCallID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23505" || pgErr.ConstraintName != "uq_calls_sip_call_id" {
			return sqlc.Call{}, err
		}
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return sqlc.Call{}, fmt.Errorf("rollback duplicate outbound call transaction: %w", rollbackErr)
		}
		existing, getErr := r.GetBySIPCallID(ctx, organizationID, sipCallID)
		if getErr != nil {
			return sqlc.Call{}, fmt.Errorf("get existing outbound call after SIP identity conflict: %w", getErr)
		}
		if existing.Direction != string(DirectionOutbound) {
			return sqlc.Call{}, fmt.Errorf("SIP call identity %q belongs to %s call %s, not outbound", sipCallID, existing.Direction, existing.ID)
		}
		return existing, nil
	}

	if _, err := tx.Exec(ctx, `
UPDATE calls
SET carrier_connection_id = $1,
    trunk_id = $2,
    trunk_endpoint_id = $3
WHERE organization_id = $4
  AND id = $5
`, route.CarrierConnectionID, route.TrunkID, route.TrunkEndpointID, organizationID, call.ID); err != nil {
		return sqlc.Call{}, fmt.Errorf("persist outbound call route attribution: %w", err)
	}
	call.CarrierConnectionID = &route.CarrierConnectionID
	call.TrunkID = &route.TrunkID
	call.TrunkEndpointID = &route.TrunkEndpointID

	if err := r.insertEvent(ctx, tx, call, EventCallInitiated); err != nil {
		return sqlc.Call{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return sqlc.Call{}, fmt.Errorf("commit outbound call transaction: %w", err)
	}
	return call, nil
}
