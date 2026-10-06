package email

import (
	"strings"
	"testing"
	"time"
)

func TestRendererOTP(t *testing.T) {
	m, err := NewRenderer().Render("otp", Data{Code: "012345", ExpiresAt: time.Date(2026, 10, 6, 12, 30, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{m.HTML, m.Text} {
		if !strings.Contains(body, "012345") || !strings.Contains(body, "12:30 UTC") {
			t.Fatal("missing code or expiry")
		}
	}
}
func TestRendererEscapesInvitation(t *testing.T) {
	m, err := NewRenderer().Render("invitation", Data{Organization: "<script>alert(1)</script>", Inviter: "Alice", Role: "member", AcceptURL: "https://app.example.com/invitation?token=abc", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(m.HTML, "<script>") || !strings.Contains(m.HTML, "&lt;script&gt;") {
		t.Fatal("organization was not escaped")
	}
}
func TestRendererRejectsInvalidData(t *testing.T) {
	for _, tc := range []struct {
		name string
		data Data
	}{{"unknown", Data{}}, {"otp", Data{Code: "bad"}}, {"invitation", Data{Organization: "Acme", AcceptURL: "javascript:alert(1)", ExpiresAt: time.Now()}}} {
		if _, err := NewRenderer().Render(tc.name, tc.data); err == nil {
			t.Fatalf("accepted %s", tc.name)
		}
	}
}

func TestRendererPreservesInvitationURL(t *testing.T) {
	url := "https://app.example.com/invitation?token=a%2Bb&organization=acme"
	m, err := NewRenderer().Render("invitation", Data{
		Organization: "Acme & Partners",
		Inviter:      "Alice <admin>",
		Role:         "member",
		AcceptURL:    url,
		ExpiresAt:    time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(m.HTML, `href="https://app.example.com/invitation?token=a%2Bb&amp;organization=acme"`) {
		t.Fatal("invitation link was not preserved with HTML attribute escaping")
	}
	if !strings.Contains(m.HTML, "Acme &amp; Partners") || !strings.Contains(m.HTML, "Alice &lt;admin&gt;") {
		t.Fatal("invitation fields were not escaped")
	}
	if !strings.Contains(m.Text, url) {
		t.Fatal("plain-text invitation URL changed")
	}
	if strings.Contains(m.HTML, "LEAMOUT_") || strings.Contains(m.HTML, "{{.") {
		t.Fatal("unresolved template placeholders")
	}
}
