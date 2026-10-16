// Package keycloak implements Leamout's OpenID Connect integration.
package keycloak

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

type Config struct {
	Issuer       string `env:"ISSUER"`
	ClientID     string `env:"CLIENT_ID" envDefault:"leamout-console"`
	ClientSecret string `env:"CLIENT_SECRET"`
}

type Client struct {
	config         oauth2.Config
	verifier       *oidc.IDTokenVerifier
	http           *http.Client
	issuer         string
	logoutEndpoint string
}

type Identity struct {
	Subject       string
	SessionID     string `json:"sid"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

type Tokens struct {
	OAuth      oauth2.Token `json:"oauth"`
	IDToken    string       `json:"id_token"`
	Nonce      string       `json:"nonce"`
	ValidUntil time.Time    `json:"valid_until"`
}

func New(ctx context.Context, cfg Config, callbackURL string, development bool) (*Client, error) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, errors.New("keycloak client ID and secret are required")
	}
	for _, value := range []string{cfg.Issuer, callbackURL} {
		u, err := url.Parse(value)
		if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
			(u.Scheme != "https" && (!development || u.Scheme != "http" || (u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1"))) {
			return nil, errors.New("keycloak issuer and callback must be absolute HTTPS URLs (localhost HTTP is allowed in development)")
		}
	}
	client := &http.Client{Timeout: 10 * time.Second}
	provider, err := oidc.NewProvider(oidc.ClientContext(ctx, client), strings.TrimRight(cfg.Issuer, "/"))
	if err != nil {
		return nil, fmt.Errorf("discover Keycloak: %w", err)
	}
	var metadata struct {
		LogoutEndpoint string `json:"end_session_endpoint"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return nil, fmt.Errorf("read Keycloak metadata: %w", err)
	}
	if metadata.LogoutEndpoint == "" {
		return nil, errors.New("keycloak logout endpoint is missing")
	}
	return &Client{
		config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  callbackURL,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		verifier:       provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		http:           client,
		issuer:         strings.TrimRight(cfg.Issuer, "/"),
		logoutEndpoint: metadata.LogoutEndpoint,
	}, nil
}

func (c *Client) Issuer() string {
	return c.issuer
}

func (c *Client) AuthorizationURL(state, nonce, verifier string, register bool) string {
	options := []oauth2.AuthCodeOption{oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)}
	if register {
		options = append(options, oauth2.SetAuthURLParam("prompt", "create"))
	}
	return c.config.AuthCodeURL(state, options...)
}

func (c *Client) Exchange(ctx context.Context, code, verifier, nonce string) (Identity, Tokens, error) {
	token, err := c.config.Exchange(oidc.ClientContext(ctx, c.http), code, oauth2.VerifierOption(verifier))
	if err != nil {
		return Identity{}, Tokens{}, errors.New("keycloak code exchange failed")
	}
	identity, tokens, actualNonce, err := c.verify(ctx, token)
	if err != nil {
		return Identity{}, Tokens{}, err
	}
	if nonce == "" || actualNonce != nonce {
		return Identity{}, Tokens{}, errors.New("invalid Keycloak nonce")
	}
	return identity, tokens, nil
}

func (c *Client) Refresh(ctx context.Context, previous Tokens, subject, sessionID string) (Tokens, error) {
	// Force an exchange when the ID token expires before the access token.
	previous.OAuth.Expiry = time.Now().Add(-time.Minute)
	token, err := c.config.TokenSource(oidc.ClientContext(ctx, c.http), &previous.OAuth).Token()
	if err != nil {
		return Tokens{}, errors.New("keycloak session refresh failed")
	}
	identity, tokens, _, err := c.verify(ctx, token)
	if err != nil {
		return Tokens{}, err
	}
	if tokens.Nonce != "" && tokens.Nonce != previous.Nonce {
		return Tokens{}, errors.New("invalid refreshed Keycloak nonce")
	}
	if identity.Subject != subject || identity.SessionID != sessionID {
		return Tokens{}, errors.New("keycloak session subject changed")
	}
	return tokens, nil
}

func (c *Client) verify(ctx context.Context, token *oauth2.Token) (Identity, Tokens, string, error) {
	raw, ok := token.Extra("id_token").(string)
	if !ok || token.RefreshToken == "" || token.AccessToken == "" || token.Expiry.IsZero() {
		return Identity{}, Tokens{}, "", errors.New("incomplete Keycloak token response")
	}
	id, err := c.verifier.Verify(oidc.ClientContext(ctx, c.http), raw)
	if err != nil {
		return Identity{}, Tokens{}, "", errors.New("invalid Keycloak ID token")
	}
	if id.AccessTokenHash != "" {
		if err := id.VerifyAccessToken(token.AccessToken); err != nil {
			return Identity{}, Tokens{}, "", errors.New("invalid Keycloak access token binding")
		}
	}
	var identity Identity
	if err := id.Claims(&identity); err != nil {
		return Identity{}, Tokens{}, "", errors.New("invalid Keycloak identity claims")
	}
	identity.Subject = id.Subject
	if identity.SessionID == "" || identity.Subject == "" || strings.TrimSpace(identity.Email) == "" || !identity.EmailVerified {
		return Identity{}, Tokens{}, "", errors.New("a verified Keycloak email and subject are required")
	}
	expiry := id.Expiry
	if token.Expiry.Before(expiry) {
		expiry = token.Expiry
	}
	return identity, Tokens{OAuth: *token, IDToken: raw, Nonce: id.Nonce, ValidUntil: expiry}, id.Nonce, nil
}

func (c *Client) LogoutURL(idToken, returnURL string) string {
	values := url.Values{"client_id": {c.config.ClientID}, "post_logout_redirect_uri": {returnURL}}
	if idToken != "" {
		values.Set("id_token_hint", idToken)
	}
	return c.logoutEndpoint + "?" + values.Encode()
}

func (c *Client) VerifyLogout(ctx context.Context, raw string) (*oidc.LogoutToken, error) {
	event, err := c.verifier.VerifyLogout(oidc.ClientContext(ctx, c.http), raw)
	if err != nil {
		return nil, errors.New("invalid Keycloak logout token")
	}
	if event.IssuedAt.IsZero() || event.IssuedAt.After(time.Now().Add(time.Minute)) || event.IssuedAt.Before(time.Now().Add(-5*time.Minute)) {
		return nil, errors.New("invalid Keycloak logout issue time")
	}
	return event, nil
}
