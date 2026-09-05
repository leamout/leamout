package usage_rating

import "time"

// Rate is a time-bounded monetary rule applied to measured usage.
type Rate struct {
	ID             string
	MeterID        string
	Currency       string
	UnitPriceMinor int64
	EffectiveFrom  time.Time
	EffectiveUntil *time.Time
}

// RatedUsage snapshots the rate and amount applied to one usage event.
type RatedUsage struct {
	UsageEventID   string
	RateID         string
	BillableUnits  int64
	UnitPriceMinor int64
	AmountMinor    int64
	Currency       string
}
