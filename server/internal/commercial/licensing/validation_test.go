package licensing

import (
	"errors"
	"testing"
	"time"
)

func TestValidateTransition(t *testing.T) {
	t.Parallel()

	allowed := [][2]Status{
		{StatusPending, StatusActive},
		{StatusPending, StatusRevoked},
		{StatusActive, StatusSuspended},
		{StatusActive, StatusExpired},
		{StatusActive, StatusRevoked},
		{StatusSuspended, StatusActive},
		{StatusSuspended, StatusExpired},
		{StatusSuspended, StatusRevoked},
		{StatusActive, StatusActive},
	}
	for _, transition := range allowed {
		if err := validateTransition(transition[0], transition[1]); err != nil {
			t.Fatalf("expected %s -> %s to be valid: %v", transition[0], transition[1], err)
		}
	}

	blocked := [][2]Status{
		{StatusPending, StatusSuspended},
		{StatusExpired, StatusActive},
		{StatusRevoked, StatusActive},
	}
	for _, transition := range blocked {
		if err := validateTransition(transition[0], transition[1]); !errors.Is(err, ErrInvalidTransition) {
			t.Fatalf("expected %s -> %s to fail with ErrInvalidTransition, got %v", transition[0], transition[1], err)
		}
	}
}

func TestNormalizeCreate(t *testing.T) {
	t.Parallel()

	issuedAt := time.Date(2026, 8, 31, 20, 0, 0, 0, time.UTC)
	expiresAt := issuedAt.Add(24 * time.Hour)
	key := " key-2026 "
	got, gotIssuedAt, err := normalizeCreate(CreateInput{SigningKeyID: &key, ExpiresAt: &expiresAt}, issuedAt)
	if err != nil {
		t.Fatalf("normalizeCreate() error = %v", err)
	}
	if gotIssuedAt != issuedAt || got.SigningKeyID == nil || *got.SigningKeyID != "key-2026" {
		t.Fatalf("unexpected normalized create: %#v, %v", got, gotIssuedAt)
	}

	invalidExpiry := issuedAt
	_, _, err = normalizeCreate(CreateInput{ExpiresAt: &invalidExpiry}, issuedAt)
	if !errors.Is(err, ErrInvalidExpiration) {
		t.Fatalf("normalizeCreate() error = %v, want %v", err, ErrInvalidExpiration)
	}
}

func TestNormalizeDeployment(t *testing.T) {
	t.Parallel()

	name := " Production Node "
	got, err := normalizeDeployment(ActivateDeploymentInput{DeploymentID: " node-01 ", Name: &name})
	if err != nil {
		t.Fatalf("normalizeDeployment() error = %v", err)
	}
	if got.DeploymentID != "node-01" || got.Name == nil || *got.Name != "Production Node" {
		t.Fatalf("unexpected normalized deployment: %#v", got)
	}

	_, err = normalizeDeployment(ActivateDeploymentInput{DeploymentID: "node 01"})
	if !errors.Is(err, ErrInvalidDeploymentID) {
		t.Fatalf("normalizeDeployment() error = %v, want %v", err, ErrInvalidDeploymentID)
	}
}
