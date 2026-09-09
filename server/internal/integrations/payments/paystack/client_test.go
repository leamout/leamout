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

func TestCreateCheckoutInitializesPaystackTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/transaction/initialize" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-secret" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		payload, _ := io.ReadAll(r.Body)
		body := string(payload)
		for _, expected := range []string{`"amount":2500`, `"currency":"GHS"`, `"reference":"invoice-1"`} {
			if !strings.Contains(body, expected) {
				t.Fatalf("body %s does not contain %s", body, expected)
			}
		}
		_, _ = w.Write([]byte(`{"status":true,"message":"Authorization URL created","data":{"id":42,"reference":"invoice-1","access_code":"access-1","authorization_url":"https://checkout.paystack.com/access-1"}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, SecretKey: "test-secret", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	session, err := client.CreateCheckout(context.Background(), paymentprovider.CheckoutRequest{
		Reference: "invoice-1", AmountMinor: 2500, Currency: "GHS", Email: "buyer@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if session.Provider != "paystack" || session.AccessCode != "access-1" || session.Reference != "invoice-1" {
		t.Fatalf("session = %+v", session)
	}
}

func TestGetPaymentVerifiesPaystackTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction/verify/invoice-1" {
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
