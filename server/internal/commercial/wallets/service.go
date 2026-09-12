package wallets

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/catalog"
)

const (
	managedNumberOperationType  = "managed_number_purchase"
	managedNumberReservationTTL = 30 * 24 * time.Hour
	managedNumberPlanCode       = "managed-numbers"
)

type managedOperations struct {
	getPrice       func(context.Context, uuid.UUID) (catalog.Price, error)
	getPlanByCode  func(context.Context, string) (catalog.Plan, error)
	listPrices     func(context.Context, uuid.UUID, bool) ([]catalog.Price, error)
	getWallet      func(context.Context, uuid.UUID, uuid.UUID) (Wallet, error)
	getByCurrency  func(context.Context, uuid.UUID, string) (Wallet, error)
	reserve        func(context.Context, uuid.UUID, uuid.UUID, ReserveInput) (Reservation, error)
	getReservation func(context.Context, uuid.UUID, uuid.UUID) (Reservation, error)
	capture        func(context.Context, uuid.UUID, uuid.UUID, int64, string) (Reservation, error)
	release        func(context.Context, uuid.UUID, uuid.UUID) (Reservation, error)
}

type Service struct {
	repo    *Repository
	managed managedOperations
	now     func() time.Time
}

func NewService(repo *Repository, catalogService *catalog.Service) *Service {
	service := &Service{repo: repo, now: time.Now}
	if repo != nil && catalogService != nil {
		service.managed = managedOperations{
			getPrice: catalogService.GetPrice, getPlanByCode: catalogService.GetPlanByCode,
			listPrices: catalogService.ListPrices, getWallet: repo.Get, getByCurrency: repo.GetByCurrency,
			reserve: repo.Reserve, getReservation: repo.GetReservation, capture: repo.Capture, release: repo.Release,
		}
	}
	return service
}

func (s *Service) List(ctx context.Context, organizationID uuid.UUID) ([]Wallet, error) { return s.repo.List(ctx, organizationID) }
func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, currency string) (Wallet, error) { return s.repo.Create(ctx, organizationID, currency) }
func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (Wallet, error) { return s.repo.Get(ctx, organizationID, id) }
func (s *Service) GetByCurrency(ctx context.Context, organizationID uuid.UUID, currency string) (Wallet, error) { return s.repo.GetByCurrency(ctx, organizationID, currency) }
func (s *Service) Balance(ctx context.Context, organizationID, id uuid.UUID) (Balance, error) { return s.repo.Balance(ctx, organizationID, id) }
func (s *Service) Post(ctx context.Context, organizationID, id uuid.UUID, input PostEntryInput) (LedgerEntry, error) {
	if input.Type == EntryCapture { return LedgerEntry{}, ErrReservationRequired }
	return s.repo.Post(ctx, organizationID, id, input)
}
func (s *Service) ListEntries(ctx context.Context, organizationID, id uuid.UUID) ([]LedgerEntry, error) { return s.repo.ListEntries(ctx, organizationID, id) }
func (s *Service) Reserve(ctx context.Context, organizationID, id uuid.UUID, input ReserveInput) (Reservation, error) {
	if err := validateReservationInput(input, s.now()); err != nil { return Reservation{}, err }
	return s.repo.Reserve(ctx, organizationID, id, input)
}
func (s *Service) GetReservation(ctx context.Context, organizationID, id uuid.UUID) (Reservation, error) { return s.repo.GetReservation(ctx, organizationID, id) }
func (s *Service) Capture(ctx context.Context, organizationID, id uuid.UUID, amount int64, key string) (Reservation, error) {
	if err := validateCaptureInput(amount); err != nil { return Reservation{}, err }
	return s.repo.Capture(ctx, organizationID, id, amount, key)
}
func (s *Service) Increase(ctx context.Context, organizationID, id uuid.UUID, input IncreaseReservationInput) (Reservation, error) {
	if input.AmountMinor <= 0 || !input.ExpiresAt.After(s.now()) { return Reservation{}, ErrInvalidMoney }
	return s.repo.Increase(ctx, organizationID, id, input)
}
func (s *Service) Release(ctx context.Context, organizationID, id uuid.UUID) (Reservation, error) { return s.repo.Release(ctx, organizationID, id) }
func (s *Service) Expire(ctx context.Context) ([]Reservation, error) { return s.repo.Expire(ctx) }

