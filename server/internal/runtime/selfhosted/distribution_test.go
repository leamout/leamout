package server

import (
	"os/exec"
	"strings"
	"testing"
)

func TestSelfHostedBinariesExcludeCloudOnlyDependencies(t *testing.T) {
	command := exec.Command("go", "list", "-deps", "../../../cmd/selfhosted", "../../../cmd/selfhosted-worker")
	output, err := command.Output()
	if err != nil {
		t.Fatalf("list self-hosted dependencies: %v", err)
	}

	forbidden := []string{
		"github.com/leamout/leamout/internal/commercial",
		"github.com/leamout/leamout/internal/integrations/carriers/commpeak",
		"github.com/leamout/leamout/internal/integrations/carriers/didww",
		"github.com/leamout/leamout/internal/integrations/payments/paystack",
		"github.com/leamout/leamout/internal/integrations/payments/stripe",
		"github.com/leamout/leamout/internal/platform/provider_diagnostics",
		"github.com/leamout/leamout/internal/telecom/edge",
		"github.com/leamout/leamout/internal/telecom/wholesale",
	}
	dependencies := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, dependency := range forbidden {
		for _, included := range dependencies {
			if included == dependency || strings.HasPrefix(included, dependency+"/") {
				t.Errorf("self-hosted binaries include cloud-only dependency %q", included)
			}
		}
	}
}
