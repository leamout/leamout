package prepaid

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
)

const (
	managedNumberOperationType  = "managed_number_purchase"
	managedNumberReservationTTL = 30 * 24 * time.Hour
)

type catalogReader interface {
	GetPrice(context.Context, uuid.UUID) (catalog.Price, error)
	ListPrices(context.Context, uuid.UUID, bool) ([]catalog.Price, error)
}

type subscriptionReader interface {
	Current(context.Context, uuid.UUID) (subscriptions.Subscription, error)
}

type authorizationWalletStore interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (Wallet, error)
	GetByCurrency(context.Context, uuid.UUID, string) (Wallet, error)
	Reserve(context.Context, uuid.UUID, uuid.UUID, ReserveInput) (Reservation, error)
	GetReservation(context.Context, uuid.UUID, uuid.UUID) (Reservation, error)
	Capture(context.Context, uuid.UUID, uuid.UUID, int64, string) (Reservation, error)
	Release(context.Context, uuid.UUID, uuid.UUID) (Reservation, error)
}

// AuthorizationService owns prepaid authorization for managed-provider obligations that have
// a fixed, server-owned customer price. Provider adapters never choose the
// customer charge and never mutate wallet value directly.
type AuthorizationService struct {
	catalog       catalogReader
	subscriptions subscriptionReader
	wallets       authorizationWalletStore
	now           func() time.Time
}

func NewAuthorizationService(catalogService catalogReader, subscriptionService subscriptionReader, walletService authorizationWalletStore) *AuthorizationService {
	return &AuthorizationService{
		catalog:       catalogService,
		subscriptions: subscriptionService,
		wallets:       walletService,
		now:           time.Now,
	}
}

