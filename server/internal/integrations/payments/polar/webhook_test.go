package polar

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func webhookHeaders(payload []byte, key []byte, timestamp int64) http.Header {
	headers := make(http.Header)
	headers.Set("webhook-id", "event-test")
	headers.Set("webhook-timestamp", fmt.Sprint(timestamp))
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("event-test." + fmt.Sprint(timestamp) + "." + string(payload)))
	headers.Set("webhook-signature", "v1,"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	return headers
}

func TestPolarWebhookSigningSchemesAndReplayWindow(t *testing.T) {
	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	key := []byte("a-test-webhook-key-with-32-bytes!!")
	secret := "whsec_" + base64.StdEncoding.EncodeToString(key)
	client := New("test-token", secret, "sandbox")
	client.now = func() time.Time {
		return now
	}
	payload := []byte(`{"type":"subscription.updated","timestamp":"2026-10-13T12:00:00Z","api_version":"2026-10","data":{"id":"subscription"}}`)
	for _, scheme := range []struct {
		name string
		key  []byte
	}{
		{
			"standard",
			key,
		},
		{
			"legacy",
			[]byte(secret),
		},
	} {
		t.Run(scheme.name, func(t *testing.T) {
			headers := webhookHeaders(payload, scheme.key, now.Unix())
			if _, err := client.ParseWebhook(payload, headers); err != nil {
				t.Fatal(err)
			}
			if _, err := client.ParseWebhook(append(payload, ' '), headers); err == nil {
				t.Fatal("tampered payload accepted")
			}
			headers.Set("webhook-id", "other-event")
			if _, err := client.ParseWebhook(payload, headers); err == nil {
				t.Fatal("tampered event identity accepted")
			}
			for _, timestamp := range []int64{
				now.Add(-6 * time.Minute).Unix(),
				now.Add(6 * time.Minute).Unix(),
			} {
				if _, err := client.ParseWebhook(payload, webhookHeaders(payload, scheme.key, timestamp)); err == nil {
					t.Fatal("webhook outside replay window accepted")
				}
			}
		})
	}
	if _, err := client.ParseWebhook(payload, make(http.Header)); err == nil {
		t.Fatal("unsigned webhook accepted")
	}
}
