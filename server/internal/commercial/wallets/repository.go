package wallets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	return &Repository{
		db:      db,
		queries: sqlc.New(db),
		now:     time.Now,
	}
}

func (r *Repository) currentTime() time.Time {
	if r.now != nil {
		return r.now()
	}
	return time.Now()
}

func (r *Repository) Create(ctx context.Context, organizationID uuid.UUID, currency string) (Wallet, error) {
	row, err := r.queries.CreateWallet(ctx, sqlc.CreateWalletParams{
		OrganizationID: organizationID,
		Currency:       currency,
	})
	if err != nil {
		return Wallet{}, mapWalletWriteError(err)
	}
	return walletFromRow(row), nil
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (Wallet, error) {
	row, err := r.queries.GetWallet(ctx, sqlc.GetWalletParams{
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Wallet{}, mapWalletReadError(err)
	}
	return walletFromRow(row), nil
}

func (r *Repository) GetByCurrency(ctx context.Context, organizationID uuid.UUID, currency string) (Wallet, error) {
	row, err := r.queries.GetWalletByCurrency(ctx, sqlc.GetWalletByCurrencyParams{
		OrganizationID: organizationID,
		Currency:       currency,
	})
	if err != nil {
		return Wallet{}, mapWalletReadError(err)
	}
	return walletFromRow(row), nil
}

func (r *Repository) Balance(ctx context.Context, organizationID, walletID uuid.UUID) (Balance, error) {
	row, err := r.queries.GetWalletBalance(ctx, sqlc.GetWalletBalanceParams{
		OrganizationID: organizationID,
		WalletID:       walletID,
	})
	if err != nil {
		return Balance{}, mapWalletReadError(err)
	}
	return Balance{
		PostedMinor:    row.PostedMinor,
		ReservedMinor:  row.ReservedMinor,
		AvailableMinor: row.AvailableMinor,
	}, nil
}

func (r *Repository) Post(ctx context.Context, organizationID, walletID uuid.UUID, input PostEntryInput) (LedgerEntry, error) {
	if input.Type == EntryCapture {
		return LedgerEntry{}, ErrReservationRequired
	}
	row, err := r.queries.CreateWalletLedgerEntry(ctx, entryParams(organizationID, walletID, input))
	if err != nil {
		return LedgerEntry{}, mapWalletWriteError(err)
	}
	return ledgerEntryFromRow(row), nil
}

func (r *Repository) ListEntries(ctx context.Context, organizationID, walletID uuid.UUID) ([]LedgerEntry, error) {
	rows, err := r.queries.ListWalletLedgerEntries(ctx, sqlc.ListWalletLedgerEntriesParams{
		WalletID:       walletID,
		OrganizationID: organizationID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]LedgerEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, ledgerEntryFromRow(row))
	}
	return result, nil
}

// Reserve serializes on the wallet row before checking available funds.
func (r *Repository) Reserve(ctx context.Context, organizationID, walletID uuid.UUID, input ReserveInput) (Reservation, error) {
	if err := validateReservationInput(input, r.currentTime()); err != nil {
		return Reservation{}, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := r.queries.WithTx(tx)
	if _, err = q.LockActiveWallet(ctx, sqlc.LockActiveWalletParams{
		ID:             walletID,
		OrganizationID: organizationID,
	}); err != nil {
		return Reservation{}, mapWalletReadError(err)
	}
	balance, err := q.GetWalletBalance(ctx, sqlc.GetWalletBalanceParams{
		OrganizationID: organizationID,
		WalletID:       walletID,
	})
	if err != nil {
		return Reservation{}, mapWalletReadError(err)
	}
	if balance.AvailableMinor < input.AmountMinor {
		return Reservation{}, ErrInsufficientFunds
	}
	row, err := q.InsertWalletReservation(ctx, sqlc.InsertWalletReservationParams{
		WalletID:       walletID,
		OrganizationID: organizationID,
		AmountMinor:    input.AmountMinor,
		OperationType:  input.OperationType,
		OperationID:    input.OperationID,
		ExpiresAt:      pgconv.NullableTimestamptz(&input.ExpiresAt),
	})
	if err != nil {
		return Reservation{}, mapWalletWriteError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Reservation{}, err
	}
	return reservationFromRow(row), nil
}

func (r *Repository) GetReservation(ctx context.Context, organizationID, id uuid.UUID) (Reservation, error) {
	row, err := r.queries.GetWalletReservation(ctx, sqlc.GetWalletReservationParams{
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Reservation{}, mapReservationReadError(err)
	}
	return reservationFromRow(row), nil
}

// Capture closes the reservation and posts its debit atomically.
func (r *Repository) Capture(ctx context.Context, organizationID, id uuid.UUID, amountMinor int64, idempotencyKey string) (Reservation, error) {
	if err := validateCaptureInput(amountMinor); err != nil {
		return Reservation{}, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := r.queries.WithTx(tx)
	row, err := q.CaptureWalletReservation(ctx, sqlc.CaptureWalletReservationParams{
		CapturedAmountMinor: &amountMinor,
		OrganizationID:      organizationID,
		ID:                  id,
	})
	if err != nil {
		return Reservation{}, mapReservationTransitionError(err)
	}
	_, err = q.CreateWalletLedgerEntry(ctx, entryParams(organizationID, row.WalletID, PostEntryInput{
		Type:           EntryCapture,
		AmountMinor:    -amountMinor,
		SourceType:     "wallet_reservation",
		SourceID:       row.ID.String(),
		IdempotencyKey: idempotencyKey,
	}))
	if err != nil {
		return Reservation{}, mapWalletWriteError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Reservation{}, err
	}
	return reservationFromRow(row), nil
}

func (r *Repository) Release(ctx context.Context, organizationID, id uuid.UUID) (Reservation, error) {
	row, err := r.queries.ReleaseWalletReservation(ctx, sqlc.ReleaseWalletReservationParams{
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Reservation{}, mapReservationTransitionError(err)
	}
	return reservationFromRow(row), nil
}

func (r *Repository) Expire(ctx context.Context) ([]Reservation, error) {
	rows, err := r.queries.ExpireWalletReservations(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Reservation, 0, len(rows))
	for _, row := range rows {
		result = append(result, reservationFromRow(row))
	}
	return result, nil
}

// Reconcile records a verified provider event and applies its wallet consequence
// in one transaction. Duplicate events cannot post a second wallet credit.
func (r *Repository) Reconcile(ctx context.Context, event paymentprovider.Event) (TopupSettlement, error) {
	if err := validateProviderEvent(event); err != nil {
		return TopupSettlement{}, err
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
		order, orderErr := q.CreateOrderFromCheckoutPayment(ctx, sqlc.CreateOrderFromCheckoutPaymentParams{
			CheckoutID:     topup.CheckoutID,
			PaymentID:      topup.PaymentID,
			OrganizationID: topup.OrganizationID,
		})
		if orderErr != nil && !errors.Is(orderErr, pgx.ErrNoRows) {
			return TopupSettlement{}, orderErr
		}
		if orderErr == nil {
			result.OrderID = order.ID
		}
		if err = q.MarkPaymentProviderEventProcessed(ctx, providerEvent.ID); err != nil {
			return TopupSettlement{}, err
		}
		if err = tx.Commit(ctx); err != nil {
			return TopupSettlement{}, err
		}
		return result, nil
	}

	now := r.currentTime().UTC()
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
		if _, err = q.CompareAndSetCheckoutState(ctx, sqlc.CompareAndSetCheckoutStateParams{
			Status:         "succeeded",
			NextAction:     "none",
			CompletedAt:    pgconv.NullableTimestamptz(&now),
			OrganizationID: topup.OrganizationID,
			ID:             topup.CheckoutID,
			ExpectedStatus: topup.CheckoutStatus,
		}); err != nil {
			return TopupSettlement{}, err
		}
		order, orderErr := q.CreateOrderFromCheckoutPayment(ctx, sqlc.CreateOrderFromCheckoutPaymentParams{
			CheckoutID:     topup.CheckoutID,
			PaymentID:      topup.PaymentID,
			OrganizationID: topup.OrganizationID,
		})
		if orderErr != nil {
			return TopupSettlement{}, orderErr
		}
		result.OrderID = order.ID
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
			Status:         status,
			OrganizationID: topup.OrganizationID,
			ID:             topup.PaymentID,
		}); err != nil {
			return TopupSettlement{}, err
		}
		if _, err = q.CompareAndSetCheckoutState(ctx, sqlc.CompareAndSetCheckoutStateParams{
			Status:         status,
			NextAction:     "none",
			CompletedAt:    pgconv.NullableTimestamptz(&now),
			OrganizationID: topup.OrganizationID,
			ID:             topup.CheckoutID,
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

func entryParams(organizationID, walletID uuid.UUID, input PostEntryInput) sqlc.CreateWalletLedgerEntryParams {
	return sqlc.CreateWalletLedgerEntryParams{
		EntryType:      string(input.Type),
		AmountMinor:    input.AmountMinor,
		SourceType:     input.SourceType,
		SourceID:       input.SourceID,
		IdempotencyKey: input.IdempotencyKey,
		Metadata:       input.Metadata,
		OccurredAt:     pgconv.NullableTimestamptz(input.OccurredAt),
		WalletID:       walletID,
		OrganizationID: organizationID,
	}
}

func walletFromRow(row sqlc.Wallet) Wallet {
	return Wallet{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		Currency:       row.Currency,
		Status:         Status(row.Status),
		CreatedAt:      pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:      pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func ledgerEntryFromRow(row sqlc.WalletLedgerEntry) LedgerEntry {
	return LedgerEntry{
		ID:             row.ID,
		WalletID:       row.WalletID,
		OrganizationID: row.OrganizationID,
		Type:           EntryType(row.EntryType),
		AmountMinor:    row.AmountMinor,
		SourceType:     row.SourceType,
		SourceID:       row.SourceID,
		IdempotencyKey: row.IdempotencyKey,
		Metadata:       row.Metadata,
		OccurredAt:     pgconv.TimestamptzToTime(row.OccurredAt),
		CreatedAt:      pgconv.TimestamptzToTime(row.CreatedAt),
	}
}

func reservationFromRow(row sqlc.WalletReservation) Reservation {
	return Reservation{
		ID:                  row.ID,
		WalletID:            row.WalletID,
		OrganizationID:      row.OrganizationID,
		AmountMinor:         row.AmountMinor,
		CapturedAmountMinor: row.CapturedAmountMinor,
		OperationType:       row.OperationType,
		OperationID:         row.OperationID,
		Status:              ReservationStatus(row.Status),
		ExpiresAt:           pgconv.TimestamptzToTime(row.ExpiresAt),
		CapturedAt:          pgconv.TimestamptzToTimePtr(row.CapturedAt),
		ReleasedAt:          pgconv.TimestamptzToTimePtr(row.ReleasedAt),
		ExpiredAt:           pgconv.TimestamptzToTimePtr(row.ExpiredAt),
		CreatedAt:           pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:           pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func mapWalletReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrWalletNotFound
	}
	return err
}

func mapReservationReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrReservationNotFound
	}
	return err
}

func mapReservationTransitionError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidReservationState
	}
	return mapWalletWriteError(err)
}

func mapWalletWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "uq_wallets_organization_currency":
			return ErrWalletExists
		case "uq_wallet_reservations_operation":
			return ErrDuplicateOperation
		case "uq_wallet_ledger_idempotency":
			return ErrDuplicateLedgerEntry
		}
		if pgErr.Code == "23514" {
			return ErrInvalidMoney
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrWalletNotFound
	}
	return err
}
