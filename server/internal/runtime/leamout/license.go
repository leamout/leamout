package leamout

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/leamout/leamout/internal/commercial/licensing"
)

type licenseKeyringFile struct {
	Version int               `json:"version"`
	Keys    map[string]string `json:"keys"`
}

func runLicense(stdout, stderr io.Writer, args []string) int {
	return runLicenseAt(stdout, stderr, args, "/var/lib/leamout/deployment.json", "/etc/leamout/license", time.Now().UTC())
}

func runLicenseAt(stdout, stderr io.Writer, args []string, statePath, licenseDir string, now time.Time) int {
	if len(args) < 1 || (args[0] != "install" && args[0] != "verify") {
		writeln(stderr, "usage: leamout license <install|verify> --artifact <path> --keyring <path>")
		return 2
	}
	action := args[0]
	var artifactPath, keyringPath string
	for index := 1; index < len(args); index++ {
		switch args[index] {
		case "--artifact", "--keyring":
			if index+1 >= len(args) {
				writef(stderr, "%s requires a path\n", args[index])
				return 2
			}
			if args[index] == "--artifact" {
				artifactPath = args[index+1]
			} else {
				keyringPath = args[index+1]
			}
			index++
		default:
			writef(stderr, "unknown license option: %s\n", args[index])
			return 2
		}
	}
	if artifactPath == "" || keyringPath == "" {
		writeln(stderr, "--artifact and --keyring are required")
		return 2
	}
	state, err := loadDeploymentState(statePath)
	if err != nil {
		writef(stderr, "load deployment identity: %v\n", err)
		return 1
	}
	artifact, err := os.ReadFile(artifactPath)
	if err != nil {
		writef(stderr, "read license artifact: %v\n", err)
		return 1
	}
	keyringBytes, err := os.ReadFile(keyringPath)
	if err != nil {
		writef(stderr, "read license keyring: %v\n", err)
		return 1
	}
	claims, err := verifyOfflineLicense(artifact, keyringBytes, state.DeploymentID, now)
	if err != nil {
		writef(stderr, "verify offline license: %v\n", err)
		return 1
	}
	if action == "install" {
		if err := os.MkdirAll(licenseDir, 0o750); err != nil {
			writef(stderr, "create license directory: %v\n", err)
			return 1
		}
		if err := writeAtomicFile(filepath.Join(licenseDir, "license.json"), artifact, 0o600); err != nil {
			writef(stderr, "install license: %v\n", err)
			return 1
		}
		if err := writeAtomicFile(filepath.Join(licenseDir, "keyring.json"), keyringBytes, 0o600); err != nil {
			writef(stderr, "install license keyring: %v\n", err)
			return 1
		}
		writeln(stdout, "✓ Offline license installed")
	} else {
		writeln(stdout, "✓ Offline license valid")
	}
	writef(stdout, "License ID: %s\nValid until: %s\n", claims.LicenseID, claims.ExpiresAt.Format(time.RFC3339))
	return 0
}

func verifyOfflineLicense(artifact, keyringBytes []byte, deploymentID string, now time.Time) (licensing.LicenseClaimsV1, error) {
	var file licenseKeyringFile
	if err := json.Unmarshal(keyringBytes, &file); err != nil {
		return licensing.LicenseClaimsV1{}, fmt.Errorf("decode keyring: %w", err)
	}
	if file.Version != 1 || len(file.Keys) == 0 {
		return licensing.LicenseClaimsV1{}, errors.New("unsupported or empty license keyring")
	}
	keys := make(map[string]ed25519.PublicKey, len(file.Keys))
	for id, encoded := range file.Keys {
		key, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil || len(key) != ed25519.PublicKeySize {
			return licensing.LicenseClaimsV1{}, fmt.Errorf("invalid public key %q", id)
		}
		keys[id] = ed25519.PublicKey(key)
	}
	keyring, err := licensing.NewKeyring(keys)
	if err != nil {
		return licensing.LicenseClaimsV1{}, err
	}
	return keyring.VerifyV1(artifact, deploymentID, now)
}

func writeAtomicFile(path string, content []byte, mode os.FileMode) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".leamout-write-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if err := temp.Chmod(mode); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err := temp.Write(content); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
