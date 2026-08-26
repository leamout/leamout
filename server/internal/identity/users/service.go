package users

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, userID uuid.UUID) (sqlc.User, error) {
	if userID == uuid.Nil {
		return sqlc.User{}, apperror.NewUnauthorized("authentication required")
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return sqlc.User{}, apperror.NewNotFound("user not found")
	}

	return user, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, name *string) (sqlc.User, error) {
	if userID == uuid.Nil {
		return sqlc.User{}, apperror.NewUnauthorized("authentication required")
	}

	if name != nil {
		value := strings.TrimSpace(*name)
		if value == "" {
			name = nil
		} else {
			name = &value
		}
	}

	user, err := s.repo.UpdateProfile(ctx, sqlc.UpdateUserProfileParams{
		ID:   userID,
		Name: pgconv.NullableString(name),
	})
	if err != nil {
		return sqlc.User{}, apperror.NewNotFound("user not found")
	}

	return user, nil
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
