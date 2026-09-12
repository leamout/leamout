package authorization

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	commercialaccess "github.com/leamout/leamout/internal/commercial/access"
	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

const (
	managedNumberOperationType  = "managed_number_purchase"
	managedNumberReservationTTL = 30 * 24 * time.Hour
)

type accessOperations struct {
	resolve func(context.Context, uuid.UUID) (commercialaccess.OrganizationAccess, error)
}

type catalogOperations struct {
	getPrice   func(context.Context, uuid.UUID) (catalog.Price, error)
	listPrices func(context.Context, uuid.UUID, bool) ([]catalog.Price, error)
}

type subscriptionOperations struct {
	current func(context.Context, uuid.UUID) (subscriptions.Subscription, error)
}

type walletOperations struct {
	get            func(context.Context, uuid.UUID, uuid.UUID) (wallets.Wallet, error)
	getByCurrency  func(context.Context, uuid.UUID, string) (wallets.Wallet, error)
	reserve        func(context.Context, uuid.UUID, uuid.UUID, wallets.ReserveInput) (wallets.Reservation, error)
	getReservation func(context.Context, uuid.UUID, uuid.UUID) (wallets.Reservation, error)
	capture        func(context.Context, uuid.UUID, uuid.UUID, int64, string) (wallets.Reservation, error)
	release        func(context.Context, uuid.UUID, uuid.UUID) (wallets.Reservation, error)
}

// Service owns product-specific prepaid authorization policy. Wallets remains
// the monetary primitive; Telecom depends only on this narrow authorization API.
type Service struct {
	access        accessOperations
	catalog       catalogOperations
	subscriptions subscriptionOperations
	wallets       walletOperations
	now           func() time.Time
}

func NewService(
	accessService *commercialaccess.Service,
	catalogService *catalog.Service,
	subscriptionService *subscriptions.Service,
	walletService *wallets.Service,
) *Service {
	if accessService == nil || catalogService == nil || subscriptionService == nil || walletService == nil {
		panic("authorization: commercial services are required")
	}
	return &Service{
		access: accessOperations{resolve: accessService.Resolve},
		catalog: catalogOperations{
			getPrice:   catalogService.GetPrice,
			listPrices: catalogService.ListPrices,
		},
		subscriptions: subscriptionOperations{current: subscriptionService.Current},
		wallets: walletOperations{
			get:            walletService.Get,
			getByCurrency:  walletService.GetByCurrency,
			reserve:        walletService.Reserve,
			getReservation: walletService.GetReservation,
			capture:        walletService.Capture,
			release:        walletService.Release,
		},
		now: time.Now,
	}
}

func (s *Service) QuoteManagedNumberPurchase(ctx context.Context, organizationID uuid.UUID) (uuid.UUID, int64, string, error) {
	planID, subscription, err := s.managedAccess(ctx, organizationID)
	if err != nil {
		return uuid.Nil, 0, "", err
	}

	acquiredPrice, err := s.catalog.getPrice(ctx, subscription.PriceID)
	if err != nil {
		return uuid.Nil, 0, "", fmt.Errorf("resolve acquired subscription price: %w", err)
	}
	currency := strings.ToUpper(strings.TrimSpace(acquiredPrice.Currency))
	if currency == "" {
		return uuid.Nil, 0, "", ErrManagedNumberPriceUnavailable
	}

	prices, err := s.catalog.listPrices(ctx, planID, true)
	if err != nil {
		return uuid.Nil, 0, "", fmt.Errorf("list managed number prices: %w", err)
	}
	var matched *catalog.Price
	for i := range prices {
		price := &prices[i]
		if price.PricingType != catalog.PricingTypeOneTime ||
			strings.ToUpper(strings.TrimSpace(price.Currency)) != currency ||
			price.AmountMinor == nil || *price.AmountMinor <= 0 {
			continue
		}
		if matched != nil {
			return uuid.Nil, 0, "", ErrManagedNumberPriceUnavailable
		}
		matched = price
	}
	if matched == nil {
		return uuid.Nil, 0, "", ErrManagedNumberPriceUnavailable
	}
	return matched.ID, *matched.AmountMinor, currency, nil
}

