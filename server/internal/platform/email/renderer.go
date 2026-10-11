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

//go:embed templates/*
var templates embed.FS

type Renderer struct{}

func NewRenderer() *Renderer { return &Renderer{} }
func (r *Renderer) Render(name string, data Data) (ses.Message, error) {
	if err := validateData(name, data); err != nil {
		return ses.Message{}, err
	}
	subjectTemplate, err := texttemplate.ParseFS(templates, "templates/"+name+".subject.txt")
	if err != nil {
		return ses.Message{}, err
	}
	html, err := template.ParseFS(templates, "templates/"+name+".html")
	if err != nil {
		return ses.Message{}, err
	}
	plain, err := texttemplate.ParseFS(templates, "templates/"+name+".txt")
	if err != nil {
		return ses.Message{}, err
	}
	values := struct {
		Data
		Expiry string
	}{Data: data, Expiry: data.ExpiresAt.UTC().Format("15:04 UTC on 02 Jan 2006")}
	var h, t, subject bytes.Buffer
	if err := html.ExecuteTemplate(&h, name+".html", values); err != nil {
		return ses.Message{}, err
	}
	if err := plain.Execute(&t, values); err != nil {
		return ses.Message{}, err
	}
	if err := subjectTemplate.Execute(&subject, values); err != nil {
		return ses.Message{}, err
	}
	subjectValue := strings.TrimSpace(subject.String())
	if subjectValue == "" || strings.ContainsAny(subjectValue, "\r\n") {
		return ses.Message{}, fmt.Errorf("invalid email subject")
	}
	return ses.Message{Subject: subjectValue, HTML: h.String(), Text: t.String()}, nil
}
