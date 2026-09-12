package authorization

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	commercialaccess "github.com/leamout/leamout/internal/commercial/access"
	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

func TestQuoteManagedNumberPurchaseRequiresManagedFeature(t *testing.T) {
	organizationID := uuid.New()
	planID := uuid.New()
	service := &Service{
		access: accessOperations{resolve: func(context.Context, uuid.UUID) (commercialaccess.OrganizationAccess, error) {
			return commercialaccess.OrganizationAccess{
				OrganizationID: organizationID,
				Standing: commercialaccess.StandingActive,
				PlanID: &planID,
				Features: map[string]bool{},
			}, nil
		}},
		now: time.Now,
	}
	_, _, _, err := service.QuoteManagedNumberPurchase(context.Background(), organizationID)
	if !errors.Is(err, ErrManagedNumberAccessRequired) {
		t.Fatalf("QuoteManagedNumberPurchase() error = %v, want %v", err, ErrManagedNumberAccessRequired)
	}
}

func TestCaptureManagedNumberPurchasePropagatesReReadFailure(t *testing.T) {
	service, organizationID, operationID, reservationID, priceID, amount := authorizationFixture(t)
	reReadErr := errors.New("database unavailable")
	reads := 0
	service.wallets.getReservation = func(context.Context, uuid.UUID, uuid.UUID) (wallets.Reservation, error) {
		reads++
		if reads > 1 {
			return wallets.Reservation{}, reReadErr
		}
		return activeReservation(organizationID, operationID, reservationID, amount), nil
	}
	service.wallets.capture = func(context.Context, uuid.UUID, uuid.UUID, int64, string) (wallets.Reservation, error) {
		return wallets.Reservation{}, wallets.ErrInvalidReservationState
	}

	err := service.CaptureManagedNumberPurchase(context.Background(), organizationID, operationID, reservationID, priceID, amount, "USD")
	if !errors.Is(err, reReadErr) {
		t.Fatalf("CaptureManagedNumberPurchase() error = %v, want re-read error", err)
	}
}

func TestReleaseManagedNumberPurchasePropagatesReReadFailure(t *testing.T) {
	organizationID := uuid.New()
	operationID := uuid.New()
	reservationID := uuid.New()
	reReadErr := errors.New("database unavailable")
	reads := 0
	service := &Service{now: time.Now}
	service.wallets.getReservation = func(context.Context, uuid.UUID, uuid.UUID) (wallets.Reservation, error) {
		reads++
		if reads > 1 {
			return wallets.Reservation{}, reReadErr
		}
		return activeReservation(organizationID, operationID, reservationID, 2500), nil
	}
	service.wallets.release = func(context.Context, uuid.UUID, uuid.UUID) (wallets.Reservation, error) {
		return wallets.Reservation{}, wallets.ErrInvalidReservationState
	}

	err := service.ReleaseManagedNumberPurchase(context.Background(), organizationID, operationID, reservationID)
	if !errors.Is(err, reReadErr) {
		t.Fatalf("ReleaseManagedNumberPurchase() error = %v, want re-read error", err)
	}
}

func authorizationFixture(t *testing.T) (*Service, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, int64) {
	t.Helper()
	organizationID := uuid.New()
	operationID := uuid.New()
	reservationID := uuid.New()
	priceID := uuid.New()
	walletID := uuid.New()
	amount := int64(2500)
	service := &Service{
		catalog: catalogOperations{getPrice: func(context.Context, uuid.UUID) (catalog.Price, error) {
			return catalog.Price{ID: priceID, PricingType: catalog.PricingTypeOneTime, Currency: "USD", AmountMinor: &amount}, nil
		}},
		wallets: walletOperations{
			get: func(context.Context, uuid.UUID, uuid.UUID) (wallets.Wallet, error) {
				return wallets.Wallet{ID: walletID, OrganizationID: organizationID, Currency: "USD", Status: wallets.StatusActive}, nil
			},
		},
		now: time.Now,
	}
	return service, organizationID, operationID, reservationID, priceID, amount
}

func activeReservation(organizationID, operationID, reservationID uuid.UUID, amount int64) wallets.Reservation {
	return wallets.Reservation{
		ID: reservationID,
		WalletID: uuid.New(),
		OrganizationID: organizationID,
		AmountMinor: amount,
		OperationType: managedNumberOperationType,
		OperationID: operationID.String(),
		Status: wallets.ReservationActive,
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

var _ = subscriptions.StatusActive
