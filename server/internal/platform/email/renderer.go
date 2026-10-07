package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	texttemplate "text/template"
)

//go:embed templates/*
var templates embed.FS
var codePattern = regexp.MustCompile(`^[0-9]{6}$`)

type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) Render(name string, data Data) (Message, error) {
	subject := ""
	switch name {
	case "otp":
		if !codePattern.MatchString(data.Code) {
			return Message{}, fmt.Errorf("invalid OTP template data")
		}
		subject = "Your Leamout sign-in code"
	case "invitation":
		u, err := url.Parse(data.AcceptURL)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || data.Organization == "" {
			return Message{}, fmt.Errorf("invalid invitation template data")
		}
		subject = "You’re invited to join an organization on Leamout"
	case "welcome":
		if data.UserName == "" {
			return Message{}, fmt.Errorf("welcome user name is required")
		}
		subject = "Welcome to Leamout"
	case "invitation-accepted":
		if data.MemberName == "" || data.Organization == "" || data.Role == "" {
			return Message{}, fmt.Errorf("invalid invitation acceptance template data")
		}
		subject = "Your Leamout invitation was accepted"
	case "security-alert":
		if data.EventName == "" || data.OccurredAt.IsZero() {
			return Message{}, fmt.Errorf("invalid security alert template data")
		}
		subject = "Leamout account security alert"
	case "api-key-expiry":
		if data.KeyName == "" || data.Organization == "" || data.KeyExpiresAt.IsZero() {
			return Message{}, fmt.Errorf("invalid API key expiry template data")
		}
		subject = "Your Leamout API key is expiring"
	case "sip-trunk-failure":
		if data.TrunkName == "" || data.Organization == "" || data.FailureReason == "" || data.OccurredAt.IsZero() {
			return Message{}, fmt.Errorf("invalid SIP trunk failure template data")
		}
		subject = "Leamout SIP trunk needs attention"
	case "voice-agent-failure":
		if data.AgentName == "" || data.Organization == "" || data.FailureReason == "" || data.OccurredAt.IsZero() {
			return Message{}, fmt.Errorf("invalid voice agent failure template data")
		}
		subject = "Leamout voice agent needs attention"
	default:
		return Message{}, fmt.Errorf("unknown email template")
	}
	if (name == "otp" || name == "invitation") && data.ExpiresAt.IsZero() {
		return Message{}, fmt.Errorf("email expiry is required")
	}
	html, err := template.ParseFS(templates, "templates/"+name+".html")
	if err != nil {
		return Message{}, err
	}
	plain, err := texttemplate.ParseFS(templates, "templates/"+name+".txt")
	if err != nil {
		return Message{}, err
	}
	var h, t bytes.Buffer
	if err := html.Execute(&h, data); err != nil {
		return Message{}, err
	}
	if err := plain.Execute(&t, data); err != nil {
		return Message{}, err
	}
	return Message{Subject: subject, HTML: h.String(), Text: t.String()}, nil
}
