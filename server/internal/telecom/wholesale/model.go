package wholesale

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCallNotFound = errors.New("managed call not found for provider CDR")
	ErrCDRConflict  = errors.New("provider CDR identity conflicts with existing record")
	ErrInvalidCDR   = errors.New("invalid provider CDR")
	ErrCDRRouteNotFound = errors.New("provider CDR route not found")
)

type CDR struct {
	Provider            string         `json:"provider"`
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

type NormalizedCDR struct {
	ProviderRecordID string
	SIPCallID        string
	StartedAt        time.Time
	DurationSeconds  int64
	Currency         string
	CostMicros       int64
	Raw              map[string]any
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

type CDRPageSource interface {
	PollCDRs(context.Context, string, time.Time, int, int) (json.RawMessage, int, error)
}

type CDRNormalizer interface {
	NormalizeCDRs(context.Context, string, json.RawMessage) ([]NormalizedCDR, error)
}

type CDRPollCursor struct {
	Provider      string
	Direction     string
	WindowDate    time.Time
	Page          int
	AttemptCount  int
	NextAttemptAt time.Time
}

type CDRPageWork struct {
	ID              uuid.UUID
	Provider        string
	Direction       string
	Raw             json.RawMessage
	ProcessAttempts int
}

type CDRPollStore interface {
	Cursor(context.Context, string, string, time.Time) (CDRPollCursor, error)
	StorePageAndAdvance(context.Context, CDRPollCursor, json.RawMessage, int, time.Time, int, time.Time) error
	Fail(context.Context, CDRPollCursor, error, time.Time) error
}

type CDRProcessingStore interface {
	ClaimCDRPages(context.Context, int32) ([]CDRPageWork, error)
	ResolveCDRRoute(context.Context, string, string) (uuid.UUID, error)
	MarkCDRPageProcessed(context.Context, uuid.UUID) error
	FailCDRPage(context.Context, uuid.UUID, error, time.Time) error
}

type CDRPollJobConfig struct {
	Provider        string
	Directions      []string
	PerPage         int
	TickInterval    time.Duration
	CurrentDayDelay time.Duration
	RetryBase       time.Duration
	RetryMax        time.Duration
	InitialLookback time.Duration
}

type CDRReconciliationJobConfig struct {
	BatchSize    int32
	TickInterval time.Duration
	RetryBase    time.Duration
	RetryMax     time.Duration
}
