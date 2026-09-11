package server

import (
	"testing"

	"github.com/leamout/leamout/internal/platform/config"
)

func TestSelfHostedCompositionRejectsCloudCredentialsBeforeConnecting(t *testing.T) {
	cfg := config.Config{
		DeploymentMode: config.DeploymentModeSelfHosted,
		DIDWW:          config.DIDWWConfig{APIKey: "cloud-provider-secret"},
	}
	if _, err := NewSelfHosted(t.Context(), cfg); err == nil {
		t.Fatal("NewSelfHosted() accepted Cloud provider credentials")
	}
}

func TestCompositionRootsRejectWrongModeBeforeConnecting(t *testing.T) {
	if _, err := NewCloud(t.Context(), config.Config{DeploymentMode: config.DeploymentModeSelfHosted}); err == nil {
		t.Fatal("NewCloud() accepted self-hosted mode")
	}
	if _, err := NewSelfHosted(t.Context(), config.Config{DeploymentMode: config.DeploymentModeCloud}); err == nil {
		t.Fatal("NewSelfHosted() accepted Cloud mode")
	}
}
