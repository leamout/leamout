package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestSignUsesTimestampedPayload(t *testing.T) {
	secret := []byte("a signing secret")
	body := []byte(`{"event":"call.created"}`)
	timestamp := time.Unix(1_700_000_000, 0)

	got := sign(secret, body, timestamp)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte("1700000000." + string(body)))
	want := "v1=" + hex.EncodeToString(mac.Sum(nil))
	if got != want {
		t.Fatalf("sign() = %q, want %q", got, want)
	}
}

func TestNormalizeEventsDeduplicatesAndRejectsInvalidNames(t *testing.T) {
	got, err := normalizeEvents([]string{" call.created ", "call.created", "call.ended"})
	if err != nil {
		t.Fatalf("normalizeEvents() error = %v", err)
	}
	if strings.Join(got, ",") != "call.created,call.ended" {
		t.Fatalf("normalizeEvents() = %v", got)
	}
	if _, err := normalizeEvents([]string{"call created"}); err == nil {
		t.Fatal("normalizeEvents() accepted an event name containing whitespace")
	}
}

func TestNormalizeURLRequiresHTTPS(t *testing.T) {
	if _, err := normalizeURL("http://example.test/hooks"); err == nil {
		t.Fatal("normalizeURL() accepted HTTP")
	}
	got, err := normalizeURL(" https://example.test/hooks ")
	if err != nil || got != "https://example.test/hooks" {
		t.Fatalf("normalizeURL() = %q, %v", got, err)
	}
}
