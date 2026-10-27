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

func TestRendererSupportsUsageAllowanceNotifications(t *testing.T) {
	cases := []struct {
		name, subject, expected string
		data                    Data
	}{
		{
			name:     "usage-allowance-warning",
			subject:  "Your Leamout runtime allowance is 80% used",
			expected: "has used 800 of 1000 runtime minutes (80%)",
			data:     Data{Organization: "Acme & Co", UsedMinutes: 800, AllowanceMinutes: 1000, UsagePercent: 80, PeriodEnd: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), BillingURL: "https://console.example.com/settings/billing", ExpiresAt: time.Now().Add(24 * time.Hour)},
		},
		{
			name:     "usage-allowance-exhausted",
			subject:  "Your Leamout runtime allowance has been exhausted",
			expected: "full runtime allowance of 1000 minutes",
			data:     Data{Organization: "Acme & Co", UsedMinutes: 1000, AllowanceMinutes: 1000, UsagePercent: 100, PeriodEnd: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), BillingURL: "https://console.example.com/settings/billing", ExpiresAt: time.Now().Add(24 * time.Hour)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message, err := NewRenderer().Render(tc.name, tc.data)
			if err != nil {
				t.Fatal(err)
			}
			if message.Subject != tc.subject {
				t.Fatalf("subject = %q, want %q", message.Subject, tc.subject)
			}
			if !strings.Contains(message.HTML, tc.expected) || !strings.Contains(message.HTML, "Acme &amp; Co") {
				t.Fatalf("rendered usage notification missing expected escaped content: %s", message.HTML)
			}
			if !strings.Contains(message.HTML, "https://console.example.com/settings/billing") {
				t.Fatal("billing URL missing from rendered message")
			}
		})
	}
}

func TestRendererRejectsInvalidUsageAllowanceNotifications(t *testing.T) {
	_, err := NewRenderer().Render("usage-allowance-warning", Data{
		Organization:     "Acme",
		UsedMinutes:      800,
		AllowanceMinutes: 1000,
		UsagePercent:     100,
		PeriodEnd:        time.Now().Add(time.Hour),
		BillingURL:       "javascript:alert(1)",
		ExpiresAt:        time.Now().Add(time.Hour),
	})
	if err == nil {
		t.Fatal("accepted invalid usage allowance warning data")
	}
}
