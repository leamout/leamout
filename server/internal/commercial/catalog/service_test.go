package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeCatalogStore struct {
	product Product
	plan    Plan
}

func (f *fakeCatalogStore) CreateProduct(context.Context, CreateProductInput) (Product, error) {
	return f.product, nil
}
func (f *fakeCatalogStore) GetProduct(context.Context, uuid.UUID) (Product, error) {
	if f.product.ID == uuid.Nil {
		return Product{}, ErrProductNotFound
	}
	return f.product, nil
}
func (f *fakeCatalogStore) GetProductByCode(context.Context, string) (Product, error) {
	return f.product, nil
}
func (f *fakeCatalogStore) ListProducts(context.Context, bool) ([]Product, error) {
	return []Product{f.product}, nil
}
func (f *fakeCatalogStore) UpdateProduct(context.Context, uuid.UUID, UpdateProductInput) (Product, error) {
	return f.product, nil
}
func (f *fakeCatalogStore) CreatePlan(context.Context, CreatePlanInput) (Plan, error) {
	return f.plan, nil
}
func (f *fakeCatalogStore) GetPlan(context.Context, uuid.UUID) (Plan, error) {
	if f.plan.ID == uuid.Nil {
		return Plan{}, ErrPlanNotFound
	}
	return f.plan, nil
}
func (f *fakeCatalogStore) GetPlanByCode(context.Context, string) (Plan, error) {
	return f.plan, nil
}
func (f *fakeCatalogStore) ListPlans(context.Context, uuid.UUID, bool) ([]Plan, error) {
	return []Plan{f.plan}, nil
}
func (f *fakeCatalogStore) UpdatePlan(context.Context, uuid.UUID, UpdatePlanInput) (Plan, error) {
	return f.plan, nil
}

func TestCreatePlanRejectsInactiveProduct(t *testing.T) {
	productID := uuid.New()
	store := &fakeCatalogStore{product: Product{ID: productID, Active: false}}
	service := NewService(store)

	_, err := service.CreatePlan(context.Background(), CreatePlanInput{
		ProductID: productID,
		Code:      "pro",
		Name:      "Pro",
	})
	if !errors.Is(err, ErrProductInactive) {
		t.Fatalf("CreatePlan() error = %v, want %v", err, ErrProductInactive)
	}
}

func TestUpdatePlanRejectsActivationUnderInactiveProduct(t *testing.T) {
	productID := uuid.New()
	planID := uuid.New()
	store := &fakeCatalogStore{
		product: Product{ID: productID, Active: false},
		plan:    Plan{ID: planID, ProductID: productID, Active: false},
	}
	service := NewService(store)
	active := true

	_, err := service.UpdatePlan(context.Background(), planID, UpdatePlanInput{Active: &active})
	if !errors.Is(err, ErrProductInactive) {
		t.Fatalf("UpdatePlan() error = %v, want %v", err, ErrProductInactive)
	}
}
