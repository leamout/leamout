package providers

import (
	"testing"

	"github.com/google/uuid"
)

func TestResolveSecretUsesPlatformCredential(t *testing.T) {
	t.Parallel()

	service := &Service{
		platform: NewPlatformCredentials(map[string]string{
			"openai": "platform-secret",
		}),
	}

	secret, err := service.resolveSecret(
		uuid.New(),
		"openai",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("resolveSecret() error = %v", err)
	}
	if secret != "platform-secret" {
		t.Fatalf("resolveSecret() = %q, want platform-secret", secret)
	}
}

func TestResolveSecretRejectsMissingPlatformCredential(t *testing.T) {
	t.Parallel()

	service := &Service{}
	if _, err := service.resolveSecret(
		uuid.New(),
		"openai",
		nil,
		nil,
	); err == nil {
		t.Fatal("expected missing platform credential error")
	}
}
