// Package kinde integrates Kinde's hosted OpenID Connect authorization server.
package kinde

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const maximumIdentityValidationAge = 10 * time.Minute

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

func (c *Client) AuthorizationURL(
	state string,
	nonce string,
	verifier string,
	register bool,
) string {
	options := []oauth2.AuthCodeOption{
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(verifier),
	}
	if register {
		options = append(options, oauth2.SetAuthURLParam("prompt", "create"))
	}

	return c.config.AuthCodeURL(state, options...)
}

// Discover verification keys on demand so an identity-provider outage cannot prevent
// voice/control-plane services from starting. Failed discovery is retried later.
func (c *Client) idVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.verifier == nil {
		provider, err := oidc.NewProvider(oidc.ClientContext(ctx, c.http), c.issuer)
		if err != nil {
			return nil, fmt.Errorf("discover Kinde: %w", err)
		}
		// Kinde can use its issuer as the ID token audience and put the OAuth
		// application identifier in azp. Audience validation is done below,
		// after go-oidc has verified issuer, signature and token expiry.
		c.verifier = provider.Verifier(&oidc.Config{SkipClientIDCheck: true})
	}

	return c.verifier, nil
}

func (c *Client) Exchange(
	ctx context.Context,
	code string,
	verifier string,
	nonce string,
) (Identity, Tokens, error) {
	if nonce == "" {
		return Identity{}, Tokens{}, errors.New("missing OIDC nonce")
	}

	access, err := c.config.Exchange(
		oidc.ClientContext(ctx, c.http),
		code,
		oauth2.VerifierOption(verifier),
	)
	if err != nil {
		return Identity{}, Tokens{}, errors.New("kinde code exchange failed")
	}

	identity, raw, actualNonce, expiration, err := c.verifyIDToken(ctx, access)
	if err != nil {
		return Identity{}, Tokens{}, err
	}
	if actualNonce != nonce {
		return Identity{}, Tokens{}, errors.New("invalid Kinde nonce")
	}
	if access.RefreshToken == "" ||
		access.AccessToken == "" ||
		access.Expiry.IsZero() {
		return Identity{}, Tokens{}, errors.New("incomplete Kinde token response")
	}

	if access.Expiry.Before(expiration) {
		expiration = access.Expiry
	}
	expiration = nextValidationDeadline(expiration)
	if !expiration.After(time.Now()) {
		return Identity{}, Tokens{}, errors.New("expired Kinde token response")
	}

	return identity, Tokens{
		OAuth:      *access,
		IDToken:    raw,
		Nonce:      actualNonce,
		ValidUntil: expiration,
	}, nil
}

func (c *Client) Refresh(
	ctx context.Context,
	previous Tokens,
	subject string,
) (Tokens, error) {
	if subject == "" || previous.OAuth.RefreshToken == "" {
		return Tokens{}, errors.New("invalid Kinde refresh context")
	}

	previous.OAuth.Expiry = time.Now().Add(-time.Minute)
	access, err := c.config.TokenSource(
		oidc.ClientContext(ctx, c.http),
		&previous.OAuth,
	).Token()
	if err != nil {
		return Tokens{}, errors.New("kinde session refresh failed")
	}
	if access.AccessToken == "" ||
		access.RefreshToken == "" ||
		access.Expiry.IsZero() {
		return Tokens{}, errors.New("incomplete Kinde refresh response")
	}

	result := Tokens{
		OAuth:      *access,
		IDToken:    previous.IDToken,
		Nonce:      previous.Nonce,
		ValidUntil: nextValidationDeadline(access.Expiry),
	}
	if !result.ValidUntil.After(time.Now()) {
		return Tokens{}, errors.New("expired Kinde refresh response")
	}

	// Kinde refresh responses can omit id_token. The rotating confidential-client
	// refresh token is then the session binding; retain previously verified identity.
	if _, ok := access.Extra("id_token").(string); ok {
		identity, raw, nonce, expiration, err := c.verifyIDToken(ctx, access)
		if err != nil {
			return Tokens{}, err
		}
		if identity.Subject != subject ||
			(nonce != "" && nonce != previous.Nonce) {
			return Tokens{}, errors.New("kinde session identity changed")
		}

		result.IDToken = raw
		if nonce != "" {
			result.Nonce = nonce
		}
		if expiration.Before(result.ValidUntil) {
			result.ValidUntil = expiration
		}
	}

	return result, nil
}

