package numbers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type providerTestRepository struct {
	*fakeNumberRepository
	accepted int
	failed   int
	complete int
}

func (r *providerTestRepository) MarkProviderOperationAccepted(context.Context, sqlc.ProviderOperation, string, []byte) error {
	r.accepted++
	return nil
}

func (r *providerTestRepository) RecordProviderOperationFailure(context.Context, uuid.UUID, error) error {
	return nil
}

func (r *providerTestRepository) FailProviderOperation(context.Context, sqlc.ProviderOperation, error) error {
	r.failed++
	return nil
}

func (r *providerTestRepository) CompleteProviderOperation(context.Context, sqlc.ProviderOperation, ProviderOperationRequest, string, []byte) error {
	r.complete++
	return nil
}

type providerStub struct {
	order       ProviderOrder
	found       bool
	number      ProviderNumber
	findCalls   int
	createCalls int
}

func (p *providerStub) FindNumberOrder(context.Context, string) (ProviderOrder, bool, error) {
	p.findCalls++
	return p.order, p.found, nil
}

func (p *providerStub) CreateNumberOrder(context.Context, ProviderOrderRequest) (ProviderOrder, error) {
	p.createCalls++
	return p.order, nil
}

func (p *providerStub) FindManagedNumber(context.Context, string) (ProviderNumber, error) {
	return p.number, nil
}

func (p *providerStub) ConfigureNumberRouting(_ context.Context, id, routingResourceID string) (ProviderNumber, error) {
	p.number.ID = id
	p.number.RoutingResourceID = routingResourceID
	return p.number, nil
}

func testProviderOperation(t *testing.T, authorization ManagedNumberPurchaseAuthorization) sqlc.ProviderOperation {
	t.Helper()
	request, err := json.Marshal(ProviderOperationRequest{
		Provider:                  "didww",
		ProviderInventoryID:       "available-1",
		ProviderProductID:         "sku-1",
		Number:                    "+15551236001",
		CountryCode:               "US",
		CarrierConnectionID:       uuid.New(),
		ProviderRoutingResourceID: "voice-in-1",
		PurchaseAuthorization:     authorization,
	})
	if err != nil {
		t.Fatal(err)
	}
	return sqlc.ProviderOperation{
		ID:                uuid.New(),
		OrganizationID:    uuid.New(),
		CarrierProviderID: uuid.New(),
		PhoneNumberID:     uuid.New(),
		OperationType:     "number_provision",
		State:             "pending",
		Request:           request,
	}
}

func TestProviderOperationFailsClosedBeforeProviderWithoutAuthorization(t *testing.T) {
	authorization := ManagedNumberPurchaseAuthorization{
		ID: uuid.New(), ReservationID: uuid.New(), PriceID: uuid.New(), AmountMinor: 2500, Currency: "USD",
	}
	repo := &providerTestRepository{fakeNumberRepository: &fakeNumberRepository{}}
	authority := &fakeManagedPurchaseAuthority{verifyErr: errors.New("authorization unavailable")}
	provider := &providerStub{}
	service := NewService(repo)
	service.SetManagedPurchaseAuthority(authority)
	service.SetManagedProvider("didww", provider)

	if err := service.ExecuteProviderOperation(context.Background(), testProviderOperation(t, authorization)); err != nil {
		t.Fatalf("ExecuteProviderOperation() error = %v", err)
	}
	if provider.findCalls != 0 || provider.createCalls != 0 {
		t.Fatalf("provider calls = find:%d create:%d, want none", provider.findCalls, provider.createCalls)
	}
	if authority.releases != 1 || repo.failed != 1 {
		t.Fatalf("release/fail = %d/%d, want 1/1", authority.releases, repo.failed)
	}
}

func TestProviderOperationCapturesBeforeCompletion(t *testing.T) {
	authorization := ManagedNumberPurchaseAuthorization{
		ID: uuid.New(), ReservationID: uuid.New(), PriceID: uuid.New(), AmountMinor: 2500, Currency: "USD",
	}
	repo := &providerTestRepository{fakeNumberRepository: &fakeNumberRepository{}}
	authority := &fakeManagedPurchaseAuthority{}
	provider := &providerStub{
		order:  ProviderOrder{ID: "order-1", Status: "completed"},
		found:  true,
		number: ProviderNumber{ID: "did-1", Number: "+15551236001", RoutingResourceID: "voice-in-1"},
	}
	service := NewService(repo)
	service.SetManagedPurchaseAuthority(authority)
	service.SetManagedProvider("didww", provider)

	if err := service.ExecuteProviderOperation(context.Background(), testProviderOperation(t, authorization)); err != nil {
		t.Fatalf("ExecuteProviderOperation() error = %v", err)
	}
	if authority.captures != 1 || repo.complete != 1 {
		t.Fatalf("capture/complete = %d/%d, want 1/1", authority.captures, repo.complete)
	}
	if authority.releases != 0 || repo.failed != 0 {
		t.Fatalf("release/fail = %d/%d, want 0/0", authority.releases, repo.failed)
	}
}

func TestProviderOperationReleasesOnTerminalProviderFailure(t *testing.T) {
	authorization := ManagedNumberPurchaseAuthorization{
		ID: uuid.New(), ReservationID: uuid.New(), PriceID: uuid.New(), AmountMinor: 2500, Currency: "USD",
	}
	repo := &providerTestRepository{fakeNumberRepository: &fakeNumberRepository{}}
	authority := &fakeManagedPurchaseAuthority{}
	provider := &providerStub{order: ProviderOrder{ID: "order-1", Status: "failed"}, found: true}
	service := NewService(repo)
	service.SetManagedPurchaseAuthority(authority)
	service.SetManagedProvider("didww", provider)

	if err := service.ExecuteProviderOperation(context.Background(), testProviderOperation(t, authorization)); err != nil {
		t.Fatalf("ExecuteProviderOperation() error = %v", err)
	}
	if authority.releases != 1 || repo.failed != 1 {
		t.Fatalf("release/fail = %d/%d, want 1/1", authority.releases, repo.failed)
	}
	if authority.captures != 0 || repo.complete != 0 {
		t.Fatalf("capture/complete = %d/%d, want 0/0", authority.captures, repo.complete)
	}
}

func TestCompletedProviderOrderKeepsCaptureWhenRoutingIdentityFails(t *testing.T) {
	authorization := ManagedNumberPurchaseAuthorization{
		ID: uuid.New(), ReservationID: uuid.New(), PriceID: uuid.New(), AmountMinor: 2500, Currency: "USD",
	}
	repo := &providerTestRepository{fakeNumberRepository: &fakeNumberRepository{}}
	authority := &fakeManagedPurchaseAuthority{}
	provider := &providerStub{
		order:  ProviderOrder{ID: "order-1", Status: "completed"},
		found:  true,
		number: ProviderNumber{ID: "did-1", Number: "+15559999999", RoutingResourceID: "voice-in-1"},
	}
	service := NewService(repo)
	service.SetManagedPurchaseAuthority(authority)
	service.SetManagedProvider("didww", provider)

	if err := service.ExecuteProviderOperation(context.Background(), testProviderOperation(t, authorization)); err != nil {
		t.Fatalf("ExecuteProviderOperation() error = %v", err)
	}
	if authority.captures != 1 || authority.releases != 0 {
		t.Fatalf("capture/release = %d/%d, want 1/0", authority.captures, authority.releases)
	}
	if repo.failed != 1 || repo.complete != 0 {
		t.Fatalf("failed/complete = %d/%d, want 1/0", repo.failed, repo.complete)
	}
}
