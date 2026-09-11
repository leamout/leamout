package checkout

import (
	"errors"

	"github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/pkg/apperror"
)

var (
	ErrProviderUnavailable = apperror.NewServiceUnavailable("payment provider is unavailable", errors.New("payment provider is not configured"))
	ErrInvalidTopup        = apperror.NewBadRequest("invalid wallet top-up")
	ErrPaymentMismatch     = apperror.NewConflict("provider payment does not match checkout")
)

type TopupCreateInput struct {
	AmountMinor int64
	Provider    Provider
	Email       string
	CallbackURL string
	MobileMoney *payments.MobileMoney
}

type TopupContinueInput struct {
	Action payments.NextAction
	Value  string
}

type TopupResult struct {
	Checkout Checkout
	Payment  payments.Payment
	Session  payments.CheckoutSession
}

type TopupDetails struct {
	Checkout Checkout
	Payment  payments.Payment
}
type TopupSettlement = payments.Settlement
