package subscriptions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, organizationID uuid.UUID) (*Subscription, error) {
	if organizationID == uuid.Nil {
		return nil, apperror.NewBadRequest("organization_id is required")
	}
	value, err := s.repo.Get(ctx, organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, apperror.NewInternal("get organization subscription", err)
	}
	return &value, nil
}

func (s *Service) Sync(
	ctx context.Context,
	organizationID uuid.UUID,
	req SyncRequest,
) (Subscription, error) {
	if organizationID == uuid.Nil {
		return Subscription{}, apperror.NewBadRequest("organization_id is required")
	}
	if err := validateSync(req); err != nil {
		return Subscription{}, err
	}
	existing, err := s.Get(ctx, organizationID)
	if err != nil {
		return Subscription{}, err
	}
	id := uuid.New()
	if existing != nil {
		id = existing.ID
	}
	value, err := s.repo.Upsert(ctx, id, organizationID, req)
	if err != nil {
		return Subscription{}, apperror.NewInternal("sync organization subscription", err)
	}
	return value, nil
}
