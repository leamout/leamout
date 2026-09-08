package wholesale

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, queries: sqlc.New(db)}
}

func (r *Repository) Reconcile(ctx context.Context, cdr CDR) (Result, error) {
	raw, err := json.Marshal(cdr.Raw)
	if err != nil {
		return Result{}, err
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	sipCallID := cdr.SIPCallID
	startedAt := pgtype.Timestamptz{Time: cdr.StartedAt, Valid: true}

	inserted := true
	record, err := queries.InsertProviderCDR(ctx, sqlc.InsertProviderCDRParams{
		CarrierProviderID:   cdr.CarrierProviderID,
		CarrierConnectionID: cdr.CarrierConnectionID,
		ProviderRecordID:    cdr.ProviderRecordID,
		Direction:           cdr.Direction,
		SipCallID:           &sipCallID,
		StartedAt:           startedAt,
		DurationSeconds:     cdr.DurationSeconds,
		Currency:            cdr.Currency,
		CostMicros:          cdr.CostMicros,
		Raw:                 raw,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		inserted = false
	} else if err != nil {
		return Result{}, err
	}
	record, err = queries.GetProviderCDRForUpdate(ctx, sqlc.GetProviderCDRForUpdateParams{
		CarrierProviderID: cdr.CarrierProviderID,
		Direction:         cdr.Direction,
		ProviderRecordID:  cdr.ProviderRecordID,
	})
	if err != nil {
		return Result{}, err
	}
	if !sameCDR(record, cdr) {
		return Result{}, ErrCDRConflict
	}
	if record.CallID != nil && record.OrganizationID != nil {
		charge, err := queries.GetWholesaleChargeByProviderCDR(ctx, record.ID)
		if err != nil {
			return Result{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return Result{}, err
		}
		return result(record.ID, *record.CallID, *record.OrganizationID, charge, true), nil
	}

	call, err := queries.FindManagedCallForProviderCDR(ctx, sqlc.FindManagedCallForProviderCDRParams{
		SipCallID:           &sipCallID,
		CarrierConnectionID: &cdr.CarrierConnectionID,
		CarrierProviderID:   cdr.CarrierProviderID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return Result{}, commitErr
		}
		return Result{}, ErrCallNotFound
	}
	if err != nil {
		return Result{}, err
	}
	if _, err := queries.MarkProviderCDRReconciled(ctx, sqlc.MarkProviderCDRReconciledParams{
		CallID:         &call.ID,
		OrganizationID: &call.OrganizationID,
		ID:             record.ID,
	}); err != nil {
		return Result{}, err
	}
	charge, err := queries.CreateWholesaleCharge(ctx, sqlc.CreateWholesaleChargeParams{
		ProviderCdrID:  record.ID,
		OrganizationID: call.OrganizationID,
		CallID:         call.ID,
		AmountMicros:   cdr.CostMicros,
		Currency:       cdr.Currency,
		OccurredAt:     startedAt,
	})
	if err != nil {
		return Result{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Result{}, err
	}
	return result(record.ID, call.ID, call.OrganizationID, charge, !inserted), nil
}

func (r *Repository) Cursor(ctx context.Context, provider, direction string, startDate time.Time) (CDRPollCursor, error) {
	row, err := r.queries.EnsureProviderCDRPollCursor(ctx, sqlc.EnsureProviderCDRPollCursorParams{
		Provider:   provider,
		Direction:  direction,
		WindowDate: pgtype.Date{Time: utcDate(startDate), Valid: true},
	})
	if err != nil {
		return CDRPollCursor{}, err
	}
	return CDRPollCursor{
		Provider:      row.Provider,
		Direction:     row.Direction,
		WindowDate:    row.WindowDate.Time,
		Page:          int(row.Page),
		AttemptCount:  int(row.AttemptCount),
		NextAttemptAt: row.NextAttemptAt.Time,
	}, nil
}

func (r *Repository) StorePageAndAdvance(ctx context.Context, cursor CDRPollCursor, raw json.RawMessage, count int, nextDate time.Time, nextPage int, nextAttemptAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	hash := sha256.Sum256(raw)

	if _, err := queries.InsertProviderCDRPage(ctx, sqlc.InsertProviderCDRPageParams{
		Provider:      cursor.Provider,
		Direction:     cursor.Direction,
		WindowDate:    pgtype.Date{Time: utcDate(cursor.WindowDate), Valid: true},
		Page:          int32(cursor.Page),
		RecordCount:   int32(count),
		PayloadSha256: hex.EncodeToString(hash[:]),
		Raw:           raw,
	}); err != nil {
		return err
	}
	if _, err := queries.AdvanceProviderCDRPollCursor(ctx, sqlc.AdvanceProviderCDRPollCursorParams{
		Provider:      cursor.Provider,
		Direction:     cursor.Direction,
		WindowDate:    pgtype.Date{Time: utcDate(nextDate), Valid: true},
		Page:          int32(nextPage),
		NextAttemptAt: pgtype.Timestamptz{Time: nextAttemptAt, Valid: true},
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) Fail(ctx context.Context, cursor CDRPollCursor, pollErr error, nextAttemptAt time.Time) error {
	message := strings.TrimSpace(pollErr.Error())
	if len(message) > 2048 {
		message = message[:2048]
	}
	_, err := r.queries.FailProviderCDRPollCursor(ctx, sqlc.FailProviderCDRPollCursorParams{
		Provider:      cursor.Provider,
		Direction:     cursor.Direction,
		NextAttemptAt: pgtype.Timestamptz{Time: nextAttemptAt, Valid: true},
		LastError:     &message,
	})
	return err
}

func sameCDR(record sqlc.ProviderCdr, cdr CDR) bool {
	var raw map[string]any
	if err := json.Unmarshal(record.Raw, &raw); err != nil {
		return false
	}
	return record.CarrierProviderID == cdr.CarrierProviderID && record.CarrierConnectionID == cdr.CarrierConnectionID &&
		record.ProviderRecordID == cdr.ProviderRecordID && record.Direction == cdr.Direction &&
		record.SipCallID != nil && *record.SipCallID == cdr.SIPCallID && record.StartedAt.Valid &&
		record.StartedAt.Time.Equal(cdr.StartedAt) && record.DurationSeconds == cdr.DurationSeconds &&
		record.Currency == cdr.Currency && record.CostMicros == cdr.CostMicros && reflect.DeepEqual(raw, cdr.Raw)
}

func result(cdrID, callID, organizationID uuid.UUID, charge sqlc.WholesaleCharge, replayed bool) Result {
	return Result{
		ProviderCDRID:  cdrID,
		CallID:         callID,
		OrganizationID: organizationID,
		ChargeID:       charge.ID,
		AmountMicros:   charge.AmountMicros,
		Currency:       charge.Currency,
		Replayed:       replayed,
	}
}