// QuoteManagedNumberPurchase resolves the one-time managed-number price from
// the organization's current plan. The price currency follows the acquired
// recurring subscription currency so customer money is never mixed across
// currencies. The first managed-number product policy intentionally has one
// flat one-time price per plan/currency; upstream DID cost remains separate.
func (s *AuthorizationService) QuoteManagedNumberPurchase(ctx context.Context, organizationID uuid.UUID) (uuid.UUID, int64, string, error) {
	if organizationID == uuid.Nil {
		return uuid.Nil, 0, "", ErrManagedNumberAuthorizationInvalid
	}

	subscription, err := s.subscriptions.Current(ctx, organizationID)
	if err != nil {
		return uuid.Nil, 0, "", err
	}
	if subscription.Status != subscriptions.StatusActive {
		return uuid.Nil, 0, "", ErrManagedNumberSubscriptionInactive
	}

	acquiredPrice, err := s.catalog.GetPrice(ctx, subscription.PriceID)
	if err != nil {
		return uuid.Nil, 0, "", fmt.Errorf("resolve acquired subscription price: %w", err)
	}
	currency := strings.ToUpper(strings.TrimSpace(acquiredPrice.Currency))
	if currency == "" {
		return uuid.Nil, 0, "", ErrManagedNumberPriceUnavailable
	}

	prices, err := s.catalog.ListPrices(ctx, subscription.PlanID, true)
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

// ReserveManagedNumberPurchase revalidates a previously quoted server-owned
// price, then reserves prepaid funds before Telecom may create the upstream
// provider obligation.
func (s *AuthorizationService) ReserveManagedNumberPurchase(
	ctx context.Context,
	organizationID, operationID, priceID uuid.UUID,
	amountMinor int64,
	currency string,
) (uuid.UUID, error) {
	if operationID == uuid.Nil || priceID == uuid.Nil || amountMinor <= 0 {
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

	wallet, err := s.wallets.GetByCurrency(ctx, organizationID, currency)
	if err != nil {
		return uuid.Nil, err
	}
	reservation, err := s.wallets.Reserve(ctx, organizationID, wallet.ID, ReserveInput{
		AmountMinor:   amountMinor,
		OperationType: managedNumberOperationType,
		OperationID:   operationID.String(),
		ExpiresAt:     s.now().Add(managedNumberReservationTTL),
	})
	if err != nil {
		return uuid.Nil, err
	}
	return reservation.ID, nil
}

// VerifyManagedNumberPurchase proves that the durable reservation still
// authorizes the exact managed-number purchase terms before a provider call.
// Captured reservations are accepted so reconciliation can finish safely after
// a capture succeeded but Telecom persistence had to retry.
func (s *AuthorizationService) VerifyManagedNumberPurchase(
	ctx context.Context,
	organizationID, operationID, reservationID, priceID uuid.UUID,
	amountMinor int64,
	currency string,
) error {
	_, err := s.authorization(ctx, organizationID, operationID, reservationID, priceID, amountMinor, currency)
	return err
}

// CaptureManagedNumberPurchase captures exactly the fixed amount that was
// reserved. The operation is idempotent when a previous capture already
// completed with the same amount.
func (s *AuthorizationService) CaptureManagedNumberPurchase(
	ctx context.Context,
	organizationID, operationID, reservationID, priceID uuid.UUID,
	amountMinor int64,
	currency string,
) error {
	reservation, err := s.authorization(ctx, organizationID, operationID, reservationID, priceID, amountMinor, currency)
	if err != nil {
		return err
	}
	if reservation.Status == ReservationCaptured {
		return nil
	}
	_, err = s.wallets.Capture(
		ctx,
		organizationID,
		reservationID,
		amountMinor,
		"managed-number-purchase:"+operationID.String(),
	)
	if errors.Is(err, ErrInvalidReservationState) {
		// Another worker may have captured the reservation after the read in
		// authorization. Re-read the durable state so concurrent reconciliation
		// remains idempotent without treating a different terminal state as
		// success.
		reservation, verifyErr := s.authorization(
			ctx, organizationID, operationID, reservationID, priceID, amountMinor, currency,
		)
		if verifyErr == nil && reservation.Status == ReservationCaptured {
			return nil
		}
	}
	return err
}

// ReleaseManagedNumberPurchase is idempotent for released/expired reservations.
// Captured value cannot be released; any correction must be a compensating
// ledger entry rather than rewriting monetary history.
func (s *AuthorizationService) ReleaseManagedNumberPurchase(
	ctx context.Context,
	organizationID, operationID, reservationID uuid.UUID,
) error {
	if organizationID == uuid.Nil || operationID == uuid.Nil || reservationID == uuid.Nil {
		return ErrManagedNumberAuthorizationInvalid
	}
	reservation, err := s.wallets.GetReservation(ctx, organizationID, reservationID)
	if err != nil {
		return err
	}
	if reservation.OperationType != managedNumberOperationType || reservation.OperationID != operationID.String() {
		return ErrManagedNumberAuthorizationInvalid
	}
	switch reservation.Status {
	case ReservationActive:
		_, err = s.wallets.Release(ctx, organizationID, reservationID)
		if errors.Is(err, ErrInvalidReservationState) {
			// A concurrent worker may have released or expired the reservation
			// after it was read above. Only those non-monetary terminal states are
			// valid idempotent outcomes; capture must continue to fail closed.
			reservation, readErr := s.wallets.GetReservation(ctx, organizationID, reservationID)
			if readErr == nil &&
				reservation.OperationType == managedNumberOperationType &&
				reservation.OperationID == operationID.String() &&
				(reservation.Status == ReservationReleased || reservation.Status == ReservationExpired) {
				return nil
			}
		}
		return err
	case ReservationReleased, ReservationExpired:
		return nil
	case ReservationCaptured:
		return ErrManagedNumberAuthorizationInvalid
	default:
		return ErrManagedNumberAuthorizationInvalid
	}
}

func (s *AuthorizationService) authorization(
	ctx context.Context,
	organizationID, operationID, reservationID, priceID uuid.UUID,
	amountMinor int64,
	currency string,
) (Reservation, error) {
	if organizationID == uuid.Nil || operationID == uuid.Nil || reservationID == uuid.Nil || priceID == uuid.Nil || amountMinor <= 0 {
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}

	price, err := s.catalog.GetPrice(ctx, priceID)
	if err != nil {
		return Reservation{}, fmt.Errorf("resolve managed number purchase price: %w", err)
	}
	if price.PricingType != catalog.PricingTypeOneTime || price.AmountMinor == nil ||
		*price.AmountMinor != amountMinor || strings.ToUpper(strings.TrimSpace(price.Currency)) != currency {
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}

	reservation, err := s.wallets.GetReservation(ctx, organizationID, reservationID)
	if err != nil {
		return Reservation{}, err
	}
	if reservation.OperationType != managedNumberOperationType || reservation.OperationID != operationID.String() || reservation.AmountMinor != amountMinor {
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}

	wallet, err := s.wallets.Get(ctx, organizationID, reservation.WalletID)
	if err != nil {
		return Reservation{}, err
	}
	if strings.ToUpper(strings.TrimSpace(wallet.Currency)) != currency {
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}

	switch reservation.Status {
	case ReservationActive:
		if !reservation.ExpiresAt.After(s.now()) {
			return Reservation{}, ErrManagedNumberAuthorizationInvalid
		}
	case ReservationCaptured:
		if reservation.CapturedAmountMinor == nil || *reservation.CapturedAmountMinor != amountMinor {
			return Reservation{}, ErrManagedNumberAuthorizationInvalid
		}
	default:
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	return reservation, nil
}
