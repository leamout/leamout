package recordings

import (
	"context"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"errors"
	"fmt"
	"log"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	platformstorage "github.com/leamout/leamout/server/internal/platform/storage"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

type ReconciliationJobConfig struct {
	Interval  time.Duration
	Grace     time.Duration
	BatchSize int32
}

func DefaultReconciliationJobConfig() ReconciliationJobConfig {
	return ReconciliationJobConfig{
		Interval:  30 * time.Second,
		Grace:     60 * time.Second,
		BatchSize: 200,
	}
}

type reconciliationRepository interface {
	ListForReconciliation(context.Context, time.Time, int32) ([]sqlc.Recording, error)
	MarkReadyForUpload(context.Context, sqlc.Recording, time.Time) (sqlc.Recording, error)
}

type ReconciliationJob struct {
	repo   reconciliationRepository
	config ReconciliationJobConfig
	now    func() time.Time
}

func NewReconciliationJob(
	repo reconciliationRepository,
	config ReconciliationJobConfig,
) (*ReconciliationJob, error) {
	if repo == nil {
		return nil, fmt.Errorf("recording reconciliation repository is required")
	}
	if config.Interval <= 0 {
		config.Interval = 30 * time.Second
	}
	if config.Grace <= 0 {
		config.Grace = 60 * time.Second
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 200
	}

	return &ReconciliationJob{
		repo:   repo,
		config: config,
		now:    time.Now,
	}, nil
}

func (j *ReconciliationJob) Run(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("recording reconciliation context is required")
	}

	j.runPass(ctx)

	ticker := time.NewTicker(j.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			j.runPass(ctx)
		}
	}
}

func (j *ReconciliationJob) runPass(ctx context.Context) {
	if err := j.Reconcile(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("recording reconciliation pass failed: %v", err)
	}
}

func (j *ReconciliationJob) Reconcile(ctx context.Context) error {
	updatedBefore := j.now().UTC().Add(-j.config.Grace)
	stale, err := j.repo.ListForReconciliation(ctx, updatedBefore, j.config.BatchSize)
	if err != nil {
		return fmt.Errorf("list recordings for reconciliation: %w", err)
	}

	for _, recording := range stale {
		_, err := j.repo.MarkReadyForUpload(ctx, recording, j.now().UTC())
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("reconcile recording %s: %w", recording.ID, err)
		}
	}

	return nil
}

// DefaultStagingPath is the shared FreeSWITCH/worker recording volume.
const DefaultStagingPath = "/var/lib/freeswitch/recordings"

type IngestionConfig struct {
	Interval    time.Duration
	Lease       time.Duration
	BaseRetry   time.Duration
	MaxRetry    time.Duration
	MaxAttempts int32
	BatchSize   int32
	StagingPath string
}

func DefaultIngestionConfig(stagingPath string) IngestionConfig {
	return IngestionConfig{
		Interval:    5 * time.Second,
		Lease:       2 * time.Minute,
		BaseRetry:   5 * time.Second,
		MaxRetry:    15 * time.Minute,
		MaxAttempts: 10,
		BatchSize:   50,
		StagingPath: stagingPath,
	}
}

type ingestionRepository interface {
	ListForUpload(context.Context, time.Time, time.Time, int32) ([]sqlc.Recording, error)
	PinUpload(
		context.Context,
		sqlc.Recording,
		*uuid.UUID,
		string,
		string,
		string,
	) (sqlc.Recording, error)
	CompleteUpload(
		context.Context,
		sqlc.Recording,
		string,
		int64,
	) (sqlc.Recording, error)
	RetryUpload(context.Context, sqlc.Recording, time.Time, string) error
	Fail(context.Context, sqlc.Recording) (sqlc.Recording, error)
}

type IngestionJob struct {
	repo    ingestionRepository
	storage *platformstorage.ObjectStorage
	config  IngestionConfig
	now     func() time.Time
}

func NewIngestionJob(
	repo ingestionRepository,
	storage *platformstorage.ObjectStorage,
	config IngestionConfig,
) (*IngestionJob, error) {
	if repo == nil || storage == nil {
		return nil, fmt.Errorf("recording ingestion requires repository and storage")
	}
	defaults := DefaultIngestionConfig(config.StagingPath)
	if config.Interval <= 0 {
		config.Interval = defaults.Interval
	}
	if config.Lease <= 0 {
		config.Lease = defaults.Lease
	}
	if config.BaseRetry <= 0 {
		config.BaseRetry = defaults.BaseRetry
	}
	if config.MaxRetry <= 0 {
		config.MaxRetry = defaults.MaxRetry
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = defaults.MaxAttempts
	}
	if config.BatchSize <= 0 {
		config.BatchSize = defaults.BatchSize
	}
	config.StagingPath = filepath.Clean(strings.TrimSpace(config.StagingPath))
	if !filepath.IsAbs(config.StagingPath) {
		return nil, fmt.Errorf("recording staging path must be absolute")
	}
	return &IngestionJob{
		repo:    repo,
		storage: storage,
		config:  config,
		now:     time.Now,
	}, nil
}

func (j *IngestionJob) Run(ctx context.Context) error {
	j.runPass(ctx)
	ticker := time.NewTicker(j.config.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			j.runPass(ctx)
		}
	}
}

func (j *IngestionJob) runPass(ctx context.Context) {
	if err := j.Ingest(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("recording ingestion pass failed: %v", err)
	}
}

