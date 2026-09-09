package paystack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

const DefaultBaseURL = "https://api.paystack.co"

var paystackReferencePattern = regexp.MustCompile(`^[A-Za-z0-9.=-]+$`)

type Config struct {
	BaseURL    string
	SecretKey  string
	HTTPClient *http.Client
}

type Client struct {
	baseURL    *url.URL
	secretKey  string
	httpClient *http.Client
}

func NewClient(config Config) (*Client, error) {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("paystack: valid base URL is required")
	}
	secretKey := strings.TrimSpace(config.SecretKey)
	if secretKey == "" {
		return nil, fmt.Errorf("paystack: secret key is required")
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{baseURL: parsed, secretKey: secretKey, httpClient: httpClient}, nil
}

type chargeRequest struct {
	Email       string            `json:"email"`
	Amount      string            `json:"amount"`
	Reference   string            `json:"reference"`
	MobileMoney mobileMoney       `json:"mobile_money"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type mobileMoney struct {
	Phone    string `json:"phone"`
	Provider string `json:"provider"`
}

type transaction struct {
	ID              json.Number `json:"id"`
	Reference       string      `json:"reference"`
	Amount          int64       `json:"amount"`
	Currency        string      `json:"currency"`
	Status          string      `json:"status"`
	Message         string      `json:"message"`
	GatewayResponse string      `json:"gateway_response"`
}

type response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    transaction `json:"data"`
}

func (c *Client) CreateCheckout(ctx context.Context, request paymentprovider.CheckoutRequest) (paymentprovider.CheckoutSession, error) {
	normalized, err := paymentprovider.ValidateCheckout(request)
	if err != nil {
		return paymentprovider.CheckoutSession{}, fmt.Errorf("paystack: %w", err)
	}
	if err := validateMobileMoney(normalized); err != nil {
		return paymentprovider.CheckoutSession{}, fmt.Errorf("paystack: %w", err)
	}
	var result response
	err = c.do(ctx, http.MethodPost, "/charge", chargeRequest{
		Email: normalized.Email, Amount: fmt.Sprint(normalized.AmountMinor), Reference: normalized.Reference,
		MobileMoney: mobileMoney{Phone: strings.TrimSpace(normalized.MobileMoney.Phone), Provider: strings.ToLower(strings.TrimSpace(normalized.MobileMoney.Provider))},
		Metadata:    normalized.Metadata,
	}, &result)
	if err != nil {
		return paymentprovider.CheckoutSession{}, err
	}
	if !result.Status || result.Data.Reference == "" {
		return paymentprovider.CheckoutSession{}, fmt.Errorf("paystack: invalid charge response: %s", result.Message)
	}
	return paymentprovider.CheckoutSession{
		Provider: "paystack", ProviderID: result.Data.ID.String(), Reference: result.Data.Reference,
		NextAction: result.Data.Status, Message: firstNonEmpty(result.Data.Message, result.Data.GatewayResponse, result.Message),
		Status: normalizeStatus(result.Data.Status),
	}, nil
}

func validateMobileMoney(request paymentprovider.CheckoutRequest) error {
	if !paystackReferencePattern.MatchString(request.Reference) {
		return fmt.Errorf("payment reference may contain only letters, numbers, period, equals, and hyphen")
	}
	if request.Currency != "GHS" {
		return fmt.Errorf("mobile money currency must be GHS")
	}
	if request.MobileMoney == nil {
		return fmt.Errorf("mobile money details are required")
	}
	phone := strings.TrimSpace(request.MobileMoney.Phone)
	if phone == "" {
		return fmt.Errorf("mobile money phone is required")
	}
	switch strings.ToLower(strings.TrimSpace(request.MobileMoney.Provider)) {
	case "mtn", "atl", "vod":
	default:
		return fmt.Errorf("mobile money provider must be mtn, atl, or vod")
	}
	return nil
}

func (c *Client) GetPayment(ctx context.Context, reference string) (paymentprovider.Payment, error) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return paymentprovider.Payment{}, fmt.Errorf("paystack: payment reference is required")
	}
	var result response
	if err := c.do(ctx, http.MethodGet, "/charge/"+url.PathEscape(reference), nil, &result); err != nil {
		return paymentprovider.Payment{}, err
	}
	if !result.Status {
		return paymentprovider.Payment{}, fmt.Errorf("paystack: charge lookup failed: %s", result.Message)
	}
	return normalizePayment(result.Data), nil
}

func (c *Client) ParseWebhook(payload []byte, headers http.Header) (paymentprovider.Event, error) {
	signature, err := hex.DecodeString(strings.TrimSpace(headers.Get("x-paystack-signature")))
	if err != nil || len(signature) == 0 {
		return paymentprovider.Event{}, fmt.Errorf("paystack: valid webhook signature is required")
	}
	mac := hmac.New(sha512.New, []byte(c.secretKey))
	_, _ = mac.Write(payload)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return paymentprovider.Event{}, fmt.Errorf("paystack: invalid webhook signature")
	}
	var envelope struct {
		Event string      `json:"event"`
		Data  transaction `json:"data"`
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		return paymentprovider.Event{}, fmt.Errorf("paystack: decode webhook: %w", err)
	}
	return paymentprovider.Event{
		Provider: "paystack", ProviderEventID: envelope.Data.ID.String(), Type: envelope.Event,
		Payment: normalizePayment(envelope.Data), Raw: append([]byte(nil), payload...),
	}, nil
}

func normalizePayment(item transaction) paymentprovider.Payment {
	status := normalizeStatus(item.Status)
	return paymentprovider.Payment{Provider: "paystack", ProviderID: item.ID.String(), Reference: item.Reference, AmountMinor: item.Amount, Currency: strings.ToUpper(item.Currency), Status: status}
}

func normalizeStatus(value string) paymentprovider.Status {
	status := paymentprovider.StatusPending
	switch value {
	case "success":
		status = paymentprovider.StatusSucceeded
	case "failed", "abandoned", "reversed":
		status = paymentprovider.StatusFailed
	case "ongoing", "processing", "pending", "queued":
		status = paymentprovider.StatusProcessing
	}
	return status
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	endpoint := c.baseURL.ResolveReference(&url.URL{Path: path})
	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("paystack: encode request: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), requestBody)
	if err != nil {
		return fmt.Errorf("paystack: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("paystack: execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("paystack: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("paystack: API status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(result); err != nil {
		return fmt.Errorf("paystack: decode response: %w", err)
	}
	return nil
}
