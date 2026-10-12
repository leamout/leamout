package agents

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/ai/providers"
)

func TestProviderBindingReadinessRequiresOrganizationIntegration(t *testing.T) {
	t.Parallel()

	issue := providerBindingReadinessIssue(
		providers.RoleRealtime,
		providers.BindingStatus{
			Role:            providers.RoleRealtime,
			Provider:        providers.ProviderOpenAI,
			ConnectionState: providers.ConnectionUnavailable,
		},
	)
	if issue.Code != "missing_organization_integration" {
		t.Fatalf("issue.Code = %q, want missing_organization_integration", issue.Code)
	}
	if issue.Field != "bindings.realtime.integration_id" {
		t.Fatalf("issue.Field = %q", issue.Field)
	}
}

func TestProviderBindingReadinessRequiresVerifiedIntegration(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	issue := providerBindingReadinessIssue(
		providers.RoleSTT,
		providers.BindingStatus{
			Role:            providers.RoleSTT,
			Provider:        providers.ProviderDeepgram,
			IntegrationID:   &id,
			ConnectionState: providers.ConnectionUnchecked,
		},
	)
	if issue.Code != "integration_not_ready" {
		t.Fatalf("issue.Code = %q, want integration_not_ready", issue.Code)
	}
	if issue.Field != "bindings.stt.integration_id" {
		t.Fatalf("issue.Field = %q", issue.Field)
	}
}
