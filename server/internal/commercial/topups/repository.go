package topups

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
	now     func() time.Time
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, queries: sqlc.New(db), now: time.Now}
}

// Reconcile records a verified provider event and applies its commercial
// consequence in one transaction. Duplicate events and later success events
// cannot post a second wallet credit.
func (r *Repository) Reconcile(ctx context.Context, event paymentprovider.Event) (Settlement, error) {
	if event.Provider == "" || event.ProviderEventID == "" || event.Type == "" ||
		event.Payment.Reference == "" || len(event.Raw) == 0 || !json.Valid(event.Raw) {
		return Settlement{}, ErrPaymentMismatch
	}

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
		topup.Amount != event.Payment.AmountMinor || topup.Currency != event.Payment.Currency ||
		event.Payment.ProviderID == "" ||
		(topup.ProviderPaymentID != nil && *topup.ProviderPaymentID != event.Payment.ProviderID) {
		return Settlement{}, ErrPaymentMismatch
	}

	if topup.ProviderPaymentID == nil {
		if _, err = q.SetPaymentProviderID(ctx, sqlc.SetPaymentProviderIDParams{
			ProviderPaymentID: &event.Payment.ProviderID,
			Status:            "processing",
			OrganizationID:    topup.OrganizationID,
			ID:                topup.PaymentID,
		}); err != nil {
			return Settlement{}, err
		}
	}

	digest := sha256.Sum256(event.Raw)
	providerEvent, err := q.InsertPaymentProviderEvent(ctx, sqlc.InsertPaymentProviderEventParams{
		PaymentID:       topup.PaymentID,
		OrganizationID:  topup.OrganizationID,
		Provider:        event.Provider,
		ProviderEventID: event.ProviderEventID,
		EventType:       event.Type,
		PayloadSha256:   hex.EncodeToString(digest[:]),
		Payload:         event.Raw,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Settlement{Applied: false}, nil
	}
	if err != nil {
		return Settlement{}, err
	}

	result := Settlement{
		OrganizationID: topup.OrganizationID,
		WalletID:       *topup.WalletID,
		PaymentID:      topup.PaymentID,
		AmountMinor:    topup.Amount,
	}
	if topup.PaymentStatus == "succeeded" {
		if err = q.MarkPaymentProviderEventProcessed(ctx, providerEvent.ID); err != nil {
			return Settlement{}, err
		}
		if err = tx.Commit(ctx); err != nil {
			return Settlement{}, err
		}
		return result, nil
	}

	now := r.now().UTC()
	switch event.Payment.Status {
	case paymentprovider.StatusSucceeded:
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{
			Status:         "succeeded",
			PaidAt:         pgconv.NullableTimestamptz(&now),
			OrganizationID: topup.OrganizationID,
			ID:             topup.PaymentID,
		}); err != nil {
			return Settlement{}, err
		}
		if _, err = q.CompareAndSetCheckoutOrderState(ctx, sqlc.CompareAndSetCheckoutOrderStateParams{
			Status:         "succeeded",
			NextAction:     "none",
			CompletedAt:    pgconv.NullableTimestamptz(&now),
			OrganizationID: topup.OrganizationID,
			ID:             topup.CheckoutOrderID,
			ExpectedStatus: topup.CheckoutStatus,
		}); err != nil {
			return Settlement{}, err
		}
		if _, err = q.CreateWalletLedgerEntry(ctx, sqlc.CreateWalletLedgerEntryParams{
			EntryType:      "topup",
			Amount:         topup.Amount,
			SourceType:     "payment",
			SourceID:       topup.PaymentID.String(),
			IdempotencyKey: fmt.Sprintf("payment:%s", topup.PaymentID),
			WalletID:       *topup.WalletID,
			OrganizationID: topup.OrganizationID,
		}); err != nil {
			return Settlement{}, err
		}
		result.Applied = true
		result.SettledAt = &now
	case paymentprovider.StatusFailed, paymentprovider.StatusCancelled:
		status := string(event.Payment.Status)
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{
			Status: status, OrganizationID: topup.OrganizationID, ID: topup.PaymentID,
		}); err != nil {
			return Settlement{}, err
		}
		if _, err = q.CompareAndSetCheckoutOrderState(ctx, sqlc.CompareAndSetCheckoutOrderStateParams{
			Status: status, NextAction: "none", CompletedAt: pgconv.NullableTimestamptz(&now),
			OrganizationID: topup.OrganizationID, ID: topup.CheckoutOrderID,
			ExpectedStatus: topup.CheckoutStatus,
		}); err != nil {
			return Settlement{}, err
		}
	}

	if err = q.MarkPaymentProviderEventProcessed(ctx, providerEvent.ID); err != nil {
		return Settlement{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Settlement{}, err
	}
	return result, nil
}
