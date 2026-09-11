package checkouts

import (
	"regexp"
	"strings"
	"time"
)

var checkoutReferencePattern = regexp.MustCompile(`^[A-Za-z0-9.=-]+$`)

func validateCreate(input CreateInput, now time.Time) error {
	validTarget := input.Type == TypeSubscription && input.PriceID != nil && input.WalletID == nil ||
		input.Type == TypeWalletTopup && input.WalletID != nil && input.PriceID == nil
	validMethod := input.Provider == ProviderStripe && input.PaymentMethod == MethodCard ||
		input.Provider == ProviderPaystack && input.PaymentMethod == MethodMobileMoney
	if !validTarget || !validMethod || input.AmountMinor <= 0 ||
		len(input.Currency) != 3 || input.Currency != strings.ToUpper(input.Currency) ||
		!checkoutReferencePattern.MatchString(input.Reference) || !input.ExpiresAt.After(now) {
		return ErrInvalidCheckout
	}
	return nil
}

func validateTransition(transition Transition) error {
	if isTerminal(transition.Expected) {
		return ErrInvalidTransition
	}
	if transition.Status != StatusPending && transition.Status != StatusProcessing && !isTerminal(transition.Status) {
		return ErrInvalidTransition
	}
	if isTerminal(transition.Status) {
		if transition.NextAction != ActionNone || transition.CompletedAt == nil {
			return ErrInvalidTransition
		}
	} else if transition.CompletedAt != nil {
		return ErrInvalidTransition
	}
	return nil
}

func isTerminal(status Status) bool {
	return status == StatusSucceeded || status == StatusFailed || status == StatusCancelled || status == StatusExpired
}
