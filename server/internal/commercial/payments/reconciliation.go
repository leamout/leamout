package payments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/commercial/purchase"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

// checkProviderEventReplay ensures an authenticated but commercially ignored
// event cannot reuse the identity of a previously persisted payment event.
func (r *Repository) checkProviderEventReplay(ctx context.Context, event paymentprovider.Event) error {
	existing, err := r.queries.GetPaymentProviderEvent(ctx, sqlc.GetPaymentProviderEventParams{
		Provider:        event.Provider,
		ProviderEventID: event.ProviderEventID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	digest := sha256.Sum256(event.Raw)
	if existing.EventType != event.Type || existing.PayloadSha256 != hex.EncodeToString(digest[:]) {
		return ErrPaymentMismatch
	}
	return nil
}

// processProviderEvent persists and applies an authenticated provider event in
// one transaction. Purchase owns checkout/order fulfillment and only requests
// prepaid credit after the durable order exists.
func (r *Repository) processProviderEvent(ctx context.Context, event paymentprovider.Event) (Settlement, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Settlement{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.queries.WithTx(tx)

	topup, err := q.LockWalletTopupByReference(ctx, event.Payment.Reference)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Settlement{}, ErrPaymentMismatch
		}
		return Settlement{}, err
	}
	if topup.WalletID == nil || topup.Provider != event.Provider ||
		topup.AmountMinor != event.Payment.AmountMinor || topup.Currency != event.Payment.Currency ||
		(topup.ProviderPaymentID != nil && *topup.ProviderPaymentID != event.Payment.ProviderID) {
		return Settlement{}, ErrPaymentMismatch
	}
	if topup.ProviderPaymentID == nil && event.Payment.ProviderID != "" {
		if _, err = q.SetPaymentProviderID(ctx, sqlc.SetPaymentProviderIDParams{
			ProviderPaymentID: &event.Payment.ProviderID, Status: "processing",
			OrganizationID: topup.OrganizationID, ID: topup.PaymentID,
		}); err != nil {
			return Settlement{}, err
		}
	}

	digest := sha256.Sum256(event.Raw)
	providerEvent, err := q.InsertPaymentProviderEvent(ctx, sqlc.InsertPaymentProviderEventParams{
		PaymentID: topup.PaymentID, OrganizationID: topup.OrganizationID, Provider: event.Provider,
		ProviderEventID: event.ProviderEventID, EventType: event.Type,
		PayloadSha256: hex.EncodeToString(digest[:]), Payload: event.Raw,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		existing, readErr := q.GetPaymentProviderEvent(ctx, sqlc.GetPaymentProviderEventParams{
			Provider: event.Provider, ProviderEventID: event.ProviderEventID,
		})
		if readErr != nil {
			return Settlement{}, readErr
		}
		if existing.PaymentID != topup.PaymentID || existing.EventType != event.Type ||
			existing.PayloadSha256 != hex.EncodeToString(digest[:]) {
			return Settlement{}, ErrPaymentMismatch
		}
		return Settlement{Applied: false}, nil
	}
	if err != nil {
		return Settlement{}, err
	}

	result := Settlement{OrganizationID: topup.OrganizationID, WalletID: *topup.WalletID,
		PaymentID: topup.PaymentID, AmountMinor: topup.AmountMinor}
	fulfillment := purchase.FulfillInput{
		OrganizationID: topup.OrganizationID, CheckoutID: topup.CheckoutID,
		PaymentID: topup.PaymentID, WalletID: topup.WalletID,
		Type:                  purchase.CheckoutWalletTopup,
		ExpectedCheckoutState: topup.CheckoutStatus, AmountMinor: topup.AmountMinor,
	}
	store := purchaseStore{queries: q}
	if topup.PaymentStatus == "succeeded" {
		orderID, orderErr := r.purchase.EnsureOrder(ctx, store, fulfillment)
		if orderErr != nil && !errors.Is(orderErr, pgx.ErrNoRows) {
			return Settlement{}, orderErr
		}
		if orderErr == nil {
			result.OrderID = orderID
		}
		if err = q.MarkPaymentProviderEventProcessed(ctx, providerEvent.ID); err != nil {
			return Settlement{}, err
		}
		if err = tx.Commit(ctx); err != nil {
			return Settlement{}, err
		}
		return result, nil
	}

	now := time.Now().UTC()
	fulfillment.CompletedAt = now
	switch event.Payment.Status {
	case paymentprovider.StatusSucceeded:
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{Status: "succeeded", PaidAt: pgconv.NullableTimestamptz(&now), OrganizationID: topup.OrganizationID, ID: topup.PaymentID}); err != nil {
			return Settlement{}, err
		}
		purchaseResult, fulfillErr := r.purchase.Fulfill(ctx, store, fulfillment)
		if fulfillErr != nil {
			return Settlement{}, fulfillErr
		}
		result.OrderID, result.Applied, result.SettledAt = purchaseResult.OrderID, purchaseResult.Applied, &now
	case paymentprovider.StatusFailed, paymentprovider.StatusCancelled:
		status := string(event.Payment.Status)
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{Status: status, OrganizationID: topup.OrganizationID, ID: topup.PaymentID}); err != nil {
			return Settlement{}, err
		}
		if err = r.purchase.Fail(ctx, store, fulfillment, purchase.CheckoutStatus(status)); err != nil {
			return Settlement{}, err
		}
	default:
		return Settlement{}, ErrPaymentMismatch
	}
	if err = q.MarkPaymentProviderEventProcessed(ctx, providerEvent.ID); err != nil {
		return Settlement{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Settlement{}, err
	}
	return result, nil
}

type purchaseStore struct{ queries *sqlc.Queries }

func (s purchaseStore) TransitionCheckout(ctx context.Context, input purchase.FulfillInput, status purchase.CheckoutStatus) error {
	_, err := s.queries.CompareAndSetCheckoutState(ctx, sqlc.CompareAndSetCheckoutStateParams{
		Status: string(status), NextAction: "none", CompletedAt: pgconv.NullableTimestamptz(&input.CompletedAt),
		OrganizationID: input.OrganizationID, ID: input.CheckoutID, ExpectedStatus: input.ExpectedCheckoutState,
	})
	return err
}

func (s purchaseStore) CreateOrder(ctx context.Context, input purchase.FulfillInput) (uuid.UUID, error) {
	order, err := s.queries.CreateOrderFromCheckoutPayment(ctx, sqlc.CreateOrderFromCheckoutPaymentParams{
		CheckoutID: input.CheckoutID, PaymentID: input.PaymentID, OrganizationID: input.OrganizationID,
	})
	return order.ID, err
}

func (s purchaseStore) CreditTopup(ctx context.Context, input purchase.FulfillInput) error {
	if input.WalletID == nil {
		return nil
	}
	_, err := s.queries.CreateWalletLedgerEntry(ctx, sqlc.CreateWalletLedgerEntryParams{
		EntryType: "topup", AmountMinor: input.AmountMinor, SourceType: "payment",
		SourceID: input.PaymentID.String(), IdempotencyKey: fmt.Sprintf("payment:%s", input.PaymentID),
		WalletID: *input.WalletID, OrganizationID: input.OrganizationID,
	})
	return err
}
