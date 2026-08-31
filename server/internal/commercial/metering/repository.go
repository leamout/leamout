package metering

import "context"

// Repository persists meters and idempotent usage events.
type Repository interface {
	GetMeter(context.Context, string) (Meter, error)
	CreateUsageEvent(context.Context, UsageEvent) (UsageEvent, error)
}
