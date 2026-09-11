package payments

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

var ErrPaymentMismatch = apperror.NewConflict("provider payment does not match checkout")

type Status string

const (
	StatusPending           Status = "pending"
	StatusProcessing        Status = "processing"
	StatusSucceeded         Status = "succeeded"
	StatusFailed            Status = "failed"
	StatusCancelled         Status = "cancelled"
	StatusRefunded          Status = "refunded"
	StatusPartiallyRefunded Status = "partially_refunded"
)

type Payment struct {
	ID             uuid.UUID
	CheckoutID     uuid.UUID
	OrganizationID uuid.UUID
	Provider       string
	ProviderID     *string
	Status         Status
	AmountMinor    int64
	Currency       string
	PaidAt         *time.Time
	Metadata       json.RawMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateInput struct {
	CheckoutID  uuid.UUID
	ProviderID  *string
	Status      Status
	AmountMinor int64
	Currency    string
	Metadata    json.RawMessage
}

// Settlement describes the durable effects of processing one provider event.
type Settlement struct {
	Applied        bool
	OrganizationID uuid.UUID
	WalletID       uuid.UUID
	PaymentID      uuid.UUID
	OrderID        uuid.UUID
	AmountMinor    int64
	SettledAt      *time.Time
}
