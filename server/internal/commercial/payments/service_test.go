package payments

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

type eventStoreStub struct {
	event       ProviderEvent
	replay      ProviderEvent
	replayError error
}

func (s *eventStoreStub) checkProviderEventReplay(_ context.Context, event ProviderEvent) error {
	s.replay = event
	return s.replayError
}

func (s *eventStoreStub) processProviderEvent(_ context.Context, event ProviderEvent) (Settlement, error) {
	s.event = event
	return Settlement{Applied: true}, nil
}

type paymentStoreStub struct {
	payment Payment
}

func (s *paymentStoreStub) Create(_ context.Context, organizationID uuid.UUID, provider string, input CreateInput) (Payment, error) {
	result := Payment{
		ID:             uuid.New(),
		CheckoutID:     input.CheckoutID,
		OrganizationID: organizationID,
		Provider:       provider,
		Status:         input.Status,
		AmountMinor:    input.AmountMinor,
		Currency:       input.Currency,
	}
	s.payment = result
	return result, nil
}

func (s *paymentStoreStub) GetByCheckout(context.Context, uuid.UUID, uuid.UUID) (Payment, error) {
	if s.payment.ID == uuid.Nil {
		return Payment{}, ErrPaymentNotFound
	}
	return s.payment, nil
}

func (s *paymentStoreStub) SetProviderID(_ context.Context, _, _ uuid.UUID, providerID string, status Status) (Payment, error) {
	s.payment.ProviderID = &providerID
	s.payment.Status = status
	return s.payment, nil
}

func (s *paymentStoreStub) UpdateStatus(_ context.Context, _, _ uuid.UUID, status Status, _ *time.Time) (Payment, error) {
	s.payment.Status = status
	return s.payment, nil
}

type providerStub struct {
	session CheckoutSession
}

func (s providerStub) CreateCheckout(context.Context, CheckoutRequest) (CheckoutSession, error) {
	return s.session, nil
}

func (providerStub) GetPayment(context.Context, string) (ProviderPayment, error) {
	return ProviderPayment{}, nil
}

func (providerStub) ParseWebhook([]byte, http.Header) (ProviderEvent, error) {
	return ProviderEvent{}, nil
}

func TestStartOwnsPaymentPersistenceAndProviderSession(t *testing.T) {
	store := &paymentStoreStub{}
	registry := NewProviderRegistry(map[string]Provider{
		"stripe": providerStub{session: CheckoutSession{
			Provider:   "stripe",
			ProviderID: "cs_123",
			Reference:  "checkout.1",
			Status:     StatusProcessing,
		}},
	})
	service := newService(store, &eventStoreStub{}, registry)

	result, err := service.Start(t.Context(), uuid.New(), StartInput{
		CheckoutID:  uuid.New(),
		Provider:    "stripe",
		Reference:   "checkout.1",
		AmountMinor: 2500,
		Currency:    "USD",
		Email:       "billing@example.com",
	})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if result.Payment.Status != StatusProcessing || result.Payment.ProviderID == nil ||
		*result.Payment.ProviderID != "cs_123" {
		t.Fatalf("payment = %+v", result.Payment)
	}
}

func TestProcessProviderEventIgnoresUnsupportedEventBeforePersistence(t *testing.T) {
	store := &eventStoreStub{}
	service := newService(nil, store, nil)
	_, err := service.ProcessProviderEvent(context.Background(), ProviderEvent{
		Provider: "stripe", ProviderEventID: "evt_1", Type: "customer.created",
		Payment: ProviderPayment{Reference: "checkout.1"}, Raw: []byte(`{"id":"evt_1"}`),
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
	service := newService(nil, store, nil)
	_, err := service.ProcessProviderEvent(context.Background(), ProviderEvent{
		Provider: "stripe", ProviderEventID: "evt_1", Type: "customer.created",
		Raw: []byte(`{"id":"evt_1","type":"customer.created"}`),
	})
	if !errors.Is(err, ErrPaymentMismatch) {
		t.Fatalf("ProcessProviderEvent() error = %v, want %v", err, ErrPaymentMismatch)
	}
}

func TestProcessProviderEventRejectsTypeStatusMismatch(t *testing.T) {
	store := &eventStoreStub{}
	service := newService(nil, store, nil)
	_, err := service.ProcessProviderEvent(context.Background(), ProviderEvent{
		Provider: "paystack", ProviderEventID: "evt_3", Type: "charge.success",
		Payment: ProviderPayment{Reference: "checkout.3", Status: StatusFailed},
		Raw:     []byte(`{"event":"charge.success"}`),
	})
	if !errors.Is(err, ErrPaymentMismatch) {
		t.Fatalf("ProcessProviderEvent() error = %v, want %v", err, ErrPaymentMismatch)
	}
}
