package topups

import (
	"errors"

	"github.com/leamout/leamout/internal/commercial/checkout"
	"github.com/leamout/leamout/internal/commercial/payments"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
	"github.com/leamout/leamout/pkg/apperror"
)

var (
	ErrProviderUnavailable = apperror.NewServiceUnavailable("payment provider is unavailable", errors.New("payment provider is not configured"))
	ErrInvalidTopup        = apperror.NewBadRequest("invalid wallet top-up")
	ErrPaymentMismatch     = apperror.NewConflict("provider payment does not match checkout")
)

type CreateInput struct {
	AmountMinor int64
	Provider    checkout.Provider
	Email       string
	CallbackURL string
	MobileMoney *paymentprovider.MobileMoney
}

type ContinueInput struct {
	Action paymentprovider.NextAction
	Value  string
}

type Checkout struct {
	Checkout checkout.Checkout
	Payment  payments.Payment
	Session  paymentprovider.CheckoutSession
}

type Details struct {
	Checkout checkout.Checkout
	Payment  payments.Payment
}

type Settlement = payments.Settlement
