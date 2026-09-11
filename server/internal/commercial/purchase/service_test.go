package purchase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fulfillmentStoreStub struct {
	transitions []CheckoutStatus
	orderID     uuid.UUID
	orders      int
	credits     int
	creditErr   error
}

func (s *fulfillmentStoreStub) TransitionCheckout(_ context.Context, _ FulfillInput, status CheckoutStatus) error {
	s.transitions = append(s.transitions, status)
	return nil
}
func (s *fulfillmentStoreStub) CreateOrder(context.Context, FulfillInput) (uuid.UUID, error) {
	s.orders++
	return s.orderID, nil
}
func (s *fulfillmentStoreStub) CreditTopup(context.Context, FulfillInput) error {
	s.credits++
	return s.creditErr
}

func TestFulfillTopupCompletesCheckoutBeforeOrderAndCredit(t *testing.T) {
	store := &fulfillmentStoreStub{orderID: uuid.New()}
	walletID := uuid.New()
	result, err := NewService().Fulfill(context.Background(), store, FulfillInput{WalletID: &walletID, Type: CheckoutWalletTopup})
	if err != nil {
		t.Fatalf("Fulfill() error = %v", err)
	}
	if !result.Applied || result.OrderID != store.orderID || store.orders != 1 || store.credits != 1 || len(store.transitions) != 1 || store.transitions[0] != CheckoutSucceeded {
		t.Fatalf("unexpected fulfillment: result=%+v store=%+v", result, store)
	}
}

func TestFulfillRollsErrorBackToTransactionOwner(t *testing.T) {
	failure := errors.New("credit failed")
	store := &fulfillmentStoreStub{orderID: uuid.New(), creditErr: failure}
	walletID := uuid.New()
	_, err := NewService().Fulfill(context.Background(), store, FulfillInput{WalletID: &walletID, Type: CheckoutWalletTopup})
	if !errors.Is(err, failure) {
		t.Fatalf("Fulfill() error = %v, want %v", err, failure)
	}
}

func TestFailedPaymentOnlyTerminatesCheckout(t *testing.T) {
	store := &fulfillmentStoreStub{}
	if err := NewService().Fail(context.Background(), store, FulfillInput{}, CheckoutFailed); err != nil {
		t.Fatalf("Fail() error = %v", err)
	}
	if store.orders != 0 || store.credits != 0 || len(store.transitions) != 1 || store.transitions[0] != CheckoutFailed {
		t.Fatalf("failed payment produced fulfillment: %+v", store)
	}
}
