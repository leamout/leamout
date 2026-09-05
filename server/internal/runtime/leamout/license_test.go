package leamout

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/licensing"
)

func TestLicenseInstallVerifiesDeploymentBindingAndPersistsArtifact(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := licensing.NewSigner("release-2026", privateKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	deploymentID := uuid.NewString()
	artifact, err := signer.SignV1(licensing.LicenseClaimsV1{
		LicenseID: uuid.New(), OrganizationID: uuid.New(), SubscriptionID: uuid.New(),
		DeploymentID: deploymentID, IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
		Features: map[string]bool{"voice.enabled": true}, Limits: map[string]int64{"voice.concurrent_calls": 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	artifactPath := filepath.Join(root, "artifact.json")
	keyringPath := filepath.Join(root, "keyring.json")
	statePath := filepath.Join(root, "deployment.json")
	licenseDir := filepath.Join(root, "license")
	state, _ := json.Marshal(deploymentState{SchemaVersion: 1, DeploymentID: deploymentID, Mode: deploymentMode, CreatedAt: now})
	keyring, _ := json.Marshal(licenseKeyringFile{Version: 1, Keys: map[string]string{"release-2026": base64.RawURLEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey))}})
	for path, content := range map[string][]byte{artifactPath: artifact, keyringPath: keyring, statePath: state} {
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	if code := runLicenseAt(&stdout, &stderr, []string{"install", "--artifact", artifactPath, "--keyring", keyringPath}, statePath, licenseDir, now); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	installed, err := os.ReadFile(filepath.Join(licenseDir, "license.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(installed, artifact) {
		t.Fatal("installed artifact changed")
	}
	if info, err := os.Stat(filepath.Join(licenseDir, "license.json")); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("license permissions: %v %v", info, err)
	}
}

func TestLicenseVerificationRejectsAnotherDeployment(t *testing.T) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	signer, _ := licensing.NewSigner("key", privateKey)
	now := time.Now().UTC()
	artifact, _ := signer.SignV1(licensing.LicenseClaimsV1{LicenseID: uuid.New(), OrganizationID: uuid.New(), SubscriptionID: uuid.New(), DeploymentID: uuid.NewString(), IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour)})
	keyring, _ := json.Marshal(licenseKeyringFile{Version: 1, Keys: map[string]string{"key": base64.RawURLEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey))}})
	if _, err := verifyOfflineLicense(artifact, keyring, uuid.NewString(), now); err == nil {
		t.Fatal("wrong deployment license accepted")
	}
}
