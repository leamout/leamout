package config

import (
	"testing"

	"github.com/leamout/leamout/server/internal/integrations/kinde"
)

func TestValidateForServerNeedsOnlyAPIDependencies(t *testing.T) {
	cfg := Config{
		DatabaseURL:           "postgres://db",
		RedisURL:              "redis://cache",
		FreeSWITCHESLPassword: "esl-secret",
		EncryptionKey:         "encryption-key",
		TURNAuthSecret:        "turn-secret",
		Kinde: kinde.Config{
			ClientSecret: "oidc-secret",
		},
	}

	if err := cfg.ValidateForServer(); err != nil {
		t.Fatalf("server should not require worker-only NATS or media tokens: %v", err)
	}

	cfg.TURNAuthSecret = ""
	if err := cfg.ValidateForServer(); err == nil {
		t.Fatal("server must reject missing TURN secret")
	}
}

func TestValidateForWorkerNeedsOnlyJobDependencies(t *testing.T) {
	cfg := Config{
		DatabaseURL:           "postgres://db",
		RedisURL:              "redis://cache",
		NATSURL:               "nats://events",
		FreeSWITCHESLPassword: "esl-secret",
		MediaControlToken:     "media-token",
		EncryptionKey:         "encryption-key",
	}

	if err := cfg.ValidateForWorker(); err != nil {
		t.Fatalf("worker should not require OIDC or TURN secrets: %v", err)
	}

	cfg.NATSURL = ""
	if err := cfg.ValidateForWorker(); err == nil {
		t.Fatal("worker must reject missing NATS URL")
	}
}

func TestValidateForBackofficeNeedsOnlyAdminDependencies(t *testing.T) {
	cfg := Config{
		DatabaseURL:   "postgres://db",
		EncryptionKey: "encryption-key",
		BackofficeKinde: kinde.Config{
			Issuer: "https://example.kinde.com", ClientID: "backoffice-client", ClientSecret: "oidc-secret",
		},
	}

	if err := cfg.ValidateForBackoffice(); err != nil {
		t.Fatalf("backoffice should not require telephony, cache or event credentials: %v", err)
	}

	cfg.BackofficeKinde.ClientSecret = ""
	if err := cfg.ValidateForBackoffice(); err == nil {
		t.Fatal("backoffice must reject missing OIDC secret")
	}
}
