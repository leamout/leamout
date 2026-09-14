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
		"database":            map[string]string{"migration": "043_create_idempotency.sql"},
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
		"runtime/migrations/043_create_idempotency.sql": "-- fixture\n",
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

func TestInstallRuntimeBundleRejectsNewerMinimumCLI(t *testing.T) {
	root := t.TempDir()
	version := "1.0.0-preview.1"
	releaseDir := stageRuntimeRelease(t, root, version)
	manifestPath := filepath.Join(releaseDir, "release-manifest.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest["minimum_cli_version"] = "1.1.0"
	content, err = json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(content, '\n'), 0o640); err != nil {
		t.Fatal(err)
	}

	err = installRuntimeBundle(releaseDir, filepath.Join(root, "runtime"), version)
	if err == nil || !strings.Contains(err.Error(), "requires leamout CLI 1.1.0 or newer") {
		t.Fatalf("expected minimum CLI rejection, got %v", err)
	}
}

func TestValidateMinimumCLIVersion(t *testing.T) {
	for _, test := range []struct {
		cli, minimum string
		wantError    bool
	}{
		{cli: "1.2.0", minimum: "1.1.9"},
		{cli: "1.2.0", minimum: "1.2.0"},
		{cli: "1.2.0", minimum: "1.2.0-preview.2"},
		{cli: "1.2.0-preview.2", minimum: "1.2.0-preview.1"},
		{cli: "1.2.0-preview.1", minimum: "1.2.0", wantError: true},
		{cli: "1.2.0-preview.1", minimum: "1.2.0-preview.2", wantError: true},
		{cli: "1.1.9", minimum: "1.2.0", wantError: true},
	} {
		err := validateMinimumCLIVersion(test.cli, test.minimum)
		if (err != nil) != test.wantError {
			t.Errorf("validateMinimumCLIVersion(%q, %q) error = %v, wantError %v", test.cli, test.minimum, err, test.wantError)
		}
	}
}

func TestCompareReleaseVersions(t *testing.T) {
	for _, test := range []struct {
		left, right string
		want        int
	}{
		{left: "1.2.0", right: "1.1.9", want: 1},
		{left: "1.2.0", right: "1.2.0", want: 0},
		{left: "1.2.0-preview.1", right: "1.2.0", want: -1},
		{left: "1.2.0-preview.2", right: "1.2.0-preview.1", want: 1},
	} {
		got, err := compareReleaseVersions(test.left, test.right)
		if err != nil || got != test.want {
			t.Errorf("compareReleaseVersions(%q, %q) = %d, %v; want %d", test.left, test.right, got, err, test.want)
		}
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

func TestInstallRuntimeBundlePreservesOlderVersionWhenStagingFails(t *testing.T) {
	root := t.TempDir()
	runtimeDir := filepath.Join(root, "runtime")
	oldVersion := "1.0.0-preview.1"
	newVersion := "1.0.0-preview.2"
	if err := installRuntimeBundle(stageRuntimeRelease(t, root, oldVersion), runtimeDir, oldVersion); err != nil {
		t.Fatal(err)
	}
	newRelease := stageRuntimeRelease(t, root, newVersion)
	archive := filepath.Join(newRelease, "leamout_runtime_"+newVersion+"_linux_amd64.tar.gz")
	if err := os.WriteFile(archive, []byte("tampered"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := installRuntimeBundle(newRelease, runtimeDir, newVersion); err == nil {
		t.Fatal("expected failed update")
	}
	installed, err := installedRuntimeVersion(runtimeDir)
	if err != nil {
		t.Fatal(err)
	}
	if installed != oldVersion {
		t.Fatalf("installed runtime = %q, want preserved %q", installed, oldVersion)
	}
}

func TestRuntimeUpdateCanCommitOrRollback(t *testing.T) {
	root := t.TempDir()
	runtimeDir := filepath.Join(root, "runtime")
	oldVersion := "1.0.0-preview.1"
	newVersion := "1.0.0-preview.2"
	if err := installRuntimeBundle(stageRuntimeRelease(t, root, oldVersion), runtimeDir, oldVersion); err != nil {
		t.Fatal(err)
	}
	if err := installRuntimeBundle(stageRuntimeRelease(t, root, newVersion), runtimeDir, newVersion); err != nil {
		t.Fatal(err)
	}
	if previous, err := installedRuntimeVersion(runtimeDir + ".previous"); err != nil || previous != oldVersion {
		t.Fatalf("previous runtime = %q, %v; want %q", previous, err, oldVersion)
	}
	if err := rollbackRuntimeUpdate(runtimeDir); err != nil {
		t.Fatal(err)
	}
	if current, err := installedRuntimeVersion(runtimeDir); err != nil || current != oldVersion {
		t.Fatalf("rolled back runtime = %q, %v; want %q", current, err, oldVersion)
	}

	if err := installRuntimeBundle(stageRuntimeRelease(t, root, newVersion), runtimeDir, newVersion); err != nil {
		t.Fatal(err)
	}
	if err := commitRuntimeUpdate(runtimeDir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(runtimeDir + ".previous"); !os.IsNotExist(err) {
		t.Fatalf("previous runtime retained after commit: %v", err)
	}
}

func TestRuntimeRollbackRecoversInterruptedFilesystemSwap(t *testing.T) {
	root := t.TempDir()
	runtimeDir := filepath.Join(root, "runtime")
	version := "1.0.0"
	if err := installRuntimeBundle(stageRuntimeRelease(t, root, version), runtimeDir, version); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(runtimeDir, runtimeDir+".previous"); err != nil {
		t.Fatal(err)
	}
	if err := rollbackRuntimeUpdate(runtimeDir); err != nil {
		t.Fatal(err)
	}
	if installed, err := installedRuntimeVersion(runtimeDir); err != nil || installed != version {
		t.Fatalf("recovered runtime = %q, %v; want %q", installed, err, version)
	}
}