// QuoteManagedNumberPurchase resolves the single active PAYG managed-number
// price. No customer subscription or entitlement state participates.
func (s *Service) QuoteManagedNumberPurchase(ctx context.Context, organizationID uuid.UUID) (uuid.UUID, int64, string, error) {
	if organizationID == uuid.Nil { return uuid.Nil, 0, "", ErrManagedNumberAuthorizationInvalid }
	if s == nil || s.managed.getPlanByCode == nil || s.managed.listPrices == nil {
		return uuid.Nil, 0, "", ErrManagedNumberPriceUnavailable
	}
	plan, err := s.managed.getPlanByCode(ctx, managedNumberPlanCode)
	if err != nil { return uuid.Nil, 0, "", fmt.Errorf("resolve managed number offer: %w", err) }
	if !plan.Active { return uuid.Nil, 0, "", ErrManagedNumberPriceUnavailable }
	prices, err := s.managed.listPrices(ctx, plan.ID, true)
	if err != nil { return uuid.Nil, 0, "", fmt.Errorf("list managed number prices: %w", err) }
	var matched *catalog.Price
	for i := range prices {
		price := &prices[i]
		currency := strings.ToUpper(strings.TrimSpace(price.Currency))
		if price.PricingType != catalog.PricingTypeOneTime || currency == "" || price.AmountMinor == nil || *price.AmountMinor <= 0 { continue }
		if matched != nil { return uuid.Nil, 0, "", ErrManagedNumberPriceUnavailable }
		matched = price
	}
	if matched == nil { return uuid.Nil, 0, "", ErrManagedNumberPriceUnavailable }
	return matched.ID, *matched.AmountMinor, strings.ToUpper(strings.TrimSpace(matched.Currency)), nil
}

