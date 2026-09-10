package payments

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Status describes the reconciliation state of a payment.
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

// Payment is provider-independent commercial payment state owned by Leamout.
type Payment struct {
	ID              uuid.UUID
	CheckoutOrderID uuid.UUID
	OrganizationID  uuid.UUID
	Provider        string
	ProviderID      *string
	Status          Status
	AmountMinor     int64
	Currency        string
	PaidAt          *time.Time
	Metadata        json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateInput struct {
	CheckoutOrderID uuid.UUID
	ProviderID      *string
	Status          Status
	AmountMinor     int64
	Currency        string
	Metadata        json.RawMessage
}
