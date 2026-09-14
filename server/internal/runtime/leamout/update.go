package leamout

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const updateStatePath = "/var/lib/leamout/update.json"

type runtimeUpdateState struct {
	SchemaVersion   int    `json:"schema_version"`
	PreviousVersion string `json:"previous_version"`
	TargetVersion   string `json:"target_version"`
	BackupPath      string `json:"backup_path"`
	Phase           string `json:"phase"`
}

func runUpdate(ctx context.Context, stdout, stderr io.Writer, args []string, version string) int {
	if len(args) == 1 && args[0] == "--rollback" {
		return recoverInterruptedUpdate(ctx, stdout, stderr)
	}
	if len(args) != 0 {
		writeln(stderr, "usage: leamout update [--rollback]")
		return 2
	}
	if version == "" || version == "dev" {
		writeln(stderr, "development CLI cannot install a production update")
		return 1
	}
	if _, err := os.Stat(updateStatePath); err == nil {
		writeln(stderr, "an interrupted update requires recovery; run: leamout update --rollback")
		return 1
	} else if !os.IsNotExist(err) {
		writef(stderr, "inspect update recovery state: %v\n", err)
		return 1
	}
	state, err := ensureDeploymentIdentity("/var/lib/leamout/deployment.json")
	if err != nil {
		writef(stderr, "load deployment identity: %v\n", err)
		return 1
	}
	if err := validateRuntimeEnv("/etc/leamout/leamout.env", state.DeploymentID); err != nil {
		writef(stderr, "validate deployment configuration: %v\n", err)
		return 1
	}
	currentVersion, err := installedRuntimeVersion("/var/lib/leamout/runtime")
	if err != nil {
		writef(stderr, "load installed runtime version: %v\n", err)
		return 1
	}
	if currentVersion == version {
		writef(stdout, "✓ Runtime %s is already installed\n", version)
		return 0
	}
	comparison, err := compareReleaseVersions(version, currentVersion)
	if err != nil {
		writef(stderr, "compare runtime versions: %v\n", err)
		return 1
	}
	if comparison < 0 {
		writef(stderr, "refusing runtime downgrade from %s to %s\n", currentVersion, version)
		return 1
	}
	backupDir := "/var/lib/leamout/backups"
	if err := os.MkdirAll(backupDir, 0o750); err != nil {
		writef(stderr, "create update backup directory: %v\n", err)
		return 1
	}
	backupPath := filepath.Join(backupDir, fmt.Sprintf("pre-update-%s-to-%s-%s.tar.gz", currentVersion, version, time.Now().UTC().Format("20060102T150405.000000000Z")))
	if code := runBackup(ctx, stdout, stderr, []string{"--output", backupPath}); code != 0 {
		writeln(stderr, "update stopped because the pre-update backup failed")
		return code
	}
	updateState := runtimeUpdateState{SchemaVersion: 1, PreviousVersion: currentVersion, TargetVersion: version, BackupPath: backupPath, Phase: "staged"}
	if err := writeRuntimeUpdateState(updateStatePath, updateState); err != nil {
		writef(stderr, "persist update recovery state: %v\n", err)
		return 1
	}
	if err := installRuntimeBundle(filepath.Join("/var/lib/leamout/releases", version), "/var/lib/leamout/runtime", version); err != nil {
		writef(stderr, "install staged runtime: %v\n", err)
		_ = os.Remove(updateStatePath)
		return 1
	}
	updateState.Phase = "pulling"
	_ = writeRuntimeUpdateState(updateStatePath, updateState)
	if code := runInstalledCompose(ctx, stdout, stderr, "pull"); code != 0 {
		if rollbackRuntimeFiles(stderr) == nil {
			_ = os.Remove(updateStatePath)
		}
		return code
	}
	updateState.Phase = "draining"
	_ = writeRuntimeUpdateState(updateStatePath, updateState)
	if err := drainInstalledRuntime(ctx, stdout, stderr, 5*time.Minute, 2*time.Second); err != nil {
		writef(stderr, "drain runtime before update: %v\n", err)
		if rollbackRuntimeFiles(stderr) == nil {
			_ = os.Remove(updateStatePath)
		}
		return 1
	}
	updateState.Phase = "applying"
	_ = writeRuntimeUpdateState(updateStatePath, updateState)
	if code := runLicensedInstalledCompose(ctx, stdout, stderr, "up", "-d", "--remove-orphans", "--wait", "--wait-timeout", "120"); code != 0 {
		_ = resumeInstalledRuntime(ctx, stderr)
		rollbackUpdatedRuntime(ctx, stdout, stderr, backupPath)
		return code
	}
	if err := resumeInstalledRuntime(ctx, stderr); err != nil {
		writeln(stderr, "runtime restarted but telecom admission could not be fully resumed")
		rollbackUpdatedRuntime(ctx, stdout, stderr, backupPath)
		return 1
	}
	if err := commitRuntimeUpdate("/var/lib/leamout/runtime"); err != nil {
		writef(stderr, "finalize runtime update: %v\n", err)
		return 1
	}
	if err := os.Remove(updateStatePath); err != nil && !os.IsNotExist(err) {
		writef(stderr, "clear update recovery state: %v\n", err)
		return 1
	}
	writeln(stdout, "✓ Runtime update installed")
	writef(stdout, "✓ Pre-update backup retained: %s\n", backupPath)
	writeln(stdout, "Run: sudo leamout doctor")
	return 0
}

