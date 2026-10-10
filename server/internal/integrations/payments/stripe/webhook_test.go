package stripe

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestParseWebhook(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_800_000_000, 0)
	client := New("sk_test_example", "whsec_example")
	client.now = func() time.Time { return now }

	payload := []byte("{\"id\":\"evt_123\",\"type\":\"customer.subscription.updated\",\"created\":1800000000,\"data\":{\"object\":{\"id\":\"sub_123\"}}}")
	message := fmt.Sprintf("%d.%s", now.Unix(), payload)
	mac := hmac.New(sha256.New, []byte("whsec_example"))
	_, _ = mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil))

	event, err := client.ParseWebhook(payload, fmt.Sprintf("t=%d,v1=%s", now.Unix(), signature))
	if err != nil {
		t.Fatalf("ParseWebhook() error = %v", err)
	}
	if event.ID != "evt_123" {
		t.Fatalf("event ID = %q", event.ID)
	}
}

func TestParseWebhookRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_800_000_000, 0)
	client := New("sk_test_example", "whsec_example")
	client.now = func() time.Time { return now }

	_, err := client.ParseWebhook(
		[]byte("{\"id\":\"evt_123\",\"type\":\"customer.subscription.updated\"}"),
		fmt.Sprintf("t=%d,v1=deadbeef", now.Unix()),
	)
	if err == nil {
		t.Fatal("expected invalid signature to be rejected")
	}
}

func TestParseWebhookRejectsStaleTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_800_000_000, 0)
	client := New("sk_test_example", "whsec_example")
	client.now = func() time.Time { return now }

	payload := []byte("{\"id\":\"evt_123\",\"type\":\"customer.subscription.updated\"}")
	old := now.Add(-10 * time.Minute).Unix()
	message := fmt.Sprintf("%d.%s", old, payload)
	mac := hmac.New(sha256.New, []byte("whsec_example"))
	_, _ = mac.Write([]byte(message))

	_, err := client.ParseWebhook(
		payload,
		fmt.Sprintf("t=%d,v1=%s", old, hex.EncodeToString(mac.Sum(nil))),
	)
	if err == nil {
		t.Fatal("expected stale webhook timestamp to be rejected")
	}
}
