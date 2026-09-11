package stripe

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	paymentprovider "github.com/leamout/leamout/internal/commercial/payments"
)

const (
	DefaultBaseURL    = "https://api.stripe.com/v1"
	DefaultAPIVersion = "2026-08-26.dahlia"
)

type Config struct {
	BaseURL          string
	SecretKey        string
	WebhookSecret    string
	WebhookTolerance time.Duration
	APIVersion       string
	HTTPClient       *http.Client
	Now              func() time.Time
}

type Client struct {
	baseURL          *url.URL
	secretKey        string
	webhookSecret    string
	webhookTolerance time.Duration
	apiVersion       string
	httpClient       *http.Client
	now              func() time.Time
}

func NewClient(config Config) (*Client, error) {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("stripe: valid base URL is required")
	}
	secretKey := strings.TrimSpace(config.SecretKey)
	if secretKey == "" {
		return nil, fmt.Errorf("stripe: secret key is required")
	}
	tolerance := config.WebhookTolerance
	if tolerance == 0 {
		tolerance = 5 * time.Minute
	}
	if tolerance < 0 {
		return nil, fmt.Errorf("stripe: webhook tolerance cannot be negative")
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &Client{
		baseURL: parsed, secretKey: secretKey, webhookSecret: strings.TrimSpace(config.WebhookSecret),
		webhookTolerance: tolerance, apiVersion: firstNonEmpty(config.APIVersion, DefaultAPIVersion),
		httpClient: httpClient, now: now,
	}, nil
}

type checkoutSession struct {
	ID            string            `json:"id"`
	ClientSecret  string            `json:"client_secret"`
	AmountTotal   int64             `json:"amount_total"`
	Currency      string            `json:"currency"`
	PaymentStatus string            `json:"payment_status"`
	Status        string            `json:"status"`
	Metadata      map[string]string `json:"metadata"`
}

func (c *Client) CreateCheckout(ctx context.Context, request paymentprovider.CheckoutRequest) (paymentprovider.CheckoutSession, error) {
	normalized, err := paymentprovider.ValidateCheckout(request)
	if err != nil {
		return paymentprovider.CheckoutSession{}, fmt.Errorf("stripe: %w", err)
	}
	values := url.Values{
		"mode":                                   {"payment"},
		"ui_mode":                                {"elements"},
		"customer_email":                         {normalized.Email},
		"payment_method_types[]":                 {"card"},
		"line_items[0][price_data][currency]":    {strings.ToLower(normalized.Currency)},
		"line_items[0][price_data][unit_amount]": {strconv.FormatInt(normalized.AmountMinor, 10)},
		"line_items[0][price_data][product_data][name]": {"Wallet top-up"},
		"line_items[0][quantity]":                       {"1"},
		"metadata[leamout_reference]":                   {normalized.Reference},
	}
	for key, value := range normalized.Metadata {
		if key = strings.TrimSpace(key); key != "" && key != "leamout_reference" {
			values.Set("metadata["+key+"]", value)
		}
	}
	var result checkoutSession
	if err := c.do(ctx, http.MethodPost, "/checkout/sessions", values, normalized.Reference, &result); err != nil {
		return paymentprovider.CheckoutSession{}, err
	}
	if result.ID == "" || result.ClientSecret == "" {
		return paymentprovider.CheckoutSession{}, fmt.Errorf("stripe: invalid Checkout Session response")
	}
	return paymentprovider.CheckoutSession{
		Provider: "stripe", ProviderID: result.ID, Reference: normalized.Reference,
		ClientSecret: result.ClientSecret, Status: normalizeStatus(result),
	}, nil
}

func (c *Client) GetPayment(ctx context.Context, providerID string) (paymentprovider.ProviderPayment, error) {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return paymentprovider.ProviderPayment{}, fmt.Errorf("stripe: Checkout Session ID is required")
	}
	var result checkoutSession
	if err := c.do(ctx, http.MethodGet, "/checkout/sessions/"+url.PathEscape(providerID), nil, "", &result); err != nil {
		return paymentprovider.ProviderPayment{}, err
	}
	return normalizePayment(result), nil
}

