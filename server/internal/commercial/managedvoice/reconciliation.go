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

// IngestProviderCDR stores the immutable provider record and reconciles it to
// an outbound call atomically. Re-delivery returns the original result without
// creating a second wholesale charge.
func (r *Repository) IngestProviderCDR(ctx context.Context, input ProviderCDR) (ReconciliationResult, error) {
	if err := validateCDR(input); err != nil {
		return ReconciliationResult{}, err
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ReconciliationResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var cdrID uuid.UUID
	duplicate := false
	err = tx.QueryRow(ctx, `INSERT INTO provider_cdrs (carrier_provider_id,carrier_connection_id,provider_record_id,direction,sip_call_id,started_at,duration_seconds,currency,cost_micros,raw) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT (carrier_provider_id,direction,provider_record_id) DO NOTHING RETURNING id`, input.ProviderID, input.CarrierConnectionID, input.ProviderRecordID, input.Direction, input.SIPCallID, input.StartedAt, input.DurationSeconds, input.Currency, input.CostMicros, input.Raw).Scan(&cdrID)
	if errors.Is(err, pgx.ErrNoRows) {
		var existingCallID, existingOrganizationID *uuid.UUID
		err = tx.QueryRow(ctx, `SELECT id,call_id,organization_id FROM provider_cdrs WHERE carrier_provider_id=$1 AND direction=$2 AND provider_record_id=$3 FOR UPDATE`, input.ProviderID, input.Direction, input.ProviderRecordID).Scan(&cdrID, &existingCallID, &existingOrganizationID)
		if err != nil {
			return ReconciliationResult{}, err
		}
		duplicate = true
		if existingCallID != nil && existingOrganizationID != nil {
			result := ReconciliationResult{CDRID: cdrID, CallID: *existingCallID, OrganizationID: *existingOrganizationID, Duplicate: true}
			return result, tx.Commit(ctx)
		}
	}
	if err != nil {
		return ReconciliationResult{}, err
	}
	var callID, organizationID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT c.id,c.organization_id FROM calls c JOIN carrier_connections cc ON cc.id=c.carrier_connection_id WHERE c.sip_call_id=$1 AND c.direction='outbound' AND c.carrier_connection_id=$2 AND cc.scope='platform' AND cc.provider_id=$3`, input.SIPCallID, input.CarrierConnectionID, input.ProviderID).Scan(&callID, &organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return ReconciliationResult{}, commitErr
		}
		return ReconciliationResult{CDRID: cdrID, Duplicate: duplicate}, ErrCDRNotMatched
	}
	if err != nil {
		return ReconciliationResult{}, err
	}
	_, err = tx.Exec(ctx, `UPDATE provider_cdrs SET call_id=$2,organization_id=$3,reconciled_at=now() WHERE id=$1; INSERT INTO wholesale_charges (provider_cdr_id,organization_id,call_id,amount_micros,currency,occurred_at) VALUES ($1,$3,$2,$4,$5,$6)`, cdrID, callID, organizationID, input.CostMicros, input.Currency, input.StartedAt)
	if err != nil {
		return ReconciliationResult{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return ReconciliationResult{}, err
	}
	return ReconciliationResult{CDRID: cdrID, CallID: callID, OrganizationID: organizationID, Duplicate: duplicate}, nil
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
