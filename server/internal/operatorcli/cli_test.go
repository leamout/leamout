package operatorcli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"version"}, &stdout, &stderr, BuildInfo{
		Version: "1.2.3",
		Commit:  "abc123",
		BuiltAt: "2026-09-03T00:00:00Z",
	})
	if code != 0 {
		t.Fatalf("Run returned %d: %s", code, stderr.String())
	}
	for _, want := range []string{"leamout 1.2.3", "commit: abc123", "built: 2026-09-03T00:00:00Z"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("version output missing %q: %s", want, stdout.String())
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"nope"}, &stdout, &stderr, BuildInfo{})
	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unknown command: nope") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestInitCreatesBaseDirectoriesWithoutRuntimeAssets(t *testing.T) {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("Phase 2 host contract is linux/amd64")
	}

	root := t.TempDir()
	configDir := filepath.Join(root, "etc")
	stateDir := filepath.Join(root, "state")
	logDir := filepath.Join(root, "log")
	t.Setenv("LEAMOUT_ROOT", filepath.Join(root, "missing-runtime"))
	t.Setenv("LEAMOUT_CONFIG_DIR", configDir)
	t.Setenv("LEAMOUT_STATE_DIR", stateDir)
	t.Setenv("LEAMOUT_LOG_DIR", logDir)

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"init"}, &stdout, &stderr, BuildInfo{})
	if code != 0 {
		t.Fatalf("init returned %d: %s", code, stderr.String())
	}
	for _, dir := range []string{configDir, stateDir, logDir} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Fatalf("expected init directory %s: %v", dir, err)
		}
	}
	if !strings.Contains(stdout.String(), "Runtime deployment assets are not installed yet") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestInitDetectsExistingDeploymentAssets(t *testing.T) {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("Phase 2 host contract is linux/amd64")
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "deploy"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "deploy", "compose.yaml"), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LEAMOUT_ROOT", root)
	t.Setenv("LEAMOUT_CONFIG_DIR", filepath.Join(root, "etc"))
	t.Setenv("LEAMOUT_STATE_DIR", filepath.Join(root, "state"))
	t.Setenv("LEAMOUT_LOG_DIR", filepath.Join(root, "log"))

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"init"}, &stdout, &stderr, BuildInfo{})
	if code != 0 {
		t.Fatalf("init returned %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Existing deployment assets detected") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}
