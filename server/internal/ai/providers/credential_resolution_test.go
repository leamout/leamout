package providers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/apperror"
)

func TestResolveSecretRequiresOrganizationIntegration(t *testing.T) {
	t.Parallel()

	service := &Service{}
	for _, credentialID := range []*uuid.UUID{nil, uuidPtr(uuid.Nil)} {
		_, err := service.resolveSecret(uuid.New(), credentialID, nil)
		if err == nil || !strings.Contains(err.Error(), "organization AI integration is required") {
			t.Fatalf("resolveSecret() error = %v, want organization integration required", err)
		}
		var appErr *apperror.AppError
		if !errors.As(err, &appErr) || appErr.Code != "CONFLICT" {
			t.Fatalf("resolveSecret() error = %v, want CONFLICT", err)
		}
	}
}

func TestUpsertBindingRequiresOrganizationIntegration(t *testing.T) {
	t.Parallel()

	service := NewService(nil, nil)
	for _, credentialID := range []*uuid.UUID{nil, uuidPtr(uuid.Nil)} {
		_, err := service.UpsertBinding(
			context.Background(),
			uuid.New(),
			uuid.New(),
			RoleLLM,
			UpsertBindingRequest{Provider: ProviderGroq, CredentialID: credentialID},
		)
		if err == nil || !strings.Contains(err.Error(), "credential_id is required for BYOA") {
			t.Fatalf("UpsertBinding() error = %v, want credential_id required", err)
		}
	}
}

func uuidPtr(value uuid.UUID) *uuid.UUID {
	return &value
}
