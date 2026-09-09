package payments

import (
	"context"
	"net/http"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type CheckoutRequest struct {
	Reference   string
	AmountMinor int64
	Currency    string
	Email       string
	CallbackURL string
	Metadata    map[string]string
}

type CheckoutSession struct {
	Provider         string
	ProviderID       string
	Reference        string
	ClientSecret     string
	AccessCode       string
	AuthorizationURL string
	Status           Status
}

type Payment struct {
	Provider    string
	ProviderID  string
	Reference   string
	AmountMinor int64
	Currency    string
	Status      Status
}

type Event struct {
	Provider        string
	ProviderEventID string
	Type            string
	Payment         Payment
	Raw             []byte
}

// Provider initializes secure provider-hosted payment collection, retrieves the
// authoritative payment state, and authenticates provider webhook payloads.
// Commercial consequences remain the responsibility of the commercial domain.
type Provider interface {
	CreateCheckout(context.Context, CheckoutRequest) (CheckoutSession, error)
	GetPayment(context.Context, string) (Payment, error)
	ParseWebhook([]byte, http.Header) (Event, error)
}
