package prepaid

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type walletServiceStub struct {
	wallet        wallets.Wallet
	reservation   wallets.Reservation
	currency      string
	reserveInput  wallets.ReserveInput
	captureAmount int64
	captureKey    string
	releaseCalls  int
	getByCurrency int
	reserveCalls  int
	captureCalls  int
}

func (s *walletServiceStub) GetByCurrency(_ context.Context, _ uuid.UUID, currency string) (wallets.Wallet, error) {
	s.getByCurrency++
	s.currency = currency
	return s.wallet, nil
}

func (s *walletServiceStub) Reserve(_ context.Context, _, _ uuid.UUID, input wallets.ReserveInput) (wallets.Reservation, error) {
	s.reserveCalls++
	s.reserveInput = input
	return s.reservation, nil
}

func (s *walletServiceStub) Capture(_ context.Context, _, _ uuid.UUID, amount int64, key string) (wallets.Reservation, error) {
	s.captureCalls++
	s.captureAmount = amount
	s.captureKey = key
	return s.reservation, nil
}

func (s *walletServiceStub) Release(context.Context, uuid.UUID, uuid.UUID) (wallets.Reservation, error) {
	s.releaseCalls++
	return s.reservation, nil
}

func TestAuthorizeReservesOrganizationCurrencyWallet(t *testing.T) {
	organizationID := uuid.New()
	walletID := uuid.New()
	authorizationID := uuid.New()
	expiresAt := time.Date(2026, time.September, 12, 12, 0, 0, 0, time.UTC)
	store := &walletServiceStub{
		wallet: wallets.Wallet{ID: walletID, OrganizationID: organizationID, Currency: "USD"},
		reservation: wallets.Reservation{
			ID: authorizationID, OrganizationID: organizationID, WalletID: walletID,
			AmountMinor: 2500, OperationType: "number_provision", OperationID: "operation-1",
			Status: wallets.ReservationActive, ExpiresAt: expiresAt,
		},
	}
	service := NewService(store)
	service.now = func() time.Time { return expiresAt.Add(-time.Minute) }

	got, err := service.Authorize(t.Context(), AuthorizeInput{
		OrganizationID: organizationID,
		Currency:       " usd ",
		AmountMinor:    2500,
		OperationType:  "number_provision",
		OperationID:    " operation-1 ",
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if store.currency != "USD" || store.reserveInput.OperationID != "operation-1" ||
		store.reserveInput.AmountMinor != 2500 || store.reserveInput.ExpiresAt != expiresAt {
		t.Fatalf("wallet authorization = currency %q, reservation %+v", store.currency, store.reserveInput)
	}
	if got.ID != authorizationID || got.Status != StatusAuthorized {
		t.Fatalf("authorization = %+v", got)
	}
}

func TestAuthorizeRejectsInvalidTermsBeforeWalletAccess(t *testing.T) {
	store := &walletServiceStub{}
	service := NewService(store)
	_, err := service.Authorize(t.Context(), AuthorizeInput{
		OrganizationID: uuid.New(), Currency: "USD", AmountMinor: 0,
		OperationType: "number_provision", OperationID: "operation-1", ExpiresAt: time.Now().Add(time.Hour),
	})
	if !errors.Is(err, ErrInvalidManagedOperation) {
		t.Fatalf("Authorize() error = %v, want %v", err, ErrInvalidManagedOperation)
	}
	if store.getByCurrency != 0 || store.reserveCalls != 0 {
		t.Fatal("invalid authorization reached wallet service")
	}
}

func TestCaptureUsesAuthorizationScopedLedgerIdentity(t *testing.T) {
	authorizationID := uuid.New()
	store := &walletServiceStub{reservation: wallets.Reservation{
		ID: authorizationID, Status: wallets.ReservationCaptured,
	}}
	service := NewService(store)

	if _, err := service.Capture(t.Context(), CaptureInput{
		OrganizationID: uuid.New(), AuthorizationID: authorizationID, AmountMinor: 1900,
	}); err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	if store.captureAmount != 1900 || store.captureKey != "managed_operation:"+authorizationID.String() {
		t.Fatalf("capture amount = %d, key = %q", store.captureAmount, store.captureKey)
	}
}

func TestReleaseRequiresOrganizationAndAuthorization(t *testing.T) {
	store := &walletServiceStub{}
	service := NewService(store)
	if _, err := service.Release(t.Context(), uuid.Nil, uuid.New()); !errors.Is(err, ErrInvalidManagedOperation) {
		t.Fatalf("Release() error = %v, want %v", err, ErrInvalidManagedOperation)
	}
	if store.releaseCalls != 0 {
		t.Fatal("invalid release reached wallet service")
	}
}
