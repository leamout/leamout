package config

import (
	"testing"

	"github.com/caarlos0/env/v11"
)

func TestEmailEnvironment(t *testing.T) {
	t.Setenv("AWS_REGION", "eu-west-1")
	t.Setenv("EMAIL_FROM", "Leamout <sender@example.com>")
	t.Setenv("SES_CONFIGURATION_SET", "transactional")
	cfg, err := env.ParseAs[EmailConfig]()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Region != "eu-west-1" || cfg.From != "Leamout <sender@example.com>" || cfg.ConfigurationSet != "transactional" {
		t.Fatal("email environment not parsed")
	}
}
