package notifications

import (
	"strings"
	"testing"
)

func TestRendererRenderAuthOTP(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}

	email, err := renderer.RenderAuthOTP(
		"user@example.com",
		"Leamout <auth@leamout.com>",
		"support@leamout.com",
		AuthOTPData{Code: "482913", ExpiresMinutes: 10},
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, body := range []string{email.HTML, email.Text} {
		if !strings.Contains(body, "482913") {
			t.Fatalf("rendered body does not contain OTP: %q", body)
		}
		if !strings.Contains(body, "10") {
			t.Fatalf("rendered body does not contain expiry: %q", body)
		}
	}
	if email.Subject != "Your Leamout verification code" {
		t.Fatalf("subject = %q", email.Subject)
	}
}

func TestRendererRejectsInvalidAuthOTPData(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}

	for name, args := range map[string]struct {
		to   string
		from string
		data AuthOTPData
	}{
		"missing recipient": {from: "auth@leamout.com", data: AuthOTPData{Code: "123456", ExpiresMinutes: 10}},
		"missing sender":    {to: "user@example.com", data: AuthOTPData{Code: "123456", ExpiresMinutes: 10}},
		"missing code":      {to: "user@example.com", from: "auth@leamout.com", data: AuthOTPData{ExpiresMinutes: 10}},
		"invalid expiry":    {to: "user@example.com", from: "auth@leamout.com", data: AuthOTPData{Code: "123456"}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := renderer.RenderAuthOTP(args.to, args.from, "", args.data); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
