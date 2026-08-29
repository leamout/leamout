package licensing

import "context"

// Repository persists commercial licenses and their lifecycle state.
type Repository interface {
	Get(context.Context, string) (License, error)
	Create(context.Context, License) (License, error)
	Update(context.Context, License) (License, error)
}
