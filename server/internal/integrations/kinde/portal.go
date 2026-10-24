package kinde

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// PortalURL requests a one-time user portal URL using an authenticated account
// API request. Browser clients must never receive the OAuth access token.
func (c *Client) PortalURL(ctx context.Context, accessToken, returnURL string) (string, error) {
	if accessToken == "" {
		return "", errors.New("missing Kinde access token")
	}
	target, err := url.Parse(returnURL)
	if err != nil || target.User != nil || target.Host == "" ||
		(target.Scheme != "https" &&
			(target.Scheme != "http" ||
				(target.Hostname() != "localhost" && target.Hostname() != "127.0.0.1"))) {
		return "", errors.New("invalid portal return URL")
	}
	endpoint := c.issuer + "/account_api/v1/portal_link"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create Kinde portal request: %w", err)
	}
	query := request.URL.Query()
	query.Set("subnav", "profile")
	query.Set("return_url", returnURL)
	request.URL.RawQuery = query.Encode()
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/json")

	httpClient := *c.http
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := httpClient.Do(request)
	if err != nil {
		return "", errors.New("kinde account portal unavailable")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return "", errors.New("kinde account portal unavailable")
	}
	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4096)).Decode(&result); err != nil {
		return "", errors.New("invalid Kinde portal response")
	}
	destination, err := url.Parse(result.URL)
	issuerURL, _ := url.Parse(c.issuer)
	if err != nil || destination.Scheme != issuerURL.Scheme ||
		destination.Host != issuerURL.Host || destination.User != nil ||
		destination.Fragment != "" || destination.Path == "" {
		return "", errors.New("invalid Kinde portal destination")
	}
	return destination.String(), nil
}
