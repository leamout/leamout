package payments

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

func ValidateCheckout(request CheckoutRequest) (CheckoutRequest, error) {
	request.Reference = strings.TrimSpace(request.Reference)
	request.Currency = strings.ToUpper(strings.TrimSpace(request.Currency))
	request.Email = strings.TrimSpace(request.Email)
	request.CallbackURL = strings.TrimSpace(request.CallbackURL)
	if request.Reference == "" {
		return CheckoutRequest{}, fmt.Errorf("payment reference is required")
	}
	if request.AmountMinor <= 0 {
		return CheckoutRequest{}, fmt.Errorf("payment amount must be positive")
	}
	if !currencyPattern.MatchString(request.Currency) {
		return CheckoutRequest{}, fmt.Errorf("payment currency must be a three-letter ISO code")
	}
	if request.Email == "" {
		return CheckoutRequest{}, fmt.Errorf("payment email is required")
	}
	if request.CallbackURL != "" {
		parsed, err := url.Parse(request.CallbackURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return CheckoutRequest{}, fmt.Errorf("payment callback URL must be an absolute HTTPS URL")
		}
	}
	return request, nil
}

func validProviderSession(provider, reference string, session CheckoutSession) bool {
	if session.Provider != provider || session.Reference != reference {
		return false
	}
	return provider != "stripe" || session.ProviderID != ""
}

func classifyProviderEvent(event ProviderEvent) (relevant, valid bool) {
	if event.Provider == "" || event.ProviderEventID == "" || len(event.Raw) == 0 || !json.Valid(event.Raw) {
		return false, false
	}
	switch event.Provider {
	case "stripe":
		relevant = event.Type == "checkout.session.completed" || event.Type == "checkout.session.expired"
	case "paystack":
		relevant = event.Type == "charge.success" || event.Type == "charge.failed"
	default:
		return false, false
	}
	if !relevant {
		return false, true
	}
	if event.Payment.Reference == "" || !eventTypeMatchesStatus(event) {
		return true, false
	}
	return true, true
}

func eventTypeMatchesStatus(event ProviderEvent) bool {
	switch event.Provider + ":" + event.Type {
	case "stripe:checkout.session.completed", "paystack:charge.success":
		return event.Payment.Status == StatusSucceeded
	case "stripe:checkout.session.expired":
		return event.Payment.Status == StatusCancelled
	case "paystack:charge.failed":
		return event.Payment.Status == StatusFailed || event.Payment.Status == StatusCancelled
	default:
		return false
	}
}

func lookupEventType(provider string, status Status) (string, bool) {
	switch provider {
	case "paystack":
		switch status {
		case StatusSucceeded:
			return "charge.success", true
		case StatusFailed, StatusCancelled:
			return "charge.failed", true
		}
	}
	return "", false
}
