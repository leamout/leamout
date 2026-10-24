//go:build kinde_live

package kinde

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// This suite contacts a real staging Kinde tenant. It is deliberately excluded
// from ordinary go test ./... and never runs against synthetic discovery.
func liveKindeConfig(t *testing.T) Config {
	t.Helper()
	cfg := Config{
		Issuer:       strings.TrimSpace(os.Getenv("KINDE_LIVE_ISSUER")),
		ClientID:     strings.TrimSpace(os.Getenv("KINDE_LIVE_CLIENT_ID")),
		ClientSecret: os.Getenv("KINDE_LIVE_CLIENT_SECRET"),
	}
	if cfg.Issuer == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
		t.Fatal("set KINDE_LIVE_ISSUER, KINDE_LIVE_CLIENT_ID and KINDE_LIVE_CLIENT_SECRET for the staging tenant")
	}
	return cfg
}

func liveCallback(t *testing.T, name string) string {
	t.Helper()
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		t.Fatalf("%s must be a registered HTTPS callback URL", name)
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		t.Fatalf("%s must be a registered HTTPS callback URL", name)
	}
	return value
}

type liveOIDCMetadata struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

func fetchLiveJSON(t *testing.T, ctx context.Context, endpoint string, target any) {
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal("construct staging provider request:", err)
	}
	httpClient := &http.Client{
		Timeout: 12 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	response, err := httpClient.Do(request)
	if err != nil {
		t.Fatal("request staging provider metadata:", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("staging provider metadata returned HTTP %d", response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(target); err != nil {
		t.Fatal("decode staging provider metadata:", err)
	}
}

func validateLiveEndpoint(t *testing.T, raw string) {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		t.Fatal("OIDC discovery advertised an invalid HTTPS endpoint")
	}
}

func liveMetadata(t *testing.T, ctx context.Context, cfg Config) liveOIDCMetadata {
	t.Helper()
	var metadata liveOIDCMetadata
	fetchLiveJSON(t, ctx, strings.TrimRight(cfg.Issuer, "/")+"/.well-known/openid-configuration", &metadata)
	if metadata.Issuer != strings.TrimRight(cfg.Issuer, "/") {
		t.Fatal("OIDC discovery issuer differs from the configured issuer")
	}
	for _, endpoint := range []string{
		metadata.AuthorizationEndpoint,
		metadata.TokenEndpoint,
		metadata.JWKSURI,
	} {
		validateLiveEndpoint(t, endpoint)
	}
	return metadata
}

func TestLiveKindeDiscovery(t *testing.T) {
	cfg := liveKindeConfig(t)
	callback := liveCallback(t, "KINDE_LIVE_CONSOLE_CALLBACK_URL")
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Second)
	defer cancel()

	client, err := New(ctx, cfg, callback, false)
	if err != nil {
		t.Fatal("staging Kinde client configuration:", err)
	}
	metadata := liveMetadata(t, ctx, cfg)
	// Fail if hosted discovery and the hard-coded production integration drift.
	if client.config.Endpoint.AuthURL != metadata.AuthorizationEndpoint ||
		client.config.Endpoint.TokenURL != metadata.TokenEndpoint {
		t.Fatal("Kinde production OAuth endpoints differ from live OIDC discovery")
	}
	provider, err := oidc.NewProvider(oidc.ClientContext(ctx, client.http), client.issuer)
	if err != nil {
		t.Fatal("OIDC library could not discover live Kinde provider:", err)
	}
	if provider.Endpoint().AuthURL != metadata.AuthorizationEndpoint ||
		provider.Endpoint().TokenURL != metadata.TokenEndpoint {
		t.Fatal("OIDC provider endpoints differ from live metadata")
	}
	if _, err := client.idVerifier(ctx); err != nil {
		t.Fatal("OIDC verifier could not initialize from live provider:", err)
	}

	var keys struct {
		Keys []struct {
			KeyID string `json:"kid"`
			Type  string `json:"kty"`
		} `json:"keys"`
	}
	fetchLiveJSON(t, ctx, metadata.JWKSURI, &keys)
	if len(keys.Keys) == 0 {
		t.Fatal("live Kinde JWKS contained no public signing keys")
	}
	for _, key := range keys.Keys {
		if key.KeyID == "" || key.Type == "" {
			t.Fatal("live Kinde JWKS contains a signing key without kid or kty")
		}
	}
}

func TestLiveKindeAuthorizationAndLogout(t *testing.T) {
	cfg := liveKindeConfig(t)
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Second)
	defer cancel()
	metadata := liveMetadata(t, ctx, cfg)

	for _, entry := range []struct {
		name string
		env  string
	}{
		{"console", "KINDE_LIVE_CONSOLE_CALLBACK_URL"},
		{"backoffice", "KINDE_LIVE_BACKOFFICE_CALLBACK_URL"},
	} {
		t.Run(entry.name, func(t *testing.T) {
			callback := liveCallback(t, entry.env)
			client, err := New(ctx, cfg, callback, false)
			if err != nil {
				t.Fatal("create staging Kinde client:", err)
			}
			verifier := oauth2.GenerateVerifier()
			for _, register := range []bool{false, true} {
				authorization, err := url.Parse(client.AuthorizationURL("staging-state", "staging-nonce", verifier, register))
				if err != nil {
					t.Fatal("parse authorization URL:", err)
				}
				expected, err := url.Parse(metadata.AuthorizationEndpoint)
				if err != nil || authorization.Scheme != expected.Scheme ||
					authorization.Host != expected.Host || authorization.Path != expected.Path {
					t.Fatal("authorization URL does not use the discovered Kinde endpoint")
				}
				values := authorization.Query()
				if values.Get("client_id") != cfg.ClientID ||
					values.Get("redirect_uri") != callback ||
					values.Get("response_type") != "code" ||
					values.Get("state") != "staging-state" ||
					values.Get("nonce") != "staging-nonce" ||
					values.Get("code_challenge") == "" ||
					values.Get("code_challenge_method") != "S256" {
					t.Fatal("authorization URL lacks the expected OAuth code, state, nonce or PKCE parameters")
				}
				scopes := strings.Fields(values.Get("scope"))
				for _, required := range []string{"openid", "profile", "email", "offline"} {
					found := false
					for _, scope := range scopes {
						if scope == required {
							found = true
						}
					}
					if !found {
						t.Fatalf("authorization URL is missing required scope %q", required)
					}
				}
				if register && values.Get("prompt") != "create" {
					t.Fatal("registration flow does not use prompt=create")
				}
				if !register && values.Get("prompt") == "create" {
					t.Fatal("normal login unexpectedly requests registration")
				}
			}

			u, _ := url.Parse(callback)
			returnURL := u.Scheme + "://" + u.Host + "/log-in"
			logout, err := url.Parse(client.LogoutURL(returnURL))
			if err != nil || logout.Scheme != u.Scheme && logout.Scheme != "https" {
				t.Fatal("invalid Kinde logout URL")
			}
			issuer, _ := url.Parse(cfg.Issuer)
			if logout.Host != issuer.Host || logout.Path != "/logout" ||
				logout.Query().Get("redirect") != returnURL {
				t.Fatal("logout endpoint or return URL is incorrect")
			}
		})
	}
}