func (s *Service) ReserveManagedNumberPurchase(ctx context.Context, organizationID, operationID, priceID uuid.UUID, amountMinor int64, currency string) (uuid.UUID, error) {
	if s == nil || s.managed.getByCurrency == nil || s.managed.reserve == nil || operationID == uuid.Nil || priceID == uuid.Nil || amountMinor <= 0 {
		return uuid.Nil, ErrManagedNumberAuthorizationInvalid
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	quotedPriceID, quotedAmount, quotedCurrency, err := s.QuoteManagedNumberPurchase(ctx, organizationID)
	if err != nil { return uuid.Nil, err }
	if quotedPriceID != priceID || quotedAmount != amountMinor || quotedCurrency != currency { return uuid.Nil, ErrManagedNumberQuoteExpired }
	wallet, err := s.managed.getByCurrency(ctx, organizationID, currency)
	if err != nil { return uuid.Nil, err }
	reservation, err := s.managed.reserve(ctx, organizationID, wallet.ID, ReserveInput{
		AmountMinor: amountMinor, OperationType: managedNumberOperationType, OperationID: operationID.String(), ExpiresAt: s.now().Add(managedNumberReservationTTL),
	})
	if err != nil { return uuid.Nil, err }
	return reservation.ID, nil
}

func (s *Service) VerifyManagedNumberPurchase(ctx context.Context, organizationID, operationID, reservationID, priceID uuid.UUID, amountMinor int64, currency string) error {
	_, err := s.authorization(ctx, organizationID, operationID, reservationID, priceID, amountMinor, currency)
	return err
}

func (s *Service) CaptureManagedNumberPurchase(ctx context.Context, organizationID, operationID, reservationID, priceID uuid.UUID, amountMinor int64, currency string) error {
	reservation, err := s.authorization(ctx, organizationID, operationID, reservationID, priceID, amountMinor, currency)
	if err != nil { return err }
	if reservation.Status == ReservationCaptured { return nil }
	_, err = s.managed.capture(ctx, organizationID, reservationID, amountMinor, "managed-number-purchase:"+operationID.String())
	if errors.Is(err, ErrInvalidReservationState) {
		reservation, verifyErr := s.authorization(ctx, organizationID, operationID, reservationID, priceID, amountMinor, currency)
		if verifyErr != nil { return verifyErr }
		if reservation.Status == ReservationCaptured { return nil }
	}
	return err
}

func (s *Service) ReleaseManagedNumberPurchase(ctx context.Context, organizationID, operationID, reservationID uuid.UUID) error {
	if s == nil || s.managed.getReservation == nil || s.managed.release == nil || organizationID == uuid.Nil || operationID == uuid.Nil || reservationID == uuid.Nil {
		return ErrManagedNumberAuthorizationInvalid
	}
	reservation, err := s.managed.getReservation(ctx, organizationID, reservationID)
	if err != nil { return err }
	if reservation.OperationType != managedNumberOperationType || reservation.OperationID != operationID.String() { return ErrManagedNumberAuthorizationInvalid }
	switch reservation.Status {
	case ReservationActive:
		_, err = s.managed.release(ctx, organizationID, reservationID)
		if errors.Is(err, ErrInvalidReservationState) {
			reservation, readErr := s.managed.getReservation(ctx, organizationID, reservationID)
			if readErr != nil { return readErr }
			if reservation.OperationType == managedNumberOperationType && reservation.OperationID == operationID.String() &&
				(reservation.Status == ReservationReleased || reservation.Status == ReservationExpired) { return nil }
		}
		return err
	case ReservationReleased, ReservationExpired:
		return nil
	default:
		return ErrManagedNumberAuthorizationInvalid
	}
}

func (s *Service) authorization(ctx context.Context, organizationID, operationID, reservationID, priceID uuid.UUID, amountMinor int64, currency string) (Reservation, error) {
	if s == nil || s.managed.getPrice == nil || s.managed.getReservation == nil || s.managed.getWallet == nil ||
		organizationID == uuid.Nil || operationID == uuid.Nil || reservationID == uuid.Nil || priceID == uuid.Nil || amountMinor <= 0 {
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" { return Reservation{}, ErrManagedNumberAuthorizationInvalid }
	price, err := s.managed.getPrice(ctx, priceID)
	if err != nil { return Reservation{}, fmt.Errorf("resolve managed number purchase price: %w", err) }
	if price.PricingType != catalog.PricingTypeOneTime || price.AmountMinor == nil || *price.AmountMinor != amountMinor || strings.ToUpper(strings.TrimSpace(price.Currency)) != currency {
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	reservation, err := s.managed.getReservation(ctx, organizationID, reservationID)
	if err != nil { return Reservation{}, err }
	if reservation.OperationType != managedNumberOperationType || reservation.OperationID != operationID.String() || reservation.AmountMinor != amountMinor {
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	wallet, err := s.managed.getWallet(ctx, organizationID, reservation.WalletID)
	if err != nil { return Reservation{}, err }
	if strings.ToUpper(strings.TrimSpace(wallet.Currency)) != currency { return Reservation{}, ErrManagedNumberAuthorizationInvalid }
	switch reservation.Status {
	case ReservationActive:
		if !reservation.ExpiresAt.After(s.now()) { return Reservation{}, ErrManagedNumberAuthorizationInvalid }
	case ReservationCaptured:
		if reservation.CapturedAmountMinor == nil || *reservation.CapturedAmountMinor != amountMinor { return Reservation{}, ErrManagedNumberAuthorizationInvalid }
	default:
		return Reservation{}, ErrManagedNumberAuthorizationInvalid
	}
	return reservation, nil
}
