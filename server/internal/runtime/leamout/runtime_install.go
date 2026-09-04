package leamout

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	hex40Pattern  = regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64Pattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	imagePattern  = regexp.MustCompile(`^.+@sha256:[0-9a-f]{64}$`)
	migrationName = regexp.MustCompile(`^[0-9]{3}_[a-z0-9_]+\.sql$`)
)

var runtimeImageTokens = map[string]string{
	"server":     "@@IMAGE_SERVER@@",
	"worker":     "@@IMAGE_WORKER@@",
	"opensips":   "@@IMAGE_OPENSIPS@@",
	"rtpengine":  "@@IMAGE_RTPENGINE@@",
	"freeswitch": "@@IMAGE_FREESWITCH@@",
	"coturn":     "@@IMAGE_COTURN@@",
	"postgres":   "@@IMAGE_POSTGRES@@",
	"redis":      "@@IMAGE_REDIS@@",
	"nats":       "@@IMAGE_NATS@@",
	"atlas":      "@@IMAGE_ATLAS@@",
}

type releaseManifest struct {
	SchemaVersion     int               `json:"schema_version"`
	ReleaseVersion    string            `json:"release_version"`
	SourceCommit      string            `json:"source_commit"`
	MinimumCLIVersion string            `json:"minimum_cli_version"`
	Database          releaseDatabase   `json:"database"`
	Images            map[string]string `json:"images"`
}

type releaseDatabase struct {
	Migration string `json:"migration"`
}

type installedRuntimeRelease struct {
	SchemaVersion  int    `json:"schema_version"`
	ReleaseVersion string `json:"release_version"`
	SourceCommit   string `json:"source_commit"`
	RuntimeSHA256  string `json:"runtime_sha256"`
	ManifestSHA256 string `json:"manifest_sha256"`
}

