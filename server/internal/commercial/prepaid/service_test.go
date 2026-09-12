package prepaid

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type catalogStub struct {
	pricesByID map[uuid.UUID]catalog.Price
	prices     []catalog.Price
}

func (s *catalogStub) GetPrice(_ context.Context, id uuid.UUID) (catalog.Price, error) {
	price, ok := s.pricesByID[id]
	if !ok {
		return catalog.Price{}, catalog.ErrPriceNotFound
	}
	return price, nil
}

func (s *catalogStub) ListPrices(context.Context, uuid.UUID, bool) ([]catalog.Price, error) {
	return append([]catalog.Price(nil), s.prices...), nil
}

type subscriptionStub struct {
	subscription subscriptions.Subscription
	err          error
}

func (s *subscriptionStub) Current(context.Context, uuid.UUID) (subscriptions.Subscription, error) {
	return s.subscription, s.err
}

type walletStub struct {
	wallet      wallets.Wallet
	reservation wallets.Reservation
	reserveIn   wallets.ReserveInput
	captures    int
	releases    int
	captureRace bool
	releaseRace bool
}

func (s *walletStub) Get(_ context.Context, _, _ uuid.UUID) (wallets.Wallet, error) {
	return s.wallet, nil
}

func (s *walletStub) GetByCurrency(_ context.Context, _ uuid.UUID, currency string) (wallets.Wallet, error) {
	if s.wallet.Currency != currency {
		return wallets.Wallet{}, wallets.ErrWalletNotFound
	}
	return s.wallet, nil
}

func (s *walletStub) Reserve(_ context.Context, organizationID, walletID uuid.UUID, input wallets.ReserveInput) (wallets.Reservation, error) {
	s.reserveIn = input
	s.reservation.OrganizationID = organizationID
	s.reservation.WalletID = walletID
	s.reservation.AmountMinor = input.AmountMinor
	s.reservation.OperationType = input.OperationType
	s.reservation.OperationID = input.OperationID
	s.reservation.ExpiresAt = input.ExpiresAt
	s.reservation.Status = wallets.ReservationActive
	return s.reservation, nil
}

func (s *walletStub) GetReservation(context.Context, uuid.UUID, uuid.UUID) (wallets.Reservation, error) {
	return s.reservation, nil
}

func (s *walletStub) Capture(_ context.Context, _ uuid.UUID, _ uuid.UUID, amount int64, _ string) (wallets.Reservation, error) {
	s.captures++
	s.reservation.Status = wallets.ReservationCaptured
	s.reservation.CapturedAmountMinor = &amount
	if s.captureRace {
		s.captureRace = false
		return wallets.Reservation{}, wallets.ErrInvalidReservationState
	}
	return s.reservation, nil
}

func (s *walletStub) Release(context.Context, uuid.UUID, uuid.UUID) (wallets.Reservation, error) {
	s.releases++
	s.reservation.Status = wallets.ReservationReleased
	if s.releaseRace {
		s.releaseRace = false
		return wallets.Reservation{}, wallets.ErrInvalidReservationState
	}
	return s.reservation, nil
}

