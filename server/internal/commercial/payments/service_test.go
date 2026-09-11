package payments

import (
	"context"
	"errors"
	"testing"

	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

type eventStoreStub struct {
	event       paymentprovider.Event
	replay      paymentprovider.Event
	replayError error
}

func (s *eventStoreStub) checkProviderEventReplay(_ context.Context, event paymentprovider.Event) error {
	s.replay = event
	return s.replayError
}

func (s *eventStoreStub) processProviderEvent(_ context.Context, event paymentprovider.Event) (Settlement, error) {
	s.event = event
	return Settlement{Applied: true}, nil
}

func TestProcessProviderEventIgnoresUnsupportedEventBeforePersistence(t *testing.T) {
	store := &eventStoreStub{}
	service := NewService(store)
	_, err := service.ProcessProviderEvent(context.Background(), paymentprovider.Event{
		Provider: "stripe", ProviderEventID: "evt_1", Type: "customer.created",
		Payment: paymentprovider.Payment{Reference: "topup.1"}, Raw: []byte(`{"id":"evt_1"}`),
	})
	if err != nil {
		t.Fatalf("ProcessProviderEvent() error = %v", err)
	}
	if store.event.ProviderEventID != "" || store.replay.ProviderEventID != "evt_1" {
		t.Fatal("unsupported event reached persistence")
	}
}

func TestProcessProviderEventRejectsConflictingIgnoredReplay(t *testing.T) {
	store := &eventStoreStub{replayError: ErrPaymentMismatch}
	service := NewService(store)
	_, err := service.ProcessProviderEvent(context.Background(), paymentprovider.Event{
		Provider: "stripe", ProviderEventID: "evt_1", Type: "customer.created",
		Raw: []byte(`{"id":"evt_1","type":"customer.created"}`),
	})
	if !errors.Is(err, ErrPaymentMismatch) {
		t.Fatalf("ProcessProviderEvent() error = %v, want %v", err, ErrPaymentMismatch)
	}
}

func TestProcessProviderEventRejectsMalformedEvent(t *testing.T) {
	service := NewService(&eventStoreStub{})
	_, err := service.ProcessProviderEvent(context.Background(), paymentprovider.Event{Provider: "unknown"})
	if !errors.Is(err, ErrPaymentMismatch) {
		t.Fatalf("ProcessProviderEvent() error = %v, want %v", err, ErrPaymentMismatch)
	}
}

func TestProcessProviderEventPassesNormalizedPaymentEvent(t *testing.T) {
	store := &eventStoreStub{}
	service := NewService(store)
	event := paymentprovider.Event{
		Provider: "paystack", ProviderEventID: "evt_2", Type: "charge.success",
		Payment: paymentprovider.Payment{Reference: "topup.2", Status: paymentprovider.StatusSucceeded}, Raw: []byte(`{"event":"charge.success"}`),
	}
	result, err := service.ProcessProviderEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("ProcessProviderEvent() error = %v", err)
	}
	if !result.Applied || store.event.ProviderEventID != event.ProviderEventID {
		t.Fatalf("event was not processed: result=%+v event=%+v", result, store.event)
	}
}

func TestProcessProviderEventRejectsTypeStatusMismatch(t *testing.T) {
	store := &eventStoreStub{}
	service := NewService(store)
	_, err := service.ProcessProviderEvent(context.Background(), paymentprovider.Event{
		Provider: "paystack", ProviderEventID: "evt_3", Type: "charge.success",
		Payment: paymentprovider.Payment{Reference: "topup.3", Status: paymentprovider.StatusFailed},
		Raw:     []byte(`{"event":"charge.success"}`),
	})
	if !errors.Is(err, ErrPaymentMismatch) {
		t.Fatalf("ProcessProviderEvent() error = %v, want %v", err, ErrPaymentMismatch)
	}
	if store.event.ProviderEventID != "" {
		t.Fatal("mismatched event reached persistence")
	}
}
