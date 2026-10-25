package kinde

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
)

func TestVerifyWebhookSignedEvents(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: key},
		(&jose.SignerOptions{}).WithHeader("kid", "staging-key"),
	)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/jwks" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{
				{Key: &key.PublicKey, KeyID: "staging-key", Algorithm: "RS256", Use: "sig"},
			},
		})
	}))
	defer server.Close()
	client, err := New(t.Context(), Config{
		Issuer: server.URL, ClientID: "client", ClientSecret: "secret",
	}, "http://localhost:8080/callback", true)
	if err != nil {
		t.Fatal(err)
	}
	makeJWT := func(eventID string, timestamp time.Time, eventType string) string {
		t.Helper()
		payload, err := json.Marshal(map[string]any{
			"event_id":  eventID,
			"type":      eventType,
			"timestamp": timestamp.UTC().Format(time.RFC3339Nano),
			"data": map[string]any{
				"user": map[string]any{
					"id":           "kp_1234567890",
					"email":        "updated@example.com",
					"first_name":   "Ada",
					"last_name":    "Lovelace",
					"is_suspended": false,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := signer.Sign(payload)
		if err != nil {
			t.Fatal(err)
		}
		token, err := encoded.CompactSerialize()
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	event, err := client.VerifyWebhook(t.Context(), makeJWT("event_test", time.Now(), "user.updated"))
	if err != nil {
		t.Fatal(err)
	}
	if event.Subject() != "kp_1234567890" ||
		event.Data.User.FirstName == nil || *event.Data.User.FirstName != "Ada" {
		t.Fatal("verified Kinde user payload did not decode")
	}
	for _, tc := range []struct {
		name string
		raw  string
	}{
		{"invalid signature", "invalid.jwt.signature"},
		{"expired", makeJWT("event_old", time.Now().Add(-72*time.Hour), "user.updated")},
		{"future", makeJWT("event_future", time.Now().Add(time.Hour), "user.updated")},
		{"bad event id", makeJWT("bad-event", time.Now(), "user.updated")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := client.VerifyWebhook(t.Context(), tc.raw); err == nil {
				t.Fatal("untrusted webhook was accepted")
			}
		})
	}
}

func TestVerifyWebhookCachesSigningKeysUnderConcurrentRequests(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: key},
		(&jose.SignerOptions{}).WithHeader("kid", "stable-key"),
	)
	if err != nil {
		t.Fatal(err)
	}
	var fetches atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fetches.Add(1)
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "stable-key", Algorithm: "RS256", Use: "sig"}},
		})
	}))
	defer server.Close()
	client, err := New(t.Context(), Config{Issuer: server.URL, ClientID: "client", ClientSecret: "secret"}, "http://localhost/callback", true)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{
		"event_id": "event_concurrent", "type": "user.deleted",
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"data":      map[string]any{"user": map[string]any{"id": "kp_123"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	jws, err := signer.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}
	token, err := jws.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 24 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := client.VerifyWebhook(t.Context(), token); err != nil {
				t.Errorf("verify cached webhook: %v", err)
			}
		}()
	}
	wg.Wait()
	if got := fetches.Load(); got != 1 {
		t.Fatalf("concurrent webhook verification fetched JWKS %d times, want 1", got)
	}
}

func TestVerifyWebhookRejectsUnknownKeyWithoutRepeatedFetch(t *testing.T) {
	var fetches atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fetches.Add(1)
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{{KeyID: "known-key", Algorithm: "RS256", Use: "sig"}},
		})
	}))
	defer server.Close()
	client, err := New(t.Context(), Config{Issuer: server.URL, ClientID: "client", ClientSecret: "secret"}, "http://localhost/callback", true)
	if err != nil {
		t.Fatal(err)
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", "unknown-key"))
	if err != nil {
		t.Fatal(err)
	}
	jws, err := signer.Sign([]byte(`{"event_id":"event_unknown","type":"user.deleted","timestamp":"2026-10-24T13:00:00Z"}`))
	if err != nil {
		t.Fatal(err)
	}
	token, err := jws.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}
	for range 20 {
		if _, err := client.VerifyWebhook(t.Context(), token); err == nil {
			t.Fatal("untrusted webhook accepted")
		}
	}
	if got := fetches.Load(); got != 1 {
		t.Fatalf("forged key IDs caused %d JWKS requests, want 1", got)
	}
}
