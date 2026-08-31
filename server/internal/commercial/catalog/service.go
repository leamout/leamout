package catalog

import (
	"context"

	"github.com/google/uuid"
)

type store interface {
	GetProduct(context.Context, uuid.UUID) (Product, error)
	GetProductByCode(context.Context, string) (Product, error)
	ListProducts(context.Context, bool) ([]Product, error)
	GetPlan(context.Context, uuid.UUID) (Plan, error)
	GetPlanByCode(context.Context, string) (Plan, error)
	ListPlans(context.Context, uuid.UUID, bool) ([]Plan, error)
}

// Service owns read-side catalog access used by commercial workflows.
type Service struct {
	store store
}

func NewService(store store) *Service {
	return &Service{store: store}
}

func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (Product, error) {
	if err := normalizeID(id); err != nil {
		return Product{}, err
	}
	return s.store.GetProduct(ctx, id)
}

func (s *Service) GetProductByCode(ctx context.Context, code string) (Product, error) {
	normalized, err := normalizeCode(code)
	if err != nil {
		return Product{}, err
	}
	return s.store.GetProductByCode(ctx, normalized)
}

func (s *Service) ListProducts(ctx context.Context, activeOnly bool) ([]Product, error) {
	return s.store.ListProducts(ctx, activeOnly)
}

func (s *Service) GetPlan(ctx context.Context, id uuid.UUID) (Plan, error) {
	if err := normalizeID(id); err != nil {
		return Plan{}, err
	}
	return s.store.GetPlan(ctx, id)
}

func (s *Service) GetPlanByCode(ctx context.Context, code string) (Plan, error) {
	normalized, err := normalizeCode(code)
	if err != nil {
		return Plan{}, err
	}
	return s.store.GetPlanByCode(ctx, normalized)
}

func (s *Service) ListPlans(ctx context.Context, productID uuid.UUID, activeOnly bool) ([]Plan, error) {
	if err := normalizeID(productID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetProduct(ctx, productID); err != nil {
		return nil, err
	}
	return s.store.ListPlans(ctx, productID, activeOnly)
}