func (c *Client) ParseWebhook(payload []byte, headers http.Header) (paymentprovider.ProviderEvent, error) {
	if c.webhookSecret == "" {
		return paymentprovider.ProviderEvent{}, fmt.Errorf("stripe: webhook secret is required")
	}
	timestamp, signatures, err := parseSignatureHeader(headers.Get("Stripe-Signature"))
	if err != nil {
		return paymentprovider.ProviderEvent{}, err
	}
	if delta := c.now().Sub(time.Unix(timestamp, 0)); delta > c.webhookTolerance || delta < -c.webhookTolerance {
		return paymentprovider.ProviderEvent{}, fmt.Errorf("stripe: webhook timestamp outside tolerance")
	}
	signed := append([]byte(strconv.FormatInt(timestamp, 10)+"."), payload...)
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	_, _ = mac.Write(signed)
	expected := mac.Sum(nil)
	verified := false
	for _, signature := range signatures {
		if hmac.Equal(signature, expected) {
			verified = true
			break
		}
	}
	if !verified {
		return paymentprovider.ProviderEvent{}, fmt.Errorf("stripe: invalid webhook signature")
	}
	var envelope struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object checkoutSession `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return paymentprovider.ProviderEvent{}, fmt.Errorf("stripe: decode webhook: %w", err)
	}
	if envelope.ID == "" || envelope.Type == "" {
		return paymentprovider.ProviderEvent{}, fmt.Errorf("stripe: webhook event identity is required")
	}
	return paymentprovider.ProviderEvent{
		Provider: "stripe", ProviderEventID: envelope.ID, Type: envelope.Type,
		Payment: normalizePayment(envelope.Data.Object), Raw: append([]byte(nil), payload...),
	}, nil
}

func parseSignatureHeader(value string) (int64, [][]byte, error) {
	var timestamp int64
	var signatures [][]byte
	for _, part := range strings.Split(value, ",") {
		key, item, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			parsed, err := strconv.ParseInt(item, 10, 64)
			if err != nil {
				return 0, nil, fmt.Errorf("stripe: invalid webhook timestamp")
			}
			timestamp = parsed
		case "v1":
			decoded, err := hex.DecodeString(item)
			if err == nil {
				signatures = append(signatures, decoded)
			}
		}
	}
	if timestamp == 0 || len(signatures) == 0 {
		return 0, nil, fmt.Errorf("stripe: valid webhook signature is required")
	}
	return timestamp, signatures, nil
}

func normalizePayment(item checkoutSession) paymentprovider.ProviderPayment {
	return paymentprovider.ProviderPayment{
		Provider: "stripe", ProviderID: item.ID, Reference: item.Metadata["leamout_reference"],
		AmountMinor: item.AmountTotal, Currency: strings.ToUpper(item.Currency), Status: normalizeStatus(item),
	}
}

func normalizeStatus(session checkoutSession) paymentprovider.Status {
	if session.PaymentStatus == "paid" {
		return paymentprovider.StatusSucceeded
	}
	switch session.Status {
	case "expired":
		return paymentprovider.StatusCancelled
	case "complete":
		return paymentprovider.StatusProcessing
	default:
		return paymentprovider.StatusPending
	}
}

func (c *Client) do(ctx context.Context, method, path string, values url.Values, idempotencyKey string, result any) error {
	endpoint := c.baseURL.ResolveReference(&url.URL{Path: strings.TrimSuffix(c.baseURL.Path, "/") + path})
	var body io.Reader
	if values != nil {
		body = strings.NewReader(values.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return fmt.Errorf("stripe: create request: %w", err)
	}
	req.SetBasicAuth(c.secretKey, "")
	req.Header.Set("Stripe-Version", c.apiVersion)
	req.Header.Set("Accept", "application/json")
	if values != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("stripe: execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("stripe: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("stripe: API status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	if err := json.NewDecoder(bytes.NewReader(payload)).Decode(result); err != nil {
		return fmt.Errorf("stripe: decode response: %w", err)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
