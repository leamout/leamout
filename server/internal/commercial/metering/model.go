package metering

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

var (
	ErrMeterNotFound       = apperror.NewNotFound("commercial meter not found")
	ErrInvalidUsageEvent   = apperror.NewBadRequest("invalid usage event")
	ErrUsageEventConflict  = apperror.NewConflict("usage event idempotency key conflicts with an existing event")
)

// Meter identifies a quantity measured from authoritative domain events.
// Recording usage does not by itself make the usage billable.
type Meter struct {
	ID        uuid.UUID
	Key       string
	Name      string
	Unit      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UsageEvent is an immutable, organization-scoped observation of measured use.
type UsageEvent struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	SubscriptionID *uuid.UUID
	MeterID        uuid.UUID
	Quantity       int64
	SourceType     string
	SourceID       string
	IdempotencyKey string
	Dimensions     json.RawMessage
	OccurredAt     time.Time
	CreatedAt      time.Time
}

type RecordInput struct {
	SubscriptionID *uuid.UUID
	MeterID        uuid.UUID
	Quantity       int64
	SourceType     string
	SourceID       string
	IdempotencyKey string
	Dimensions     json.RawMessage
	OccurredAt     time.Time
}

type RecordResult struct {
	Event    UsageEvent
	Replayed bool
}
