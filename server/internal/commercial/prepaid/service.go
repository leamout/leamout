package prepaid

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type walletService interface {
	GetByCurrency(context.Context, uuid.UUID, string) (wallets.Wallet, error)
	Reserve(context.Context, uuid.UUID, uuid.UUID, wallets.ReserveInput) (wallets.Reservation, error)
	Capture(context.Context, uuid.UUID, uuid.UUID, int64, string) (wallets.Reservation, error)
	Release(context.Context, uuid.UUID, uuid.UUID) (wallets.Reservation, error)
}

// Service is the prepaid authorization boundary for managed-provider work.
// Telecom supplies already-resolved customer terms and must receive an
// authorization before it creates an upstream obligation.
type Service struct {
	wallets walletService
	now     func() time.Time
}

func NewService(walletsService walletService) *Service {
	return &Service{wallets: walletsService, now: time.Now}
}

func (s *Service) Authorize(ctx context.Context, input AuthorizeInput) (Authorization, error) {
	input, err := normalizeAuthorization(input, s.now().UTC())
	if err != nil {
		return Authorization{}, err
	}
	wallet, err := s.wallets.GetByCurrency(ctx, input.OrganizationID, input.Currency)
	if err != nil {
		return Authorization{}, err
	}
	reservation, err := s.wallets.Reserve(ctx, input.OrganizationID, wallet.ID, wallets.ReserveInput{
		AmountMinor:   input.AmountMinor,
		OperationType: input.OperationType,
		OperationID:   input.OperationID,
		ExpiresAt:     input.ExpiresAt,
	})
	if err != nil {
		return Authorization{}, err
	}
	return authorizationFromReservation(reservation), nil
}

func (s *Service) Capture(ctx context.Context, input CaptureInput) (Authorization, error) {
	if err := validateCapture(input); err != nil {
		return Authorization{}, err
	}
	reservation, err := s.wallets.Capture(
		ctx,
		input.OrganizationID,
		input.AuthorizationID,
		input.AmountMinor,
		"managed_operation:"+input.AuthorizationID.String(),
	)
	if err != nil {
		return Authorization{}, err
	}
	return authorizationFromReservation(reservation), nil
}

func (s *Service) Release(ctx context.Context, organizationID, authorizationID uuid.UUID) (Authorization, error) {
	if err := validateRelease(organizationID, authorizationID); err != nil {
		return Authorization{}, err
	}
	reservation, err := s.wallets.Release(ctx, organizationID, authorizationID)
	if err != nil {
		return Authorization{}, err
	}
	return authorizationFromReservation(reservation), nil
}

func authorizationFromReservation(reservation wallets.Reservation) Authorization {
	return Authorization{
		ID:             reservation.ID,
		OrganizationID: reservation.OrganizationID,
		AmountMinor:    reservation.AmountMinor,
		CapturedMinor:  reservation.CapturedAmountMinor,
		OperationType:  reservation.OperationType,
		OperationID:    reservation.OperationID,
		Status:         authorizationStatus(reservation.Status),
		ExpiresAt:      reservation.ExpiresAt,
	}
}

func authorizationStatus(status wallets.ReservationStatus) Status {
	switch status {
	case wallets.ReservationActive:
		return StatusAuthorized
	case wallets.ReservationCaptured:
		return StatusCaptured
	case wallets.ReservationReleased:
		return StatusReleased
	case wallets.ReservationExpired:
		return StatusExpired
	default:
		return Status(status)
	}
}