func installRuntimeBundle(releaseDir, runtimeDir, version string) error {
	if version == "" || version == "dev" {
		return errors.New("installed CLI does not identify a production release")
	}

	manifestPath := filepath.Join(releaseDir, "release-manifest.json")
	manifest, manifestSHA, err := loadReleaseManifest(manifestPath, version)
	if err != nil {
		return err
	}

	archiveName := fmt.Sprintf("leamout_runtime_%s_linux_amd64.tar.gz", version)
	expectedSHA, err := checksumFor(filepath.Join(releaseDir, "checksums.txt"), archiveName)
	if err != nil {
		return err
	}
	archivePath := filepath.Join(releaseDir, archiveName)
	actualSHA, err := hashFile(archivePath)
	if err != nil {
		return fmt.Errorf("hash staged runtime bundle: %w", err)
	}
	if actualSHA != expectedSHA {
		return fmt.Errorf("staged runtime bundle checksum mismatch: got %s, want %s", actualSHA, expectedSHA)
	}

	exists, err := pathExists(runtimeDir)
	if err != nil {
		return fmt.Errorf("inspect installed runtime: %w", err)
	}
	if exists {
		installedVersion, err := installedRuntimeVersion(runtimeDir)
		if err != nil {
			return err
		}
		if installedVersion == manifest.ReleaseVersion {
			return validateInstalledRuntime(runtimeDir, manifest, expectedSHA, manifestSHA)
		}
		if err := os.RemoveAll(runtimeDir); err != nil {
			return fmt.Errorf("remove runtime version %q before installing %q: %w", installedVersion, manifest.ReleaseVersion, err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(runtimeDir), 0o750); err != nil {
		return fmt.Errorf("create runtime parent directory: %w", err)
	}
	workdir, err := os.MkdirTemp(filepath.Dir(runtimeDir), ".runtime-install-")
	if err != nil {
		return fmt.Errorf("create runtime staging directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(workdir) }()
	if err := os.Chmod(workdir, 0o750); err != nil {
		return fmt.Errorf("secure runtime staging directory: %w", err)
	}

	if err := extractRuntimeArchive(archivePath, workdir); err != nil {
		return err
	}
	stagedRuntime := filepath.Join(workdir, "runtime")
	if err := renderRuntimeCompose(stagedRuntime, manifest.Images); err != nil {
		return err
	}
	if err := validateRuntimeAssets(stagedRuntime, manifest.Database.Migration); err != nil {
		return err
	}

	releaseState := installedRuntimeRelease{
		SchemaVersion:  1,
		ReleaseVersion: manifest.ReleaseVersion,
		SourceCommit:   manifest.SourceCommit,
		RuntimeSHA256:  expectedSHA,
		ManifestSHA256: manifestSHA,
	}
	releaseBytes, err := json.MarshalIndent(releaseState, "", "  ")
	if err != nil {
		return fmt.Errorf("encode installed runtime metadata: %w", err)
	}
	releaseBytes = append(releaseBytes, '\n')
	if err := os.WriteFile(filepath.Join(stagedRuntime, "release.json"), releaseBytes, 0o640); err != nil {
		return fmt.Errorf("write installed runtime metadata: %w", err)
	}

	if err := os.Rename(stagedRuntime, runtimeDir); err != nil {
		return fmt.Errorf("install production runtime: %w", err)
	}
	if err := validateInstalledRuntime(runtimeDir, manifest, expectedSHA, manifestSHA); err != nil {
		if removeErr := os.RemoveAll(runtimeDir); removeErr != nil {
			return fmt.Errorf("validate installed runtime: %w; rollback runtime: %w", err, removeErr)
		}
		return fmt.Errorf("validate installed runtime: %w", err)
	}
	return nil
}

func installedRuntimeVersion(runtimeDir string) (string, error) {
	content, err := os.ReadFile(filepath.Join(runtimeDir, "release.json"))
	if err != nil {
		return "", fmt.Errorf("read installed runtime metadata: %w", err)
	}
	var installed installedRuntimeRelease
	if err := json.Unmarshal(content, &installed); err != nil {
		return "", fmt.Errorf("decode installed runtime metadata: %w", err)
	}
	if installed.SchemaVersion != 1 {
		return "", fmt.Errorf("unsupported installed runtime metadata schema: %d", installed.SchemaVersion)
	}
	if installed.ReleaseVersion == "" {
		return "", errors.New("installed runtime metadata has no release version")
	}
	return installed.ReleaseVersion, nil
}

func loadReleaseManifest(path, version string) (releaseManifest, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return releaseManifest{}, "", fmt.Errorf("read trusted release manifest: %w", err)
	}
	manifestSHA := sha256.Sum256(content)
	var manifest releaseManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return releaseManifest{}, "", fmt.Errorf("decode trusted release manifest: %w", err)
	}
	if manifest.SchemaVersion != 1 {
		return releaseManifest{}, "", fmt.Errorf("unsupported release manifest schema: %d", manifest.SchemaVersion)
	}
	if manifest.ReleaseVersion != version {
		return releaseManifest{}, "", fmt.Errorf("release manifest version %q does not match installed CLI %q", manifest.ReleaseVersion, version)
	}
	if !hex40Pattern.MatchString(manifest.SourceCommit) {
		return releaseManifest{}, "", errors.New("release manifest source commit is invalid")
	}
	if !migrationName.MatchString(manifest.Database.Migration) {
		return releaseManifest{}, "", errors.New("release manifest database migration boundary is invalid")
	}
	if err := validateManifestImages(manifest.Images); err != nil {
		return releaseManifest{}, "", err
	}
	return manifest, hex.EncodeToString(manifestSHA[:]), nil
}

func validateManifestImages(images map[string]string) error {
	if len(images) != len(runtimeImageTokens) {
		return fmt.Errorf("release manifest contains %d runtime images, want %d", len(images), len(runtimeImageTokens))
	}
	for name := range images {
		if _, ok := runtimeImageTokens[name]; !ok {
			return fmt.Errorf("release manifest contains unsupported runtime image %q", name)
		}
	}
	for name := range runtimeImageTokens {
		image := images[name]
		if !imagePattern.MatchString(image) {
			return fmt.Errorf("release manifest image %q is not immutable", name)
		}
	}
	return nil
}

func checksumFor(path, filename string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read trusted release checksums: %w", err)
	}
	for line := range strings.SplitSeq(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[1] != filename {
			continue
		}
		if !hex64Pattern.MatchString(fields[0]) {
			return "", fmt.Errorf("invalid checksum for %s", filename)
		}
		return fields[0], nil
	}
	return "", fmt.Errorf("trusted release checksums do not contain %s", filename)
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func extractRuntimeArchive(archivePath, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open staged runtime bundle: %w", err)
	}
	defer func() { _ = file.Close() }()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open staged runtime bundle gzip stream: %w", err)
	}
	defer func() { _ = gzipReader.Close() }()

	root, err := os.OpenRoot(destination)
	if err != nil {
		return fmt.Errorf("open runtime staging root: %w", err)
	}
	defer func() { _ = root.Close() }()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read staged runtime bundle: %w", err)
		}

		clean := filepath.Clean(filepath.FromSlash(header.Name))
		if clean == "." || !filepath.IsLocal(clean) {
			return fmt.Errorf("runtime bundle contains unsafe path %q", header.Name)
		}
		if clean != "runtime" && !strings.HasPrefix(clean, "runtime"+string(os.PathSeparator)) {
			return fmt.Errorf("runtime bundle contains unexpected top-level path %q", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := root.MkdirAll(clean, 0o750); err != nil {
				return fmt.Errorf("create runtime directory %s: %w", clean, err)
			}
		case tar.TypeReg:
			if header.Size < 0 || header.Size > 128<<20 {
				return fmt.Errorf("runtime bundle file %q has unsupported size %d", header.Name, header.Size)
			}
			parent := filepath.Dir(clean)
			if parent != "." {
				if err := root.MkdirAll(parent, 0o750); err != nil {
					return fmt.Errorf("create runtime file parent %s: %w", clean, err)
				}
			}
			output, err := root.OpenFile(clean, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
			if err != nil {
				return fmt.Errorf("create runtime file %s: %w", clean, err)
			}
			_, copyErr := io.CopyN(output, tarReader, header.Size)
			closeErr := output.Close()
			if copyErr != nil {
				return fmt.Errorf("extract runtime file %s: %w", clean, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("close runtime file %s: %w", clean, closeErr)
			}
		default:
			return fmt.Errorf("runtime bundle contains unsupported entry %q", header.Name)
		}
	}
	return nil
}

func renderRuntimeCompose(runtimeDir string, images map[string]string) error {
	templatePath := filepath.Join(runtimeDir, "compose.yaml.tmpl")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read runtime compose template: %w", err)
	}
	rendered := string(content)
	for name, token := range runtimeImageTokens {
		if !strings.Contains(rendered, token) {
			return fmt.Errorf("runtime compose template is missing %s", token)
		}
		rendered = strings.ReplaceAll(rendered, token, images[name])
	}
	if strings.Contains(rendered, "@@IMAGE_") {
		return errors.New("runtime compose contains unresolved image tokens")
	}
	for line := range strings.SplitSeq(rendered, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "build:") {
			return errors.New("production runtime contains a build directive")
		}
	}
	for _, frontend := range []string{"web", "console", "waitlist"} {
		if strings.Contains(rendered, "\n  "+frontend+":") {
			return fmt.Errorf("production runtime contains hosted frontend service %q", frontend)
		}
	}
	if err := os.WriteFile(filepath.Join(runtimeDir, "compose.yaml"), []byte(rendered), 0o640); err != nil {
		return fmt.Errorf("write rendered runtime compose: %w", err)
	}
	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("remove runtime compose template: %w", err)
	}
	return nil
}

