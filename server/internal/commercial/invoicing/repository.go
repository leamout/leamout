package invoicing

import "context"

// Repository persists invoices and immutable invoice item snapshots within an organization boundary.
type Repository interface {
	Get(context.Context, string, string) (Invoice, error)
	Create(context.Context, Invoice) (Invoice, error)
	AddItem(context.Context, string, Item) (Item, error)
}
