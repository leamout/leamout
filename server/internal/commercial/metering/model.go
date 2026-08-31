package metering

import "time"

// Meter identifies a commercial quantity that can be measured from authoritative domain events.
type Meter struct {
	ID     string
	Key    string
	Unit   string
	Active bool
}

// UsageEvent is an idempotent commercial record of measured organization usage.
type UsageEvent struct {
	ID             string
	OrganizationID string
	MeterID        string
	SourceEventID  string
	Quantity       int64
	OccurredAt     time.Time
}
