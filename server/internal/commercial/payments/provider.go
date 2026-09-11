package payments

import (
	"context"
	"net/http"
)

type NextAction string

const (
	NextActionNone                 NextAction = "none"
	NextActionWait                 NextAction = "wait"
	NextActionAuthorizeMobileMoney NextAction = "authorize_mobile_money"
	NextActionSubmitOTP            NextAction = "submit_otp"
	NextActionSubmitPhone          NextAction = "submit_phone"
	NextActionUnsupported          NextAction = "unsupported"
)

type CheckoutRequest struct {
	Reference   string
	AmountMinor int64
	Currency    string
	Email       string
	CallbackURL string
	Metadata    map[string]string
	MobileMoney *MobileMoney
}

type MobileMoney struct {
	Phone    string
	Provider string
}

type CheckoutSession struct {
	Provider     string
	ProviderID   string
	Reference    string
	ClientSecret string
	NextAction   NextAction
	Message      string
	Status       Status
}

type ContinueCheckoutRequest struct {
	Reference string
	Action    NextAction
	Value     string
}

// ContinuationProvider handles provider-requested input without persisting
// sensitive challenge values such as a PIN or OTP.
type ContinuationProvider interface {
	ContinueCheckout(context.Context, ContinueCheckoutRequest) (CheckoutSession, error)
	GetCheckout(context.Context, string) (CheckoutSession, error)
}

type ProviderPayment struct {
	Provider    string
	ProviderID  string
	Reference   string
	AmountMinor int64
	Currency    string
	Status      Status
}

type ProviderEvent struct {
	Provider        string
	ProviderEventID string
	Type            string
	Payment         ProviderPayment
	Raw             []byte
}

// Provider initializes payment collection, retrieves the authoritative payment
// state, and authenticates provider webhook payloads.
// Commercial consequences remain the responsibility of the commercial domain.
type Provider interface {
	CreateCheckout(context.Context, CheckoutRequest) (CheckoutSession, error)
	GetPayment(context.Context, string) (ProviderPayment, error)
	ParseWebhook([]byte, http.Header) (ProviderEvent, error)
}
