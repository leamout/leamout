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
	if session.Provider != "paystack" || session.NextAction != paymentprovider.NextActionWait || session.Reference != "invoice-1" {
		t.Fatalf("session = %+v", session)
	}
}

func TestContinueCheckoutSubmitsOTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/charge/submit_otp" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		payload, _ := io.ReadAll(r.Body)
		for _, expected := range []string{`"otp":"123456"`, `"reference":"invoice-1"`} {
			if !strings.Contains(string(payload), expected) {
				t.Fatalf("body %s does not contain %s", payload, expected)
			}
		}
		_, _ = w.Write([]byte(`{"status":true,"message":"Charge attempted","data":{"id":42,"reference":"invoice-1","status":"success","gateway_response":"Approved"}}`))
	}))
	defer server.Close()
	client, _ := NewClient(Config{BaseURL: server.URL, SecretKey: "test-secret", HTTPClient: server.Client()})
	session, err := client.ContinueCheckout(context.Background(), paymentprovider.ContinueCheckoutRequest{
		Reference: "invoice-1", Action: paymentprovider.NextActionSubmitOTP, Value: "123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	if session.Status != paymentprovider.StatusSucceeded || session.NextAction != paymentprovider.NextActionNone {
		t.Fatalf("session = %+v", session)
	}
}

func TestCreateCheckoutMapsMobileMoneyAuthorization(t *testing.T) {
	if action := nextAction("pay_offline"); action != paymentprovider.NextActionAuthorizeMobileMoney {
		t.Fatalf("action = %s", action)
	}
}

func TestGetCheckoutReturnsProviderContinuation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":true,"message":"Charge attempted","data":{"id":42,"reference":"invoice-1","status":"pay_offline","display_text":"Approve the payment on your phone"}}`))
	}))
	defer server.Close()
	client, _ := NewClient(Config{BaseURL: server.URL, SecretKey: "test-secret", HTTPClient: server.Client()})
	session, err := client.GetCheckout(context.Background(), "invoice-1")
	if err != nil {
		t.Fatal(err)
	}
	if session.NextAction != paymentprovider.NextActionAuthorizeMobileMoney || session.Message != "Approve the payment on your phone" {
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
	header := signedHeader(payload)
	client, _ := NewClient(Config{SecretKey: "test-secret"})
	event, err := client.ParseWebhook(payload, header)
	if err != nil {
		t.Fatal(err)
	}
	if event.ProviderEventID != "charge.success:42" || event.Type != "charge.success" || event.Payment.Reference != "invoice-1" || event.Payment.Status != paymentprovider.StatusSucceeded {
		t.Fatalf("event = %+v", event)
	}
	header.Set("x-paystack-signature", strings.Repeat("0", sha512.Size*2))
	if _, err := client.ParseWebhook(payload, header); err == nil {
		t.Fatal("expected invalid signature error")
	}
}

func TestParseWebhookRejectsMissingIdentity(t *testing.T) {
	client, _ := NewClient(Config{SecretKey: "test-secret"})
	for _, payload := range [][]byte{
		[]byte(`{"data":{"id":42}}`),
		[]byte(`{"event":"   ","data":{"id":42}}`),
		[]byte(`{"event":"charge.success","data":{}}`),
	} {
		if _, err := client.ParseWebhook(payload, signedHeader(payload)); err == nil || !strings.Contains(err.Error(), "event identity") {
			t.Fatalf("payload %s error = %v", payload, err)
		}
	}
}

func signedHeader(payload []byte) http.Header {
	mac := hmac.New(sha512.New, []byte("test-secret"))
	_, _ = mac.Write(payload)
	header := http.Header{}
	header.Set("x-paystack-signature", hex.EncodeToString(mac.Sum(nil)))
	return header
}
