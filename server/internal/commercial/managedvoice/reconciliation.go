package managedvoice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leamout/leamout/internal/database/sqlc"
)

var ErrCDRNotMatched = errors.New("provider CDR does not match a managed call")

type ProviderCDR struct {
	ProviderID          uuid.UUID
	CarrierConnectionID uuid.UUID
	ProviderRecordID    string
	Direction           string
	SIPCallID           string
	StartedAt           time.Time
	DurationSeconds     int64
	Currency            string
	CostMicros          int64
	Raw                 json.RawMessage
}

type ReconciliationResult struct {
	CDRID, CallID, OrganizationID uuid.UUID
	Duplicate                     bool
}

func (r *Repository) IngestProviderCDR(ctx context.Context, input ProviderCDR) (ReconciliationResult, error) {
	if err := validateCDR(input); err != nil {
		return ReconciliationResult{}, err
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ReconciliationResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.queries.WithTx(tx)

	cdr, err := q.InsertProviderCDR(ctx, sqlc.InsertProviderCDRParams{
		CarrierProviderID:   input.ProviderID,
		CarrierConnectionID: input.CarrierConnectionID,
		ProviderRecordID:    input.ProviderRecordID,
		Direction:           input.Direction,
		SipCallID:           &input.SIPCallID,
		StartedAt:           pgtype.Timestamptz{Time: input.StartedAt, Valid: true},
		DurationSeconds:     input.DurationSeconds,
		Currency:            input.Currency,
		CostMicros:          input.CostMicros,
		Raw:                 input.Raw,
	})
	duplicate := false
	if errors.Is(err, pgx.ErrNoRows) {
		cdr, err = q.GetProviderCDRForUpdate(ctx, sqlc.GetProviderCDRForUpdateParams{
			CarrierProviderID: input.ProviderID,
			Direction:         input.Direction,
			ProviderRecordID:  input.ProviderRecordID,
		})
		if err != nil {
			return ReconciliationResult{}, err
		}
		duplicate = true
		if cdr.CallID != nil && cdr.OrganizationID != nil {
			result := ReconciliationResult{CDRID: cdr.ID, CallID: *cdr.CallID, OrganizationID: *cdr.OrganizationID, Duplicate: true}
			return result, tx.Commit(ctx)
		}
	} else if err != nil {
		return ReconciliationResult{}, err
	}

	call, err := q.FindManagedCallForProviderCDR(ctx, sqlc.FindManagedCallForProviderCDRParams{
		SipCallID:           &input.SIPCallID,
		CarrierConnectionID: &input.CarrierConnectionID,
		CarrierProviderID:   input.ProviderID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return ReconciliationResult{}, commitErr
		}
		return ReconciliationResult{CDRID: cdr.ID, Duplicate: duplicate}, ErrCDRNotMatched
	}
	if err != nil {
		return ReconciliationResult{}, err
	}

	if _, err = q.MarkProviderCDRReconciled(ctx, sqlc.MarkProviderCDRReconciledParams{
		ID:             cdr.ID,
		CallID:         &call.ID,
		OrganizationID: &call.OrganizationID,
	}); err != nil {
		return ReconciliationResult{}, err
	}

	if _, err = q.CreateWholesaleCharge(ctx, sqlc.CreateWholesaleChargeParams{
		ProviderCdrID:  cdr.ID,
		OrganizationID: call.OrganizationID,
		CallID:         call.ID,
		AmountMicros:   input.CostMicros,
		Currency:       input.Currency,
		OccurredAt:     pgtype.Timestamptz{Time: input.StartedAt, Valid: true},
	}); err != nil {
		return ReconciliationResult{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return ReconciliationResult{}, err
	}
	return ReconciliationResult{CDRID: cdr.ID, CallID: call.ID, OrganizationID: call.OrganizationID, Duplicate: duplicate}, nil
}

func validateCDR(in ProviderCDR) error {
	if in.ProviderID == uuid.Nil || in.CarrierConnectionID == uuid.Nil || strings.TrimSpace(in.ProviderRecordID) == "" || strings.TrimSpace(in.SIPCallID) == "" || in.StartedAt.IsZero() {
		return fmt.Errorf("provider, connection, record, SIP call ID, and start time are required")
	}
	if in.Direction != "termination" && in.Direction != "origination" {
		return fmt.Errorf("invalid CDR direction")
	}
	if len(in.Currency) != 3 || strings.ToUpper(in.Currency) != in.Currency {
		return fmt.Errorf("invalid CDR currency")
	}
	if in.DurationSeconds < 0 || in.CostMicros < 0 {
		return fmt.Errorf("CDR duration and cost must be non-negative")
	}
	if !json.Valid(in.Raw) {
		return fmt.Errorf("raw CDR must be valid JSON")
	}
	var obj map[string]any
	if json.Unmarshal(in.Raw, &obj) != nil || obj == nil {
		return fmt.Errorf("raw CDR must be a JSON object")
	}
	return nil
}
