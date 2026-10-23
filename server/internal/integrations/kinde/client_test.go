package kinde

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
)

// Exercise discovery, JWKS verification, token exchange and rotation over HTTP.
func TestKindeExchangeAndRefresh(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", "test"))
	if err != nil {
		t.Fatal(err)
	}
	var issuer string
	var mutate func(map[string]any)
	refreshCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			writeJSON(t, w, map[string]any{
				"issuer":                                issuer,
				"authorization_endpoint":                issuer + "/oauth2/auth",
				"token_endpoint":                        issuer + "/oauth2/token",
				"jwks_uri":                              issuer + "/keys",
				"end_session_endpoint":                  issuer + "/logout",
				"id_token_signing_alg_values_supported": []string{"RS256"},
			})
		case "/keys":
			writeJSON(t, w, jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "test", Algorithm: "RS256", Use: "sig"}}})
		case "/token":
			if err := r.ParseForm(); err != nil {
				t.Error(err)
				return
			}
			clientID, secret, ok := r.BasicAuth()
			if !ok || clientID != "console" || secret != "secret" {
				t.Error("missing confidential client authentication")
			}
			if r.Form.Get("grant_type") == "authorization_code" && r.Form.Get("code_verifier") != "verifier" {
				t.Error("missing PKCE verifier")
			}
			if r.Form.Get("grant_type") == "refresh_token" {
				refreshCount++
				if r.Form.Get("refresh_token") != "refresh-one" {
					t.Error("wrong refresh token")
				}
			}
			claims := map[string]any{
				"iss": issuer, "aud": "console", "sub": "subject-one", "sid": "session-one",
				"exp": time.Now().Add(time.Minute).Unix(), "iat": time.Now().Unix(),
				"nonce": "nonce", "email": "user@example.com", "email_verified": true,
			}
			if mutate != nil {
				mutate(claims)
			}
			raw := signClaims(t, signer, claims)
			writeJSON(t, w, map[string]any{
				"access_token": "access", "refresh_token": "refresh-one", "token_type": "Bearer", "expires_in": 60, "id_token": raw,
			})
		default:
			http.NotFound(w, r)
		}
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	issuer = server.URL
	client, err := New(t.Context(), Config{Issuer: issuer, ClientID: "console", ClientSecret: "secret"}, "http://localhost:3001/api/v1/auth/callback", true)
	if err != nil {
		t.Fatal(err)
	}
	authorization, err := url.Parse(client.AuthorizationURL("state", "nonce", "verifier", false))
	if err != nil {
		t.Fatal(err)
	}
	if authorization.Query().Get("code_challenge_method") != "S256" || authorization.Query().Get("state") != "state" || authorization.Query().Get("nonce") != "nonce" || !strings.Contains(authorization.Query().Get("scope"), "offline") {
		t.Fatal("authorization request is missing state, nonce or PKCE")
	}
	identity, tokens, err := client.Exchange(t.Context(), "code", "verifier", "nonce")
	if err != nil {
		t.Fatal(err)
	}
	if identity.Subject != "subject-one" {
		t.Fatal("wrong identity")
	}
	if _, err := client.Refresh(t.Context(), tokens, identity.Subject); err != nil {
		t.Fatal(err)
	}
	if refreshCount != 1 {
		t.Fatal("refresh did not exchange tokens")
	}
	if _, err := client.Refresh(t.Context(), tokens, "another-subject"); err == nil {
		t.Fatal("accepted changed subject")
	}
	if _, _, err := client.Exchange(t.Context(), "code", "verifier", "wrong-nonce"); err == nil {
		t.Fatal("accepted wrong nonce")
	}
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"issuer", func(c map[string]any) { c["iss"] = "https://wrong.example.com" }},
		{"audience", func(c map[string]any) { c["aud"] = "another-client" }},
		{"expired", func(c map[string]any) { c["exp"] = time.Now().Add(-time.Minute).Unix() }},
		{"unverified", func(c map[string]any) { c["email_verified"] = false }},
		{"subject", func(c map[string]any) { c["sub"] = "" }},
		
	} {
		t.Run(test.name, func(t *testing.T) {
			mutate = test.mutate
			if _, _, err := client.Exchange(t.Context(), "code", "verifier", "nonce"); err == nil {
				t.Fatal("accepted invalid identity")
			}
		})
	}
	mutate = nil
	if got := client.LogoutURL("https://console.example.com/log-in"); !strings.Contains(got, "/logout?redirect=") {
		t.Fatal("Kinde logout redirect is not formed")
	}

}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Error(err)
	}
}

func signClaims(t *testing.T, signer jose.Signer, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := signed.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestProductionRejectsInsecureConfiguration(t *testing.T) {
	if _, err := New(t.Context(), Config{Issuer: "http://localhost:8180", ClientID: "console", ClientSecret: "secret"}, "https://console.example.com/callback", false); err == nil {
		t.Fatal("accepted insecure production issuer")
	}
}
