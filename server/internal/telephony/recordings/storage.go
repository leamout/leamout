package recordings

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	s3integration "github.com/leamout/leamout/server/internal/integrations/s3"
	platformstorage "github.com/leamout/leamout/server/internal/platform/storage"
)

type ObjectStore interface {
	Put(context.Context, string, string, io.Reader, int64) error
	PlaybackURL(context.Context, string) (string, time.Time, error)
	Delete(context.Context, string) error
	Bucket() string
}

type ObjectStorage struct {
	integrations *platformstorage.Service

	// Test hooks; production always resolves tenant integrations and uses S3.
	resolveActive func(context.Context, uuid.UUID) (platformstorage.ResolvedIntegration, error)
	resolveByID   func(context.Context, uuid.UUID, uuid.UUID) (platformstorage.ResolvedIntegration, error)
	newClient     func(context.Context, platformstorage.ResolvedIntegration) (ObjectStore, error)
}

func NewObjectStorage(integrations *platformstorage.Service) *ObjectStorage {
	if integrations == nil {
		panic("recordings: storage integration service is required")
	}
	return &ObjectStorage{integrations: integrations}
}

func (s *ObjectStorage) ResolveUpload(
	ctx context.Context,
	recording sqlc.Recording,
) (*uuid.UUID, string, string, error) {
	client, integrationID, err := s.clientForRecording(ctx, recording)
	if err != nil {
		return nil, "", "", err
	}
	return integrationID, "s3", client.Bucket(), nil
}

func (s *ObjectStorage) Put(
	ctx context.Context,
	recording sqlc.Recording,
	key string,
	contentType string,
	reader io.Reader,
	size int64,
) (*uuid.UUID, string, string, error) {
	client, integrationID, err := s.clientForRecording(ctx, recording)
	if err != nil {
		return nil, "", "", err
	}
	if err := validateManagedRecordingKey(recording.OrganizationID, key); err != nil {
		return nil, "", "", err
	}
	if err := client.Put(ctx, key, contentType, reader, size); err != nil {
		return nil, "", "", err
	}
	return integrationID, "s3", client.Bucket(), nil
}

func (s *ObjectStorage) PlaybackURL(
	ctx context.Context,
	recording sqlc.Recording,
) (string, time.Time, error) {
	if recording.StorageKey == nil ||
		recording.StorageProvider == nil ||
		*recording.StorageProvider != "s3" {
		return "", time.Time{}, fmt.Errorf("recording has no S3 object")
	}
	client, _, err := s.clientForRecording(ctx, recording)
	if err != nil {
		return "", time.Time{}, err
	}
	if recording.StorageBucket == nil || *recording.StorageBucket != client.Bucket() {
		return "", time.Time{}, fmt.Errorf("recording S3 bucket is invalid")
	}
	if err := validateManagedRecordingKey(recording.OrganizationID, *recording.StorageKey); err != nil {
		return "", time.Time{}, err
	}
	return client.PlaybackURL(ctx, *recording.StorageKey)
}

func (s *ObjectStorage) Delete(ctx context.Context, recording sqlc.Recording) error {
	if recording.StorageKey == nil ||
		recording.StorageProvider == nil ||
		*recording.StorageProvider != "s3" {
		return fmt.Errorf("recording has no S3 object")
	}
	client, _, err := s.clientForRecording(ctx, recording)
	if err != nil {
		return err
	}
	if recording.StorageBucket == nil || *recording.StorageBucket != client.Bucket() {
		return fmt.Errorf("recording S3 bucket is invalid")
	}
	if err := validateManagedRecordingKey(recording.OrganizationID, *recording.StorageKey); err != nil {
		return err
	}
	return client.Delete(ctx, *recording.StorageKey)
}

func (s *ObjectStorage) clientForRecording(
	ctx context.Context,
	recording sqlc.Recording,
) (ObjectStore, *uuid.UUID, error) {
	var (
		resolved platformstorage.ResolvedIntegration
		err      error
	)
	if recording.StorageIntegrationID != nil {
		if s.resolveByID != nil {
			resolved, err = s.resolveByID(ctx, recording.OrganizationID, *recording.StorageIntegrationID)
		} else if s.integrations != nil {
			resolved, err = s.integrations.ResolveByID(ctx, recording.OrganizationID, *recording.StorageIntegrationID)
		} else {
			return nil, nil, fmt.Errorf("recording storage resolver is unavailable")
		}
		if err != nil {
			return nil, nil, fmt.Errorf("resolve recording BYOS integration: %w", err)
		}
	} else {
		// A pinned key without an integration belongs to managed storage.
		// Never silently reinterpret that key against a new customer bucket.
		if recording.StorageKey != nil {
			return nil, nil, fmt.Errorf("legacy managed recording requires BYOS migration")
		}
		if s.resolveActive != nil {
			resolved, err = s.resolveActive(ctx, recording.OrganizationID)
		} else if s.integrations != nil {
			resolved, err = s.integrations.ResolveRecording(ctx, recording.OrganizationID)
		} else {
			return nil, nil, fmt.Errorf("recording storage resolver is unavailable")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, fmt.Errorf("organization must configure active BYOS storage: %w", err)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("resolve organization BYOS storage: %w", err)
		}
	}

	if resolved.ID == uuid.Nil || resolved.OrganizationID != recording.OrganizationID ||
		(recording.StorageIntegrationID != nil && resolved.ID != *recording.StorageIntegrationID) {
		return nil, nil, fmt.Errorf("recording storage integration scope mismatch")
	}

	var client ObjectStore
	if s.newClient != nil {
		client, err = s.newClient(ctx, resolved)
	} else {
		client, err = customClient(ctx, resolved)
	}
	if err != nil {
		return nil, nil, err
	}
	if client == nil {
		return nil, nil, fmt.Errorf("recording object store unavailable")
	}
	id := resolved.ID
	return client, &id, nil
}

func customClient(ctx context.Context, resolved platformstorage.ResolvedIntegration) (ObjectStore, error) {
	client, err := s3integration.New(ctx, s3integration.Config{
		Endpoint:     resolved.EndpointURL,
		Region:       resolved.Region,
		Bucket:       resolved.Bucket,
		AccessKey:    resolved.AccessKeyID,
		SecretKey:    resolved.SecretAccessKey,
		UsePathStyle: resolved.UsePathStyle,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize recording S3 integration: %w", err)
	}
	return client, nil
}
