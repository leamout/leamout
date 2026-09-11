package topups

import (
	"errors"

	"github.com/leamout/leamout/internal/commercial/payments"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	checkouts "github.com/leamout/leamout/internal/commercial/purchase/checkouts"
	"github.com/leamout/leamout/pkg/apperror"
)

var (
	ErrProviderUnavailable = apperror.NewServiceUnavailable("payment provider is unavailable", errors.New("payment provider is not configured"))
	ErrInvalidTopup        = apperror.NewBadRequest("invalid wallet top-up")
	ErrPaymentMismatch     = apperror.NewConflict("provider payment does not match checkout")
)

type CreateInput struct {
	AmountMinor int64
	Provider    checkouts.Provider
	Email       string
	CallbackURL string
	MobileMoney *commercialpayments.MobileMoney
}

type ContinueInput struct {
	Action commercialpayments.NextAction
	Value  string
}

type Checkout struct {
	Checkout checkouts.Checkout
	Payment  payments.Payment
	Session  commercialpayments.CheckoutSession
}

type Details struct {
	Checkout checkouts.Checkout
	Payment  payments.Payment
}

type Settlement = payments.Settlement
