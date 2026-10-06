package config

import (
	"testing"

	"github.com/caarlos0/env/v11"
)

func TestEmailEnvironment(t *testing.T) {
	t.Setenv("AWS_REGION", "eu-west-1")
	t.Setenv("FROM_EMAIL", "Leamout <sender@example.com>")
	t.Setenv("SES_CONFIGURATION_SET", "transactional")
	t.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
	cfg, err := env.ParseAs[AWSConfig]()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Region != "eu-west-1" || cfg.FromEmail != "Leamout <sender@example.com>" || cfg.ConfigurationSet != "transactional" || cfg.AccessKey != "test-access-key" || cfg.SecretKey != "test-secret-key" {
		t.Fatal("email environment not parsed")
	}
}
