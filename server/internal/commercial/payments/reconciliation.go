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
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// checkProviderEventReplay ensures an authenticated but commercially ignored
// event cannot reuse the identity of a previously persisted payment event.
func (r *Repository) checkProviderEventReplay(ctx context.Context, event ProviderEvent) error {
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
// one transaction. A successful top-up completes its checkout and posts one
// idempotent prepaid credit; payment settlement does not create an order.
func (r *Repository) processProviderEvent(ctx context.Context, event ProviderEvent) (Settlement, error) {
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
	case StatusSucceeded:
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{Status: "succeeded", PaidAt: pgconv.NullableTimestamptz(&now), OrganizationID: topup.OrganizationID, ID: topup.PaymentID}); err != nil {
			return Settlement{}, err
		}
		if err = transitionCheckout(ctx, q, topup.OrganizationID, topup.CheckoutID, topup.CheckoutStatus, "succeeded", now); err != nil {
			return Settlement{}, err
		}
		if _, err = q.CreateWalletLedgerEntry(ctx, sqlc.CreateWalletLedgerEntryParams{
			EntryType: "topup", AmountMinor: topup.AmountMinor, SourceType: "payment",
			SourceID: topup.PaymentID.String(), IdempotencyKey: fmt.Sprintf("payment:%s", topup.PaymentID),
			WalletID: *topup.WalletID, OrganizationID: topup.OrganizationID,
		}); err != nil {
			return Settlement{}, err
		}
		result.Applied, result.SettledAt = true, &now
	case StatusFailed, StatusCancelled:
		status := string(event.Payment.Status)
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{Status: status, OrganizationID: topup.OrganizationID, ID: topup.PaymentID}); err != nil {
			return Settlement{}, err
		}
		if err = transitionCheckout(ctx, q, topup.OrganizationID, topup.CheckoutID, topup.CheckoutStatus, status, now); err != nil {
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

func transitionCheckout(ctx context.Context, q *sqlc.Queries, organizationID, checkoutID uuid.UUID, expected, status string, completedAt time.Time) error {
	_, err := q.CompareAndSetCheckoutState(ctx, sqlc.CompareAndSetCheckoutStateParams{
		Status: status, NextAction: "none", CompletedAt: pgconv.NullableTimestamptz(&completedAt),
		OrganizationID: organizationID, ID: checkoutID, ExpectedStatus: expected,
	})
	return err
}
