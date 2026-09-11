package payments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

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
// one transaction. It updates payment state only; Checkout owns the commercial
// consequence of that settlement.
func (r *Repository) processProviderEvent(ctx context.Context, event ProviderEvent) (Settlement, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Settlement{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.queries.WithTx(tx)

	payment, err := q.LockPaymentByReference(ctx, event.Payment.Reference)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Settlement{}, ErrPaymentMismatch
		}
		return Settlement{}, err
	}
	if payment.Provider != event.Provider ||
		payment.AmountMinor != event.Payment.AmountMinor ||
		payment.Currency != event.Payment.Currency ||
		(payment.ProviderPaymentID != nil && *payment.ProviderPaymentID != event.Payment.ProviderID) {
		return Settlement{}, ErrPaymentMismatch
	}

	if payment.ProviderPaymentID == nil && event.Payment.ProviderID != "" {
		if _, err = q.SetPaymentProviderID(ctx, sqlc.SetPaymentProviderIDParams{
			ProviderPaymentID: &event.Payment.ProviderID,
			Status:            "processing",
			OrganizationID:    payment.OrganizationID,
			ID:                payment.PaymentID,
		}); err != nil {
			return Settlement{}, err
		}
	}

	result := Settlement{
		CheckoutID:     payment.CheckoutID,
		OrganizationID: payment.OrganizationID,
		PaymentID:      payment.PaymentID,
		Provider:       payment.Provider,
		Status:         Status(payment.PaymentStatus),
		AmountMinor:    payment.AmountMinor,
		Currency:       payment.Currency,
	}

	digest := sha256.Sum256(event.Raw)
	providerEvent, err := q.InsertPaymentProviderEvent(ctx, sqlc.InsertPaymentProviderEventParams{
		PaymentID:       payment.PaymentID,
		OrganizationID:  payment.OrganizationID,
		Provider:        event.Provider,
		ProviderEventID: event.ProviderEventID,
		EventType:       event.Type,
		PayloadSha256:   hex.EncodeToString(digest[:]),
		Payload:         event.Raw,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		existing, readErr := q.GetPaymentProviderEvent(ctx, sqlc.GetPaymentProviderEventParams{
			Provider:        event.Provider,
			ProviderEventID: event.ProviderEventID,
		})
		if readErr != nil {
			return Settlement{}, readErr
		}
		if existing.PaymentID != payment.PaymentID || existing.EventType != event.Type ||
			existing.PayloadSha256 != hex.EncodeToString(digest[:]) {
			return Settlement{}, ErrPaymentMismatch
		}
		if err = tx.Commit(ctx); err != nil {
			return Settlement{}, err
		}
		return result, nil
	}
	if err != nil {
		return Settlement{}, err
	}

	currentStatus := Status(payment.PaymentStatus)
	if currentStatus == StatusSucceeded || currentStatus == StatusFailed || currentStatus == StatusCancelled {
		if currentStatus != event.Payment.Status {
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

	now := time.Now().UTC()
	switch event.Payment.Status {
	case StatusSucceeded:
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{
			Status:         "succeeded",
			PaidAt:         pgconv.NullableTimestamptz(&now),
			OrganizationID: payment.OrganizationID,
			ID:             payment.PaymentID,
		}); err != nil {
			return Settlement{}, err
		}
		result.Status = StatusSucceeded
		result.SettledAt = &now
		result.Applied = true
	case StatusFailed, StatusCancelled:
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{
			Status:         string(event.Payment.Status),
			OrganizationID: payment.OrganizationID,
			ID:             payment.PaymentID,
		}); err != nil {
			return Settlement{}, err
		}
		result.Status = event.Payment.Status
		result.SettledAt = &now
		result.Applied = true
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
