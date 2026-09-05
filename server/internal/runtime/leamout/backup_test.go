package leamout

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBackupArchiveRoundTrip(t *testing.T) {
	root := t.TempDir()
	state := deploymentState{SchemaVersion: 1, DeploymentID: uuid.NewString(), Mode: deploymentMode, CreatedAt: time.Now().UTC()}
	stateBytes, _ := jsonMarshal(state)
	statePath := filepath.Join(root, "state.json")
	envPath := filepath.Join(root, "env")
	if err := os.WriteFile(statePath, stateBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(envPath, []byte("SECRET=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(root, "backup.tar.gz")
	if err := createBackupArchive(archive, state, map[string]string{"deployment.json": statePath, "leamout.env": envPath}, []byte("SELECT 1;\n"), time.Now()); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(archive); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("backup permissions: %v %v", info, err)
	}
	destination := filepath.Join(root, "restore")
	if err := os.Mkdir(destination, 0o750); err != nil {
		t.Fatal(err)
	}
	manifest, err := extractBackupArchive(archive, destination)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.DeploymentID != state.DeploymentID {
		t.Fatal("deployment ID mismatch")
	}
	database, _ := os.ReadFile(filepath.Join(destination, "database.sql"))
	if !bytes.Equal(database, []byte("SELECT 1;\n")) {
		t.Fatal("database dump changed")
	}
}

func jsonMarshal(value any) ([]byte, error) {
	content, err := json.Marshal(value)
	return append(content, '\n'), err
}
