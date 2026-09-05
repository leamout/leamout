package usage_rating

import (
	"context"
	"time"
)

// Repository resolves durable rates applicable to organization usage.
type Repository interface {
	Resolve(context.Context, string, string, time.Time) (Rate, error)
}
