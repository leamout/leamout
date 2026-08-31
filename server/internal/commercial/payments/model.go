package payments

import "time"

// Status describes the reconciliation state of a payment.
type Status string

const (
	StatusPending   Status = "pending"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

// Payment is provider-independent commercial payment state owned by Leamout.
type Payment struct {
	ID             string
	OrganizationID string
	InvoiceID      string
	Provider       string
	ProviderID     string
	Status         Status
	AmountMinor    int64
	Currency       string
	OccurredAt     time.Time
}
