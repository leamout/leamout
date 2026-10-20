package egress

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseHTTPSRejectsUnsafeDestinations(t *testing.T) {
	for _, endpoint := range []string{
		"http://example.com/hook",
		"https://user:secret@example.com/hook",
		"https://example.com/hook#fragment",
		"https://127.0.0.1/hook",
		"https://10.0.0.2/hook",
		"https://169.254.169.254/latest/meta-data/",
		"https://100.64.1.2/hook",
		"https://[::1]/hook",
		"https://[::ffff:127.0.0.1]/hook",
	} {
		if _, err := ParseHTTPS(endpoint); err == nil {
			t.Errorf("ParseHTTPS(%q) accepted unsafe destination", endpoint)
		}
	}
	for _, endpoint := range []string{"https://example.com/hook", "https://8.8.8.8/hook"} {
		if _, err := ParseHTTPS(endpoint); err != nil {
			t.Errorf("ParseHTTPS(%q): %v", endpoint, err)
		}
	}
}

func TestValidateAddressRejectsInternalNetworks(t *testing.T) {
	for _, raw := range []string{
		"127.0.0.1", "10.0.0.1", "172.16.0.1",
		"192.168.1.1", "169.254.169.254", "100.64.0.1",
		"::1", "fc00::1", "fe80::1", "::ffff:127.0.0.1",
	} {
		if err := ValidateAddress(netip.MustParseAddr(raw)); err == nil {
			t.Errorf("ValidateAddress(%q) accepted private destination", raw)
		}
	}
	if err := ValidateAddress(netip.MustParseAddr("8.8.8.8")); err != nil {
		t.Fatalf("ValidateAddress(public): %v", err)
	}
}

func TestProxyCannotBypassWebhookDestinationPolicy(t *testing.T) {
	var forwarded atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		forwarded.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer proxy.Close()

	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("https_proxy", proxy.URL)
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("http_proxy", proxy.URL)
	t.Setenv("ALL_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")

	transport := NewTransport()
	defer transport.CloseIdleConnections()
	if transport.Proxy != nil {
		t.Fatal("outbound webhook transport must not use an environment proxy")
	}
	client := &http.Client{Transport: transport, Timeout: time.Second}

	// A proxied CONNECT could otherwise ask the proxy to reach this private
	// origin; direct dialing must reject it before making any network request.
	request, err := http.NewRequest(http.MethodPost, "https://10.20.30.40/internal", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(request)
	if err == nil || !strings.Contains(err.Error(), "non-public address") {
		t.Fatalf("private destination response error = %v, want policy rejection", err)
	}
	if got := forwarded.Load(); got != 0 {
		t.Fatalf("proxy received %d requests to private destination, want zero", got)
	}
}
