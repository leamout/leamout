package checkout

import (
	"context"
	"errors"

	"github.com/google/uuid"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

// CompletePayment applies the commercial effect of a settled wallet top-up.
func (s *Service) CompletePayment(ctx context.Context, settlement commercialpayments.Settlement) error {
	if settlement.CheckoutID == uuid.Nil || settlement.OrganizationID == uuid.Nil ||
		settlement.PaymentID == uuid.Nil || settlement.AmountMinor <= 0 || settlement.Currency == "" {
		return ErrPaymentMismatch
	}

	checkoutRecord, err := s.checkouts.get(ctx, settlement.OrganizationID, settlement.CheckoutID)
	if err != nil {
		return err
	}
	if checkoutRecord.Type != TypeWalletTopup || checkoutRecord.AmountMinor != settlement.AmountMinor ||
		checkoutRecord.Currency != settlement.Currency ||
		(checkoutRecord.Provider != "" && string(checkoutRecord.Provider) != settlement.Provider) {
		return ErrPaymentMismatch
	}

	target, ok := checkoutStatusForPayment(settlement.Status)
	if !ok {
		return ErrPaymentMismatch
	}
	if checkoutRecord.Status == target {
		return nil
	}
	if isTerminal(checkoutRecord.Status) || checkoutRecord.Status != StatusProcessing {
		return ErrInvalidTransition
	}

	settledAt := s.now().UTC()
	if settlement.SettledAt != nil {
		settledAt = settlement.SettledAt.UTC()
	}

	if target == StatusSucceeded {
		if checkoutRecord.WalletID == nil {
			return ErrInvalidCheckout
		}
		_, err = s.wallets.post(ctx, checkoutRecord.OrganizationID, *checkoutRecord.WalletID, wallets.PostEntryInput{
			Type:           wallets.EntryTopup,
			AmountMinor:    checkoutRecord.AmountMinor,
			SourceType:     "checkout",
			SourceID:       checkoutRecord.ID.String(),
			IdempotencyKey: "checkout:" + checkoutRecord.ID.String(),
			Metadata:       checkoutRecord.Metadata,
			OccurredAt:     &settledAt,
		})
		if err != nil && !errors.Is(err, wallets.ErrDuplicateLedgerEntry) {
			return err
		}
	}

	_, err = s.checkouts.transition(ctx, checkoutRecord.OrganizationID, checkoutRecord.ID, Transition{
		Expected:    StatusProcessing,
		Status:      target,
		NextAction:  ActionNone,
		CompletedAt: &settledAt,
	})
	if errors.Is(err, ErrInvalidTransition) {
		current, readErr := s.checkouts.get(ctx, checkoutRecord.OrganizationID, checkoutRecord.ID)
		if readErr != nil {
			return readErr
		}
		if current.Status == target {
			return nil
		}
	}
	return err
}

func checkoutStatusForPayment(status commercialpayments.Status) (Status, bool) {
	switch status {
	case commercialpayments.StatusSucceeded:
		return StatusSucceeded, true
	case commercialpayments.StatusFailed:
		return StatusFailed, true
	case commercialpayments.StatusCancelled:
		return StatusCancelled, true
	default:
		return "", false
	}
}
