package plans

import (
	"context"
	"errors"
	"strings"

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

func (s *Service) List(ctx context.Context) ([]Plan, error) {
	values, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, apperror.NewInternal("list subscription plans", err)
	}
	for i := range values {
		entitlements, err := s.repo.ListEntitlements(ctx, values[i].ID)
		if err != nil {
			return nil, apperror.NewInternal("list subscription plan entitlements", err)
		}
		values[i].Entitlements = entitlements
	}
	return values, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Plan, error) {
	if id == uuid.Nil {
		return Plan{}, apperror.NewBadRequest("plan_id is required")
	}
	value, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, apperror.NewNotFound("subscription plan not found")
	}
	if err != nil {
		return Plan{}, apperror.NewInternal("get subscription plan", err)
	}
	entitlements, err := s.repo.ListEntitlements(ctx, value.ID)
	if err != nil {
		return Plan{}, apperror.NewInternal("list subscription plan entitlements", err)
	}
	value.Entitlements = entitlements
	return value, nil
}

func (s *Service) GetByCode(ctx context.Context, code string) (Plan, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Plan{}, apperror.NewBadRequest("plan code is required")
	}
	value, err := s.repo.GetByCode(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, apperror.NewNotFound("subscription plan not found")
	}
	if err != nil {
		return Plan{}, apperror.NewInternal("get subscription plan", err)
	}
	entitlements, err := s.repo.ListEntitlements(ctx, value.ID)
	if err != nil {
		return Plan{}, apperror.NewInternal("list subscription plan entitlements", err)
	}
	value.Entitlements = entitlements
	return value, nil
}

func (s *Service) SetEntitlement(
	ctx context.Context,
	planID uuid.UUID,
	capability string,
	enabled bool,
) error {
	if planID == uuid.Nil {
		return apperror.NewBadRequest("plan_id is required")
	}
	capability = strings.TrimSpace(capability)
	if capability == "" {
		return apperror.NewBadRequest("capability is required")
	}
	if err := s.repo.SetEntitlement(ctx, planID, capability, enabled); err != nil {
		return apperror.NewInternal("set subscription plan entitlement", err)
	}
	return nil
}

func (s *Service) Upsert(ctx context.Context, id uuid.UUID, req UpsertRequest) (Plan, error) {
	if id == uuid.Nil {
		id = uuid.New()
	}
	normalized, err := validateUpsert(req)
	if err != nil {
		return Plan{}, err
	}
	value, err := s.repo.Upsert(ctx, id, normalized)
	if err != nil {
		return Plan{}, apperror.NewInternal("upsert subscription plan", err)
	}
	return value, nil
}
