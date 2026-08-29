package calls

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	sipCallID string,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallInitiated, func(q *sqlc.Queries) (sqlc.Call, error) {
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
	return r.mutateWithEvent(ctx, EventCallRinging, func(q *sqlc.Queries) (sqlc.Call, error) {
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

func (r *Repository) MarkActive(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallActive, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallActive(ctx, sqlc.MarkCallActiveParams{
			OrganizationID: organizationID,
			ID:             id,
		})
	})
}

func (r *Repository) MarkHeld(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallHeld, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallHeld(ctx, sqlc.MarkCallHeldParams{
			OrganizationID: organizationID,
			ID:             id,
		})
	})
}

func (r *Repository) MarkResumed(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.mutateWithEvent(ctx, EventCallResumed, func(q *sqlc.Queries) (sqlc.Call, error) {
		return q.MarkCallResumed(ctx, sqlc.MarkCallResumedParams{
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
	return r.mutateWithEvent(ctx, EventCallCompleted, func(q *sqlc.Queries) (sqlc.Call, error) {
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
	return r.mutateWithEvent(ctx, EventCallFailed, func(q *sqlc.Queries) (sqlc.Call, error) {
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
	return r.mutateWithEvent(ctx, EventCallCancelled, func(q *sqlc.Queries) (sqlc.Call, error) {
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
	mutation callMutation,
) (sqlc.Call, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.Call{}, fmt.Errorf("begin call transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := r.queries.WithTx(tx)
	call, err := mutation(queries)
	if err != nil {
		return sqlc.Call{}, err
	}

	occurredAt := time.Now().UTC()
	event := callDomainEvent(call, eventType, occurredAt)
	if _, err := r.outbox.WithTx(tx).Insert(ctx, outbox.Event{
		Subject:       string(eventType),
		AggregateType: "call",
		AggregateID:   call.ID,
		Payload:       event,
		Headers: map[string]string{
			"event_type":      string(eventType),
			"organization_id": call.OrganizationID.String(),
			"schema_version":  "1",
		},
	}); err != nil {
		return sqlc.Call{}, fmt.Errorf("insert call outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Call{}, fmt.Errorf("commit call transaction: %w", err)
	}
	return call, nil
}

func callDomainEvent(call sqlc.Call, eventType CallEventType, occurredAt time.Time) CallEvent {
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
		Status:         statusForState(call.State),
		MediaState:     CallMediaState(call.MediaState),
		HangupReason:   hangupReason,
		OccurredAt:     occurredAt,
	}
}
