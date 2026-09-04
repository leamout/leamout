package calls

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/modules/outbox"
)

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
	outbox  *outbox.Repository
}

func NewRepository(db *pgxpool.Pool) *Repository {
	if db == nil {
		panic("calls: database is required")
	}
	queries := sqlc.New(db)
	return &Repository{
		db:      db,
		queries: queries,
		outbox:  outbox.NewRepository(queries),
	}
}

func (r *Repository) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	req CreateCallRequest,
	route RouteAttribution,
	sipCallID string,
) (sqlc.Call, error) {
	return r.createWithEvent(ctx, organizationID, sipCallID, DirectionOutbound, EventCallInitiated, func(q *sqlc.Queries) (sqlc.Call, error) {
		call, err := q.CreateCall(ctx, sqlc.CreateCallParams{
			OrganizationID: organizationID,
			ApplicationID:  req.ApplicationID,
			Direction:      string(DirectionOutbound),
			State:          "initiating",
			FromUri:        req.From,
			ToUri:          req.To,
			SipCallID:      &sipCallID,
		})
		if err != nil {
			return sqlc.Call{}, err
		}
		return q.SetCallRouteAttribution(ctx, sqlc.SetCallRouteAttributionParams{
			CarrierConnectionID: &route.CarrierConnectionID,
			TrunkID:             &route.TrunkID,
			TrunkEndpointID:     &route.TrunkEndpointID,
			OrganizationID:      organizationID,
			ID:                  call.ID,
		})
	})
}

func (r *Repository) CreateInbound(ctx context.Context, event InboundCallEvent) (sqlc.Call, error) {
	applicationID := event.ApplicationID
	channelID := event.ChannelID
	return r.createWithEvent(ctx, event.OrganizationID, channelID, DirectionInbound, EventCallRinging, func(q *sqlc.Queries) (sqlc.Call, error) {
		call, err := q.CreateCall(ctx, sqlc.CreateCallParams{
			OrganizationID: event.OrganizationID,
			ApplicationID:  &applicationID,
			Direction:      string(DirectionInbound),
			State:          "ringing",
			FromUri:        event.From,
			ToUri:          event.To,
			SipCallID:      &channelID,
		})
		if err != nil {
			return sqlc.Call{}, err
		}
		return q.SetCallRouteAttribution(ctx, sqlc.SetCallRouteAttributionParams{
			CarrierConnectionID: &event.CarrierConnectionID,
			OrganizationID:      event.OrganizationID,
			ID:                  call.ID,
		})
	})
}

func (r *Repository) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.queries.GetCall(ctx, sqlc.GetCallParams{
		OrganizationID: organizationID,
		ID:             id,
	})
}

func (r *Repository) GetBySIPCallID(
	ctx context.Context,
	organizationID uuid.UUID,
	sipCallID string,
) (sqlc.Call, error) {
	return r.queries.GetCallBySIPCallID(ctx, sqlc.GetCallBySIPCallIDParams{
		OrganizationID: organizationID,
		SipCallID:      &sipCallID,
	})
}

func (r *Repository) GetBySIPCallIDGlobal(ctx context.Context, sipCallID string) (sqlc.Call, error) {
	return r.queries.GetCallBySIPCallIDGlobal(ctx, &sipCallID)
}

func (r *Repository) List(
	ctx context.Context,
	organizationID uuid.UUID,
	state *string,
	offset int32,
	limit int32,
) ([]sqlc.Call, error) {
	return r.queries.ListCalls(ctx, sqlc.ListCallsParams{
		OrganizationID: organizationID,
		State:          state,
		PageOffset:     offset,
		PageLimit:      limit,
	})
}

func (r *Repository) ListForReconciliation(
	ctx context.Context,
	updatedBefore time.Time,
	batchSize int32,
) ([]sqlc.Call, error) {
	return r.queries.ListCallsForReconciliation(ctx, sqlc.ListCallsForReconciliationParams{
		UpdatedBefore: pgtype.Timestamptz{
			Time:  updatedBefore,
			Valid: true,
		},
		BatchSize: batchSize,
	})
}

// CarrierDailySeconds derives usage from durable call timestamps. Active calls
// contribute through the current database time, so API restarts cannot reset
// or undercount the daily allowance.
func (r *Repository) CarrierDailySeconds(ctx context.Context, carrierID uuid.UUID, day time.Time) (int64, error) {
	var seconds int64
	err := r.db.QueryRow(ctx, `
SELECT COALESCE(SUM(GREATEST(0, EXTRACT(EPOCH FROM (COALESCE(ended_at, NOW()) - GREATEST(answered_at, $2))))), 0)::BIGINT
FROM calls
WHERE carrier_connection_id = $1
  AND answered_at IS NOT NULL
  AND answered_at < $2 + INTERVAL '1 day'
  AND COALESCE(ended_at, NOW()) >= $2`, carrierID, day).Scan(&seconds)
	return seconds, err
}