func (c *Client) verifyIDToken(
	ctx context.Context,
	access *oauth2.Token,
) (Identity, string, string, time.Time, error) {
	raw, ok := access.Extra("id_token").(string)
	if !ok || raw == "" || access.AccessToken == "" {
		return Identity{}, "", "", time.Time{}, errors.New("missing Kinde ID or access token")
	}

	verifier, err := c.idVerifier(ctx)
	if err != nil {
		return Identity{}, "", "", time.Time{}, err
	}

	id, err := verifier.Verify(oidc.ClientContext(ctx, c.http), raw)
	if err != nil {
		return Identity{}, "", "", time.Time{}, errors.New("invalid Kinde ID token")
	}
	if id.AccessTokenHash != "" {
		if err := id.VerifyAccessToken(access.AccessToken); err != nil {
			return Identity{}, "", "", time.Time{}, errors.New("invalid Kinde access token binding")
		}
	}

	var identity Identity
	if err := id.Claims(&identity); err != nil {
		return Identity{}, "", "", time.Time{}, errors.New("invalid Kinde identity claims")
	}
	identity.Subject = id.Subject
	if identity.Subject == "" ||
		strings.TrimSpace(identity.Email) == "" ||
		!identity.EmailVerified {
		return Identity{}, "", "", time.Time{}, errors.New("a verified Kinde email and subject are required")
	}
	if identity.Name == "" {
		identity.Name = strings.TrimSpace(identity.GivenName + " " + identity.FamilyName)
	}
	if err := c.validateAudience(id); err != nil {
		return Identity{}, "", "", time.Time{}, err
	}

	return identity, raw, id.Nonce, id.Expiry, nil
}

func (c *Client) LogoutURL(returnURL string) string {
	return c.issuer + "/logout?redirect=" + url.QueryEscape(returnURL)
}

func (c *Client) validateAudience(id *oidc.IDToken) error {
	var claims struct {
		AuthorizedParty string `json:"azp"`
	}
	if err := id.Claims(&claims); err != nil {
		return errors.New("invalid Kinde authorized party")
	}

	hasClientAudience := false
	hasIssuerAudience := false
	for _, audience := range id.Audience {
		switch audience {
		case c.config.ClientID:
			hasClientAudience = true
		case c.issuer:
			hasIssuerAudience = true
		}
	}

	// Standard OIDC ID tokens target the client ID. Kinde can instead use
	// the issuer as audience, in which case azp MUST identify our app.
	// Multi-audience tokens also require an explicit matching azp.
	if claims.AuthorizedParty != "" && claims.AuthorizedParty != c.config.ClientID {
		return errors.New("invalid Kinde authorized party")
	}
	if hasClientAudience && len(id.Audience) == 1 {
		return nil
	}
	if (hasClientAudience || hasIssuerAudience) &&
		claims.AuthorizedParty == c.config.ClientID {
		return nil
	}
	return errors.New("invalid Kinde ID token audience")
}

// nextValidationDeadline bounds the period during which a local Leamout
// session can avoid refreshing its provider credentials, even if Kinde issues
// longer-lived access tokens. Kinde remains the session's authority.
func nextValidationDeadline(providerExpiry time.Time) time.Time {
	deadline := time.Now().Add(maximumIdentityValidationAge)
	if providerExpiry.Before(deadline) {
		return providerExpiry
	}
	return deadline
}

// PortalURL requests a one-time user portal URL using an authenticated account
// API request. Browser clients must never receive the OAuth access token.
func (c *Client) PortalURL(ctx context.Context, accessToken, returnURL string) (string, error) {
	if accessToken == "" {
		return "", errors.New("missing Kinde access token")
	}
	target, err := url.Parse(returnURL)
	if err != nil || target.User != nil || target.Host == "" ||
		(target.Scheme != "https" && !(target.Scheme == "http" &&
			(target.Hostname() == "localhost" || target.Hostname() == "127.0.0.1"))) {
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
