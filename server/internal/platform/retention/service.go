package retention

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Service struct {
	repo          *Repository
	retentionDays func(context.Context, uuid.UUID) (int64, bool, error)
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}
func (s *Service) SetPlanRetentionResolver(
	resolve func(context.Context, uuid.UUID) (int64, bool, error),
) {
	s.retentionDays = resolve
}

func (s *Service) List(ctx context.Context, organizationID uuid.UUID) ([]Policy, error) {
	if organizationID == uuid.Nil {
		return nil, apperror.NewBadRequest("organization_id is required")
	}
	values, err := s.repo.List(ctx, organizationID)
	if err != nil {
		return nil, apperror.NewInternal("list retention policies", err)
	}
	return values, nil
}
func (s *Service) Upsert(ctx context.Context, organizationID uuid.UUID, resource string, req UpsertRequest) (Policy, error) {
	if organizationID == uuid.Nil {
		return Policy{}, apperror.NewBadRequest("organization_id is required")
	}
	if err := validateResource(resource); err != nil {
		return Policy{}, err
	}
	if err := validateUpsert(req); err != nil {
		return Policy{}, err
	}
	if s.retentionDays == nil {
		return Policy{}, apperror.NewServiceUnavailable(
			"commercial plan limit service is unavailable",
			nil,
		)
	}
	maxDays, limited, err := s.retentionDays(ctx, organizationID)
	if err != nil {
		return Policy{}, err
	}
	if limited && int64(req.RetentionDays) > maxDays {
		return Policy{}, apperror.NewPaymentRequired(
			"retention_days exceeds the organization plan limit",
		)
	}
	value, err := s.repo.Upsert(ctx, organizationID, resource, req)
	if err != nil {
		return Policy{}, apperror.NewInternal("upsert retention policy", err)
	}
	return value, nil
}
func (s *Service) Delete(ctx context.Context, organizationID uuid.UUID, resource string) error {
	if organizationID == uuid.Nil {
		return apperror.NewBadRequest("organization_id is required")
	}
	if err := validateResource(resource); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, organizationID, resource); err != nil {
		return apperror.NewInternal("delete retention policy", err)
	}
	return nil
}
