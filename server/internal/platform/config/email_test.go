package config

import (
	"github.com/caarlos0/env/v11"
	"testing"
)

func TestEmailEnvironment(t *testing.T) {
	t.Setenv("EMAIL_ENABLED", "true")
	t.Setenv("AWS_REGION", "eu-west-1")
	t.Setenv("EMAIL_FROM", "Leamout <sender@example.com>")
	t.Setenv("EMAIL_REPLY_TO", "support@example.com")
	t.Setenv("SES_CONFIGURATION_SET", "transactional")
	cfg, err := env.ParseAs[EmailConfig]()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || cfg.Region != "eu-west-1" || cfg.From != "Leamout <sender@example.com>" || cfg.ReplyTo != "support@example.com" || cfg.ConfigurationSet != "transactional" {
		t.Fatal("email environment not parsed")
	}
}
