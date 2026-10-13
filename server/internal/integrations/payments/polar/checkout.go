package polar

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

func (c *Client) CreateCheckoutSession(ctx context.Context, params CheckoutParams) (CheckoutSession, error) {
	if err := validateUUID(params.ProductID); err != nil {
		return CheckoutSession{}, err
	}
	if err := validateUUID(params.OrganizationID); err != nil {
		return CheckoutSession{}, err
	}
	input := struct {
		Products           []string          `json:"products"`
		ExternalCustomerID string            `json:"external_customer_id"`
		SuccessURL         string            `json:"success_url"`
		ReturnURL          string            `json:"return_url"`
		Metadata           map[string]string `json:"metadata"`
	}{
		Products: []string{
			params.ProductID,
		},
		ExternalCustomerID: params.OrganizationID,
		SuccessURL:         params.SuccessURL,
		ReturnURL:          params.ReturnURL,
		Metadata: map[string]string{
			"organization_id": params.OrganizationID,
			"plan_code":       params.PlanCode,
		},
	}
	var result CheckoutSession
	err := c.doJSON(ctx, http.MethodPost, "/v1/checkouts/", input, &result)
	if err == nil && result.URL == "" {
		return CheckoutSession{}, fmt.Errorf("polar checkout URL is missing")
	}
	return result, err
}

func (c *Client) CreatePortalSession(ctx context.Context, params PortalParams) (PortalSession, error) {
	if err := validateUUID(params.CustomerID); err != nil {
		return PortalSession{}, err
	}
	input := struct {
		CustomerID string `json:"customer_id"`
		ReturnURL  string `json:"return_url"`
	}{
		CustomerID: params.CustomerID,
		ReturnURL:  params.ReturnURL,
	}
	var result PortalSession
	err := c.doJSON(ctx, http.MethodPost, "/v1/customer-sessions/", input, &result)
	if err == nil && result.URL == "" {
		return PortalSession{}, fmt.Errorf("polar customer portal URL is missing")
	}
	return result, err
}

func (c *Client) GetSubscription(ctx context.Context, id string) (Subscription, error) {
	if err := validateUUID(id); err != nil {
		return Subscription{}, err
	}
	var result Subscription
	err := c.doJSON(ctx, http.MethodGet, "/v1/subscriptions/"+url.PathEscape(id), nil, &result)
	return result, err
}

func validateUUID(value string) error {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return fmt.Errorf("nonempty UUID is required")
	}
	return nil
}
