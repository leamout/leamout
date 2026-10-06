package config

import (
	"testing"

	"github.com/caarlos0/env/v11"
)

func TestEmailEnvironment(t *testing.T) {
	t.Setenv("AWS_REGION", "eu-west-1")
	t.Setenv("AWS_FROM_EMAIL", "Leamout <sender@example.com>")
	t.Setenv("AWS_CONFIGURATION_SET", "transactional")
	t.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret-key")
	cfg, err := env.ParseAs[struct {
		AWS AWSConfig `envPrefix:"AWS_"`
	}]()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AWS.Region != "eu-west-1" || cfg.AWS.FromEmail != "Leamout <sender@example.com>" || cfg.AWS.ConfigurationSet != "transactional" || cfg.AWS.AccessKey != "test-access-key" || cfg.AWS.SecretKey != "test-secret-key" {
		t.Fatal("email environment not parsed")
	}
}