func validateRuntimeAssets(runtimeDir, migration string) error {
	for _, path := range []string{
		filepath.Join(runtimeDir, "compose.yaml"),
		filepath.Join(runtimeDir, "coturn", "turnserver.conf"),
		filepath.Join(runtimeDir, "migrations", "atlas.sum"),
		filepath.Join(runtimeDir, "migrations", migration),
	} {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("runtime asset missing: %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("runtime asset is not a regular file: %s", path)
		}
	}
	return nil
}

func validateInstalledRuntime(runtimeDir string, manifest releaseManifest, runtimeSHA, manifestSHA string) error {
	if err := validateRuntimeAssets(runtimeDir, manifest.Database.Migration); err != nil {
		return err
	}
	content, err := os.ReadFile(filepath.Join(runtimeDir, "release.json"))
	if err != nil {
		return fmt.Errorf("read installed runtime metadata: %w", err)
	}
	var installed installedRuntimeRelease
	if err := json.Unmarshal(content, &installed); err != nil {
		return fmt.Errorf("decode installed runtime metadata: %w", err)
	}
	if installed.SchemaVersion != 1 {
		return fmt.Errorf("unsupported installed runtime metadata schema: %d", installed.SchemaVersion)
	}
	if installed.ReleaseVersion != manifest.ReleaseVersion || installed.SourceCommit != manifest.SourceCommit {
		return errors.New("installed runtime identity does not match trusted release manifest")
	}
	if installed.RuntimeSHA256 != runtimeSHA || installed.ManifestSHA256 != manifestSHA {
		return errors.New("installed runtime integrity metadata does not match trusted release metadata")
	}
	compose, err := os.ReadFile(filepath.Join(runtimeDir, "compose.yaml"))
	if err != nil {
		return fmt.Errorf("read installed runtime compose: %w", err)
	}
	if strings.Contains(string(compose), "@@IMAGE_") {
		return errors.New("installed runtime compose contains unresolved image tokens")
	}
	for name, image := range manifest.Images {
		if !strings.Contains(string(compose), image) {
			return fmt.Errorf("installed runtime compose is missing manifest image %q", name)
		}
	}
	return nil
}
