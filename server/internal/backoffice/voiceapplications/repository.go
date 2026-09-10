package voiceapplications

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(queries *sqlc.Queries) *Repository { return &Repository{queries: queries} }
func (r *Repository) List(ctx context.Context) ([]VoiceApplication, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	return r.queries.ListBackofficeVoiceApplications(ctx)
}
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	return r.queries.GetBackofficeVoiceApplication(ctx, id)
}
