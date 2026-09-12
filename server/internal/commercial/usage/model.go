package usage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

var (
	ErrInvalidEvent  = apperror.NewBadRequest("invalid usage event")
	ErrEventConflict = apperror.NewConflict("usage event idempotency key conflicts with an existing event")
)

// Event is an immutable, organization-scoped observation of measured use.
// Recording an event does not by itself make the usage billable.
type Event struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
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
	MeterID        uuid.UUID
	Quantity       int64
	SourceType     string
	SourceID       string
	IdempotencyKey string
	Dimensions     json.RawMessage
	OccurredAt     time.Time
}

type RecordResult struct {
	Event    Event
	Replayed bool
}
