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

func NewRenderer() *Renderer { return &Renderer{} }
func (r *Renderer) Render(name string, data Data) (Message, error) {
	subject := ""
	switch name {
	case "verification-code":
		if !codePattern.MatchString(data.Code) {
			return Message{}, fmt.Errorf("invalid OTP template data")
		}
		subject = "Your Leamout verification code"
	case "organization-invitation":
		u, err := url.Parse(data.AcceptURL)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || data.Organization == "" {
			return Message{}, fmt.Errorf("invalid invitation template data")
		}
		subject = "You’re invited to join an organization on Leamout"
	case "password-changed":
		u, err := url.Parse(data.RecoveryURL)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || data.ChangedAt == "" {
			return Message{}, fmt.Errorf("invalid password change template data")
		}
		subject = "Your Leamout password was changed"
	default:
		return Message{}, fmt.Errorf("unknown email template")
	}
	if data.ExpiresAt.IsZero() {
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
	values := struct {
		Data
		Expiry string
	}{Data: data, Expiry: data.ExpiresAt.UTC().Format("15:04 UTC on 02 Jan 2006")}
	var h, t bytes.Buffer
	if err := html.ExecuteTemplate(&h, name+".html", values); err != nil {
		return Message{}, err
	}
	if err := plain.Execute(&t, values); err != nil {
		return Message{}, err
	}
	return Message{Subject: subject, HTML: h.String(), Text: t.String()}, nil
}
