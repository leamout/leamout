package numbers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/modules/audit"
)

type fakeNumberRepository struct {
	createdCustomer       CreateRequest
	providerRequest       CreateRequest
	providerSelectionID   string
	providerAuthorization ManagedNumberPurchaseAuthorization
	providerCreateErr     error
	getNumber             sqlc.PhoneNumber
	getForRelease         sqlc.PhoneNumber
	releaseCalls          int
	selectionOrgID        uuid.UUID
	selectionCandidate    ManagedNumberCandidate
	selectionID           string
}

func (f *fakeNumberRepository) CreateCustomer(_ context.Context, _ uuid.UUID, req CreateRequest) (sqlc.PhoneNumber, error) {
	f.createdCustomer = req
	return sqlc.PhoneNumber{Number: req.Number, CountryCode: req.CountryCode, CarrierConnectionID: req.CarrierConnectionID, Status: "active"}, nil
}

func (f *fakeNumberRepository) CreateProvider(_ context.Context, _ uuid.UUID, req CreateRequest, selectionID string, authorization ManagedNumberPurchaseAuthorization) (sqlc.PhoneNumber, error) {
	f.providerRequest = req
	f.providerSelectionID = selectionID
	f.providerAuthorization = authorization
	if f.providerCreateErr != nil {
		return sqlc.PhoneNumber{}, f.providerCreateErr
	}
	return sqlc.PhoneNumber{CarrierConnectionID: req.CarrierConnectionID, Status: "provisioning"}, nil
}

func (f *fakeNumberRepository) List(context.Context, uuid.UUID) ([]sqlc.PhoneNumber, error) {
	return nil, nil
}
func (f *fakeNumberRepository) Get(context.Context, uuid.UUID, uuid.UUID) (sqlc.PhoneNumber, error) {
	return f.getNumber, nil
}
func (f *fakeNumberRepository) GetForRelease(context.Context, uuid.UUID, uuid.UUID) (sqlc.PhoneNumber, error) {
	return f.getForRelease, nil
}
func (f *fakeNumberRepository) Update(context.Context, uuid.UUID, uuid.UUID, UpdateRequest) (sqlc.PhoneNumber, error) {
	return f.getNumber, nil
}
func (f *fakeNumberRepository) ReleaseCustomer(context.Context, uuid.UUID, uuid.UUID) (sqlc.PhoneNumber, error) {
	f.releaseCalls++
	return f.getForRelease, nil
}
func (f *fakeNumberRepository) SetCarrierConnection(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, audit.Event) (sqlc.PhoneNumber, error) {
	return f.getNumber, nil
}
func (f *fakeNumberRepository) SaveManagedSelection(_ context.Context, organizationID uuid.UUID, candidate ManagedNumberCandidate) (string, error) {
	f.selectionOrgID = organizationID
	f.selectionCandidate = candidate
	return f.selectionID, nil
}
func (f *fakeNumberRepository) LoadManagedSelection(_ context.Context, organizationID uuid.UUID, selectionID string) (ManagedNumberCandidate, error) {
	f.selectionOrgID = organizationID
	if selectionID != f.selectionID && f.selectionID != "" {
		return ManagedNumberCandidate{}, ErrSelectionNotFound
	}
	return f.selectionCandidate, nil
}

type fakeManagedInventory struct {
	request    AvailableSearchRequest
	candidates []ManagedNumberCandidate
}

func (f *fakeManagedInventory) SearchAvailable(_ context.Context, request AvailableSearchRequest) ([]ManagedNumberCandidate, error) {
	f.request = request
	return f.candidates, nil
}

type fakeManagedPurchaseAuthority struct {
	priceID       uuid.UUID
	amountMinor   int64
	currency      string
	reservationID uuid.UUID
	reserveErr    error
	verifyErr     error
	captureErr    error
	releaseErr    error
	reserves      int
	verifies      int
	captures      int
	releases      int
}

func (f *fakeManagedPurchaseAuthority) QuoteManagedNumberPurchase(context.Context, uuid.UUID) (uuid.UUID, int64, string, error) {
	return f.priceID, f.amountMinor, f.currency, nil
}
func (f *fakeManagedPurchaseAuthority) ReserveManagedNumberPurchase(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int64, string) (uuid.UUID, error) {
	f.reserves++
	return f.reservationID, f.reserveErr
}
func (f *fakeManagedPurchaseAuthority) VerifyManagedNumberPurchase(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, int64, string) error {
	f.verifies++
	return f.verifyErr
}
func (f *fakeManagedPurchaseAuthority) CaptureManagedNumberPurchase(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, int64, string) error {
	f.captures++
	return f.captureErr
}
func (f *fakeManagedPurchaseAuthority) ReleaseManagedNumberPurchase(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	f.releases++
	return f.releaseErr
}

func TestCreateRegistersCustomerNumberWithoutType(t *testing.T) {
	repo := &fakeNumberRepository{}
	service := NewService(repo)
	_, err := service.Create(context.Background(), uuid.New(), CreateRequest{Number: " +233201234567 ", CountryCode: " gh "})
	if err != nil {
		t.Fatal(err)
	}
	if repo.createdCustomer.Number != "+233201234567" || repo.createdCustomer.CountryCode != "GH" {
		t.Fatalf("created=%+v", repo.createdCustomer)
	}
}

