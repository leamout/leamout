package email

import (
	"strings"
	"testing"
	"time"
)

func TestNotificationTemplates(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 30, 0, 0, time.UTC)
	data := Data{
		UserName:      "Alex <admin>",
		MemberName:    "Alex <admin>",
		Organization:  "Acme & Partners",
		Role:          "member",
		EventName:     "Password <changed>",
		OccurredAt:    now,
		KeyName:       "Production <key>",
		KeyExpiresAt:  now.Add(24 * time.Hour),
		TrunkName:     "Primary <trunk>",
		AgentName:     "Support <agent>",
		FailureReason: "Provider <unavailable>",
		ExpiresAt:     now.Add(time.Hour),
	}
	cases := []struct {
		name string
		text string
		html string
	}{
		{"welcome", "Alex <admin>", "Alex &lt;admin&gt;"},
		{"invitation-accepted", "Alex <admin>", "Alex &lt;admin&gt;"},
		{"security-alert", "Password <changed>", "Password &lt;changed&gt;"},
		{"api-key-expiry", "Production <key>", "Production &lt;key&gt;"},
		{"sip-trunk-failure", "Provider <unavailable>", "Provider &lt;unavailable&gt;"},
		{"voice-agent-failure", "Support <agent>", "Support &lt;agent&gt;"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message, err := NewRenderer().Render(tc.name, data)
			if err != nil {
				t.Fatal(err)
			}
			if message.Subject == "" || !strings.Contains(message.HTML, tc.html) || !strings.Contains(message.Text, tc.text) {
				t.Fatal("missing subject or notification data")
			}
			if strings.Contains(message.HTML, "LEAMOUT_") || strings.Contains(message.HTML, "{{.") {
				t.Fatal("unresolved placeholder")
			}
			if strings.Contains(message.HTML, "13:30 UTC") || strings.Contains(message.Text, "13:30 UTC") {
				t.Fatal("delivery deadline leaked into notification content")
			}
			if tc.name == "api-key-expiry" && (!strings.Contains(message.HTML, "08 Oct 2026") || !strings.Contains(message.Text, "08 Oct 2026")) {
				t.Fatal("API key expiry was not rendered independently of delivery expiry")
			}
			if _, err := NewRenderer().Render(tc.name, Data{ExpiresAt: data.ExpiresAt}); err == nil {
				t.Fatal("accepted missing notification data")
			}
		})
	}
}
