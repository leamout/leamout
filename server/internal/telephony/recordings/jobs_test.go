package recordings

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	platformstorage "github.com/leamout/leamout/server/internal/platform/storage"
)

func TestNewReconciliationJobRequiresRepository(t *testing.T) {
	_, err := NewReconciliationJob(nil, ReconciliationJobConfig{})
	if err == nil {
		t.Fatal("expected missing repository error")
	}
}

func TestNewReconciliationJobDefaults(t *testing.T) {
	job, err := NewReconciliationJob(&Repository{}, ReconciliationJobConfig{})
	if err != nil {
		t.Fatal(err)
	}
	want := DefaultReconciliationJobConfig()
	if job.config.Interval != want.Interval ||
		job.config.Grace != want.Grace ||
		job.config.BatchSize != want.BatchSize {
		t.Fatalf("job config = %+v, want %+v", job.config, want)
	}
}

func TestNewIngestionJobRequiresDependencies(t *testing.T) {
	_, err := NewIngestionJob(nil, &platformstorage.ObjectStorage{}, DefaultIngestionConfig(DefaultStagingPath))
	if err == nil {
		t.Fatal("expected missing repository error")
	}

	_, err = NewIngestionJob(&Repository{}, nil, DefaultIngestionConfig(DefaultStagingPath))
	if err == nil {
		t.Fatal("expected missing storage error")
	}
}

func TestNewIngestionJobDefaults(t *testing.T) {
	job, err := NewIngestionJob(
		&Repository{},
		&platformstorage.ObjectStorage{},
		IngestionConfig{StagingPath: DefaultStagingPath},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := DefaultIngestionConfig(DefaultStagingPath)
	if job.config != want {
		t.Fatalf("job config = %+v, want %+v", job.config, want)
	}
}

func TestNewIngestionJobRejectsRelativeStagingPath(t *testing.T) {
	_, err := NewIngestionJob(
		&Repository{},
		&platformstorage.ObjectStorage{},
		IngestionConfig{StagingPath: "recordings"},
	)
	if err == nil {
		t.Fatal("expected relative staging path to be rejected")
	}
}

func TestIngestionSafeSourcePathRejectsTraversal(t *testing.T) {
	job := &IngestionJob{
		config: IngestionConfig{StagingPath: DefaultStagingPath},
	}
	source := "/var/lib/freeswitch/recordings/../outside.wav"
	_, err := job.safeSourcePath(sqlc.Recording{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		SourcePath:     &source,
	})
	if err == nil {
		t.Fatal("expected source path outside staging directory to be rejected")
	}
}

func TestDefaultIngestionConfigUsesBoundedRetries(t *testing.T) {
	config := DefaultIngestionConfig(DefaultStagingPath)
	if config.Interval <= 0 ||
		config.Lease <= 0 ||
		config.BaseRetry <= 0 ||
		config.MaxRetry < config.BaseRetry ||
		config.MaxAttempts <= 0 ||
		config.BatchSize <= 0 {
		t.Fatalf("unexpected ingestion defaults: %+v", config)
	}
	if config.MaxRetry > time.Hour {
		t.Fatalf("maximum retry interval too long: %s", config.MaxRetry)
	}
}

