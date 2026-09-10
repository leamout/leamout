package topups

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/checkout"
	"github.com/leamout/leamout/internal/commercial/payments"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
	"github.com/leamout/leamout/pkg/apperror"
)

var (
	ErrProviderUnavailable = apperror.NewServiceUnavailable("payment provider is unavailable", errors.New("payment provider is not configured"))
	ErrInvalidTopup        = apperror.NewBadRequest("invalid wallet top-up")
	ErrPaymentMismatch     = apperror.NewConflict("provider payment does not match checkout order")
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
	Order   checkout.Order
	Payment payments.Payment
	Session paymentprovider.CheckoutSession
}

type Details struct {
	Order   checkout.Order
	Payment payments.Payment
}

type Settlement struct {
	Applied        bool
	OrganizationID uuid.UUID
	WalletID       uuid.UUID
	PaymentID      uuid.UUID
	AmountMinor    int64
	SettledAt      *time.Time
}
