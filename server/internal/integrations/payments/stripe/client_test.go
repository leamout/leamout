package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	paymentprovider "github.com/leamout/leamout/internal/commercial/payments"
)

func TestCreateCheckoutCreatesCardCheckoutSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/checkout/sessions" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		username, _, _ := r.BasicAuth()
		if username != "sk_test" || r.Header.Get("Idempotency-Key") != "wallet-topup-1" ||
			r.Header.Get("Stripe-Version") != DefaultAPIVersion {
			t.Fatalf("authentication/idempotency headers are incorrect")
		}
		payload, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(payload))
		if values.Get("mode") != "payment" || values.Get("ui_mode") != "elements" ||
			values.Get("line_items[0][price_data][unit_amount]") != "2500" ||
			values.Get("line_items[0][price_data][currency]") != "usd" ||
			values.Get("payment_method_types[]") != "card" {
			t.Fatalf("form = %v", values)
		}
		_, _ = w.Write([]byte(`{"id":"cs_1","client_secret":"cs_1_secret","amount_total":2500,"currency":"usd","payment_status":"unpaid","status":"open","metadata":{"leamout_reference":"wallet-topup-1"}}`))
	}))
	defer server.Close()
	client, err := NewClient(Config{BaseURL: server.URL + "/v1", SecretKey: "sk_test", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	session, err := client.CreateCheckout(context.Background(), paymentprovider.CheckoutRequest{
		Reference: "wallet-topup-1", AmountMinor: 2500, Currency: "USD", Email: "buyer@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if session.ProviderID != "cs_1" || session.ClientSecret != "cs_1_secret" || session.Status != paymentprovider.StatusPending {
		t.Fatalf("session = %+v", session)
	}
}

func TestParseWebhookAuthenticatesStripePayloadAndTimestamp(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	payload := []byte(`{"id":"evt_1","type":"checkout.session.completed","data":{"object":{"id":"cs_1","amount_total":2500,"currency":"usd","payment_status":"paid","status":"complete","metadata":{"leamout_reference":"wallet-topup-1"}}}}`)
	mac := hmac.New(sha256.New, []byte("whsec_test"))
	_, _ = mac.Write(append([]byte("1800000000."), payload...))
	header := http.Header{"Stripe-Signature": []string{"t=1800000000,v1=" + hex.EncodeToString(mac.Sum(nil))}}
	client, _ := NewClient(Config{SecretKey: "sk_test", WebhookSecret: "whsec_test", Now: func() time.Time { return now }})
	event, err := client.ParseWebhook(payload, header)
	if err != nil {
		t.Fatal(err)
	}
	if event.ProviderEventID != "evt_1" || event.Payment.Reference != "wallet-topup-1" || event.Payment.Status != paymentprovider.StatusSucceeded {
		t.Fatalf("event = %+v", event)
	}
	header.Set("Stripe-Signature", "t=1700000000,v1="+strings.Repeat("0", sha256.Size*2))
	if _, err := client.ParseWebhook(payload, header); err == nil || !strings.Contains(err.Error(), "timestamp") {
		t.Fatalf("expected timestamp error, got %v", err)
	}
}