func TestCreateProvisionsSelectionWithoutType(t *testing.T) {
	priceID := uuid.New()
	reservationID := uuid.New()
	carrierConnectionID := uuid.New()
	repo := &fakeNumberRepository{
		selectionID: "sel_test",
		selectionCandidate: ManagedNumberCandidate{
			Provider: "didww", ProviderInventoryID: "available-1", ProviderProductID: "sku-1",
			Number: "+15551236001", CountryCode: "US", ChannelsIncludedCount: 2,
			PriceID: priceID, PriceAmountMinor: 2500, PriceCurrency: "USD",
		},
	}
	authority := &fakeManagedPurchaseAuthority{priceID: priceID, amountMinor: 2500, currency: "USD", reservationID: reservationID}
	service := NewService(repo)
	service.SetManagedPurchaseAuthority(authority)
	result, err := service.Create(context.Background(), uuid.New(), CreateRequest{SelectionID: " sel_test ", CarrierConnectionID: &carrierConnectionID})
	if err != nil {
		t.Fatal(err)
	}
	if repo.providerSelectionID != "sel_test" || repo.providerRequest.CarrierConnectionID == nil || *repo.providerRequest.CarrierConnectionID != carrierConnectionID {
		t.Fatalf("request=%+v selection=%q", repo.providerRequest, repo.providerSelectionID)
	}
	if repo.providerAuthorization.ReservationID != reservationID || repo.providerAuthorization.PriceID != priceID {
		t.Fatalf("authorization=%+v", repo.providerAuthorization)
	}
	if authority.reserves != 1 || authority.releases != 0 {
		t.Fatalf("reserve/release=%d/%d, want 1/0", authority.reserves, authority.releases)
	}
	if result.Status != "provisioning" {
		t.Fatalf("status=%q", result.Status)
	}
}

func TestProviderCreateReleasesReservationWhenPersistenceFails(t *testing.T) {
	priceID := uuid.New()
	repo := &fakeNumberRepository{
		selectionID: "sel_test",
		selectionCandidate: ManagedNumberCandidate{
			Provider: "didww", ProviderInventoryID: "available-1", ProviderProductID: "sku-1",
			Number: "+15551236001", CountryCode: "US", ChannelsIncludedCount: 2,
			PriceID: priceID, PriceAmountMinor: 2500, PriceCurrency: "USD",
		},
		providerCreateErr: errors.New("database unavailable"),
	}
	authority := &fakeManagedPurchaseAuthority{reservationID: uuid.New()}
	service := NewService(repo)
	service.SetManagedPurchaseAuthority(authority)
	_, err := service.Create(context.Background(), uuid.New(), CreateRequest{SelectionID: "sel_test"})
	if err == nil {
		t.Fatal("Create() unexpectedly succeeded")
	}
	if authority.reserves != 1 || authority.releases != 1 {
		t.Fatalf("reserve/release=%d/%d, want 1/1", authority.reserves, authority.releases)
	}
}

func TestSelectionCreateRejectsCustomerNumberFields(t *testing.T) {
	service := NewService(&fakeNumberRepository{})
	_, err := service.Create(context.Background(), uuid.New(), CreateRequest{SelectionID: "sel_test", Number: "+233201234567"})
	if err == nil {
		t.Fatal("selection create accepted caller-supplied number")
	}
}

func TestSearchAvailableReturnsOpaqueCustomerResponse(t *testing.T) {
	priceID := uuid.New()
	repo := &fakeNumberRepository{selectionID: "sel_test"}
	inventory := &fakeManagedInventory{candidates: []ManagedNumberCandidate{{Provider: "didww", ProviderInventoryID: "available-1", ProviderProductID: "sku-1", Number: "+12125550100", CountryCode: "US", ChannelsIncludedCount: 2}}}
	authority := &fakeManagedPurchaseAuthority{priceID: priceID, amountMinor: 2500, currency: "USD"}
	service := NewService(repo)
	service.SetManagedAcquisition(inventory)
	service.SetManagedPurchaseAuthority(authority)
	result, err := service.SearchAvailable(context.Background(), uuid.New(), AvailableSearchRequest{CountryCode: " us ", Contains: "+212"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].SelectionID != "sel_test" {
		t.Fatalf("result=%+v", result)
	}
	if result[0].Price.AmountMinor != 2500 || result[0].Price.Currency != "USD" {
		t.Fatalf("price=%+v", result[0].Price)
	}
	if repo.selectionCandidate.PriceID != priceID || repo.selectionCandidate.PriceAmountMinor != 2500 || repo.selectionCandidate.PriceCurrency != "USD" {
		t.Fatalf("stored selection quote=%+v", repo.selectionCandidate)
	}
}

func TestReleaseRejectsProviderOwnedNumber(t *testing.T) {
	providerID := uuid.New()
	repo := &fakeNumberRepository{getForRelease: sqlc.PhoneNumber{ProviderID: &providerID, Status: "active"}}
	service := NewService(repo)
	if err := service.Release(context.Background(), uuid.New(), uuid.New()); err == nil {
		t.Fatal("release accepted provider-owned number")
	}
	if repo.releaseCalls != 0 {
		t.Fatalf("release calls=%d", repo.releaseCalls)
	}
}

func TestResponseHasNoTypeAndExposesCustomerCarrier(t *testing.T) {
	connectionID := uuid.New()
	got := response(sqlc.PhoneNumber{CarrierConnectionID: &connectionID})
	if got.CarrierConnectionID == nil || *got.CarrierConnectionID != connectionID {
		t.Fatalf("response=%+v", got)
	}
}
