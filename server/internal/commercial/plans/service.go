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
		return nil, apperror.NewInternal("list plans", err)
	}
	return values, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Plan, error) {
	if id == uuid.Nil {
		return Plan{}, apperror.NewBadRequest("plan_id is required")
	}
	value, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, apperror.NewNotFound("plan not found")
	}
	if err != nil {
		return Plan{}, apperror.NewInternal("get plan", err)
	}
	return value, nil
}

func (s *Service) GetByCode(ctx context.Context, code string) (Plan, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Plan{}, apperror.NewBadRequest("plan code is required")
	}
	value, err := s.repo.GetByCode(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return Plan{}, apperror.NewNotFound("plan not found")
	}
	if err != nil {
		return Plan{}, apperror.NewInternal("get plan", err)
	}
	return value, nil
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
		return Plan{}, apperror.NewInternal("upsert plan", err)
	}
	return value, nil
}
