package subscriptions

import "context"

// Repository persists commercial subscription state.
type Repository interface {
	Get(context.Context, string) (Subscription, error)
	Create(context.Context, Subscription) (Subscription, error)
	Update(context.Context, Subscription) (Subscription, error)
}