// These network tests exercise an actual deployed Leamout staging environment,
// without creating a user, running a browser or consuming a refresh token.
func TestLiveKindeDeployedEntryPoints(t *testing.T) {
	console := strings.TrimSpace(os.Getenv("KINDE_LIVE_CONSOLE_URL"))
	backoffice := strings.TrimSpace(os.Getenv("KINDE_LIVE_BACKOFFICE_URL"))
	if console == "" && backoffice == "" {
		t.Skip("set both KINDE_LIVE_CONSOLE_URL and KINDE_LIVE_BACKOFFICE_URL to test staging deployments")
	}
	if console == "" || backoffice == "" {
		t.Fatal("both staging application origins must be provided")
	}
	cfg := liveKindeConfig(t)
	for _, entry := range []struct {
		name       string
		origin     string
		loginPath  string
		accountURL string
		callback   string
	}{
		{"console", console, "/api/v1/auth/login", "/api/v1/auth/account", liveCallback(t, "KINDE_LIVE_CONSOLE_CALLBACK_URL")},
		{"backoffice", backoffice, "/auth/login", "/auth/account", liveCallback(t, "KINDE_LIVE_BACKOFFICE_CALLBACK_URL")},
	} {
		t.Run(entry.name, func(t *testing.T) {
			origin, err := url.Parse(entry.origin)
			if err != nil || origin.Scheme != "https" || origin.Host == "" ||
				origin.User != nil || origin.Path != "" || origin.RawQuery != "" {
				t.Fatal("staging app URL must be an HTTPS origin")
			}
			client := &http.Client{
				Timeout: 12 * time.Second,
				CheckRedirect: func(*http.Request, []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			for _, path := range []string{entry.loginPath, entry.loginPath + "?register=true"} {
				request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, entry.origin+path, nil)
				if err != nil {
					t.Fatal(err)
				}
				response, err := client.Do(request)
				if err != nil {
					t.Fatal("staging login route unavailable:", err)
				}
				_ = response.Body.Close()
				if response.StatusCode != http.StatusFound {
					t.Fatalf("staging login route returned HTTP %d, expected 302", response.StatusCode)
				}
				location, err := url.Parse(response.Header.Get("Location"))
				issuer, _ := url.Parse(cfg.Issuer)
				if err != nil || location.Scheme != issuer.Scheme ||
					location.Host != issuer.Host || location.Path != "/oauth2/auth" ||
					location.Query().Get("redirect_uri") != entry.callback {
					t.Fatal("staging login did not redirect to the expected Kinde OAuth endpoint")
				}
				cookieState := ""
				for _, cookie := range response.Cookies() {
					if cookie.Name == "leamout-oidc-state" {
						if !cookie.HttpOnly || !cookie.Secure {
							t.Fatal("staging login state cookie lacks Secure or HttpOnly")
						}
						cookieState = cookie.Value
					}
				}
				if cookieState == "" || cookieState != location.Query().Get("state") {
					t.Fatal("staging login did not bind OIDC state to the browser cookie")
				}
				if strings.Contains(path, "register=true") &&
					location.Query().Get("prompt") != "create" {
					t.Fatal("staging registration route did not request prompt=create")
				}
			}
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, entry.origin+entry.accountURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := client.Do(request)
			if err != nil {
				t.Fatal("staging account route unavailable:", err)
			}
			_ = response.Body.Close()
			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("unauthenticated staging account route returned HTTP %d, expected 401", response.StatusCode)
			}
		})
	}
}