func writeRuntimeUpdateState(path string, state runtimeUpdateState) error {
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomicFile(path, append(content, '\n'), 0o600)
}

func loadRuntimeUpdateState(path string) (runtimeUpdateState, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return runtimeUpdateState{}, err
	}
	var state runtimeUpdateState
	if err := json.Unmarshal(content, &state); err != nil {
		return runtimeUpdateState{}, err
	}
	if state.SchemaVersion != 1 || state.PreviousVersion == "" || state.TargetVersion == "" || state.BackupPath == "" || state.Phase == "" {
		return runtimeUpdateState{}, fmt.Errorf("update recovery state is incomplete")
	}
	return state, nil
}

func recoverInterruptedUpdate(ctx context.Context, stdout, stderr io.Writer) int {
	state, err := loadRuntimeUpdateState(updateStatePath)
	if err != nil {
		writef(stderr, "load update recovery state: %v\n", err)
		return 1
	}
	currentVersion, err := installedRuntimeVersion("/var/lib/leamout/runtime")
	if err != nil {
		if rollbackErr := rollbackRuntimeUpdate("/var/lib/leamout/runtime"); rollbackErr != nil {
			writef(stderr, "load installed runtime version: %v; restore previous runtime: %v\n", err, rollbackErr)
			return 1
		}
		currentVersion = state.PreviousVersion
	}
	if currentVersion == state.PreviousVersion {
		if state.Phase == "applying" {
			if code := runRestore(ctx, stdout, stderr, []string{"--force", state.BackupPath}); code != 0 {
				writeln(stderr, "restore pre-update backup failed; recovery state was preserved")
				return code
			}
		} else {
			if err := os.Remove(updateStatePath); err != nil {
				writef(stderr, "clear update recovery state: %v\n", err)
				return 1
			}
			writeln(stdout, "✓ Cleared interrupted update; the previous runtime was still active")
			return 0
		}
		if err := os.Remove(updateStatePath); err != nil {
			writef(stderr, "clear update recovery state: %v\n", err)
			return 1
		}
		writeln(stdout, "✓ Restored the pre-update backup for the previous runtime")
		return 0
	}
	if currentVersion != state.TargetVersion {
		writef(stderr, "installed runtime %s does not match update recovery state\n", currentVersion)
		return 1
	}
	if err := rollbackRuntimeUpdate("/var/lib/leamout/runtime"); err != nil {
		writef(stderr, "restore previous runtime files: %v\n", err)
		return 1
	}
	if code := runRestore(ctx, stdout, stderr, []string{"--force", state.BackupPath}); code != 0 {
		writeln(stderr, "restore pre-update backup failed; recovery state was preserved")
		return code
	}
	if err := os.Remove(updateStatePath); err != nil {
		writef(stderr, "clear update recovery state: %v\n", err)
		return 1
	}
	writef(stdout, "✓ Rolled back interrupted update from %s to %s\n", state.TargetVersion, state.PreviousVersion)
	return 0
}

func rollbackRuntimeFiles(stderr io.Writer) error {
	if err := rollbackRuntimeUpdate("/var/lib/leamout/runtime"); err != nil {
		writef(stderr, "restore previous runtime files: %v\n", err)
		return err
	}
	return nil
}

func rollbackUpdatedRuntime(ctx context.Context, stdout, stderr io.Writer, backupPath string) {
	if err := rollbackRuntimeUpdate("/var/lib/leamout/runtime"); err != nil {
		writef(stderr, "restore previous runtime files: %v\n", err)
		return
	}
	if code := runRestore(ctx, stdout, stderr, []string{"--force", backupPath}); code != 0 {
		writeln(stderr, "automatic update rollback failed; the pre-update backup was preserved")
		return
	}
	_ = os.Remove(updateStatePath)
	writeln(stderr, "update failed; previous runtime and database were restored")
}