func (j *IngestionJob) Ingest(ctx context.Context) error {
	now := j.now().UTC()
	items, err := j.repo.ListForUpload(
		ctx,
		now,
		now.Add(j.config.Lease),
		j.config.BatchSize,
	)
	if err != nil {
		return fmt.Errorf("claim recording uploads: %w", err)
	}
	for _, item := range items {
		if err := j.ingestOne(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (j *IngestionJob) ingestOne(
	ctx context.Context,
	recording sqlc.Recording,
) error {
	path, err := j.safeSourcePath(recording)
	if err != nil {
		return j.retryOrFail(ctx, recording, err)
	}
	file, err := os.Open(path)
	if err != nil {
		return j.retryOrFail(
			ctx,
			recording,
			fmt.Errorf("open staged recording: %w", err),
		)
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return j.retryOrFail(
			ctx,
			recording,
			fmt.Errorf("stat staged recording: %w", err),
		)
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 {
		return j.retryOrFail(
			ctx,
			recording,
			fmt.Errorf("staged recording is not a non-empty regular file"),
		)
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if ext == "" {
		ext = "bin"
	}

	key := ""
	if recording.StorageKey != nil {
		key = *recording.StorageKey
	} else {
		key, err = recordingObjectKey(
			recording.OrganizationID,
			recording.ID,
			j.now().UTC(),
			ext,
		)
		if err != nil {
			return j.retryOrFail(ctx, recording, err)
		}

		destination, resolveErr := j.storage.ResolveRecording(ctx, recording.OrganizationID)
		if resolveErr != nil {
			return j.retryOrFail(ctx, recording, resolveErr)
		}
		recording, err = j.repo.PinUpload(
			ctx,
			recording,
			&destination.IntegrationID,
			key,
			destination.Provider,
			destination.Bucket,
		)
		if err != nil {
			return j.retryOrFail(
				ctx,
				recording,
				fmt.Errorf("pin recording upload destination: %w", err),
			)
		}
	}

	contentType := mime.TypeByExtension("." + ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	integrationID, bucket, pinnedKey, err := recordingStorageDestination(recording)
	if err != nil {
		return j.retryOrFail(ctx, recording, err)
	}
	if pinnedKey != key {
		return j.retryOrFail(ctx, recording, fmt.Errorf("recording key differs from pinned BYOS destination"))
	}
	if err := j.storage.Put(
		ctx,
		recording.OrganizationID,
		integrationID,
		bucket,
		key,
		contentType,
		file,
		info.Size(),
	); err != nil {
		return j.retryOrFail(ctx, recording, err)
	}
	if _, err := j.repo.CompleteUpload(ctx, recording, ext, info.Size()); err != nil {
		return fmt.Errorf("complete recording upload %s: %w", recording.ID, err)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("remove staged recording %s: %v", recording.ID, err)
	}
	return nil
}

func (j *IngestionJob) safeSourcePath(
	recording sqlc.Recording,
) (string, error) {
	if recording.SourcePath == nil {
		return "", fmt.Errorf("recording source path is missing")
	}
	path := filepath.Clean(strings.TrimSpace(*recording.SourcePath))
	relative, err := filepath.Rel(j.config.StagingPath, path)
	if err != nil || relative == "." || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("recording source path is outside staging directory")
	}
	resolvedRoot, err := filepath.EvalSymlinks(j.config.StagingPath)
	if err != nil {
		return "", fmt.Errorf("resolve recording staging directory: %w", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve staged recording: %w", err)
	}
	resolvedRelative, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || resolvedRelative == "." || resolvedRelative == ".." ||
		strings.HasPrefix(resolvedRelative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("recording source path resolves outside staging directory")
	}
	return resolvedPath, nil
}

func (j *IngestionJob) retryOrFail(
	ctx context.Context,
	recording sqlc.Recording,
	cause error,
) error {
	if recording.UploadAttempts+1 >= j.config.MaxAttempts {
		if _, err := j.repo.Fail(ctx, recording); err != nil &&
			!errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("fail recording upload %s: %w", recording.ID, err)
		}
		return nil
	}
	delay := j.config.BaseRetry
	for attempt := int32(0); attempt < recording.UploadAttempts &&
		delay < j.config.MaxRetry; attempt++ {
		delay *= 2
	}
	if delay > j.config.MaxRetry {
		delay = j.config.MaxRetry
	}
	message := cause.Error()
	if len(message) > 1000 {
		message = message[:1000]
	}
	if err := j.repo.RetryUpload(
		ctx,
		recording,
		j.now().UTC().Add(delay),
		message,
	); err != nil {
		return fmt.Errorf("schedule recording upload retry %s: %w", recording.ID, err)
	}
	return nil
}

func recordingObjectKey(
	organizationID uuid.UUID,
	recordingID uuid.UUID,
	occurredAt time.Time,
	extension string,
) (string, error) {
	prefix, err := recordingStoragePrefix(organizationID)
	if err != nil {
		return "", err
	}
	if recordingID == uuid.Nil {
		return "", fmt.Errorf("recording id is required")
	}
	extension = strings.TrimPrefix(strings.TrimSpace(extension), ".")
	if extension == "" {
		extension = "bin"
	}
	if strings.ContainsAny(extension, `/\\`) {
		return "", fmt.Errorf("recording extension is invalid")
	}

	return prefix + occurredAt.UTC().Format("2006/01/02") + "/" +
		recordingID.String() + "." + extension, nil
}

