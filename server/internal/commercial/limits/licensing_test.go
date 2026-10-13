package limits

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/commercial/licensing"
)

// Nil Cloud services and DB prove self-hosted decisions have no billing dependency.
func TestSelfHostedDoesNotReadCloudLimits(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	now := time.Now().UTC()
	claims := licensing.Claims{Version: 1, LicenseID: "id", CustomerID: "customer", DeploymentID: "deployment", ValidFrom: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), GraceUntil: now.Add(time.Hour), MaxConcurrentCalls: 50, MaxNodes: 3}
	payload, _ := json.Marshal(claims)
	data, _ := json.Marshal(licensing.Document{KeyID: "key", Payload: base64.StdEncoding.EncodeToString(payload), Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, payload))})
	path := filepath.Join(t.TempDir(), "license.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	license, err := licensing.NewService(licensing.Config{Mode: licensing.SelfHosted, DeploymentID: "deployment", File: path, PublicKeys: map[string]string{"key": base64.StdEncoding.EncodeToString(public)}})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(nil, nil)
	service.SetLicenseService(license)
	ctx := context.Background()
	organization := uuid.New()
	limit, limited, err := service.MaxConcurrentCalls(ctx, organization)
	if err != nil || !limited || limit != 50 {
		t.Fatalf("concurrency=%d,%v,%v", limit, limited, err)
	}
	for _, resolve := range []func(context.Context, uuid.UUID) (int64, bool, error){service.RetentionDays, service.MaxVoiceAgents, service.MaxSIPTrunks} {
		_, limited, err := resolve(ctx, organization)
		if err != nil || limited {
			t.Fatalf("unexpected Cloud restriction: %v %v", limited, err)
		}
	}
	usage, err := service.RuntimeUsage(ctx, organization)
	if err != nil || !usage.Unlimited {
		t.Fatalf("runtime usage=%+v %v", usage, err)
	}
	if err := service.RequireRuntimeAllowance(ctx, organization); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.MaxConcurrentCalls(ctx, organization); err == nil {
		t.Fatal("unlicensed runtime admitted")
	}
	if err := service.RequireRuntimeAllowance(ctx, organization); err == nil {
		t.Fatal("unlicensed Voice Agent admitted")
	}
}
