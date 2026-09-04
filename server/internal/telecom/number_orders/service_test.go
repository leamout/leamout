package number_orders

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type fakeRepository struct {
	created        CreateInput
	order          sqlc.NumberOrder
	markError      error
	failed         Failure
	failedExpected Status
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

func (f *fakeRepository) MarkPurchasing(context.Context, uuid.UUID) (sqlc.NumberOrder, error) {
	return f.order, f.markError
}

func (f *fakeRepository) MarkPurchased(context.Context, uuid.UUID, string) (sqlc.NumberOrder, error) {
	return f.order, f.markError
}

func (f *fakeRepository) MarkPersisting(context.Context, uuid.UUID, *string) (sqlc.NumberOrder, error) {
	return f.order, f.markError
}

func (f *fakeRepository) MarkConfiguring(context.Context, uuid.UUID, *uuid.UUID) (sqlc.NumberOrder, error) {
	return f.order, f.markError
}

func (f *fakeRepository) MarkCompleted(context.Context, uuid.UUID) (sqlc.NumberOrder, error) {
	return f.order, f.markError
}

func (f *fakeRepository) MarkFailed(_ context.Context, _ uuid.UUID, expected Status, failure Failure) (sqlc.NumberOrder, error) {
	f.failedExpected = expected
	f.failed = failure
	return f.order, f.markError
}

func (f *fakeRepository) ListRecoverable(context.Context, int32) ([]sqlc.NumberOrder, error) {
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

func TestTransitionConflictOnInvalidDatabaseState(t *testing.T) {
	service := NewService(&fakeRepository{markError: pgx.ErrNoRows})
	if _, err := service.BeginPurchase(context.Background(), uuid.New()); err == nil {
		t.Fatal("BeginPurchase() accepted an invalid transition")
	}
}

func TestFailRequiresMatchingStage(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.Fail(context.Background(), uuid.New(), StatusPersisting, Failure{
		Stage:   StageConfiguring,
		Message: "routing failed",
	})
	if err == nil {
		t.Fatal("Fail() accepted a mismatched failure stage")
	}
}

func TestFailRecordsRecoverableStage(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	_, err := service.Fail(context.Background(), uuid.New(), StatusConfiguring, Failure{
		Stage:   StageConfiguring,
		Code:    "provider_unavailable",
		Message: " routing failed ",
	})
	if err != nil {
		t.Fatalf("Fail() error = %v", err)
	}
	if repo.failedExpected != StatusConfiguring {
		t.Fatalf("expected status = %q", repo.failedExpected)
	}
	if repo.failed.Stage != StageConfiguring || repo.failed.Message != "routing failed" {
		t.Fatalf("failure = %#v", repo.failed)
	}
}

func TestListRecoverableValidatesBatchSize(t *testing.T) {
	service := NewService(&fakeRepository{})
	for _, limit := range []int32{0, -1, 501} {
		if _, err := service.ListRecoverable(context.Background(), limit); err == nil {
			t.Fatalf("ListRecoverable(%d) accepted invalid limit", limit)
		}
	}
}
