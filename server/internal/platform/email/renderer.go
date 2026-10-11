package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
	texttemplate "text/template"

	"github.com/leamout/leamout/server/internal/integrations/ses"
)

//go:embed templates/*.html
var templates embed.FS

type Renderer struct{}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) Render(name string, data Data) (ses.Message, error) {
	if err := validateData(name, data); err != nil {
		return ses.Message{}, err
	}

	source, err := templates.ReadFile("templates/" + name + ".html")
	if err != nil {
		return ses.Message{}, err
	}

	// Render the subject as text so HTML escaping does not alter the header.
	subjectTemplate, err := texttemplate.New(name).Parse(string(source))
	if err != nil {
		return ses.Message{}, err
	}

	bodyTemplate, err := template.New(name).Parse(string(source))
	if err != nil {
		return ses.Message{}, err
	}

	values := struct {
		Data
		Expiry string
	}{
		Data:   data,
		Expiry: data.ExpiresAt.UTC().Format("15:04 UTC on 02 Jan 2006"),
	}

	var subject bytes.Buffer
	if err := subjectTemplate.ExecuteTemplate(&subject, "subject", values); err != nil {
		return ses.Message{}, err
	}

	subjectValue := strings.TrimSpace(subject.String())
	if subjectValue == "" || strings.ContainsAny(subjectValue, "\r\n") {
		return ses.Message{}, fmt.Errorf("invalid email subject")
	}

	var body bytes.Buffer
	if err := bodyTemplate.ExecuteTemplate(&body, "body", values); err != nil {
		return ses.Message{}, err
	}

	return ses.Message{
		Subject: subjectValue,
		HTML:    body.String(),
	}, nil
}
