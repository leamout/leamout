package catalog

import "context"

// Repository owns durable product and plan persistence.
type Repository interface {
	GetProduct(context.Context, string) (Product, error)
	GetPlan(context.Context, string) (Plan, error)
}
