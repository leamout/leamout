package leamout

import (
	"path/filepath"
	"testing"
)

func TestRuntimeUpdateStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "update.json")
	want := runtimeUpdateState{
		SchemaVersion:   1,
		PreviousVersion: "1.0.0",
		TargetVersion:   "1.1.0",
		BackupPath:      "/var/lib/leamout/backups/pre-update.tar.gz",
		Phase:           "draining",
	}
	if err := writeRuntimeUpdateState(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := loadRuntimeUpdateState(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("loadRuntimeUpdateState() = %#v, want %#v", got, want)
	}
}

func TestRuntimeUpdateStateRejectsIncompleteState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "update.json")
	if err := writeAtomicFile(path, []byte(`{"schema_version":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadRuntimeUpdateState(path); err == nil {
		t.Fatal("expected incomplete update state to be rejected")
	}
}
