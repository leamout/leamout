package didww

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultBaseURL    = "https://api.didww.com"
	DefaultAPIVersion = "2026-04-16"
	jsonAPIMediaType  = "application/vnd.api+json"
)

type Config struct {
	BaseURL    string
	APIKey     string
	APIVersion string
	HTTPClient *http.Client
}

type Client struct {
	baseURL    *url.URL
	apiKey     string
	apiVersion string
	httpClient *http.Client
}

func NewClient(config Config) (*Client, error) {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("didww: valid base URL is required")
	}
	apiKey := strings.TrimSpace(config.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("didww: API key is required")
	}
	apiVersion := strings.TrimSpace(config.APIVersion)
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		baseURL:    parsed,
		apiKey:     apiKey,
		apiVersion: apiVersion,
		httpClient: httpClient,
	}, nil
}

func (c *Client) do(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
	body any,
	result any,
) error {
	endpoint := c.baseURL.ResolveReference(&url.URL{Path: path})
	if len(query) > 0 {
		endpoint.RawQuery = query.Encode()
	}

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("didww: encode request: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), requestBody)
	if err != nil {
		return fmt.Errorf("didww: create request: %w", err)
	}
	req.Header.Set("Accept", jsonAPIMediaType)
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("X-DIDWW-API-Version", c.apiVersion)
	if body != nil {
		req.Header.Set("Content-Type", jsonAPIMediaType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("didww: execute request: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("didww: read response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return newAPIError(resp.StatusCode, payload)
	}
	if result == nil || len(payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, result); err != nil {
		return fmt.Errorf("didww: decode response: %w", err)
	}
	return nil
}
