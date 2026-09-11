package purchase

import (
	"context"

	"github.com/google/uuid"
)

// FulfillmentStore is a transaction-scoped persistence boundary supplied by
// Payments. Implementations must apply every method in the same transaction.
type FulfillmentStore interface {
	TransitionCheckout(context.Context, FulfillInput, CheckoutStatus) error
	CreateOrder(context.Context, FulfillInput) (uuid.UUID, error)
	CreditTopup(context.Context, FulfillInput) error
}

// Service owns checkout-to-order fulfillment semantics. It does not interpret
// payment-provider payloads or authorize wallet operations.
type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) Fulfill(ctx context.Context, store FulfillmentStore, input FulfillInput) (Result, error) {
	if err := store.TransitionCheckout(ctx, input, CheckoutSucceeded); err != nil {
		return Result{}, err
	}
	orderID, err := store.CreateOrder(ctx, input)
	if err != nil {
		return Result{}, err
	}
	if input.Type == CheckoutWalletTopup {
		if err = store.CreditTopup(ctx, input); err != nil {
			return Result{}, err
		}
	}
	return Result{OrderID: orderID, Applied: true}, nil
}

func (s *Service) Fail(ctx context.Context, store FulfillmentStore, input FulfillInput, status CheckoutStatus) error {
	if status != CheckoutFailed && status != CheckoutCancelled {
		return nil
	}
	return store.TransitionCheckout(ctx, input, status)
}

// EnsureOrder repairs durable purchase history without repeating target
// fulfillment when payment settlement was already applied.
func (s *Service) EnsureOrder(ctx context.Context, store FulfillmentStore, input FulfillInput) (uuid.UUID, error) {
	return store.CreateOrder(ctx, input)
}
