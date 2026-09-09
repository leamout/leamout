package payments

import "time"

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
	ID              string
	CheckoutOrderID string
	OrganizationID  string
	Provider        string
	ProviderID      string
	Status          Status
	AmountMinor     int64
	Currency        string
	OccurredAt      time.Time
}