func TestManagedNumberPurchaseQuoteReserveAndCapture(t *testing.T) {
	now := time.Date(2026, 9, 12, 8, 0, 0, 0, time.UTC)
	organizationID := uuid.New()
	planID := uuid.New()
	recurringPriceID := uuid.New()
	purchasePriceID := uuid.New()
	walletID := uuid.New()
	reservationID := uuid.New()
	operationID := uuid.New()
	amount := int64(2500)
	recurringAmount := int64(10000)

	catalogService := &catalogStub{
		pricesByID: map[uuid.UUID]catalog.Price{
			recurringPriceID: {
				ID: recurringPriceID, PlanID: planID, PricingType: catalog.PricingTypeRecurring,
				Currency: "USD", AmountMinor: &recurringAmount,
			},
			purchasePriceID: {
				ID: purchasePriceID, PlanID: planID, PricingType: catalog.PricingTypeOneTime,
				Currency: "USD", AmountMinor: &amount,
			},
		},
		prices: []catalog.Price{{
			ID: purchasePriceID, PlanID: planID, PricingType: catalog.PricingTypeOneTime,
			Currency: "USD", AmountMinor: &amount, Active: true, EffectiveFrom: now.Add(-time.Hour),
		}},
	}
	subscriptionService := &subscriptionStub{subscription: subscriptions.Subscription{
		ID: uuid.New(), OrganizationID: organizationID, PlanID: planID, PriceID: recurringPriceID,
		Status: subscriptions.StatusActive,
	}}
	walletService := &walletStub{
		wallet:      wallets.Wallet{ID: walletID, OrganizationID: organizationID, Currency: "USD", Status: wallets.StatusActive},
		reservation: wallets.Reservation{ID: reservationID},
	}
	service := NewService(catalogService, subscriptionService, walletService)
	service.now = func() time.Time { return now }

	priceID, quotedAmount, currency, err := service.QuoteManagedNumberPurchase(context.Background(), organizationID)
	if err != nil {
		t.Fatalf("QuoteManagedNumberPurchase() error = %v", err)
	}
	if priceID != purchasePriceID || quotedAmount != amount || currency != "USD" {
		t.Fatalf("quote = (%s, %d, %s), want (%s, %d, USD)", priceID, quotedAmount, currency, purchasePriceID, amount)
	}

	gotReservationID, err := service.ReserveManagedNumberPurchase(
		context.Background(), organizationID, operationID, priceID, quotedAmount, currency,
	)
	if err != nil {
		t.Fatalf("ReserveManagedNumberPurchase() error = %v", err)
	}
	if gotReservationID != reservationID {
		t.Fatalf("reservation id = %s, want %s", gotReservationID, reservationID)
	}
	if walletService.reserveIn.OperationType != managedNumberOperationType || walletService.reserveIn.OperationID != operationID.String() {
		t.Fatalf("reservation operation = (%q, %q)", walletService.reserveIn.OperationType, walletService.reserveIn.OperationID)
	}
	if walletService.reserveIn.AmountMinor != amount || !walletService.reserveIn.ExpiresAt.Equal(now.Add(managedNumberReservationTTL)) {
		t.Fatalf("reservation terms = (%d, %s)", walletService.reserveIn.AmountMinor, walletService.reserveIn.ExpiresAt)
	}

	if err := service.CaptureManagedNumberPurchase(
		context.Background(), organizationID, operationID, reservationID, priceID, amount, currency,
	); err != nil {
		t.Fatalf("CaptureManagedNumberPurchase() error = %v", err)
	}
	if walletService.captures != 1 {
		t.Fatalf("capture count = %d, want 1", walletService.captures)
	}

	if err := service.CaptureManagedNumberPurchase(
		context.Background(), organizationID, operationID, reservationID, priceID, amount, currency,
	); err != nil {
		t.Fatalf("idempotent CaptureManagedNumberPurchase() error = %v", err)
	}
	if walletService.captures != 1 {
		t.Fatalf("capture count after replay = %d, want 1", walletService.captures)
	}
}

func TestManagedNumberPurchaseQuoteRequiresActiveSubscription(t *testing.T) {
	service := NewService(
		&catalogStub{},
		&subscriptionStub{subscription: subscriptions.Subscription{Status: subscriptions.StatusPastDue}},
		&walletStub{},
	)
	_, _, _, err := service.QuoteManagedNumberPurchase(context.Background(), uuid.New())
	if !errors.Is(err, ErrManagedNumberSubscriptionInactive) {
		t.Fatalf("QuoteManagedNumberPurchase() error = %v, want %v", err, ErrManagedNumberSubscriptionInactive)
	}
}

