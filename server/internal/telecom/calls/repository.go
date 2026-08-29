package calls

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	req CreateCallRequest,
	sipCallID string,
) (sqlc.Call, error) {
	return r.queries.CreateCall(ctx, sqlc.CreateCallParams{
		OrganizationID: organizationID,
		ApplicationID:  req.ApplicationID,
		Direction:      string(DirectionOutbound),
		State:          "initiating",
		FromUri:        req.From,
		ToUri:          req.To,
		SipCallID:      &sipCallID,
	})
}

func (r *Repository) CreateInbound(ctx context.Context, event InboundCallEvent) (sqlc.Call, error) {
	applicationID := event.ApplicationID
	channelID := event.ChannelID
	return r.queries.CreateCall(ctx, sqlc.CreateCallParams{
		OrganizationID: event.OrganizationID,
		ApplicationID:  &applicationID,
		Direction:      string(DirectionInbound),
		State:          "ringing",
		FromUri:        event.From,
		ToUri:          event.To,
		SipCallID:      &channelID,
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
	return r.queries.MarkCallRinging(ctx, sqlc.MarkCallRingingParams{
		OrganizationID: organizationID,
		ID:             id,
	})
}

func (r *Repository) MarkAnswered(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.queries.MarkCallAnswered(ctx, sqlc.MarkCallAnsweredParams{
		OrganizationID: organizationID,
		ID:             id,
	})
}

func (r *Repository) MarkActive(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Call, error) {
	return r.queries.MarkCallActive(ctx, sqlc.MarkCallActiveParams{
		OrganizationID: organizationID,
		ID:             id,
	})
}

func (r *Repository) MarkCompleted(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
	hangupReason *string,
) (sqlc.Call, error) {
	return r.queries.MarkCallCompleted(ctx, sqlc.MarkCallCompletedParams{
		OrganizationID: organizationID,
		ID:             id,
		HangupReason:   hangupReason,
	})
}

func (r *Repository) MarkFailed(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
	hangupReason *string,
) (sqlc.Call, error) {
	return r.queries.MarkCallFailed(ctx, sqlc.MarkCallFailedParams{
		OrganizationID: organizationID,
		ID:             id,
		HangupReason:   hangupReason,
	})
}

func (r *Repository) MarkCancelled(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
	hangupReason *string,
) (sqlc.Call, error) {
	return r.queries.MarkCallCancelled(ctx, sqlc.MarkCallCancelledParams{
		OrganizationID: organizationID,
		ID:             id,
		HangupReason:   hangupReason,
	})
}
