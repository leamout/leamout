package calls

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for calls.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]Call, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	rows, err := r.queries.ListBackofficeCalls(ctx)
	if err != nil {
		return nil, err
	}
	calls := make([]Call, 0, len(rows))
	for _, row := range rows {
		calls = append(calls, Call{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			Organization:   row.OrganizationName,
			From:           row.FromUri,
			To:             row.ToUri,
			Direction:      row.Direction,
			Duration:       formatDuration(row.DurationSeconds),
			Status:         row.State,
			CreatedAt:      row.CreatedAt,
		})
	}
	return calls, nil
}

func (r *Repository) Get(ctx context.Context, callID uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeCall(ctx, callID)
	if err != nil {
		return Detail{}, err
	}
	return Detail{
		Call: Call{ID: row.ID, OrganizationID: row.OrganizationID, Organization: row.OrganizationName,
			From: row.FromUri, To: row.ToUri, Direction: row.Direction,
			Duration: formatDuration(row.DurationSeconds), Status: row.State, CreatedAt: row.CreatedAt},
		MediaState: row.MediaState, SIPCallID: row.SipCallID,
		ApplicationID: row.ApplicationID, Application: row.ApplicationName,
		CarrierConnectionID: row.CarrierConnectionID, CarrierConnection: row.CarrierConnectionName,
		ProviderID: row.ProviderID, Provider: row.ProviderName,
		TrunkID: row.TrunkID, Trunk: row.TrunkName, TrunkEndpointID: row.TrunkEndpointID,
		HangupReason: row.HangupReason, RecordingCount: row.RecordingCount,
		RecordingStatus: row.RecordingStatus, StartedAt: row.StartedAt, AnsweredAt: row.AnsweredAt,
		EndedAt: row.EndedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}
