package email

import (
	"strings"
	"testing"
	"time"
)

func TestRendererEscapesInvitation(t *testing.T) {
	m, err := NewRenderer().Render("organization-invitation", Data{Organization: "<script>alert(1)</script>", Inviter: "Alice", Role: "member", AcceptURL: "https://app.example.com/invitation?token=abc", ExpiresAt: time.Now().Add(time.Hour)})
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
	}{{"unknown", Data{}}, {"organization-invitation", Data{Organization: "Acme", AcceptURL: "javascript:alert(1)", ExpiresAt: time.Now()}}} {
		if _, err := NewRenderer().Render(tc.name, tc.data); err == nil {
			t.Fatalf("accepted %s", tc.name)
		}
	}
}

func TestInvitationPreservesDynamicData(t *testing.T) {
	data := Data{Organization: `Acme {{.Code}} & Co`, Inviter: `Alice <Admin>`, Role: "member", AcceptURL: "https://app.example.com/invitation?token=abc&next=console", ExpiresAt: time.Now().Add(time.Hour)}
	message, err := NewRenderer().Render("organization-invitation", data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(message.HTML, "Acme {{.Code}} &amp; Co") || strings.Contains(message.HTML, "<Admin>") {
		t.Fatal("dynamic text was not safely escaped")
	}
	if !strings.Contains(message.HTML, "token=abc&amp;next=console") {
		t.Fatal("dynamic URL was not escaped")
	}

	if strings.Contains(message.HTML, "SENTINEL") {
		t.Fatal("export placeholder leaked")
	}
}

func TestSubjectComesFromTemplate(t *testing.T) {
	data := Data{Organization: "Acme", Inviter: "Alice", Role: "member", AcceptURL: "https://console.example.com/invitations/accept", ExpiresAt: time.Now().Add(time.Hour)}
	message, err := NewRenderer().Render("organization-invitation", data)
	if err != nil {
		t.Fatal(err)
	}
	if message.Subject != "You’re invited to join Acme on Leamout" {
		t.Fatal("subject did not render organization name")
	}
	data.Organization = "Acme\r\nBcc: other@example.com"
	if _, err := NewRenderer().Render("organization-invitation", data); err == nil {
		t.Fatal("accepted a multiline subject")
	}
	if _, err := NewRenderer().Render("../organization-invitation", data); err == nil {
		t.Fatal("accepted a template path")
	}
}

func TestLocalDevelopmentLinks(t *testing.T) {
	for _, name := range []string{"organization-invitation"} {
		t.Run(name, func(t *testing.T) {
			link := "http://localhost:3000/forgot-password"
			data := Data{
				Organization: "Acme",
				AcceptURL:    link,
				ExpiresAt:    time.Now().Add(time.Hour),
			}
			message, err := NewRenderer().Render(name, data)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(message.HTML, link) {
				t.Fatal("local URL was not preserved")
			}
		})
	}
}
