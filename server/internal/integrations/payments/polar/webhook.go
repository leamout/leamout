package polar

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const webhookTolerance = 5 * time.Minute

func (c *Client) ParseWebhook(payload []byte, headers http.Header) (Event, error) {
	if c == nil || c.webhookSecret == "" {
		return Event{}, fmt.Errorf("polar webhook secret is not configured")
	}
	id := headers.Get("webhook-id")
	rawTimestamp := headers.Get("webhook-timestamp")
	seconds, err := strconv.ParseInt(rawTimestamp, 10, 64)
	if err != nil || seconds <= 0 || id == "" {
		return Event{}, fmt.Errorf("invalid Polar webhook identity or timestamp")
	}
	delta := c.now().Sub(time.Unix(seconds, 0))
	if delta > webhookTolerance || delta < -webhookTolerance {
		return Event{}, fmt.Errorf("polar webhook timestamp is outside tolerance")
	}
	// Polar documents Standard Webhooks for new secrets and full-string HMAC
	// for pre-September-2026 secrets. Like its SDK, accept either verified key.
	keys := [][]byte{
		[]byte(c.webhookSecret),
	}
	if decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(c.webhookSecret, "whsec_")); err == nil && len(decoded) > 0 {
		keys = append(keys, decoded)
	}
	message := id + "." + rawTimestamp + "." + string(payload)
	verified := false
	for _, key := range keys {
		mac := hmac.New(sha256.New, key)
		_, _ = mac.Write([]byte(message))
		expected := mac.Sum(nil)
		for _, candidate := range strings.Fields(headers.Get("webhook-signature")) {
			version, encoded, ok := strings.Cut(candidate, ",")
			if !ok || version != "v1" {
				continue
			}
			signature, decodeErr := base64.StdEncoding.DecodeString(encoded)
			if decodeErr == nil && hmac.Equal(expected, signature) {
				verified = true
			}
		}
	}
	if !verified {
		return Event{}, fmt.Errorf("polar webhook signature mismatch")
	}
	var event Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return Event{}, fmt.Errorf("decode Polar webhook: %w", err)
	}
	if event.Type == "" || event.Timestamp.IsZero() || len(event.Data) == 0 || (event.APIVersion != "" && event.APIVersion != APIVersion) {
		return Event{}, fmt.Errorf("invalid or unsupported Polar event")
	}
	return event, nil
}
