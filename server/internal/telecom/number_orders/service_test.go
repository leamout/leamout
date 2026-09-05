package number_orders

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type fakeRepository struct {
	created CreateInput
	order   sqlc.NumberOrder
}

func (f *fakeRepository) Create(_ context.Context, input CreateInput) (sqlc.NumberOrder, error) {
	f.created = input
	return f.order, nil
}

func (f *fakeRepository) Get(context.Context, uuid.UUID, uuid.UUID) (sqlc.NumberOrder, error) {
	return f.order, nil
}

func (f *fakeRepository) List(context.Context, uuid.UUID) ([]sqlc.NumberOrder, error) {
	return []sqlc.NumberOrder{f.order}, nil
}

func TestCreateNormalizesTrustedSelection(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	_, err := service.Create(context.Background(), CreateInput{
		OrganizationID:      uuid.New(),
		ProviderID:          uuid.New(),
		ProviderInventoryID: " inventory-1 ",
		ProviderProductID:   " sku-1 ",
		Number:              " +233201234567 ",
		CountryCode:         " gh ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repo.created.ProviderInventoryID != "inventory-1" {
		t.Fatalf("provider_inventory_id = %q", repo.created.ProviderInventoryID)
	}
	if repo.created.ProviderProductID != "sku-1" {
		t.Fatalf("provider_product_id = %q", repo.created.ProviderProductID)
	}
	if repo.created.Number != "+233201234567" {
		t.Fatalf("number = %q", repo.created.Number)
	}
	if repo.created.CountryCode != "GH" {
		t.Fatalf("country_code = %q", repo.created.CountryCode)
	}
}

func TestCreateRejectsInvalidSelection(t *testing.T) {
	service := NewService(&fakeRepository{})
	base := CreateInput{
		OrganizationID:      uuid.New(),
		ProviderID:          uuid.New(),
		ProviderInventoryID: "inventory-1",
		ProviderProductID:   "sku-1",
		Number:              "+233201234567",
		CountryCode:         "GH",
	}

	tests := []struct {
		name   string
		mutate func(*CreateInput)
	}{
		{"missing provider", func(input *CreateInput) { input.ProviderID = uuid.Nil }},
		{"missing inventory", func(input *CreateInput) { input.ProviderInventoryID = " " }},
		{"missing product", func(input *CreateInput) { input.ProviderProductID = " " }},
		{"invalid number", func(input *CreateInput) { input.Number = "233201234567" }},
		{"invalid country", func(input *CreateInput) { input.CountryCode = "GHA" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := base
			tt.mutate(&input)
			if _, err := service.Create(context.Background(), input); err == nil {
				t.Fatal("Create() accepted invalid selection")
			}
		})
	}
}
