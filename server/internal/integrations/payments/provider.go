package payments

import (
	"context"
	"net/http"
)

type Status string
type NextAction string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

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

// Provider initializes payment collection, retrieves the authoritative payment
// state, and authenticates provider webhook payloads.
// Commercial consequences remain the responsibility of the commercial domain.
type Provider interface {
	CreateCheckout(context.Context, CheckoutRequest) (CheckoutSession, error)
	GetPayment(context.Context, string) (Payment, error)
	ParseWebhook([]byte, http.Header) (Event, error)
}
