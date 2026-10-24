package kinde

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPortalURL(t *testing.T) {
	var host string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/account_api/v1/portal_link" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer provider-access-token" {
			t.Errorf("authorization = %q", got)
		}
		if got := r.URL.Query().Get("subnav"); got != "profile" {
			t.Errorf("subnav = %q", got)
		}
		if got := r.URL.Query().Get("return_url"); got != "https://console.example.com/settings/profile" {
			t.Errorf("return_url = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"url": host + "/portal?token=one-time",
		}); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	host = server.URL

	client, err := New(t.Context(), Config{Issuer: server.URL, ClientID: "client", ClientSecret: "secret"}, "http://localhost/callback", true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.PortalURL(t.Context(), "provider-access-token", "https://console.example.com/settings/profile")
	if err != nil {
		t.Fatal(err)
	}
	if got != server.URL+"/portal?token=one-time" {
		t.Fatalf("portal URL = %q", got)
	}
}

func TestPortalURLRejectsUntrustedHost(t *testing.T) {
	for _, candidate := range []string{
		"https://evil.example.com/portal",
		"http://tenant.kinde.com/portal",
		"https://tenant.kinde.com.evil.test/portal",
		"https://user@tenant.kinde.com/portal",
		"javascript:alert(1)",
		"/relative",
	} {
		t.Run(candidate, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]string{"url": candidate})
			}))
			defer server.Close()
			client, err := New(t.Context(), Config{Issuer: server.URL, ClientID: "client", ClientSecret: "secret"}, "http://localhost/callback", true)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.PortalURL(t.Context(), "token", "https://console.example.com/settings/profile"); err == nil {
				t.Fatal("accepted unsafe Kinde portal destination")
			}
		})
	}
}

func TestPortalURLDoesNotFollowRedirect(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("authorization header must never reach redirected destination")
	}))
	defer target.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/token-stealing-target", http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	client, err := New(t.Context(), Config{Issuer: server.URL, ClientID: "client", ClientSecret: "secret"}, "http://localhost/callback", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.PortalURL(context.Background(), "secret-token", "https://console.example.com/settings/profile"); err == nil {
		t.Fatal("accepted redirected account API response")
	}
}

func TestPortalURLRejectsMissingAccessTokenAndBadReturnURL(t *testing.T) {
	client := &Client{issuer: "https://tenant.kinde.com"}
	if _, err := client.PortalURL(t.Context(), "", "https://console.example.com"); err == nil {
		t.Fatal("accepted missing access token")
	}
	for _, invalid := range []string{"javascript:alert(1)", "//evil.example.com", "http://evil.example.com", "https://user@evil.example.com"} {
		if _, err := client.PortalURL(t.Context(), "token", invalid); err == nil {
			t.Fatalf("accepted invalid return URL: %s", invalid)
		}
	}
}

func TestPortalURLReturnParameterIsEncoded(t *testing.T) {
	urlString := "https://console.example.com/settings/profile"
	parsed, err := url.Parse(urlString)
	if err != nil || parsed.String() != urlString {
		t.Fatal("invalid fixture")
	}
}
