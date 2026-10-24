// Package kinde integrates Kinde's hosted OpenID Connect authorization server.
package kinde

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func New(
	_ context.Context,
	cfg Config,
	callbackURL string,
	development bool,
) (*Client, error) {
	if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, errors.New("kinde client ID and secret are required")
	}

	for _, value := range []string{cfg.Issuer, callbackURL} {
		u, err := url.Parse(value)
		if err != nil ||
			u.Host == "" ||
			u.User != nil ||
			u.RawQuery != "" ||
			u.Fragment != "" ||
			(u.Scheme != "https" &&
				(!development ||
					u.Scheme != "http" ||
					(u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1"))) {
			return nil, errors.New("kinde issuer and callback must be absolute HTTPS URLs (localhost HTTP is allowed in development)")
		}
	}

	issuer := strings.TrimRight(cfg.Issuer, "/")
	parsed, _ := url.Parse(issuer)
	if parsed.Path != "" {
		return nil, errors.New("kinde issuer must be an origin without a path")
	}

	return &Client{
		config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  issuer + "/oauth2/auth",
				TokenURL: issuer + "/oauth2/token",
			},
			RedirectURL: callbackURL,
			Scopes: []string{
				oidc.ScopeOpenID,
				"profile",
				"email",
				"offline",
			},
		},
		http:   &http.Client{Timeout: 10 * time.Second},
		issuer: issuer,
	}, nil
}

func (c *Client) Issuer() string {
	return c.issuer
}
