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
	_, signingPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := licensing.NewSigner("release-2026", signingPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	deploymentPublicKey, deploymentPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	deploymentPublicKeyEncoded := base64.RawURLEncoding.EncodeToString(deploymentPublicKey)
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	deploymentID := uuid.NewString()
	artifact, err := signer.SignV1(licensing.LicenseClaimsV1{
		LicenseID: uuid.New(), OrganizationID: uuid.New(), DeploymentID: deploymentID,
		DeploymentPublicKey: deploymentPublicKeyEncoded,
		IssuedAt:            now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	artifactPath := filepath.Join(root, "artifact.json")
	keyringPath := filepath.Join(root, "keyring.json")
	statePath := filepath.Join(root, "deployment.json")
	identityKeyPath := filepath.Join(root, "deployment.key")
	licenseDir := filepath.Join(root, "license")
	state, _ := json.Marshal(deploymentState{SchemaVersion: 1, DeploymentID: deploymentID, PublicKey: deploymentPublicKeyEncoded, Mode: deploymentMode, CreatedAt: now})
	keyring, _ := json.Marshal(licenseKeyringFile{Version: 1, Keys: map[string]string{"release-2026": base64.RawURLEncoding.EncodeToString(signingPrivateKey.Public().(ed25519.PublicKey))}})
	files := map[string][]byte{
		artifactPath:    artifact,
		keyringPath:     keyring,
		statePath:       state,
		identityKeyPath: []byte(base64.RawURLEncoding.EncodeToString(deploymentPrivateKey) + "\n"),
	}
	for path, content := range files {
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

func TestValidateInstalledLicenseRequiresCurrentInstalledArtifact(t *testing.T) {
	_, signingPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	deploymentPublicKey, deploymentPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	deploymentPublicKeyEncoded := base64.RawURLEncoding.EncodeToString(deploymentPublicKey)
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	deploymentID := uuid.NewString()
	signer, err := licensing.NewSigner("release-2026", signingPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := signer.SignV1(licensing.LicenseClaimsV1{
		LicenseID: uuid.New(), OrganizationID: uuid.New(), DeploymentID: deploymentID,
		DeploymentPublicKey: deploymentPublicKeyEncoded,
		IssuedAt:            now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := json.Marshal(licenseKeyringFile{
		Version: 1,
		Keys: map[string]string{
			"release-2026": base64.RawURLEncoding.EncodeToString(signingPrivateKey.Public().(ed25519.PublicKey)),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	statePath := filepath.Join(root, "deployment.json")
	licenseDir := filepath.Join(root, "license")
	if err := os.MkdirAll(licenseDir, 0o750); err != nil {
		t.Fatal(err)
	}
	state, err := json.Marshal(deploymentState{
		SchemaVersion: 1,
		DeploymentID:  deploymentID,
		PublicKey:     deploymentPublicKeyEncoded,
		Mode:          deploymentMode,
		CreatedAt:     now,
	})
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		statePath:                                 state,
		filepath.Join(root, "deployment.key"):     []byte(base64.RawURLEncoding.EncodeToString(deploymentPrivateKey) + "\n"),
		filepath.Join(licenseDir, "license.json"): artifact,
		filepath.Join(licenseDir, "keyring.json"): keyring,
	}
	for path, content := range files {
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := validateInstalledLicense(statePath, licenseDir, now); err != nil {
		t.Fatalf("valid installed license rejected: %v", err)
	}
	if _, err := validateInstalledLicense(statePath, licenseDir, now.Add(2*time.Hour)); err == nil {
		t.Fatal("expired installed license accepted")
	}
	if err := os.Remove(filepath.Join(licenseDir, "license.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := validateInstalledLicense(statePath, licenseDir, now); err == nil {
		t.Fatal("missing installed license accepted")
	}
}

func TestLicenseVerificationRejectsAnotherDeployment(t *testing.T) {
	_, signingPrivateKey, _ := ed25519.GenerateKey(rand.Reader)
	deploymentPublicKey, _, _ := ed25519.GenerateKey(rand.Reader)
	deploymentPublicKeyEncoded := base64.RawURLEncoding.EncodeToString(deploymentPublicKey)
	signer, _ := licensing.NewSigner("key", signingPrivateKey)
	now := time.Now().UTC()
	artifact, _ := signer.SignV1(licensing.LicenseClaimsV1{
		LicenseID: uuid.New(), OrganizationID: uuid.New(), DeploymentID: uuid.NewString(),
		DeploymentPublicKey: deploymentPublicKeyEncoded,
		IssuedAt:            now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
	})
	keyring, _ := json.Marshal(licenseKeyringFile{Version: 1, Keys: map[string]string{"key": base64.RawURLEncoding.EncodeToString(signingPrivateKey.Public().(ed25519.PublicKey))}})
	if _, err := verifyOfflineLicense(artifact, keyring, uuid.NewString(), deploymentPublicKeyEncoded, now); err == nil {
		t.Fatal("wrong deployment license accepted")
	}
}

func TestLicenseVerificationRejectsAnotherDeploymentKey(t *testing.T) {
	_, signingPrivateKey, _ := ed25519.GenerateKey(rand.Reader)
	deploymentPublicKey, _, _ := ed25519.GenerateKey(rand.Reader)
	otherPublicKey, _, _ := ed25519.GenerateKey(rand.Reader)
	deploymentPublicKeyEncoded := base64.RawURLEncoding.EncodeToString(deploymentPublicKey)
	signer, _ := licensing.NewSigner("key", signingPrivateKey)
	now := time.Now().UTC()
	deploymentID := uuid.NewString()
	artifact, _ := signer.SignV1(licensing.LicenseClaimsV1{
		LicenseID: uuid.New(), OrganizationID: uuid.New(), DeploymentID: deploymentID,
		DeploymentPublicKey: deploymentPublicKeyEncoded,
		IssuedAt:            now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour),
	})
	keyring, _ := json.Marshal(licenseKeyringFile{Version: 1, Keys: map[string]string{"key": base64.RawURLEncoding.EncodeToString(signingPrivateKey.Public().(ed25519.PublicKey))}})
	if _, err := verifyOfflineLicense(artifact, keyring, deploymentID, base64.RawURLEncoding.EncodeToString(otherPublicKey), now); err == nil {
		t.Fatal("wrong deployment key license accepted")
	}
}
