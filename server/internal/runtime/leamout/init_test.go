package leamout

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInitCreatesDurableDeploymentStateAndSecrets(t *testing.T) {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("Self-Hosted Production v0.1 host contract is linux/amd64")
	}

	root := t.TempDir()
	configDir := filepath.Join(root, "etc", "leamout")
	stateDir := filepath.Join(root, "var", "lib", "leamout")
	logDir := filepath.Join(root, "var", "log", "leamout")

	var stdout, stderr bytes.Buffer
	code := runInitAt(&stdout, &stderr, configDir, stateDir, logDir)
	if code != 0 {
		t.Fatalf("init returned %d: %s", code, stderr.String())
	}

	statePath := filepath.Join(stateDir, "deployment.json")
	envPath := filepath.Join(configDir, "leamout.env")
	for _, path := range []string{statePath, envPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("%s mode = %o, want 600", path, got)
		}
	}

	stateBytes, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	var state deploymentState
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		t.Fatal(err)
	}
	if state.DeploymentID == "" || state.Mode != deploymentMode || state.SchemaVersion != deploymentStateSchemaVersion {
		t.Fatalf("unexpected deployment state: %+v", state)
	}

	envBytes, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	values := parseEnv(envBytes)
	if values["APP_ENV"] != "production" {
		t.Fatalf("APP_ENV = %q", values["APP_ENV"])
	}
	if values["LEAMOUT_DEPLOYMENT_ID"] != state.DeploymentID {
		t.Fatalf("deployment ID mismatch")
	}
	for _, key := range []string{
		"FREESWITCH_ESL_PASSWORD",
		"TURN_AUTH_SECRET",
		"POSTGRES_PASSWORD",
		"REDIS_PASSWORD",
		"NATS_PASSWORD",
	} {
		if len(values[key]) != 64 {
			t.Fatalf("%s length = %d, want 64 hex chars", key, len(values[key]))
		}
	}
	carrierKey, err := base64.RawURLEncoding.DecodeString(values["CARRIER_CREDENTIAL_ENCRYPTION_KEY"])
	if err != nil || len(carrierKey) != 32 {
		t.Fatalf("invalid carrier encryption key: len=%d err=%v", len(carrierKey), err)
	}

	for _, dir := range []string{
		configDir,
		filepath.Join(configDir, "certs"),
		filepath.Join(configDir, "license"),
		stateDir,
		filepath.Join(stateDir, "backups"),
		logDir,
	} {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			t.Fatalf("expected directory %s: %v", dir, err)
		}
		if got := info.Mode().Perm(); got != 0o750 {
			t.Fatalf("%s mode = %o, want 750", dir, got)
		}
	}

	for _, secret := range []string{
		values["TURN_AUTH_SECRET"],
		values["POSTGRES_PASSWORD"],
		values["REDIS_PASSWORD"],
		values["NATS_PASSWORD"],
	} {
		if strings.Contains(stdout.String(), secret) {
			t.Fatalf("init output leaked generated secret: %s", stdout.String())
		}
	}
	if !strings.Contains(stdout.String(), "Deployment identity created") {
		t.Fatalf("unexpected init output: %s", stdout.String())
	}
}

func TestInitIsIdempotentAndPreservesIdentityAndSecrets(t *testing.T) {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("Self-Hosted Production v0.1 host contract is linux/amd64")
	}

	root := t.TempDir()
	configDir := filepath.Join(root, "etc", "leamout")
	stateDir := filepath.Join(root, "var", "lib", "leamout")
	logDir := filepath.Join(root, "var", "log", "leamout")

	run := func() string {
		var stdout, stderr bytes.Buffer
		if code := runInitAt(&stdout, &stderr, configDir, stateDir, logDir); code != 0 {
			t.Fatalf("init returned %d: %s", code, stderr.String())
		}
		return stdout.String()
	}

	run()
	statePath := filepath.Join(stateDir, "deployment.json")
	envPath := filepath.Join(configDir, "leamout.env")
	stateBefore, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	envBefore, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}

	stdout := run()
	stateAfter, _ := os.ReadFile(statePath)
	envAfter, _ := os.ReadFile(envPath)
	if !bytes.Equal(stateBefore, stateAfter) {
		t.Fatal("deployment identity changed on repeated init")
	}
	if !bytes.Equal(envBefore, envAfter) {
		t.Fatal("deployment secrets changed on repeated init")
	}
	if !strings.Contains(stdout, "Existing deployment identity and secrets preserved") {
		t.Fatalf("unexpected repeat init output: %s", stdout)
	}
}

func TestInitRejectsPartialInitialization(t *testing.T) {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("Self-Hosted Production v0.1 host contract is linux/amd64")
	}

	root := t.TempDir()
	configDir := filepath.Join(root, "etc", "leamout")
	stateDir := filepath.Join(root, "var", "lib", "leamout")
	logDir := filepath.Join(root, "var", "log", "leamout")
	if err := os.MkdirAll(stateDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "deployment.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := runInitAt(&stdout, &stderr, configDir, stateDir, logDir)
	if code == 0 {
		t.Fatalf("init unexpectedly succeeded: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "incomplete Leamout initialization") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}
