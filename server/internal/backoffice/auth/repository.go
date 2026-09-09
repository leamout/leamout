package auth

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

func (r *Repository) GetUser(ctx context.Context, id uuid.UUID) (PlatformUser, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return PlatformUser{}, err
	}
	return PlatformUser{
		ID:              user.ID,
		IsPlatformAdmin: user.IsPlatformAdmin,
	}, nil
}
