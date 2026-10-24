package kinde

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const maximumIdentityValidationAge = 10 * time.Minute

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

func (c *Client) LogoutURL(returnURL string) string {
	return c.issuer + "/logout?redirect=" + url.QueryEscape(returnURL)
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
