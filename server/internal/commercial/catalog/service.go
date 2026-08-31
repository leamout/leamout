package catalog

import (
	"context"

	"github.com/google/uuid"
)

type store interface {
	CreateProduct(context.Context, CreateProductInput) (Product, error)
	GetProduct(context.Context, uuid.UUID) (Product, error)
	GetProductByCode(context.Context, string) (Product, error)
	ListProducts(context.Context, bool) ([]Product, error)
	UpdateProduct(context.Context, uuid.UUID, UpdateProductInput) (Product, error)
	CreatePlan(context.Context, CreatePlanInput) (Plan, error)
	GetPlan(context.Context, uuid.UUID) (Plan, error)
	GetPlanByCode(context.Context, string) (Plan, error)
	ListPlans(context.Context, uuid.UUID, bool) ([]Plan, error)
	UpdatePlan(context.Context, uuid.UUID, UpdatePlanInput) (Plan, error)
}

// Service owns catalog business rules for reusable products and plans.
type Service struct {
	store store
}

func NewService(store store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateProduct(ctx context.Context, input CreateProductInput) (Product, error) {
	normalized, err := normalizeCreateProduct(input)
	if err != nil {
		return Product{}, err
	}
	return s.store.CreateProduct(ctx, normalized)
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

func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, input UpdateProductInput) (Product, error) {
	if err := normalizeID(id); err != nil {
		return Product{}, err
	}
	normalized, err := normalizeUpdateProduct(input)
	if err != nil {
		return Product{}, err
	}
	return s.store.UpdateProduct(ctx, id, normalized)
}

func (s *Service) CreatePlan(ctx context.Context, input CreatePlanInput) (Plan, error) {
	normalized, err := normalizeCreatePlan(input)
	if err != nil {
		return Plan{}, err
	}
	product, err := s.store.GetProduct(ctx, normalized.ProductID)
	if err != nil {
		return Plan{}, err
	}
	if !product.Active {
		return Plan{}, ErrProductInactive
	}
	return s.store.CreatePlan(ctx, normalized)
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

func (s *Service) UpdatePlan(ctx context.Context, id uuid.UUID, input UpdatePlanInput) (Plan, error) {
	if err := normalizeID(id); err != nil {
		return Plan{}, err
	}
	normalized, err := normalizeUpdatePlan(input)
	if err != nil {
		return Plan{}, err
	}
	plan, err := s.store.GetPlan(ctx, id)
	if err != nil {
		return Plan{}, err
	}
	if normalized.Active != nil && *normalized.Active {
		product, err := s.store.GetProduct(ctx, plan.ProductID)
		if err != nil {
			return Plan{}, err
		}
		if !product.Active {
			return Plan{}, ErrProductInactive
		}
	}
	return s.store.UpdatePlan(ctx, id, normalized)
}
