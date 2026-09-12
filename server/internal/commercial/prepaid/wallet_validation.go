package prepaid

import "time"

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
