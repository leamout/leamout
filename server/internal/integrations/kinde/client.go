// Package kinde integrates Kinde's hosted OpenID Connect authorization server.
package kinde

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func New(_ context.Context, cfg Config, callbackURL string, development bool) (*Client, error) {
	if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, errors.New("Kinde client ID and secret are required")
	}
	for _, value := range []string{cfg.Issuer, callbackURL} {
		u, err := url.Parse(value)
		if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
			(u.Scheme != "https" && (!development || u.Scheme != "http" ||
				(u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1"))) {
			return nil, errors.New("Kinde issuer and callback must be absolute HTTPS URLs (localhost HTTP is allowed in development)")
		}
	}
	issuer := strings.TrimRight(cfg.Issuer, "/")
	parsed, _ := url.Parse(issuer)
	if parsed.Path != "" {
		return nil, errors.New("Kinde issuer must be an origin without a path")
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
			Scopes:      []string{oidc.ScopeOpenID, "profile", "email", "offline"},
		},
		http:   &http.Client{Timeout: 10 * time.Second},
		issuer: issuer,
	}, nil
}

func (c *Client) Issuer() string { return c.issuer }

func (c *Client) AuthorizationURL(state, nonce, verifier string, register bool) string {
	options := []oauth2.AuthCodeOption{oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)}
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
		c.verifier = provider.Verifier(&oidc.Config{ClientID: c.config.ClientID})
	}
	return c.verifier, nil
}

func (c *Client) Exchange(ctx context.Context, code, verifier, nonce string) (Identity, Tokens, error) {
	if nonce == "" {
		return Identity{}, Tokens{}, errors.New("missing OIDC nonce")
	}
	access, err := c.config.Exchange(oidc.ClientContext(ctx, c.http), code, oauth2.VerifierOption(verifier))
	if err != nil {
		return Identity{}, Tokens{}, errors.New("Kinde code exchange failed")
	}
	identity, raw, actualNonce, expiration, err := c.verifyIDToken(ctx, access)
	if err != nil {
		return Identity{}, Tokens{}, err
	}
	if actualNonce != nonce {
		return Identity{}, Tokens{}, errors.New("invalid Kinde nonce")
	}
	if access.RefreshToken == "" || access.AccessToken == "" || access.Expiry.IsZero() {
		return Identity{}, Tokens{}, errors.New("incomplete Kinde token response")
	}
	if access.Expiry.Before(expiration) {
		expiration = access.Expiry
	}
	return identity, Tokens{OAuth: *access, IDToken: raw, Nonce: actualNonce, ValidUntil: expiration}, nil
}

func (c *Client) Refresh(ctx context.Context, previous Tokens, subject string) (Tokens, error) {
	if subject == "" || previous.OAuth.RefreshToken == "" {
		return Tokens{}, errors.New("invalid Kinde refresh context")
	}
	previous.OAuth.Expiry = time.Now().Add(-time.Minute)
	access, err := c.config.TokenSource(oidc.ClientContext(ctx, c.http), &previous.OAuth).Token()
	if err != nil {
		return Tokens{}, errors.New("Kinde session refresh failed")
	}
	if access.AccessToken == "" || access.RefreshToken == "" || access.Expiry.IsZero() {
		return Tokens{}, errors.New("incomplete Kinde refresh response")
	}
	result := Tokens{
		OAuth: *access, IDToken: previous.IDToken, Nonce: previous.Nonce,
		ValidUntil: access.Expiry,
	}
	// Kinde refresh responses can omit id_token. The rotating confidential-client
	// refresh token is then the session binding; retain previously verified identity.
	if _, ok := access.Extra("id_token").(string); ok {
		identity, raw, nonce, expiration, err := c.verifyIDToken(ctx, access)
		if err != nil {
			return Tokens{}, err
		}
		if identity.Subject != subject || (nonce != "" && nonce != previous.Nonce) {
			return Tokens{}, errors.New("Kinde session identity changed")
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

func (c *Client) verifyIDToken(ctx context.Context, access *oauth2.Token) (Identity, string, string, time.Time, error) {
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
	if identity.Subject == "" || strings.TrimSpace(identity.Email) == "" || !identity.EmailVerified {
		return Identity{}, "", "", time.Time{}, errors.New("a verified Kinde email and subject are required")
	}
	if identity.Name == "" {
		identity.Name = strings.TrimSpace(identity.GivenName + " " + identity.FamilyName)
	}
	return identity, raw, id.Nonce, id.Expiry, nil
}

func (c *Client) LogoutURL(returnURL string) string {
	return c.issuer + "/logout?redirect=" + url.QueryEscape(returnURL)
}
