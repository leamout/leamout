package kinde

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
)

const (
	webhookMaxAge = 48 * time.Hour
	webhookClockSkew = 5 * time.Minute
)

// ErrWebhookUnavailable indicates that Kinde's signing keys could not be fetched.
// Webhook delivery should be retried rather than accepted in this case.
var ErrWebhookUnavailable = errors.New("kinde webhook signing keys are unavailable")

// WebhookUser contains only provider profile data used by Leamout.
type WebhookUser struct {
	ID          string  `json:"id"`
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	IsSuspended *bool   `json:"is_suspended"`
}

// WebhookEvent is a verified, signed Kinde identity change.
type WebhookEvent struct {
	EventID   string    `json:"event_id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		ID   string      `json:"id"`
		User WebhookUser `json:"user"`
	} `json:"data"`
}

func (e WebhookEvent) Subject() string {
	if e.Data.User.ID != "" {
		return e.Data.User.ID
	}
	return e.Data.ID
}

// VerifyWebhook validates Kinde's RS256 JWT against the issuer's published JWKS,
// then checks the signed event identity and replay time window. Webhook JWTs are
// not OIDC ID tokens and cannot be validated using the OAuth client audience.
func (c *Client) VerifyWebhook(ctx context.Context, raw string) (WebhookEvent, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) == 0 || len(raw) > 65536 {
		return WebhookEvent{}, errors.New("invalid Kinde webhook token length")
	}
	signed, err := jose.ParseSigned(raw, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil || len(signed.Signatures) != 1 ||
		signed.Signatures[0].Header.KeyID == "" {
		return WebhookEvent{}, errors.New("invalid Kinde webhook signature header")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.issuer+"/.well-known/jwks", nil)
	if err != nil {
		return WebhookEvent{}, ErrWebhookUnavailable
	}
	httpClient := *c.http
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return WebhookEvent{}, fmt.Errorf("%w: %v", ErrWebhookUnavailable, err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return WebhookEvent{}, ErrWebhookUnavailable
	}
	var keys jose.JSONWebKeySet
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&keys); err != nil {
		return WebhookEvent{}, ErrWebhookUnavailable
	}

	var payload []byte
	valid := false
	for _, key := range keys.Key(signed.Signatures[0].Header.KeyID) {
		if !key.Valid() || (key.Algorithm != "" && key.Algorithm != string(jose.RS256)) ||
			(key.Use != "" && key.Use != "sig") {
			continue
		}
		payload, err = signed.Verify(key.Key)
		if err == nil {
			valid = true
			break
		}
	}
	if !valid {
		return WebhookEvent{}, errors.New("invalid Kinde webhook signature")
	}
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return WebhookEvent{}, errors.New("invalid Kinde webhook claims")
	}
	now := time.Now()
	if !strings.HasPrefix(event.EventID, "event_") ||
		len(event.EventID) > 128 ||
		event.Timestamp.IsZero() ||
		event.Timestamp.Before(now.Add(-webhookMaxAge)) ||
		event.Timestamp.After(now.Add(webhookClockSkew)) {
		return WebhookEvent{}, errors.New("invalid or expired Kinde webhook event")
	}
	switch event.Type {
	case "user.updated", "user.deleted":
		if subject := event.Subject(); !strings.HasPrefix(subject, "kp_") &&
			!strings.HasPrefix(subject, "kp:") {
			return WebhookEvent{}, errors.New("invalid Kinde user identity")
		}
		if event.Data.ID != "" && event.Data.User.ID != "" &&
			event.Data.ID != event.Data.User.ID {
			return WebhookEvent{}, errors.New("conflicting Kinde user identities")
		}
	}
	return event, nil
}
