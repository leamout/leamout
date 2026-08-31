package catalog

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Service owns read-side catalog access used by commercial workflows.
type Service struct {
	repo *Repository
	now  func() time.Time
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (Product, error) {
	if err := normalizeID(id); err != nil {
		return Product{}, err
	}
	return s.repo.GetProduct(ctx, id)
}

func (s *Service) GetProductByCode(ctx context.Context, code string) (Product, error) {
	normalized, err := normalizeCode(code)
	if err != nil {
		return Product{}, err
	}
	return s.repo.GetProductByCode(ctx, normalized)
}

func (s *Service) ListProducts(ctx context.Context, activeOnly bool) ([]Product, error) {
	return s.repo.ListProducts(ctx, activeOnly)
}

func (s *Service) GetPlan(ctx context.Context, id uuid.UUID) (Plan, error) {
	if err := normalizeID(id); err != nil {
		return Plan{}, err
	}
	return s.repo.GetPlan(ctx, id)
}

func (s *Service) GetPlanByCode(ctx context.Context, code string) (Plan, error) {
	normalized, err := normalizeCode(code)
	if err != nil {
		return Plan{}, err
	}
	return s.repo.GetPlanByCode(ctx, normalized)
}

func (s *Service) ListPlans(ctx context.Context, productID uuid.UUID, activeOnly bool) ([]Plan, error) {
	if err := normalizeID(productID); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProduct(ctx, productID); err != nil {
		return nil, err
	}
	return s.repo.ListPlans(ctx, productID, activeOnly)
}

func (s *Service) GetPrice(ctx context.Context, id uuid.UUID) (Price, error) {
	if err := normalizeID(id); err != nil {
		return Price{}, err
	}
	return s.repo.GetPrice(ctx, id)
}

func (s *Service) ListPrices(ctx context.Context, planID uuid.UUID, activeOnly bool) ([]Price, error) {
	if err := normalizeID(planID); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetPlan(ctx, planID); err != nil {
		return nil, err
	}
	return s.repo.ListPrices(ctx, planID, activeOnly, s.now())
}
