package wallets

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

type TopupRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
	now     func() time.Time
}

func NewTopupRepository(db *pgxpool.Pool) *TopupRepository {
	return &TopupRepository{db: db, queries: sqlc.New(db), now: time.Now}
}

// Reconcile records a verified provider event and applies its commercial
// consequence in one transaction. Duplicate events and later success events
// cannot post a second wallet credit.
func (r *TopupRepository) Reconcile(ctx context.Context, event paymentprovider.Event) (TopupSettlement, error) {
	if event.Provider == "" || event.ProviderEventID == "" || event.Type == "" ||
		event.Payment.Reference == "" || len(event.Raw) == 0 || !json.Valid(event.Raw) {
		return TopupSettlement{}, ErrPaymentMismatch
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return TopupSettlement{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.queries.WithTx(tx)

	topup, err := q.LockWalletTopupByReference(ctx, event.Payment.Reference)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TopupSettlement{}, ErrPaymentMismatch
		}
		return TopupSettlement{}, err
	}
	if topup.WalletID == nil || topup.Provider != event.Provider ||
		topup.AmountMinor != event.Payment.AmountMinor || topup.Currency != event.Payment.Currency ||
		(topup.ProviderPaymentID != nil && *topup.ProviderPaymentID != event.Payment.ProviderID) {
		return TopupSettlement{}, ErrPaymentMismatch
	}

	if topup.ProviderPaymentID == nil && event.Payment.ProviderID != "" {
		if _, err = q.SetPaymentProviderID(ctx, sqlc.SetPaymentProviderIDParams{
			ProviderPaymentID: &event.Payment.ProviderID,
			Status:            "processing",
			OrganizationID:    topup.OrganizationID,
			ID:                topup.PaymentID,
		}); err != nil {
			return TopupSettlement{}, err
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
		return TopupSettlement{Applied: false}, nil
	}
	if err != nil {
		return TopupSettlement{}, err
	}

	result := TopupSettlement{
		OrganizationID: topup.OrganizationID,
		WalletID:       *topup.WalletID,
		PaymentID:      topup.PaymentID,
		AmountMinor:    topup.AmountMinor,
	}
	if topup.PaymentStatus == "succeeded" {
		if err = q.MarkPaymentProviderEventProcessed(ctx, providerEvent.ID); err != nil {
			return TopupSettlement{}, err
		}
		if err = tx.Commit(ctx); err != nil {
			return TopupSettlement{}, err
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
			return TopupSettlement{}, err
		}
		if _, err = q.CompareAndSetCheckoutOrderState(ctx, sqlc.CompareAndSetCheckoutOrderStateParams{
			Status:         "succeeded",
			NextAction:     "none",
			CompletedAt:    pgconv.NullableTimestamptz(&now),
			OrganizationID: topup.OrganizationID,
			ID:             topup.CheckoutOrderID,
			ExpectedStatus: topup.CheckoutStatus,
		}); err != nil {
			return TopupSettlement{}, err
		}
		if _, err = q.CreateWalletLedgerEntry(ctx, sqlc.CreateWalletLedgerEntryParams{
			EntryType:      "topup",
			AmountMinor:    topup.AmountMinor,
			SourceType:     "payment",
			SourceID:       topup.PaymentID.String(),
			IdempotencyKey: fmt.Sprintf("payment:%s", topup.PaymentID),
			WalletID:       *topup.WalletID,
			OrganizationID: topup.OrganizationID,
		}); err != nil {
			return TopupSettlement{}, err
		}
		result.Applied = true
		result.SettledAt = &now
	case paymentprovider.StatusFailed, paymentprovider.StatusCancelled:
		status := string(event.Payment.Status)
		if _, err = q.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{
			Status: status, OrganizationID: topup.OrganizationID, ID: topup.PaymentID,
		}); err != nil {
			return TopupSettlement{}, err
		}
		if _, err = q.CompareAndSetCheckoutOrderState(ctx, sqlc.CompareAndSetCheckoutOrderStateParams{
			Status: status, NextAction: "none", CompletedAt: pgconv.NullableTimestamptz(&now),
			OrganizationID: topup.OrganizationID, ID: topup.CheckoutOrderID,
			ExpectedStatus: topup.CheckoutStatus,
		}); err != nil {
			return TopupSettlement{}, err
		}
	}

	if err = q.MarkPaymentProviderEventProcessed(ctx, providerEvent.ID); err != nil {
		return TopupSettlement{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return TopupSettlement{}, err
	}
	return result, nil
}
