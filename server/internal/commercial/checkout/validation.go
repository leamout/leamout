package checkout

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var checkoutReferencePattern = regexp.MustCompile(`^[A-Za-z0-9.=-]+$`)

func validateIntent(input CreateParams) error {
	if input.WalletID == nil || *input.WalletID == uuid.Nil || input.AmountMinor <= 0 || !validMetadata(input.Metadata) {
		return ErrInvalidCheckout
	}
	return nil
}

func validateCreate(input CreateInput, now time.Time) error {
	if input.Type != TypeWalletTopup || input.WalletID == nil || input.PriceID != nil || input.AmountMinor <= 0 ||
		len(input.Currency) != 3 || input.Currency != strings.ToUpper(input.Currency) ||
		!checkoutReferencePattern.MatchString(input.Reference) || !input.ExpiresAt.After(now) || !validMetadata(input.Metadata) {
		return ErrInvalidCheckout
	}
	return nil
}

func validateConfirm(input ConfirmInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return ErrInvalidCheckout
	}
	switch input.PaymentMethod {
	case MethodCard:
		if input.MobileMoney != nil { return ErrInvalidCheckout }
	case MethodMobileMoney:
		if input.MobileMoney == nil || strings.TrimSpace(input.MobileMoney.Phone) == "" || strings.TrimSpace(input.MobileMoney.Provider) == "" {
			return ErrInvalidCheckout
		}
	default:
		return ErrInvalidCheckout
	}
	return nil
}

func validateTransition(transition Transition) error {
	if isTerminal(transition.Expected) { return ErrInvalidTransition }
	if transition.Status != StatusPending && transition.Status != StatusProcessing && !isTerminal(transition.Status) { return ErrInvalidTransition }
	if isTerminal(transition.Status) {
		if transition.NextAction != ActionNone || transition.CompletedAt == nil { return ErrInvalidTransition }
	} else if transition.CompletedAt != nil { return ErrInvalidTransition }
	return nil
}

func providerForMethod(method PaymentMethod) (Provider, bool) {
	switch method {
	case MethodCard: return ProviderStripe, true
	case MethodMobileMoney: return ProviderPaystack, true
	default: return "", false
	}
}

func validMetadata(raw json.RawMessage) bool {
	if len(raw) == 0 { return true }
	var value map[string]any
	return json.Unmarshal(raw, &value) == nil && value != nil
}

func isTerminal(status Status) bool {
	return status == StatusSucceeded || status == StatusFailed || status == StatusCancelled || status == StatusExpired
}
