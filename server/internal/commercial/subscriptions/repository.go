package subscriptions

import "context"

// Repository persists subscription state within an organization boundary.
type Repository interface {
	Get(context.Context, string, string) (Subscription, error)
	Current(context.Context, string) (Subscription, error)
	Create(context.Context, Subscription) (Subscription, error)
	Update(context.Context, Subscription) (Subscription, error)
}
