package kinde

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

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