func TestLiveKindeConfiguredIssuerIsNotPlaceholder(t *testing.T) {
	cfg := liveKindeConfig(t)
	parsed, err := url.Parse(cfg.Issuer)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Host == "" ||
		strings.HasPrefix(parsed.Hostname(), "your-tenant.") {
		t.Fatal("KINDE_LIVE_ISSUER must be a real HTTPS staging tenant")
	}
	if _, err := New(context.Background(), cfg, liveCallback(t, "KINDE_LIVE_CONSOLE_CALLBACK_URL"), false); err != nil {
		t.Fatal(fmt.Errorf("invalid staging client setup: %w", err))
	}
}

// TestLiveKindeAuthorizationCodeExchange is a manual, single-use staging flow.
// Its optional inputs are deliberately not wired to GitHub Actions secrets.
// Run only after generating the PKCE verifier, nonce and code in a real login.
func TestLiveKindeAuthorizationCodeExchange(t *testing.T) {
	code := strings.TrimSpace(os.Getenv("KINDE_LIVE_AUTH_CODE"))
	verifier := strings.TrimSpace(os.Getenv("KINDE_LIVE_PKCE_VERIFIER"))
	nonce := strings.TrimSpace(os.Getenv("KINDE_LIVE_NONCE"))
	if code == "" && verifier == "" && nonce == "" {
		t.Skip("real authorization-code exchange requires single-use KINDE_LIVE_AUTH_CODE, KINDE_LIVE_PKCE_VERIFIER and KINDE_LIVE_NONCE")
	}
	if code == "" || verifier == "" || nonce == "" {
		t.Fatal("provide all three single-use authorization-code inputs")
	}
	cfg := liveKindeConfig(t)
	callback := liveCallback(t, "KINDE_LIVE_CONSOLE_CALLBACK_URL")
	client, err := New(t.Context(), cfg, callback, false)
	if err != nil {
		t.Fatal("initialize staging client:", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 35*time.Second)
	defer cancel()
	identity, tokens, err := client.Exchange(ctx, code, verifier, nonce)
	if err != nil {
		t.Fatal("live Kinde authorization code was rejected:", err)
	}
	if identity.Subject == "" || !identity.EmailVerified ||
		tokens.OAuth.RefreshToken == "" || !tokens.ValidUntil.After(time.Now()) {
		t.Fatal("live identity lacks a verified subject, refresh token or valid session deadline")
	}
	// Refresh is deliberately opt-in: it consumes/rotates the provider token,
	// which makes a persisted copy unsafe to reuse in a second run.
	if os.Getenv("KINDE_LIVE_TEST_REFRESH") == "true" {
		refreshed, err := client.Refresh(ctx, tokens, identity.Subject)
		if err != nil {
			t.Fatal("live refresh-token rotation failed:", err)
		}
		if refreshed.OAuth.RefreshToken == "" || !refreshed.ValidUntil.After(time.Now()) {
			t.Fatal("live refresh returned invalid credentials")
		}
		tokens = refreshed
	}
	// The portal is a separate optional capability of the staging tenant.
	if os.Getenv("KINDE_LIVE_TEST_PORTAL") == "true" {
		returnURL := strings.TrimSpace(os.Getenv("KINDE_LIVE_PORTAL_RETURN_URL"))
		if returnURL == "" {
			t.Fatal("set KINDE_LIVE_PORTAL_RETURN_URL for a registered return URL")
		}
		if _, err := client.PortalURL(ctx, tokens.OAuth.AccessToken, returnURL); err != nil {
			t.Fatal("live Kinde user portal request failed:", err)
		}
	}
	// Do not print, persist or commit code, ID token, access token, refresh
	// token, subject, profile information or the one-time portal URL.
}
