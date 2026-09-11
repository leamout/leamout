package payments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

// ProcessProviderEvent persists and applies an authenticated provider event in
// one transaction. Payment settlement owns checkout/order transitions and only
// posts prepaid credit after the durable order exists.
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
	if topup.PaymentStatus == "succeeded" {
		order, orderErr := q.CreateOrderFromCheckoutPayment(ctx, sqlc.CreateOrderFromCheckoutPaymentParams{
			CheckoutID: topup.CheckoutID, PaymentID: topup.PaymentID, OrganizationID: topup.OrganizationID,
		})
		if orderErr != nil && !errors.Is(orderErr, pgx.ErrNoRows) {
			return Settlement{}, orderErr
		}
		if orderErr == nil {
			result.OrderID = order.ID
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
	switch event.Payment.Status {
	case paymentprovider.StatusSucceeded:
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{Status: "succeeded", PaidAt: pgconv.NullableTimestamptz(&now), OrganizationID: topup.OrganizationID, ID: topup.PaymentID}); err != nil {
			return Settlement{}, err
		}
		if _, err = q.CompareAndSetCheckoutState(ctx, sqlc.CompareAndSetCheckoutStateParams{Status: "succeeded", NextAction: "none", CompletedAt: pgconv.NullableTimestamptz(&now), OrganizationID: topup.OrganizationID, ID: topup.CheckoutID, ExpectedStatus: topup.CheckoutStatus}); err != nil {
			return Settlement{}, err
		}
		order, orderErr := q.CreateOrderFromCheckoutPayment(ctx, sqlc.CreateOrderFromCheckoutPaymentParams{CheckoutID: topup.CheckoutID, PaymentID: topup.PaymentID, OrganizationID: topup.OrganizationID})
		if orderErr != nil {
			return Settlement{}, orderErr
		}
		result.OrderID = order.ID
		if _, err = q.CreateWalletLedgerEntry(ctx, sqlc.CreateWalletLedgerEntryParams{EntryType: "topup", AmountMinor: topup.AmountMinor, SourceType: "payment", SourceID: topup.PaymentID.String(), IdempotencyKey: fmt.Sprintf("payment:%s", topup.PaymentID), WalletID: *topup.WalletID, OrganizationID: topup.OrganizationID}); err != nil {
			return Settlement{}, err
		}
		result.Applied, result.SettledAt = true, &now
	case paymentprovider.StatusFailed, paymentprovider.StatusCancelled:
		status := string(event.Payment.Status)
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{Status: status, OrganizationID: topup.OrganizationID, ID: topup.PaymentID}); err != nil {
			return Settlement{}, err
		}
		if _, err = q.CompareAndSetCheckoutState(ctx, sqlc.CompareAndSetCheckoutStateParams{Status: status, NextAction: "none", CompletedAt: pgconv.NullableTimestamptz(&now), OrganizationID: topup.OrganizationID, ID: topup.CheckoutID, ExpectedStatus: topup.CheckoutStatus}); err != nil {
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
