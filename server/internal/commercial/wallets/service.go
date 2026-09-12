package wallets

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Service owns prepaid value. Product-specific pricing and authorization policy
// lives above this package.
type Service struct {
	repo *Repository
	now  func() time.Time
}

func NewService(repo *Repository) *Service { return &Service{repo: repo, now: time.Now} }

func (s *Service) List(ctx context.Context, organizationID uuid.UUID) ([]Wallet, error) {
	return s.repo.List(ctx, organizationID)
}
func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, currency string) (Wallet, error) {
	return s.repo.Create(ctx, organizationID, currency)
}
func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (Wallet, error) {
	return s.repo.Get(ctx, organizationID, id)
}
func (s *Service) GetByCurrency(ctx context.Context, organizationID uuid.UUID, currency string) (Wallet, error) {
	return s.repo.GetByCurrency(ctx, organizationID, currency)
}
func (s *Service) Balance(ctx context.Context, organizationID, id uuid.UUID) (Balance, error) {
	return s.repo.Balance(ctx, organizationID, id)
}
func (s *Service) Post(ctx context.Context, organizationID, id uuid.UUID, input PostEntryInput) (LedgerEntry, error) {
	if input.Type == EntryCapture {
		return LedgerEntry{}, ErrReservationRequired
	}
	return s.repo.Post(ctx, organizationID, id, input)
}
func (s *Service) ListEntries(ctx context.Context, organizationID, id uuid.UUID) ([]LedgerEntry, error) {
	return s.repo.ListEntries(ctx, organizationID, id)
}
func (s *Service) Reserve(ctx context.Context, organizationID, id uuid.UUID, input ReserveInput) (Reservation, error) {
	if err := validateReservationInput(input, s.now()); err != nil {
		return Reservation{}, err
	}
	return s.repo.Reserve(ctx, organizationID, id, input)
}
func (s *Service) GetReservation(ctx context.Context, organizationID, id uuid.UUID) (Reservation, error) {
	return s.repo.GetReservation(ctx, organizationID, id)
}
func (s *Service) Capture(ctx context.Context, organizationID, id uuid.UUID, amount int64, key string) (Reservation, error) {
	if err := validateCaptureInput(amount); err != nil {
		return Reservation{}, err
	}
	return s.repo.Capture(ctx, organizationID, id, amount, key)
}
func (s *Service) IncreaseTo(ctx context.Context, organizationID, id uuid.UUID, input IncreaseReservationInput) (Reservation, error) {
	if input.TargetAmountMinor <= 0 || !input.ExpiresAt.After(s.now()) {
		return Reservation{}, ErrInvalidMoney
	}
	return s.repo.IncreaseTo(ctx, organizationID, id, input)
}
func (s *Service) Release(ctx context.Context, organizationID, id uuid.UUID) (Reservation, error) {
	return s.repo.Release(ctx, organizationID, id)
}
func (s *Service) Expire(ctx context.Context) ([]Reservation, error) { return s.repo.Expire(ctx) }
