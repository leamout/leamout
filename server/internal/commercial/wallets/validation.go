package wallets

import (
	"encoding/json"
	"time"

	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

func validateReservationInput(input ReserveInput, now time.Time) error {
	if input.AmountMinor <= 0 || !input.ExpiresAt.After(now) {
		return ErrInvalidMoney
	}
	return nil
}

func validateCaptureInput(amountMinor int64) error {
	if amountMinor <= 0 {
		return ErrInvalidMoney
	}
	return nil
}

func validateProviderEvent(event paymentprovider.Event) error {
	if event.Provider == "" || event.ProviderEventID == "" || event.Type == "" ||
		event.Payment.Reference == "" || len(event.Raw) == 0 || !json.Valid(event.Raw) {
		return ErrPaymentMismatch
	}
	return nil
}
