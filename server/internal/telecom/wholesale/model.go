package wholesale

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCallNotFound = errors.New("managed call not found for provider CDR")
	ErrCDRConflict  = errors.New("provider CDR identity conflicts with existing record")
	ErrInvalidCDR   = errors.New("invalid provider CDR")
)

type CDR struct {
	CarrierProviderID   uuid.UUID      `json:"carrier_provider_id"`
	CarrierConnectionID uuid.UUID      `json:"carrier_connection_id"`
	ProviderRecordID    string         `json:"provider_record_id"`
	Direction           string         `json:"direction"`
	SIPCallID           string         `json:"sip_call_id"`
	StartedAt           time.Time      `json:"started_at"`
	DurationSeconds     int64          `json:"duration_seconds"`
	Currency            string         `json:"currency"`
	CostMicros          int64          `json:"cost_micros"`
	Raw                 map[string]any `json:"raw"`
}

type Result struct {
	ProviderCDRID  uuid.UUID `json:"provider_cdr_id"`
	CallID         uuid.UUID `json:"call_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	ChargeID       uuid.UUID `json:"wholesale_charge_id"`
	AmountMicros   int64     `json:"amount_micros"`
	Currency       string    `json:"currency"`
	Replayed       bool      `json:"replayed"`
}
