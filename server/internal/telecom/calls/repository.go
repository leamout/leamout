package calls

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	if db == nil {
		panic("calls: database is required")
	}
	return &Repository{db: db, queries: sqlc.New(db)}
}

func (r *Repository) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	req CreateCallRequest,
	sipCallID string,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallInitiated, StatusInitiated, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.CreateCall(ctx, sqlc.CreateCallParams{
			OrganizationID: organizationID,
			ApplicationID:  req.ApplicationID,
			Direction:      string(DirectionOutbound),
			State:          "initiating",
			FromUri:        req.From,
			ToUri:          req.To,
			SipCallID:      &sipCallID,
		})
	})
}

func (r *Repository) CreateInbound(ctx context.Context, event InboundCallEvent) (sqlc.Call, error) {
	applicationID := event.ApplicationID
	channelID := event.ChannelID
	return r.mutateWithEvent(ctx, EventCallRinging, StatusRinging, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.CreateCall(ctx, sqlc.CreateCallParams{
			OrganizationID: event.OrganizationID,
			ApplicationID:  &applicationID,
			Direction:      string(DirectionInbound),
			State:          "ringing",
			FromUri:        event.From,
			ToUri:          event.To,
			SipCallID:      &channelID,
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

func (r *Repository) GetBySIPCallID(ctx context.Context, organizationID uuid.UUID, sipCallID string) (sqlc.Call, error) {
	return r.queries.GetCallBySIPCallID(ctx, sqlc.GetCallBySIPCallIDParams{
		OrganizationID: organizationID,
		SipCallID:      &sipCallID,
	})
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

func (r *Repository) MarkRinging(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallRinging, StatusRinging, func(q *sqlc.Queries) (sqlc.Call, error) {
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
	return r.mutateWithEvent(ctx, EventCallAnswered, StatusAnswered, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallAnswered(ctx, sqlc.MarkCallAnsweredParams{
			OrganizationID: organizationID,
			ID:             id,
		})
	})
}

func (r *Repository) MarkActive(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallActive, StatusActive, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallActive(ctx, sqlc.MarkCallActiveParams{
			OrganizationID: organizationID,
			ID:             id,
		})
	})
}

func (r *Repository) MarkCompleted(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
	hangupReason *string,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallCompleted, StatusCompleted, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallCompleted(ctx, sqlc.MarkCallCompletedParams{
			OrganizationID: organizationID,
			ID:             id,
			HangupReason:   hangupReason,
		})
	})
}

func (r *Repository) MarkFailed(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
	hangupReason *string,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallFailed, StatusFailed, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallFailed(ctx, sqlc.MarkCallFailedParams{
			OrganizationID: organizationID,
			ID:             id,
			HangupReason:   hangupReason,
		})
	})
}

func (r *Repository) MarkCancelled(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
	hangupReason *string,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallCancelled, StatusCancelled, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallCancelled(ctx, sqlc.MarkCallCancelledParams{
			OrganizationID: organizationID,
			ID:             id,
			HangupReason:   hangupReason,
		})
	})
}

type callMutation func(*sqlc.Queries) (sqlc.Call, error)

func (r *Repository) mutateWithEvent(
	ctx context.Context,
	eventType CallEventType,
	status CallStatus,
	mutation callMutation,
) (sqlc.Call, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.Call{}, fmt.Errorf("begin call transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	call, err := mutation(r.queries.WithTx(tx))
	if err != nil {
		return sqlc.Call{}, err
	}

	eventID := uuid.New()
	occurredAt := time.Now().UTC()
	event := callDomainEvent(call, eventType, status, occurredAt)
	payload, err := json.Marshal(event)
	if err != nil {
		return sqlc.Call{}, fmt.Errorf("marshal call event: %w", err)
	}
	headers, err := json.Marshal(map[string]string{
		"event_id":        eventID.String(),
		"event_type":      string(eventType),
		"organization_id": call.OrganizationID.String(),
		"schema_version":  "1",
	})
	if err != nil {
		return sqlc.Call{}, fmt.Errorf("marshal call event headers: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO call_events (id, organization_id, call_id, event_type, payload, occurred_at)
VALUES ($1, $2, $3, $4, $5::jsonb, $6)
`, eventID, call.OrganizationID, call.ID, string(eventType), payload, occurredAt); err != nil {
		return sqlc.Call{}, fmt.Errorf("insert call event: %w", err)
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO outbox_events (id, subject, aggregate_type, aggregate_id, payload, headers)
VALUES ($1, $2, 'call', $3, $4::jsonb, $5::jsonb)
`, eventID, string(eventType), call.ID, payload, headers); err != nil {
		return sqlc.Call{}, fmt.Errorf("insert call outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Call{}, fmt.Errorf("commit call transaction: %w", err)
	}
	return call, nil
}

func callDomainEvent(call sqlc.Call, eventType CallEventType, status CallStatus, occurredAt time.Time) CallEvent {
	applicationID := ""
	if call.ApplicationID != nil {
		applicationID = call.ApplicationID.String()
	}
	sipCallID := ""
	if call.SipCallID != nil {
		sipCallID = *call.SipCallID
	}
	hangupReason := ""
	if call.HangupReason != nil {
		hangupReason = *call.HangupReason
	}

	return CallEvent{
		EventType:      eventType,
		CallID:         call.ID.String(),
		SIPCallID:      sipCallID,
		ApplicationID:  applicationID,
		OrganizationID: call.OrganizationID.String(),
		From:           call.FromUri,
		To:             call.ToUri,
		Direction:      CallDirection(call.Direction),
		Status:         status,
		HangupReason:   hangupReason,
		OccurredAt:     occurredAt,
	}
}
