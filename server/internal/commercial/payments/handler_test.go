package payments

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type webhookProviderStub struct{ event ProviderEvent }

func (s webhookProviderStub) CreateCheckout(context.Context, CheckoutRequest) (CheckoutSession, error) {
	return CheckoutSession{}, nil
}
func (s webhookProviderStub) GetPayment(context.Context, string) (ProviderPayment, error) {
	return ProviderPayment{}, nil
}
func (s webhookProviderStub) ParseWebhook([]byte, http.Header) (ProviderEvent, error) {
	return s.event, nil
}

func TestWebhookAuthenticatesAdapterBeforeCommercialProcessing(t *testing.T) {
	event := ProviderEvent{Provider: "stripe", ProviderEventID: "evt_1", Type: "checkout.session.completed", Payment: ProviderPayment{Reference: "topup.1", Status: StatusSucceeded}, Raw: []byte(`{"id":"evt_1"}`)}
	store := &eventStoreStub{}
	registry := NewProviderRegistry(map[string]Provider{"stripe": webhookProviderStub{event: event}})
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(NewService(store), registry))
	request := httptest.NewRequest(http.MethodPost, "/payment-webhooks/stripe", bytes.NewBufferString(`{}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if store.event.ProviderEventID != event.ProviderEventID {
		t.Fatalf("processed event = %+v", store.event)
	}
}
