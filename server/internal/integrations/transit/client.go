package transit

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
	availableNumbersPath = "/internal/v1/transit/numbers/available"
	numberOrdersPath     = "/internal/v1/transit/number-orders"
)

type Config struct {
	BaseURL      string
	Token        string
	DeploymentID string
	HTTPClient   *http.Client
}

type Client struct {
	baseURL      *url.URL
	token        string
	deploymentID string
	httpClient   *http.Client
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("transit returned %d: %s", e.StatusCode, e.Message)
}

func (e *HTTPError) Retryable() bool {
	if e == nil {
		return true
	}
	return e.StatusCode == http.StatusRequestTimeout ||
		e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode >= http.StatusInternalServerError
}

func NewClient(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("transit base URL is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Transit base URL: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, fmt.Errorf("transit base URL must use http or https")
	}
	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		return nil, fmt.Errorf("transit token is required")
	}
	deploymentID := strings.TrimSpace(cfg.DeploymentID)
	if deploymentID == "" {
		return nil, fmt.Errorf("transit deployment ID is required")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		baseURL:      parsed,
		token:        token,
		deploymentID: deploymentID,
		httpClient:   httpClient,
	}, nil
}

func (c *Client) SearchAvailable(ctx context.Context, req AvailableNumberSearchRequest) (AvailableNumberSearchResponse, error) {
	var response AvailableNumberSearchResponse
	if err := c.doJSON(ctx, http.MethodPost, availableNumbersPath, req, &response); err != nil {
		return AvailableNumberSearchResponse{}, err
	}
	return response, nil
}

func (c *Client) ExecuteNumberOrder(ctx context.Context, req ExecuteNumberOrderRequest) (ExecuteNumberOrderResponse, error) {
	var response ExecuteNumberOrderResponse
	if err := c.doJSON(ctx, http.MethodPost, numberOrdersPath, req, &response); err != nil {
		return ExecuteNumberOrderResponse{}, err
	}
	return response, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, requestBody any, responseBody any) error {
	if c == nil || c.baseURL == nil || c.httpClient == nil {
		return fmt.Errorf("transit client is not initialized")
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("encode Transit request: %w", err)
	}
	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + path

	httpRequest, err := http.NewRequestWithContext(ctx, method, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create Transit request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.token)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("X-Leamout-Deployment-ID", c.deploymentID)

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("transit request failed: %w", err)
	}
	defer func() { _ = httpResponse.Body.Close() }()

	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		payload, _ := io.ReadAll(io.LimitReader(httpResponse.Body, 4096))
		message := strings.TrimSpace(string(payload))
		if message == "" {
			message = http.StatusText(httpResponse.StatusCode)
		}
		return &HTTPError{StatusCode: httpResponse.StatusCode, Message: message}
	}
	if responseBody == nil {
		return nil
	}
	decoder := json.NewDecoder(httpResponse.Body)
	if err := decoder.Decode(responseBody); err != nil {
		return fmt.Errorf("decode Transit response: %w", err)
	}
	return nil
}
