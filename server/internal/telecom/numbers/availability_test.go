package numbers

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeManagedInventory struct {
	request    AvailableSearchRequest
	candidates []ManagedNumberCandidate
	err        error
}

func (f *fakeManagedInventory) SearchAvailable(_ context.Context, request AvailableSearchRequest) ([]ManagedNumberCandidate, error) {
	f.request = request
	return f.candidates, f.err
}

type fakeSelectionStore struct {
	organizationID uuid.UUID
	candidate      ManagedNumberCandidate
	selectionID    string
	err            error
}

func (f *fakeSelectionStore) Save(_ context.Context, organizationID uuid.UUID, candidate ManagedNumberCandidate) (string, error) {
	f.organizationID = organizationID
	f.candidate = candidate
	return f.selectionID, f.err
}

func TestSearchAvailableReturnsOpaqueCustomerResponse(t *testing.T) {
	inventory := &fakeManagedInventory{candidates: []ManagedNumberCandidate{{
		Provider:              "didww",
		ProviderInventoryID:   "available-1",
		ProviderProductID:     "sku-1",
		Number:                "+12125550100",
		CountryCode:           "US",
		ChannelsIncludedCount: 2,
	}}}
	selections := &fakeSelectionStore{selectionID: "sel_test"}
	service := NewService(&fakeNumberRepository{})
	service.SetManagedAcquisition(inventory, selections)
	organizationID := uuid.New()

	result, err := service.SearchAvailable(context.Background(), organizationID, AvailableSearchRequest{
		CountryCode: " us ",
		Contains:    "+212",
	})
	if err != nil {
		t.Fatalf("SearchAvailable() error = %v", err)
	}
	if inventory.request.CountryCode != "US" || inventory.request.Contains != "212" {
		t.Fatalf("normalized request = %#v", inventory.request)
	}
	if selections.organizationID != organizationID {
		t.Fatalf("selection organization = %s", selections.organizationID)
	}
	if selections.candidate.ProviderInventoryID != "available-1" || selections.candidate.ProviderProductID != "sku-1" {
		t.Fatalf("stored selection = %#v", selections.candidate)
	}
	if len(result) != 1 {
		t.Fatalf("result count = %d", len(result))
	}
	if result[0].SelectionID != "sel_test" || result[0].Number != "+12125550100" || result[0].CountryCode != "US" || !result[0].VoiceEnabled {
		t.Fatalf("public result = %#v", result[0])
	}
}

func TestSearchAvailableRequiresConfiguredInventory(t *testing.T) {
	service := NewService(&fakeNumberRepository{})
	_, err := service.SearchAvailable(context.Background(), uuid.New(), AvailableSearchRequest{CountryCode: "US"})
	if err == nil {
		t.Fatal("SearchAvailable accepted an unconfigured managed inventory")
	}
}
