package commpeak

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.commpeak.com"

type Config struct {
	BaseURL       string
	Authorization string
	HTTPClient    *http.Client
}

type Client struct {
	baseURL       *url.URL
	authorization string
	httpClient    *http.Client
}

func NewClient(config Config) (*Client, error) {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("commpeak: valid base URL is required")
	}
	authorization := strings.TrimSpace(config.Authorization)
	if authorization == "" {
		return nil, fmt.Errorf("commpeak: Authorization credential is required")
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		baseURL:       parsed,
		authorization: authorization,
		httpClient:    httpClient,
	}, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	endpoint := c.baseURL.ResolveReference(&url.URL{Path: path})
	if len(query) > 0 {
		endpoint.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("commpeak: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.authorization)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("commpeak: execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("commpeak: read response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newAPIError(resp.StatusCode, payload)
	}
	return payload, nil
}
