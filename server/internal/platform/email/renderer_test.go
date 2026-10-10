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

func TestGeneratedInvitationPreservesDynamicData(t *testing.T) {
	data := Data{Organization: `Acme {{.Code}} & Co`, Inviter: `Alice <Admin>`, Role: "member", AcceptURL: "https://app.example.com/invitation?token=abc&next=console", ExpiresAt: time.Now().Add(time.Hour)}
	message, err := NewRenderer().Render("invitation", data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(message.HTML, "Acme {{.Code}} &amp; Co") || strings.Contains(message.HTML, "<Admin>") {
		t.Fatal("dynamic text was not safely escaped")
	}
	if !strings.Contains(message.HTML, "token=abc&amp;next=console") {
		t.Fatal("dynamic URL was not escaped")
	}
	if !strings.Contains(message.Text, data.AcceptURL) || !strings.Contains(message.Text, data.Organization) {
		t.Fatal("plain text lost dynamic data")
	}
	if strings.Contains(message.HTML, "SENTINEL") || strings.Contains(message.Text, "SENTINEL") {
		t.Fatal("export placeholder leaked")
	}
}
