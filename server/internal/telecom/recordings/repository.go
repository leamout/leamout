package recordings

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

func (r *Repository) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Recording, error) {
	return r.queries.GetRecording(ctx, sqlc.GetRecordingParams{
		OrganizationID: organizationID,
		ID:             id,
	})
}

func (r *Repository) List(
	ctx context.Context,
	organizationID uuid.UUID,
	offset int32,
	limit int32,
) ([]sqlc.Recording, error) {
	return r.queries.ListRecordings(ctx, sqlc.ListRecordingsParams{
		OrganizationID: organizationID,
		PageOffset:     offset,
		PageLimit:      limit,
	})
}

func (r *Repository) Delete(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.Recording, error) {
	return r.queries.DeleteRecording(ctx, sqlc.DeleteRecordingParams{
		OrganizationID: organizationID,
		ID:             id,
	})
}
