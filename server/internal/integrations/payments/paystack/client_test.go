package paystack

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

func TestCreateCheckoutCreatesMobileMoneyCharge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/charge" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-secret" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		payload, _ := io.ReadAll(r.Body)
		body := string(payload)
		for _, expected := range []string{`"amount":"2500"`, `"reference":"invoice-1"`, `"mobile_money":{"phone":"0240000000","provider":"mtn"}`} {
			if !strings.Contains(body, expected) {
				t.Fatalf("body %s does not contain %s", body, expected)
			}
		}
		_, _ = w.Write([]byte(`{"status":true,"message":"Charge attempted","data":{"id":42,"reference":"invoice-1","status":"pending","message":"Authorize the payment on your phone"}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, SecretKey: "test-secret", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	session, err := client.CreateCheckout(context.Background(), paymentprovider.CheckoutRequest{
		Reference: "invoice-1", AmountMinor: 2500, Currency: "GHS", Email: "buyer@example.com",
		MobileMoney: &paymentprovider.MobileMoney{Phone: "0240000000", Provider: "mtn"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if session.Provider != "paystack" || session.NextAction != "pending" || session.Reference != "invoice-1" {
		t.Fatalf("session = %+v", session)
	}
}

func TestGetPaymentChecksPaystackCharge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/charge/invoice-1" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":true,"data":{"id":42,"reference":"invoice-1","amount":2500,"currency":"GHS","status":"success"}}`))
	}))
	defer server.Close()
	client, _ := NewClient(Config{BaseURL: server.URL, SecretKey: "test-secret", HTTPClient: server.Client()})
	payment, err := client.GetPayment(context.Background(), "invoice-1")
	if err != nil {
		t.Fatal(err)
	}
	if payment.Status != paymentprovider.StatusSucceeded || payment.AmountMinor != 2500 || payment.Currency != "GHS" {
		t.Fatalf("payment = %+v", payment)
	}
}

func TestCreateCheckoutRequiresMobileMoneyFields(t *testing.T) {
	client, _ := NewClient(Config{SecretKey: "test-secret"})
	base := paymentprovider.CheckoutRequest{Reference: "invoice-1", AmountMinor: 2500, Currency: "GHS", Email: "buyer@example.com"}
	if _, err := client.CreateCheckout(context.Background(), base); err == nil || !strings.Contains(err.Error(), "details are required") {
		t.Fatalf("missing details error = %v", err)
	}
	base.MobileMoney = &paymentprovider.MobileMoney{Phone: "0240000000", Provider: "unknown"}
	if _, err := client.CreateCheckout(context.Background(), base); err == nil || !strings.Contains(err.Error(), "mtn, atl, or vod") {
		t.Fatalf("provider error = %v", err)
	}
}

func TestParseWebhookAuthenticatesPaystackPayload(t *testing.T) {
	payload := []byte(`{"event":"charge.success","data":{"id":42,"reference":"invoice-1","amount":2500,"currency":"GHS","status":"success"}}`)
	mac := hmac.New(sha512.New, []byte("test-secret"))
	_, _ = mac.Write(payload)
	header := http.Header{}
	header.Set("x-paystack-signature", hex.EncodeToString(mac.Sum(nil)))
	client, _ := NewClient(Config{SecretKey: "test-secret"})
	event, err := client.ParseWebhook(payload, header)
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != "charge.success" || event.Payment.Reference != "invoice-1" || event.Payment.Status != paymentprovider.StatusSucceeded {
		t.Fatalf("event = %+v", event)
	}
	header.Set("x-paystack-signature", strings.Repeat("0", sha512.Size*2))
	if _, err := client.ParseWebhook(payload, header); err == nil {
		t.Fatal("expected invalid signature error")
	}
}
