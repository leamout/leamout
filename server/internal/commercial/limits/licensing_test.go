package limits

import (
	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/platform/deployment"
	"testing"
)

func TestSelfHostedDoesNotReadCloudLimits(t *testing.T) {
	service := NewService(nil, nil)
	service.SetDeployment(deployment.Config{Mode: deployment.SelfHosted})
	org := uuid.New()
	value, limited, err := service.MaxConcurrentCalls(t.Context(), org)
	if err != nil || limited || value != 0 {
		t.Fatalf("unexpected self-hosted quota: %d %v %v", value, limited, err)
	}
	usage, err := service.RuntimeUsage(t.Context(), org)
	if err != nil || !usage.Unlimited {
		t.Fatalf("unexpected usage: %+v %v", usage, err)
	}
	if err := service.RequireRuntimeAllowance(t.Context(), org); err != nil {
		t.Fatal(err)
	}
}
