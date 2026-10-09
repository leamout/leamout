package audit

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, organizationID uuid.UUID, req ListRequest) ([]Event, error) {
	if err := validateFilters(req); err != nil {
		return nil, err
	}

	if organizationID == uuid.Nil {
		return nil, apperror.NewBadRequest("organization id is required")
	}
	if req.Limit < 1 || req.Limit > 100 {
		return nil, apperror.NewBadRequest("limit must be between 1 and 100")
	}
	if req.Offset < 0 {
		return nil, apperror.NewBadRequest("offset cannot be negative")
	}
	items, err := s.repo.List(ctx, organizationID, req)
	if err != nil {
		return nil, apperror.NewInternal("list audit events", err)
	}
	return items, nil
}
