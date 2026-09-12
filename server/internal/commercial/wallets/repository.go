package wallets

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, queries: sqlc.New(db)}
}

func (r *Repository) Create(ctx context.Context, organizationID uuid.UUID, currency string) (Wallet, error) {
	row, err := r.queries.CreateWallet(ctx, sqlc.CreateWalletParams{OrganizationID: organizationID, Currency: currency})
	if err != nil {
		return Wallet{}, mapWalletWriteError(err)
	}
	return walletFromRow(row), nil
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (Wallet, error) {
	row, err := r.queries.GetWallet(ctx, sqlc.GetWalletParams{OrganizationID: organizationID, ID: id})
	if err != nil {
		return Wallet{}, mapWalletReadError(err)
	}
	return walletFromRow(row), nil
}

func (r *Repository) List(ctx context.Context, organizationID uuid.UUID) ([]Wallet, error) {
	rows, err := r.queries.ListWallets(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	result := make([]Wallet, 0, len(rows))
	for _, row := range rows {
		result = append(result, walletFromRow(row))
	}
	return result, nil
}

func (r *Repository) GetByCurrency(ctx context.Context, organizationID uuid.UUID, currency string) (Wallet, error) {
	row, err := r.queries.GetWalletByCurrency(ctx, sqlc.GetWalletByCurrencyParams{OrganizationID: organizationID, Currency: currency})
	if err != nil {
		return Wallet{}, mapWalletReadError(err)
	}
	return walletFromRow(row), nil
}

func (r *Repository) Balance(ctx context.Context, organizationID, walletID uuid.UUID) (Balance, error) {
	row, err := r.queries.GetWalletBalance(ctx, sqlc.GetWalletBalanceParams{OrganizationID: organizationID, WalletID: walletID})
	if err != nil {
		return Balance{}, mapWalletReadError(err)
	}
	return Balance{PostedMinor: row.PostedMinor, ReservedMinor: row.ReservedMinor, AvailableMinor: row.AvailableMinor}, nil
}

func (r *Repository) Post(ctx context.Context, organizationID, walletID uuid.UUID, input PostEntryInput) (LedgerEntry, error) {
	row, err := r.queries.CreateWalletLedgerEntry(ctx, entryParams(organizationID, walletID, input))
	if err != nil {
		return LedgerEntry{}, mapWalletWriteError(err)
	}
	return ledgerEntryFromRow(row), nil
}

func (r *Repository) ListEntries(ctx context.Context, organizationID, walletID uuid.UUID) ([]LedgerEntry, error) {
	rows, err := r.queries.ListWalletLedgerEntries(ctx, sqlc.ListWalletLedgerEntriesParams{WalletID: walletID, OrganizationID: organizationID})
	if err != nil {
		return nil, err
	}
	result := make([]LedgerEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, ledgerEntryFromRow(row))
	}
	return result, nil
}

func (r *Repository) Reserve(ctx context.Context, organizationID, walletID uuid.UUID, input ReserveInput) (Reservation, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.queries.WithTx(tx)
	if _, err = q.LockActiveWallet(ctx, sqlc.LockActiveWalletParams{ID: walletID, OrganizationID: organizationID}); err != nil {
		return Reservation{}, mapWalletReadError(err)
	}
	balance, err := q.GetWalletBalance(ctx, sqlc.GetWalletBalanceParams{OrganizationID: organizationID, WalletID: walletID})
	if err != nil {
		return Reservation{}, mapWalletReadError(err)
	}
	if balance.AvailableMinor < input.AmountMinor {
		return Reservation{}, ErrInsufficientFunds
	}
	row, err := q.InsertWalletReservation(ctx, sqlc.InsertWalletReservationParams{
		WalletID: walletID, OrganizationID: organizationID, AmountMinor: input.AmountMinor,
		OperationType: input.OperationType, OperationID: input.OperationID,
		ExpiresAt: pgconv.NullableTimestamptz(&input.ExpiresAt),
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
	row, err := r.queries.GetWalletReservation(ctx, sqlc.GetWalletReservationParams{OrganizationID: organizationID, ID: id})
	if err != nil {
		return Reservation{}, mapReservationReadError(err)
	}
	return reservationFromRow(row), nil
}

func (r *Repository) Capture(ctx context.Context, organizationID, id uuid.UUID, amountMinor int64, idempotencyKey string) (Reservation, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.queries.WithTx(tx)
	row, err := q.CaptureWalletReservation(ctx, sqlc.CaptureWalletReservationParams{CapturedAmountMinor: &amountMinor, OrganizationID: organizationID, ID: id})
	if err != nil {
		return Reservation{}, mapReservationTransitionError(err)
	}
	_, err = q.CreateWalletLedgerEntry(ctx, entryParams(organizationID, row.WalletID, PostEntryInput{
		Type: EntryCapture, AmountMinor: -amountMinor, SourceType: "wallet_reservation",
		SourceID: row.ID.String(), IdempotencyKey: idempotencyKey,
	}))
	if err != nil {
		return Reservation{}, mapWalletWriteError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Reservation{}, err
	}
	return reservationFromRow(row), nil
}

// IncreaseTo moves an active reservation to a target total hold. Replaying the
// same target is a no-op, and stale retries cannot reduce the amount or expiry.
func (r *Repository) IncreaseTo(ctx context.Context, organizationID, id uuid.UUID, input IncreaseReservationInput) (Reservation, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.queries.WithTx(tx)

	reservation, err := q.GetWalletReservation(ctx, sqlc.GetWalletReservationParams{OrganizationID: organizationID, ID: id})
	if err != nil {
		return Reservation{}, mapReservationReadError(err)
	}
	if _, err = q.LockActiveWallet(ctx, sqlc.LockActiveWalletParams{ID: reservation.WalletID, OrganizationID: organizationID}); err != nil {
		return Reservation{}, mapWalletReadError(err)
	}
	reservation, err = q.GetWalletReservation(ctx, sqlc.GetWalletReservationParams{OrganizationID: organizationID, ID: id})
	if err != nil {
		return Reservation{}, mapReservationReadError(err)
	}
	if reservation.Status != ReservationActive || !reservation.ExpiresAt.After(input.ExpiresAt.Add(-input.ExpiresAt.Sub(input.ExpiresAt))) {
		// The SQL transition below remains the source of truth for expiry/state;
		// this branch only handles obvious backward retries before balance work.
	}
	if input.TargetAmountMinor < reservation.AmountMinor || input.ExpiresAt.Before(reservation.ExpiresAt) {
		return Reservation{}, ErrReservationRegression
	}
	if input.TargetAmountMinor == reservation.AmountMinor && input.ExpiresAt.Equal(reservation.ExpiresAt) {
		return reservationFromRow(reservation), nil
	}
	additional := input.TargetAmountMinor - reservation.AmountMinor
	if additional > 0 {
		balance, balanceErr := q.GetWalletBalance(ctx, sqlc.GetWalletBalanceParams{OrganizationID: organizationID, WalletID: reservation.WalletID})
		if balanceErr != nil {
			return Reservation{}, mapWalletReadError(balanceErr)
		}
		if balance.AvailableMinor < additional {
			return Reservation{}, ErrInsufficientFunds
		}
	}
	row, err := q.IncreaseWalletReservation(ctx, sqlc.IncreaseWalletReservationParams{
		IncrementMinor: input.TargetAmountMinor,
		ExpiresAt: pgconv.NullableTimestamptz(&input.ExpiresAt),
		OrganizationID: organizationID,
		ID: id,
	})
	if err != nil {
		return Reservation{}, mapReservationTransitionError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return Reservation{}, err
	}
	return reservationFromRow(row), nil
}

func (r *Repository) Release(ctx context.Context, organizationID, id uuid.UUID) (Reservation, error) {
	row, err := r.queries.ReleaseWalletReservation(ctx, sqlc.ReleaseWalletReservationParams{OrganizationID: organizationID, ID: id})
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

func entryParams(organizationID, walletID uuid.UUID, input PostEntryInput) sqlc.CreateWalletLedgerEntryParams {
	return sqlc.CreateWalletLedgerEntryParams{EntryType: string(input.Type), AmountMinor: input.AmountMinor, SourceType: input.SourceType,
		SourceID: input.SourceID, IdempotencyKey: input.IdempotencyKey, Metadata: input.Metadata,
		OccurredAt: pgconv.NullableTimestamptz(input.OccurredAt), WalletID: walletID, OrganizationID: organizationID}
}

func walletFromRow(row sqlc.Wallet) Wallet {
	return Wallet{ID: row.ID, OrganizationID: row.OrganizationID, Currency: row.Currency, Status: Status(row.Status),
		CreatedAt: pgconv.TimestamptzToTime(row.CreatedAt), UpdatedAt: pgconv.TimestamptzToTime(row.UpdatedAt)}
}

func ledgerEntryFromRow(row sqlc.WalletLedgerEntry) LedgerEntry {
	return LedgerEntry{ID: row.ID, WalletID: row.WalletID, OrganizationID: row.OrganizationID, Type: EntryType(row.EntryType),
		AmountMinor: row.AmountMinor, SourceType: row.SourceType, SourceID: row.SourceID, IdempotencyKey: row.IdempotencyKey,
		Metadata: row.Metadata, OccurredAt: pgconv.TimestamptzToTime(row.OccurredAt), CreatedAt: pgconv.TimestamptzToTime(row.CreatedAt)}
}

func reservationFromRow(row sqlc.WalletReservation) Reservation {
	return Reservation{ID: row.ID, WalletID: row.WalletID, OrganizationID: row.OrganizationID, AmountMinor: row.AmountMinor,
		CapturedAmountMinor: row.CapturedAmountMinor, OperationType: row.OperationType, OperationID: row.OperationID,
		Status: ReservationStatus(row.Status), ExpiresAt: pgconv.TimestamptzToTime(row.ExpiresAt), CapturedAt: pgconv.TimestamptzToTimePtr(row.CapturedAt),
		ReleasedAt: pgconv.TimestamptzToTimePtr(row.ReleasedAt), ExpiredAt: pgconv.TimestamptzToTimePtr(row.ExpiredAt),
		CreatedAt: pgconv.TimestamptzToTime(row.CreatedAt), UpdatedAt: pgconv.TimestamptzToTime(row.UpdatedAt)}
}

func mapWalletReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) { return ErrWalletNotFound }
	return err
}
func mapReservationReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) { return ErrReservationNotFound }
	return err
}
func mapReservationTransitionError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) { return ErrInvalidReservationState }
	return mapWalletWriteError(err)
}
func mapWalletWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "uq_wallets_organization_currency": return ErrWalletExists
		case "uq_wallet_reservations_operation": return ErrDuplicateOperation
		case "uq_wallet_ledger_idempotency": return ErrDuplicateLedgerEntry
		}
		if pgErr.Code == "23514" { return ErrInvalidMoney }
	}
	if errors.Is(err, pgx.ErrNoRows) { return ErrWalletNotFound }
	return err
}
