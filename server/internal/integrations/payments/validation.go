package payments

import (
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
