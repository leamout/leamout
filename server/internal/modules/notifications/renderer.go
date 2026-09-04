package notifications

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type AuthOTPData struct {
	Code           string
	ExpiresMinutes int
}

type Renderer struct {
	authOTPHTML *template.Template
}

func NewRenderer() (*Renderer, error) {
	htmlTemplate, err := template.ParseFS(templateFS, "templates/auth_otp.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse auth OTP HTML template: %w", err)
	}
	return &Renderer{authOTPHTML: htmlTemplate}, nil
}

func (r *Renderer) RenderAuthOTP(to, from, replyTo string, data AuthOTPData) (Email, error) {
	if strings.TrimSpace(to) == "" {
		return Email{}, fmt.Errorf("recipient email is required")
	}
	if strings.TrimSpace(from) == "" {
		return Email{}, fmt.Errorf("sender email is required")
	}
	if strings.TrimSpace(data.Code) == "" {
		return Email{}, fmt.Errorf("authentication code is required")
	}
	if data.ExpiresMinutes <= 0 {
		return Email{}, fmt.Errorf("authentication code expiry must be positive")
	}

	var htmlBody bytes.Buffer
	if err := r.authOTPHTML.Execute(&htmlBody, data); err != nil {
		return Email{}, fmt.Errorf("render auth OTP HTML: %w", err)
	}

	return Email{
		To:      to,
		From:    from,
		ReplyTo: replyTo,
		Subject: "Your Leamout verification code",
		HTML:    htmlBody.String(),
	}, nil
}
