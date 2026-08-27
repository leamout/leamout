package calls

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(queries *sqlc.Queries) *Repository { return &Repository{queries: queries} }

func (r *Repository) Create(ctx context.Context, organizationID uuid.UUID, req CreateCallRequest, sipCallID string) (sqlc.Call, error) {
	return r.queries.CreateCall(ctx, sqlc.CreateCallParams{OrganizationID: organizationID, ApplicationID: req.ApplicationID, Direction: string(DirectionOutbound), State: "initiating", FromUri: req.From, ToUri: req.To, SipCallID: &sipCallID})
}
func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Call, error) {
	return r.queries.GetCall(ctx, sqlc.GetCallParams{OrganizationID: organizationID, ID: id})
}
func (r *Repository) List(ctx context.Context, organizationID uuid.UUID, state *string, offset, limit int32) ([]sqlc.Call, error) {
	return r.queries.ListCalls(ctx, sqlc.ListCallsParams{OrganizationID: organizationID, State: state, PageOffset: offset, PageLimit: limit})
}
func (r *Repository) MarkAnswered(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Call, error) {
	return r.queries.MarkCallAnswered(ctx, sqlc.MarkCallAnsweredParams{OrganizationID: organizationID, ID: id})
}
func (r *Repository) MarkCompleted(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Call, error) {
	return r.queries.MarkCallCompleted(ctx, sqlc.MarkCallCompletedParams{OrganizationID: organizationID, ID: id})
}