func (s *Service) ReserveManagedNumberPurchase(
	ctx context.Context,
	organizationID, operationID, priceID uuid.UUID,
	amountMinor int64,
	currency string,
) (uuid.UUID, error) {
	if organizationID == uuid.Nil || operationID == uuid.Nil || priceID == uuid.Nil || amountMinor <= 0 {
		return uuid.Nil, ErrManagedNumberAuthorizationInvalid
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	quotedPriceID, quotedAmount, quotedCurrency, err := s.QuoteManagedNumberPurchase(ctx, organizationID)
	if err != nil {
		return uuid.Nil, err
	}
	if quotedPriceID != priceID || quotedAmount != amountMinor || quotedCurrency != currency {
		return uuid.Nil, ErrManagedNumberQuoteExpired
	}
	wallet, err := s.wallets.getByCurrency(ctx, organizationID, currency)
	if err != nil {
		return uuid.Nil, err
	}
	reservation, err := s.wallets.reserve(ctx, organizationID, wallet.ID, wallets.ReserveInput{
		AmountMinor: amountMinor, OperationType: managedNumberOperationType,
		OperationID: operationID.String(), ExpiresAt: s.now().Add(managedNumberReservationTTL),
	})
	if err != nil {
		return uuid.Nil, err
	}
	return reservation.ID, nil
}

func (s *Service) VerifyManagedNumberPurchase(
	ctx context.Context,
	organizationID, operationID, reservationID, priceID uuid.UUID,
	amountMinor int64,
	currency string,
) error {
	if _, _, err := s.managedAccess(ctx, organizationID); err != nil {
		return err
	}
	_, err := s.authorization(ctx, organizationID, operationID, reservationID, priceID, amountMinor, currency)
	return err
}

func (s *Service) CaptureManagedNumberPurchase(
	ctx context.Context,
	organizationID, operationID, reservationID, priceID uuid.UUID,
	amountMinor int64,
	currency string,
) error {
	reservation, err := s.authorization(ctx, organizationID, operationID, reservationID, priceID, amountMinor, currency)
	if err != nil {
		return err
	}
	if reservation.Status == wallets.ReservationCaptured {
		return nil
	}
	_, err = s.wallets.capture(ctx, organizationID, reservationID, amountMinor, "managed-number-purchase:"+operationID.String())
	if errors.Is(err, wallets.ErrInvalidReservationState) {
		reservation, verifyErr := s.authorization(ctx, organizationID, operationID, reservationID, priceID, amountMinor, currency)
		if verifyErr != nil {
			return verifyErr
		}
		if reservation.Status == wallets.ReservationCaptured {
			return nil
		}
	}
	return err
}

func (s *Service) ReleaseManagedNumberPurchase(ctx context.Context, organizationID, operationID, reservationID uuid.UUID) error {
	if organizationID == uuid.Nil || operationID == uuid.Nil || reservationID == uuid.Nil {
		return ErrManagedNumberAuthorizationInvalid
	}
	reservation, err := s.wallets.getReservation(ctx, organizationID, reservationID)
	if err != nil {
		return err
	}
	if reservation.OperationType != managedNumberOperationType || reservation.OperationID != operationID.String() {
		return ErrManagedNumberAuthorizationInvalid
	}
	switch reservation.Status {
	case wallets.ReservationActive:
		_, err = s.wallets.release(ctx, organizationID, reservationID)
		if errors.Is(err, wallets.ErrInvalidReservationState) {
			reservation, readErr := s.wallets.getReservation(ctx, organizationID, reservationID)
			if readErr != nil {
				return readErr
			}
			if reservation.OperationType == managedNumberOperationType && reservation.OperationID == operationID.String() &&
				(reservation.Status == wallets.ReservationReleased || reservation.Status == wallets.ReservationExpired) {
				return nil
			}
		}
		return err
	case wallets.ReservationReleased, wallets.ReservationExpired:
		return nil
	case wallets.ReservationCaptured:
		return ErrManagedNumberAuthorizationInvalid
	default:
		return ErrManagedNumberAuthorizationInvalid
	}
}

func (s *Service) managedAccess(ctx context.Context, organizationID uuid.UUID) (uuid.UUID, subscriptions.Subscription, error) {
	if organizationID == uuid.Nil {
		return uuid.Nil, subscriptions.Subscription{}, ErrManagedNumberAuthorizationInvalid
	}
	state, err := s.access.resolve(ctx, organizationID)
	if err != nil {
		return uuid.Nil, subscriptions.Subscription{}, err
	}
	if state.Standing != commercialaccess.StandingActive || !state.Enabled(ManagedVoiceFeature) || state.PlanID == nil {
		return uuid.Nil, subscriptions.Subscription{}, ErrManagedNumberAccessRequired
	}
	current, err := s.subscriptions.current(ctx, organizationID)
	if err != nil {
		return uuid.Nil, subscriptions.Subscription{}, err
	}
	if current.Status != subscriptions.StatusActive || current.PlanID != *state.PlanID {
		return uuid.Nil, subscriptions.Subscription{}, ErrManagedNumberAccessRequired
	}
	return current.PlanID, current, nil
}

func (s *Service) authorization(
	ctx context.Context,
	organizationID, operationID, reservationID, priceID uuid.UUID,
	amountMinor int64,
	currency string,
) (wallets.Reservation, error) {
	if organizationID == uuid.Nil || operationID == uuid.Nil || reservationID == uuid.Nil || priceID == uuid.Nil || amountMinor <= 0 {
		return wallets.Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return wallets.Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	price, err := s.catalog.getPrice(ctx, priceID)
	if err != nil {
		return wallets.Reservation{}, fmt.Errorf("resolve managed number purchase price: %w", err)
	}
	if price.PricingType != catalog.PricingTypeOneTime || price.AmountMinor == nil || *price.AmountMinor != amountMinor ||
		strings.ToUpper(strings.TrimSpace(price.Currency)) != currency {
		return wallets.Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	reservation, err := s.wallets.getReservation(ctx, organizationID, reservationID)
	if err != nil {
		return wallets.Reservation{}, err
	}
	if reservation.OperationType != managedNumberOperationType || reservation.OperationID != operationID.String() || reservation.AmountMinor != amountMinor {
		return wallets.Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	wallet, err := s.wallets.get(ctx, organizationID, reservation.WalletID)
	if err != nil {
		return wallets.Reservation{}, err
	}
	if strings.ToUpper(strings.TrimSpace(wallet.Currency)) != currency {
		return wallets.Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	switch reservation.Status {
	case wallets.ReservationActive:
		if !reservation.ExpiresAt.After(s.now()) {
			return wallets.Reservation{}, ErrManagedNumberAuthorizationInvalid
		}
	case wallets.ReservationCaptured:
		if reservation.CapturedAmountMinor == nil || *reservation.CapturedAmountMinor != amountMinor {
			return wallets.Reservation{}, ErrManagedNumberAuthorizationInvalid
		}
	default:
		return wallets.Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	return reservation, nil
}
