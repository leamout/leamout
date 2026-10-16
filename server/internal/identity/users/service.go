package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, userID uuid.UUID) (sqlc.User, error) {
	if err := validateUserID(userID); err != nil {
		return sqlc.User{}, err
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return sqlc.User{}, apperror.NewNotFound("user not found")
	}

	return user, nil
}

// Delete disables the user rather than physically removing the row. This
// preserves referential integrity while making the user immediately
// ineligible for authentication through the existing database predicates.
func (s *Service) Delete(ctx context.Context, userID uuid.UUID) error {
	if err := validateUserID(userID); err != nil {
		return err
	}

	if err := s.repo.Disable(ctx, userID); err != nil {
		return apperror.NewNotFound("user not found")
	}

	return nil
}

func validateUserID(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return apperror.NewUnauthorized("authentication required")
	}

	return nil
}

func toResponse(user sqlc.User) Response {
	return Response{
		ID:            user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Name:          user.Name,
		CreatedAt:     pgconv.TimestamptzToTime(user.CreatedAt),
		UpdatedAt:     pgconv.TimestamptzToTime(user.UpdatedAt),
	}
}
