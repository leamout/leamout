package leamout

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stageRuntimeRelease(t *testing.T, stateDir, version string) string {
	t.Helper()
	releaseDir := filepath.Join(stateDir, "releases", version)
	if err := os.MkdirAll(releaseDir, 0o750); err != nil {
		t.Fatal(err)
	}

	images := map[string]string{}
	for name := range runtimeImageTokens {
		digest := sha256.Sum256([]byte(name))
		images[name] = fmt.Sprintf("registry.example/leamout/%s@sha256:%s", name, hex.EncodeToString(digest[:]))
	}
	manifest := map[string]any{
		"schema_version":      1,
		"release_version":     version,
		"channel":             "preview",
		"source_commit":       strings.Repeat("1", 40),
		"minimum_cli_version": version,
		"supported_hosts":     []map[string]string{{"os": "ubuntu", "version": "24.04", "arch": "amd64"}},
		"database":            map[string]string{"migration": "039_create_idempotency.sql"},
		"cli_artifacts": []map[string]string{{
			"os": "linux", "arch": "amd64",
			"filename": "leamout_" + version + "_linux_amd64.tar.gz",
			"sha256":   strings.Repeat("2", 64),
		}},
		"images": images,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	manifestBytes = append(manifestBytes, '\n')
	if err := os.WriteFile(filepath.Join(releaseDir, "release-manifest.json"), manifestBytes, 0o640); err != nil {
		t.Fatal(err)
	}

	archiveName := "leamout_runtime_" + version + "_linux_amd64.tar.gz"
	archivePath := filepath.Join(releaseDir, archiveName)
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)

	compose := "services:\n"
	for name, token := range runtimeImageTokens {
		compose += fmt.Sprintf("  %s:\n    image: %s\n", name, token)
	}
	files := map[string]string{
		"runtime/compose.yaml.tmpl":                     compose,
		"runtime/coturn/turnserver.conf":                "listening-port=3478\n",
		"runtime/migrations/atlas.sum":                  "h1:test\n",
		"runtime/migrations/039_create_idempotency.sql": "-- fixture\n",
	}
	for name, content := range files {
		header := &tar.Header{Name: name, Mode: 0o640, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	digest, err := hashFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	checksums := fmt.Sprintf("%s  %s\n", digest, archiveName)
	if err := os.WriteFile(filepath.Join(releaseDir, "checksums.txt"), []byte(checksums), 0o640); err != nil {
		t.Fatal(err)
	}
	return releaseDir
}

func TestInstallRuntimeBundleRendersManifestImages(t *testing.T) {
	root := t.TempDir()
	version := "1.0.0-preview.1"
	releaseDir := stageRuntimeRelease(t, root, version)
	runtimeDir := filepath.Join(root, "runtime")

	if err := installRuntimeBundle(releaseDir, runtimeDir, version); err != nil {
		t.Fatal(err)
	}
	compose, err := os.ReadFile(filepath.Join(runtimeDir, "compose.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(compose), "@@IMAGE_") {
		t.Fatalf("runtime contains unresolved images: %s", compose)
	}
	if _, err := os.Stat(filepath.Join(runtimeDir, "release.json")); err != nil {
		t.Fatal(err)
	}
	if err := installRuntimeBundle(releaseDir, runtimeDir, version); err != nil {
		t.Fatalf("repeated runtime installation failed: %v", err)
	}
}

func TestInstallRuntimeBundleRejectsTamperedArchive(t *testing.T) {
	root := t.TempDir()
	version := "1.0.0-preview.1"
	releaseDir := stageRuntimeRelease(t, root, version)
	archive := filepath.Join(releaseDir, "leamout_runtime_"+version+"_linux_amd64.tar.gz")
	if err := os.WriteFile(archive, []byte("tampered"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := installRuntimeBundle(releaseDir, filepath.Join(root, "runtime"), version); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch, got %v", err)
	}
}

func TestInstallRuntimeBundleReplacesOlderVersion(t *testing.T) {
	root := t.TempDir()
	runtimeDir := filepath.Join(root, "runtime")
	oldVersion := "1.0.0-preview.1"
	newVersion := "1.0.0-preview.2"

	if err := installRuntimeBundle(stageRuntimeRelease(t, root, oldVersion), runtimeDir, oldVersion); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtimeDir, "old-version-only"), []byte("stale\n"), 0o640); err != nil {
		t.Fatal(err)
	}

	if err := installRuntimeBundle(stageRuntimeRelease(t, root, newVersion), runtimeDir, newVersion); err != nil {
		t.Fatalf("upgrade runtime installation failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(runtimeDir, "old-version-only")); !os.IsNotExist(err) {
		t.Fatalf("old runtime was not replaced: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(runtimeDir, "release.json"))
	if err != nil {
		t.Fatal(err)
	}
	var installed installedRuntimeRelease
	if err := json.Unmarshal(content, &installed); err != nil {
		t.Fatal(err)
	}
	if installed.ReleaseVersion != newVersion {
		t.Fatalf("installed runtime version = %q, want %q", installed.ReleaseVersion, newVersion)
	}
}
