package conferences

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(queries *sqlc.Queries) *Repository { return &Repository{queries: queries} }
func (r *Repository) Create(ctx context.Context, org uuid.UUID, req CreateRequest) (sqlc.Conference, error) {
	return r.queries.CreateConference(ctx, sqlc.CreateConferenceParams{OrganizationID: org, ApplicationID: req.ApplicationID, Name: req.Name, StartedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}})
}
func (r *Repository) Get(ctx context.Context, org, id uuid.UUID) (sqlc.Conference, error) {
	return r.queries.GetConference(ctx, sqlc.GetConferenceParams{OrganizationID: org, ID: id})
}
func (r *Repository) List(ctx context.Context, org uuid.UUID, offset, limit int32) ([]sqlc.Conference, error) {
	return r.queries.ListConferences(ctx, sqlc.ListConferencesParams{OrganizationID: org, PageOffset: offset, PageLimit: limit})
}
func (r *Repository) SetState(ctx context.Context, org, id uuid.UUID, state string) (sqlc.Conference, error) {
	return r.queries.UpdateConferenceState(ctx, sqlc.UpdateConferenceStateParams{OrganizationID: org, ID: id, State: state})
}
func (r *Repository) CreateParticipant(ctx context.Context, org, conferenceID uuid.UUID, req AddParticipantRequest) (sqlc.ConferenceParticipant, error) {
	return r.queries.CreateConferenceParticipant(ctx, sqlc.CreateConferenceParticipantParams{OrganizationID: org, ConferenceID: conferenceID, CallParticipantID: req.CallParticipantID, State: "joined"})
}
func (r *Repository) GetParticipant(ctx context.Context, org, id uuid.UUID) (sqlc.ConferenceParticipant, error) {
	return r.queries.GetConferenceParticipant(ctx, sqlc.GetConferenceParticipantParams{OrganizationID: org, ID: id})
}
func (r *Repository) ListParticipants(ctx context.Context, org, conferenceID uuid.UUID) ([]sqlc.ConferenceParticipant, error) {
	return r.queries.ListConferenceParticipants(ctx, sqlc.ListConferenceParticipantsParams{OrganizationID: org, ConferenceID: conferenceID})
}
func (r *Repository) SetParticipant(ctx context.Context, org, id uuid.UUID, state string, muted, deaf *bool) (sqlc.ConferenceParticipant, error) {
	return r.queries.UpdateConferenceParticipantState(ctx, sqlc.UpdateConferenceParticipantStateParams{OrganizationID: org, ID: id, State: state, Muted: muted, Deaf: deaf})
}
