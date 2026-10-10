package stripe

import (
	"fmt"
	"net/url"
	"strings"
)

func validateCheckoutParams(params CheckoutParams) error {
	if err := validateStripeID("price", params.PriceID, "price_"); err != nil {
		return err
	}
	if strings.TrimSpace(params.OrganizationID) == "" {
		return fmt.Errorf("organization id is required")
	}
	if strings.TrimSpace(params.PlanCode) == "" {
		return fmt.Errorf("plan code is required")
	}
	if err := validateRedirectURL("success URL", params.SuccessURL); err != nil {
		return err
	}
	return validateRedirectURL("cancel URL", params.CancelURL)
}

func validatePortalParams(params PortalParams) error {
	if err := validateStripeID("customer", params.CustomerID, "cus_"); err != nil {
		return err
	}
	return validateRedirectURL("return URL", params.ReturnURL)
}

func validateStripeID(name, value, prefix string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s id is required", name)
	}
	if !strings.HasPrefix(value, prefix) {
		return fmt.Errorf("invalid %s id", name)
	}
	return nil
}

func validateRedirectURL(name, value string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("invalid %s", name)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("invalid %s scheme", name)
	}
	return nil
}
