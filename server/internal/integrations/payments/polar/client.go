package polar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	accessToken   string
	webhookSecret string
	baseURL       string
	httpClient    *http.Client
	now           func() time.Time
}

func New(accessToken, webhookSecret, environment string) *Client {
	baseURL := "https://api.polar.sh"
	if environment == "sandbox" {
		baseURL = "https://sandbox-api.polar.sh"
	}
	return &Client{
		accessToken:   strings.TrimSpace(accessToken),
		webhookSecret: strings.TrimSpace(webhookSecret),
		baseURL:       baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
			// Never forward bearer tokens across redirects.
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		now: time.Now,
	}
}

// NewWithHTTPClient allows callers to supply transport configuration while
// retaining the standard API hosts, request timeout, and redirect protection.
func NewWithHTTPClient(accessToken, webhookSecret, environment string, httpClient *http.Client) *Client {
	c := New(accessToken, webhookSecret, environment)
	if httpClient == nil {
		return c
	}
	configuredClient := *httpClient
	configuredClient.CheckRedirect = c.httpClient.CheckRedirect
	if configuredClient.Timeout == 0 {
		configuredClient.Timeout = c.httpClient.Timeout
	}
	c.httpClient = &configuredClient
	return c
}

func (c *Client) Configured() bool {
	return c != nil && c.accessToken != "" && c.webhookSecret != ""
}

func (c *Client) Environment() string {
	if c != nil && c.baseURL == "https://sandbox-api.polar.sh" {
		return "sandbox"
	}
	return "production"
}

func (c *Client) doJSON(ctx context.Context, method, path string, input, output any) error {
	if c == nil || c.accessToken == "" {
		return fmt.Errorf("polar access token is not configured")
	}
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode Polar request: %w", err)
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create Polar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Polar-Version", APIVersion)
	req.Header.Set("Accept", "application/json")
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send Polar request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
		}
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(output); err != nil {
		return fmt.Errorf("decode Polar response: %w", err)
	}
	return nil
}
