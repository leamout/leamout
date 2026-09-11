package checkout

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type checkoutStoreStub struct {
	creates     int
	transitions int
}

func (s *checkoutStoreStub) Create(_ context.Context, organizationID uuid.UUID, input CreateInput) (Checkout, error) {
	s.creates++
	return Checkout{OrganizationID: organizationID, Type: input.Type}, nil
}
func (s *checkoutStoreStub) Get(context.Context, uuid.UUID, uuid.UUID) (Checkout, error) {
	return Checkout{}, nil
}
func (s *checkoutStoreStub) GetByReference(context.Context, string) (Checkout, error) {
	return Checkout{}, nil
}
func (s *checkoutStoreStub) Transition(context.Context, uuid.UUID, uuid.UUID, Transition) (Checkout, error) {
	s.transitions++
	return Checkout{}, nil
}
func (s *checkoutStoreStub) ClaimRefresh(context.Context, uuid.UUID, uuid.UUID, time.Time) (Checkout, error) {
	return Checkout{}, nil
}
func (s *checkoutStoreStub) Expire(context.Context) ([]Checkout, error) { return nil, nil }

func TestServiceRejectsInvalidCheckoutBeforePersistence(t *testing.T) {
	store := &checkoutStoreStub{}
	service := NewService(store)
	_, err := service.Create(context.Background(), uuid.New(), CreateInput{})
	if !errors.Is(err, ErrInvalidCheckout) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidCheckout)
	}
	if store.creates != 0 {
		t.Fatal("invalid checkout reached repository")
	}
}

func TestServiceRejectsInvalidTransitionBeforePersistence(t *testing.T) {
	store := &checkoutStoreStub{}
	service := NewService(store)
	_, err := service.Transition(context.Background(), uuid.New(), uuid.New(), Transition{Expected: StatusSucceeded, Status: StatusProcessing})
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("Transition() error = %v, want %v", err, ErrInvalidTransition)
	}
	if store.transitions != 0 {
		t.Fatal("invalid transition reached repository")
	}
}
