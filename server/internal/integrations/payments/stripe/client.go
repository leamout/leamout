package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiBaseURL = "https://api.stripe.com"

type Client struct {
	secretKey     string
	webhookSecret string
	httpClient    *http.Client
	now           func() time.Time
}

func New(secretKey, webhookSecret string) *Client {
	return &Client{
		secretKey:     strings.TrimSpace(secretKey),
		webhookSecret: strings.TrimSpace(webhookSecret),
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		now:           time.Now,
	}
}

func (c *Client) Configured() bool {
	return c != nil && c.secretKey != "" && c.webhookSecret != ""
}

func (c *Client) doForm(ctx context.Context, method, path string, values url.Values, out any) error {
	if c == nil || c.secretKey == "" {
		return fmt.Errorf("stripe secret key is not configured")
	}

	var body io.Reader
	if values != nil {
		body = strings.NewReader(values.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, apiBaseURL+path, body)
	if err != nil {
		return fmt.Errorf("create Stripe request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	if values != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send Stripe request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var payload struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload)
		if payload.Error.Message != "" {
			return fmt.Errorf("Stripe %s: %s", payload.Error.Type, payload.Error.Message)
		}
		return fmt.Errorf("Stripe request failed with status %d", resp.StatusCode)
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out); err != nil {
		return fmt.Errorf("decode Stripe response: %w", err)
	}
	return nil
}
