package stripe

import (
	"context"
	"net/http"
	"net/url"
)

type CheckoutParams struct {
	PriceID        string
	CustomerID     string
	OrganizationID string
	PlanCode       string
	SuccessURL     string
	CancelURL      string
}

type CheckoutSession struct {
	ID           string            `json:"id"`
	URL          string            `json:"url"`
	Customer     string            `json:"customer"`
	Subscription string            `json:"subscription"`
	Metadata     map[string]string `json:"metadata"`
}

type PortalParams struct {
	CustomerID string
	ReturnURL  string
}

type PortalSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func (c *Client) CreateCheckoutSession(ctx context.Context, params CheckoutParams) (CheckoutSession, error) {
	if err := validateCheckoutParams(params); err != nil {
		return CheckoutSession{}, err
	}

	values := url.Values{
		"mode":                                         {"subscription"},
		"line_items[0][price]":                         {params.PriceID},
		"line_items[0][quantity]":                      {"1"},
		"client_reference_id":                          {params.OrganizationID},
		"metadata[organization_id]":                    {params.OrganizationID},
		"metadata[plan_code]":                          {params.PlanCode},
		"subscription_data[metadata][organization_id]": {params.OrganizationID},
		"subscription_data[metadata][plan_code]":       {params.PlanCode},
		"success_url":                                  {params.SuccessURL},
		"cancel_url":                                   {params.CancelURL},
	}
	if params.CustomerID != "" {
		values.Set("customer", params.CustomerID)
	}

	var session CheckoutSession
	err := c.doForm(ctx, http.MethodPost, "/v1/checkout/sessions", values, &session)
	return session, err
}

func (c *Client) CreatePortalSession(ctx context.Context, params PortalParams) (PortalSession, error) {
	if err := validatePortalParams(params); err != nil {
		return PortalSession{}, err
	}

	values := url.Values{
		"customer":   {params.CustomerID},
		"return_url": {params.ReturnURL},
	}
	var session PortalSession
	err := c.doForm(ctx, http.MethodPost, "/v1/billing_portal/sessions", values, &session)
	return session, err
}

func (c *Client) GetSubscription(ctx context.Context, id string) (Subscription, error) {
	if err := validateStripeID("subscription", id, "sub_"); err != nil {
		return Subscription{}, err
	}
	var subscription Subscription
	err := c.doForm(ctx, http.MethodGet, "/v1/subscriptions/"+url.PathEscape(id), nil, &subscription)
	return subscription, err
}
