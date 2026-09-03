package leamout

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateInstalledRuntimeFiles(runtimeDir string) error {
	for _, path := range []string{
		filepath.Join(runtimeDir, "compose.yaml"),
		filepath.Join(runtimeDir, "release.json"),
		filepath.Join(runtimeDir, "coturn", "turnserver.conf"),
		filepath.Join(runtimeDir, "migrations", "atlas.sum"),
	} {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("required runtime asset missing: %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("required runtime asset is not a regular file: %s", path)
		}
	}

	metadata, err := os.ReadFile(filepath.Join(runtimeDir, "release.json"))
	if err != nil {
		return err
	}
	var release installedRuntimeRelease
	if err := json.Unmarshal(metadata, &release); err != nil {
		return fmt.Errorf("decode installed runtime metadata: %w", err)
	}
	if release.SchemaVersion != 1 || release.ReleaseVersion == "" || release.SourceCommit == "" {
		return errors.New("installed runtime metadata is incomplete")
	}

	compose, err := os.ReadFile(filepath.Join(runtimeDir, "compose.yaml"))
	if err != nil {
		return err
	}
	if strings.Contains(string(compose), "@@IMAGE_") {
		return errors.New("installed runtime contains unresolved image references")
	}
	return nil
}
