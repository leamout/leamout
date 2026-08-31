package invoicing

import "time"

// Status describes the lifecycle of an invoice.
type Status string

const (
	StatusDraft Status = "draft"
	StatusOpen  Status = "open"
	StatusPaid  Status = "paid"
	StatusVoid  Status = "void"
)

// Invoice is the durable commercial statement for an organization and billing period.
type Invoice struct {
	ID             string
	OrganizationID string
	Status         Status
	Currency       string
	PeriodStartsAt time.Time
	PeriodEndsAt   time.Time
	TotalMinor     int64
}

// Item snapshots a rated charge on an invoice so historical pricing is immutable.
type Item struct {
	ID             string
	InvoiceID      string
	UsageEventID   *string
	Description    string
	Quantity       int64
	UnitPriceMinor int64
	AmountMinor    int64
	Currency       string
}