// InboundCallLimits revalidates the trusted SIP metadata against the current
// number/application/carrier relationship and returns limits from the resolved
// carrier connection. Platform connections are intentionally not required to
// belong to the call's organization; managed ingress tenancy comes from the DID.
func (r *Repository) InboundCallLimits(ctx context.Context, event InboundCallEvent) (CallLimits, error) {
	row, err := r.queries.GetInboundCallContext(ctx, sqlc.GetInboundCallContextParams{
		PhoneNumberID:       event.PhoneNumberID,
		OrganizationID:      event.OrganizationID,
		CalledNumber:        event.To,
		CarrierConnectionID: event.CarrierConnectionID,
		ApplicationID:       event.ApplicationID,
	})
	if err != nil {
		return CallLimits{}, err
	}
	return CallLimits{
		CarrierConnectionID: event.CarrierConnectionID,
		MaxCPS:              row.MaxCps,
		MaxConcurrent:       row.MaxConcurrentCalls,
		MaxDailyMinutes:     row.MaxDailyMinutes,
	}, nil
}

func (r *Repository) MarkRinging(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallRinging, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallRinging(ctx, sqlc.MarkCallRingingParams{
			OrganizationID: organizationID,
			ID:             id,
		})
	})
}

func (r *Repository) MarkAnswered(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallAnswered, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallAnswered(ctx, sqlc.MarkCallAnsweredParams{
			OrganizationID: organizationID,
			ID:             id,
		})
	})
}

func (r *Repository) MarkActive(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallActive, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallActive(ctx, sqlc.MarkCallActiveParams{OrganizationID: organizationID, ID: id})
	})
}

func (r *Repository) MarkHeld(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallHeld, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallHeld(ctx, sqlc.MarkCallHeldParams{OrganizationID: organizationID, ID: id})
	})
}

func (r *Repository) MarkResumed(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallResumed, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallResumed(ctx, sqlc.MarkCallResumedParams{OrganizationID: organizationID, ID: id})
	})
}

func (r *Repository) MarkCompleted(ctx context.Context, organizationID, id uuid.UUID, reason *string) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallCompleted, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallCompleted(ctx, sqlc.MarkCallCompletedParams{
			OrganizationID: organizationID,
			ID:             id,
			HangupReason:   reason,
		})
	})
}

func (r *Repository) MarkFailed(ctx context.Context, organizationID, id uuid.UUID, reason *string) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallFailed, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallFailed(ctx, sqlc.MarkCallFailedParams{
			OrganizationID: organizationID,
			ID:             id,
			HangupReason:   reason,
		})
	})
}

func (r *Repository) MarkCancelled(ctx context.Context, organizationID, id uuid.UUID, reason *string) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallCancelled, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallCancelled(ctx, sqlc.MarkCallCancelledParams{
			OrganizationID: organizationID,
			ID:             id,
			HangupReason:   reason,
		})
	})
}

func (r *Repository) createWithEvent(
	ctx context.Context,
	organizationID uuid.UUID,
	sipCallID string,
	direction CallDirection,
	eventType CallEventType,
	mutation func(*sqlc.Queries) (sqlc.Call, error),
) (sqlc.Call, error) {
	var result sqlc.Call
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		call, err := mutation(q)
		if err != nil {
			return err
		}
		if err := r.publishCallEvent(ctx, q, call, sipCallID, direction, eventType); err != nil {
			return err
		}
		result = call
		return nil
	})
	return result, err
}

func (r *Repository) mutateWithEvent(
	ctx context.Context,
	eventType CallEventType,
	mutation func(*sqlc.Queries) (sqlc.Call, error),
) (sqlc.Call, error) {
	var result sqlc.Call
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		call, err := mutation(q)
		if err != nil {
			return err
		}
		if err := r.publishCallEvent(ctx, q, call, stringValue(call.SipCallID), CallDirection(call.Direction), eventType); err != nil {
			return err
		}
		result = call
		return nil
	})
	return result, err
}

func (r *Repository) publishCallEvent(
	ctx context.Context,
	q *sqlc.Queries,
	call sqlc.Call,
	sipCallID string,
	direction CallDirection,
	eventType CallEventType,
) error {
	payload := CallEvent{
		EventType:      eventType,
		CallID:         call.ID.String(),
		SIPCallID:      sipCallID,
		ApplicationID:  uuidString(call.ApplicationID),
		OrganizationID: call.OrganizationID.String(),
		From:           call.FromUri,
		To:             call.ToUri,
		Direction:      direction,
		Status:         statusForState(call.State),
		MediaState:     CallMediaState(call.MediaState),
		HangupReason:   stringValue(call.HangupReason),
		DurationSec:    durationSeconds(call),
		OccurredAt:     time.Now().UTC(),
	}
	_, err := r.outbox.PublishCallEvent(ctx, q, call.OrganizationID, call.ID, string(eventType), payload)
	return err
}

func durationSeconds(call sqlc.Call) int {
	if !call.AnsweredAt.Valid || !call.EndedAt.Valid {
		return 0
	}
	seconds := int(call.EndedAt.Time.Sub(call.AnsweredAt.Time).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}

func uuidString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func readCall(call sqlc.Call, err error) (sqlc.Call, error) {
	if err == nil {
		return call, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Call{}, fmt.Errorf("call not found: %w", err)
	}
	return sqlc.Call{}, fmt.Errorf("read call: %w", err)
}

func writeError(err error, action string) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return fmt.Errorf("%s: invalid reference: %w", action, err)
		case "23505":
			return fmt.Errorf("%s: duplicate call: %w", action, err)
		case "23514":
			return fmt.Errorf("%s: invalid call state: %w", action, err)
		}
	}
	return fmt.Errorf("%s: %w", action, err)
}
