package checkout

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/catalog"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/prepaid"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
)

// CompletePayment applies the commercial effect of a settled payment. Payments
// owns provider state; Checkout owns what a successful payment means.
func (s *Service) CompletePayment(ctx context.Context, settlement commercialpayments.Settlement) error {
	if settlement.CheckoutID == uuid.Nil || settlement.OrganizationID == uuid.Nil ||
		settlement.PaymentID == uuid.Nil || settlement.AmountMinor <= 0 || settlement.Currency == "" {
		return ErrPaymentMismatch
	}

	checkoutRecord, err := s.repo.Get(ctx, settlement.OrganizationID, settlement.CheckoutID)
	if err != nil {
		return err
	}
	if checkoutRecord.AmountMinor != settlement.AmountMinor ||
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
		if err = s.fulfillSucceededCheckout(ctx, checkoutRecord, settledAt); err != nil {
			return err
		}
	}

	_, err = s.repo.Transition(
		ctx,
		checkoutRecord.OrganizationID,
		checkoutRecord.ID,
		Transition{
			Expected:    StatusProcessing,
			Status:      target,
			NextAction:  ActionNone,
			CompletedAt: &settledAt,
		},
	)
	if errors.Is(err, ErrInvalidTransition) {
		current, readErr := s.repo.Get(ctx, checkoutRecord.OrganizationID, checkoutRecord.ID)
		if readErr == nil && current.Status == target {
			return nil
		}
	}
	return err
}

func (s *Service) fulfillSucceededCheckout(
	ctx context.Context,
	checkoutRecord Checkout,
	settledAt time.Time,
) error {
	switch checkoutRecord.Type {
	case TypeWalletTopup:
		if checkoutRecord.WalletID == nil {
			return ErrInvalidCheckout
		}
		_, err := s.wallets.Post(
			ctx,
			checkoutRecord.OrganizationID,
			*checkoutRecord.WalletID,
			prepaid.PostEntryInput{
				Type:           prepaid.EntryTopup,
				AmountMinor:    checkoutRecord.AmountMinor,
				SourceType:     "checkout",
				SourceID:       checkoutRecord.ID.String(),
				IdempotencyKey: "checkout:" + checkoutRecord.ID.String(),
				Metadata:       checkoutRecord.Metadata,
				OccurredAt:     &settledAt,
			},
		)
		if errors.Is(err, prepaid.ErrDuplicateLedgerEntry) {
			return nil
		}
		return err

	case TypeSubscription:
		if checkoutRecord.PriceID == nil {
			return ErrInvalidCheckout
		}
		price, err := s.catalog.GetPrice(ctx, *checkoutRecord.PriceID)
		if err != nil {
			return err
		}
		if price.PricingType != catalog.PricingTypeRecurring || price.BillingInterval == nil {
			return ErrInvalidCheckout
		}
		renewsAt := renewalTime(settledAt, *price.BillingInterval)
		if renewsAt == nil {
			return ErrInvalidCheckout
		}
		status := subscriptions.StatusActive
		_, err = s.subscriptions.Create(ctx, checkoutRecord.OrganizationID, subscriptions.CreateInput{
			PriceID:  *checkoutRecord.PriceID,
			Status:   &status,
			StartsAt: &settledAt,
			RenewsAt: renewsAt,
		})
		if errors.Is(err, subscriptions.ErrCurrentSubscriptionExists) {
			current, readErr := s.subscriptions.Current(ctx, checkoutRecord.OrganizationID)
			if readErr == nil && current.PriceID == *checkoutRecord.PriceID {
				return nil
			}
		}
		return err

	default:
		return ErrInvalidCheckout
	}
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

func renewalTime(start time.Time, interval catalog.BillingInterval) *time.Time {
	var value time.Time
	switch interval {
	case catalog.BillingIntervalMonth:
		value = start.AddDate(0, 1, 0)
	case catalog.BillingIntervalYear:
		value = start.AddDate(1, 0, 0)
	default:
		return nil
	}
	return &value
}