func TestManagedNumberPurchaseRejectsStaleQuote(t *testing.T) {
	organizationID := uuid.New()
	planID := uuid.New()
	recurringPriceID := uuid.New()
	purchasePriceID := uuid.New()
	recurringAmount := int64(10000)
	purchaseAmount := int64(2500)
	catalogService := &catalogStub{
		pricesByID: map[uuid.UUID]catalog.Price{
			recurringPriceID: {ID: recurringPriceID, PlanID: planID, PricingType: catalog.PricingTypeRecurring, Currency: "USD", AmountMinor: &recurringAmount},
		},
		prices: []catalog.Price{{ID: purchasePriceID, PlanID: planID, PricingType: catalog.PricingTypeOneTime, Currency: "USD", AmountMinor: &purchaseAmount}},
	}
	service := NewService(
		catalogService,
		&subscriptionStub{subscription: subscriptions.Subscription{PlanID: planID, PriceID: recurringPriceID, Status: subscriptions.StatusActive}},
		&walletStub{},
	)

	_, err := service.ReserveManagedNumberPurchase(
		context.Background(), organizationID, uuid.New(), purchasePriceID, purchaseAmount+1, "USD",
	)
	if !errors.Is(err, ErrManagedNumberQuoteExpired) {
		t.Fatalf("ReserveManagedNumberPurchase() error = %v, want %v", err, ErrManagedNumberQuoteExpired)
	}
}

func TestReleaseManagedNumberPurchaseIsIdempotent(t *testing.T) {
	organizationID := uuid.New()
	operationID := uuid.New()
	reservationID := uuid.New()
	walletService := &walletStub{
		reservation: wallets.Reservation{
			ID: reservationID, OrganizationID: organizationID,
			OperationType: managedNumberOperationType, OperationID: operationID.String(),
			Status: wallets.ReservationReleased,
		},
	}
	service := NewService(&catalogStub{}, &subscriptionStub{}, walletService)

	if err := service.ReleaseManagedNumberPurchase(context.Background(), organizationID, operationID, reservationID); err != nil {
		t.Fatalf("ReleaseManagedNumberPurchase() error = %v", err)
	}
	if walletService.releases != 0 {
		t.Fatalf("release count = %d, want 0", walletService.releases)
	}
}

func TestCaptureManagedNumberPurchaseIsIdempotentAcrossWorkers(t *testing.T) {
	service, walletService, organizationID, operationID, reservationID, priceID, amount := managedNumberAuthorizationFixture(t)
	walletService.captureRace = true

	if err := service.CaptureManagedNumberPurchase(
		context.Background(), organizationID, operationID, reservationID, priceID, amount, "USD",
	); err != nil {
		t.Fatalf("CaptureManagedNumberPurchase() concurrent replay error = %v", err)
	}
	if walletService.captures != 1 {
		t.Fatalf("capture count = %d, want 1", walletService.captures)
	}
}

func TestReleaseManagedNumberPurchaseIsIdempotentAcrossWorkers(t *testing.T) {
	organizationID := uuid.New()
	operationID := uuid.New()
	reservationID := uuid.New()
	walletService := &walletStub{
		reservation: wallets.Reservation{
			ID: reservationID, OrganizationID: organizationID,
			OperationType: managedNumberOperationType, OperationID: operationID.String(),
			Status: wallets.ReservationActive,
		},
		releaseRace: true,
	}
	service := NewService(&catalogStub{}, &subscriptionStub{}, walletService)

	if err := service.ReleaseManagedNumberPurchase(context.Background(), organizationID, operationID, reservationID); err != nil {
		t.Fatalf("ReleaseManagedNumberPurchase() concurrent replay error = %v", err)
	}
	if walletService.releases != 1 {
		t.Fatalf("release count = %d, want 1", walletService.releases)
	}
}

func managedNumberAuthorizationFixture(t *testing.T) (*Service, *walletStub, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, int64) {
	t.Helper()
	organizationID := uuid.New()
	operationID := uuid.New()
	reservationID := uuid.New()
	priceID := uuid.New()
	walletID := uuid.New()
	amount := int64(2500)
	walletService := &walletStub{
		wallet: wallets.Wallet{ID: walletID, OrganizationID: organizationID, Currency: "USD", Status: wallets.StatusActive},
		reservation: wallets.Reservation{
			ID: reservationID, WalletID: walletID, OrganizationID: organizationID,
			AmountMinor: amount, OperationType: managedNumberOperationType, OperationID: operationID.String(),
			Status: wallets.ReservationActive, ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	service := NewService(
		&catalogStub{pricesByID: map[uuid.UUID]catalog.Price{
			priceID: {ID: priceID, PricingType: catalog.PricingTypeOneTime, Currency: "USD", AmountMinor: &amount},
		}},
		&subscriptionStub{},
		walletService,
	)
	return service, walletService, organizationID, operationID, reservationID, priceID, amount
}
