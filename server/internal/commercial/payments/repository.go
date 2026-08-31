package payments

import "context"

// Repository persists payment reconciliation state within an organization boundary.
type Repository interface {
	Get(context.Context, string, string) (Payment, error)
	Upsert(context.Context, Payment) (Payment, error)
}
