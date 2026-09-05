package leamout

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type backupManifest struct {
	SchemaVersion int       `json:"schema_version"`
	DeploymentID  string    `json:"deployment_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func runBackup(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	output := ""
	if len(args) == 2 && args[0] == "--output" {
		output = args[1]
	} else if len(args) != 0 {
		writeln(stderr, "usage: leamout backup [--output <path>]")
		return 2
	}
	state, err := loadDeploymentState("/var/lib/leamout/deployment.json")
	if err != nil {
		writef(stderr, "load deployment identity: %v\n", err)
		return 1
	}
	if output == "" {
		output = fmt.Sprintf("leamout-backup-%s-%s.tar.gz", state.DeploymentID, time.Now().UTC().Format("20060102T150405Z"))
	}
	cmd := installedComposeCommand(ctx, "exec", "-T", "postgres", "pg_dump", "-U", "leamout", "-d", "leamout", "--clean", "--if-exists")
	database, err := cmd.Output()
	if err != nil {
		writef(stderr, "dump database: %v\n", err)
		return 1
	}
	files := map[string]string{
		"deployment.json": "/var/lib/leamout/deployment.json",
		"leamout.env":     "/etc/leamout/leamout.env",
	}
	for _, name := range []string{"license.json", "keyring.json"} {
		path := filepath.Join("/etc/leamout/license", name)
		if _, err := os.Stat(path); err == nil {
			files[filepath.Join("license", name)] = path
		}
	}
	if err := createBackupArchive(output, state, files, database, time.Now().UTC()); err != nil {
		writef(stderr, "create backup: %v\n", err)
		return 1
	}
	writef(stdout, "✓ Backup created: %s\n", output)
	return 0
}

func runRestore(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	if len(args) != 2 || args[0] != "--force" {
		writeln(stderr, "usage: leamout restore --force <backup.tar.gz>")
		return 2
	}
	temp, err := os.MkdirTemp("/var/lib/leamout", ".restore-*")
	if err != nil {
		writef(stderr, "create restore staging directory: %v\n", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(temp) }()
	manifest, err := extractBackupArchive(args[1], temp)
	if err != nil {
		writef(stderr, "validate backup: %v\n", err)
		return 1
	}
	if state, err := loadDeploymentState("/var/lib/leamout/deployment.json"); err == nil && state.DeploymentID != manifest.DeploymentID {
		writeln(stderr, "backup belongs to another deployment; refusing restore")
		return 1
	}
	if code := runInstalledCompose(ctx, stdout, stderr, "down"); code != 0 {
		return code
	}
	if err := restoreConfiguration(temp); err != nil {
		writef(stderr, "restore configuration: %v\n", err)
		return 1
	}
	if code := runInstalledCompose(ctx, stdout, stderr, "up", "-d", "postgres"); code != 0 {
		return code
	}
	database, err := os.Open(filepath.Join(temp, "database.sql"))
	if err != nil {
		writef(stderr, "open database backup: %v\n", err)
		return 1
	}
	defer func() { _ = database.Close() }()
	cmd := installedComposeCommand(ctx, "exec", "-T", "postgres", "psql", "-v", "ON_ERROR_STOP=1", "-U", "leamout", "-d", "leamout")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = database, stdout, stderr
	if code := exitCode(cmd.Run(), stderr); code != 0 {
		return code
	}
	if code := runInstalledCompose(ctx, stdout, stderr, "up", "-d"); code != 0 {
		return code
	}
	writeln(stdout, "✓ Backup restored")
	return 0
}

func installedComposeCommand(ctx context.Context, args ...string) *exec.Cmd {
	base := []string{"compose", "--env-file", "/etc/leamout/leamout.env", "-f", "/var/lib/leamout/runtime/compose.yaml"}
	return exec.CommandContext(ctx, "docker", append(base, args...)...)
}

func createBackupArchive(output string, state deploymentState, files map[string]string, database []byte, now time.Time) error {
	if err := os.MkdirAll(filepath.Dir(output), 0o750); err != nil {
		return err
	}
	file, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = os.Remove(output)
		}
	}()
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	manifest, _ := json.Marshal(backupManifest{SchemaVersion: 1, DeploymentID: state.DeploymentID, CreatedAt: now.UTC()})
	entries := map[string][]byte{"manifest.json": append(manifest, '\n'), "database.sql": database}
	for name, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		entries[name] = content
	}
	for name, content := range entries {
		if !filepath.IsLocal(name) {
			return fmt.Errorf("unsafe backup entry %q", name)
		}
		if err := tw.WriteHeader(&tar.Header{Name: filepath.ToSlash(name), Mode: 0o600, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			return err
		}
		if _, err := tw.Write(content); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func extractBackupArchive(path, destination string) (backupManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return backupManifest{}, err
	}
	defer func() { _ = file.Close() }()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return backupManifest{}, err
	}
	defer func() { _ = gz.Close() }()
	allowed := map[string]bool{"manifest.json": true, "database.sql": true, "deployment.json": true, "leamout.env": true, "license/license.json": true, "license/keyring.json": true}
	tr := tar.NewReader(gz)
	seen := map[string]bool{}
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return backupManifest{}, err
		}
		name := filepath.Clean(filepath.FromSlash(h.Name))
		if !filepath.IsLocal(name) || !allowed[name] || h.Typeflag != tar.TypeReg || seen[name] {
			return backupManifest{}, fmt.Errorf("invalid backup entry %q", h.Name)
		}
		seen[name] = true
		target := filepath.Join(destination, name)
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return backupManifest{}, err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return backupManifest{}, err
		}
		_, copyErr := io.Copy(out, io.LimitReader(tr, 1<<30))
		closeErr := out.Close()
		if copyErr != nil {
			return backupManifest{}, copyErr
		}
		if closeErr != nil {
			return backupManifest{}, closeErr
		}
	}
	for _, required := range []string{"manifest.json", "database.sql", "deployment.json", "leamout.env"} {
		if !seen[required] {
			return backupManifest{}, fmt.Errorf("backup is missing %s", required)
		}
	}
	content, err := os.ReadFile(filepath.Join(destination, "manifest.json"))
	if err != nil {
		return backupManifest{}, err
	}
	var manifest backupManifest
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return backupManifest{}, err
	}
	if manifest.SchemaVersion != 1 || manifest.DeploymentID == "" {
		return backupManifest{}, errors.New("unsupported or incomplete backup manifest")
	}
	state, err := loadDeploymentState(filepath.Join(destination, "deployment.json"))
	if err != nil {
		return backupManifest{}, err
	}
	if state.DeploymentID != manifest.DeploymentID {
		return backupManifest{}, errors.New("backup deployment identity mismatch")
	}
	return manifest, nil
}

func restoreConfiguration(source string) error {
	for src, dst := range map[string]string{
		filepath.Join(source, "deployment.json"): "/var/lib/leamout/deployment.json",
		filepath.Join(source, "leamout.env"):     "/etc/leamout/leamout.env",
	} {
		content, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		if err := writeAtomicFile(dst, content, 0o600); err != nil {
			return err
		}
	}
	for _, name := range []string{"license.json", "keyring.json"} {
		src := filepath.Join(source, "license", name)
		if content, err := os.ReadFile(src); err == nil {
			if err := os.MkdirAll("/etc/leamout/license", 0o750); err != nil {
				return err
			}
			if err := writeAtomicFile(filepath.Join("/etc/leamout/license", name), content, 0o600); err != nil {
				return err
			}
		}
	}
	return nil
}
