package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	s3integration "github.com/leamout/leamout/server/internal/integrations/s3"
)

// ObjectStorage performs operations against organization-owned S3-compatible
// buckets. It never selects a platform-owned fallback object store.
type ObjectStorage struct {
	integrations *Service
}

func NewObjectStorage(integrations *Service) *ObjectStorage {
	if integrations == nil {
		panic("storage: integration service is required")
	}
	return &ObjectStorage{integrations: integrations}
}

// ResolveRecording selects an active customer destination for a new upload.
// Callers must pin its integration ID and bucket before writing an object.
func (s *ObjectStorage) ResolveRecording(
	ctx context.Context,
	organizationID uuid.UUID,
) (Destination, error) {
	resolved, err := s.integrations.ResolveRecording(ctx, organizationID)
	if err != nil {
		return Destination{}, fmt.Errorf("resolve active recording storage: %w", err)
	}
	if err := validateResolvedDestination(resolved, organizationID); err != nil {
		return Destination{}, err
	}
	if resolved.Status != StatusActive {
		return Destination{}, fmt.Errorf("recording storage integration is not active")
	}
	return Destination{
		IntegrationID: resolved.ID,
		Provider:      ProviderS3,
		Bucket:        resolved.Bucket,
	}, nil
}

// Put writes to the destination already pinned to a recording. Historical
// integration IDs remain resolvable after the integration is disabled.
func (s *ObjectStorage) Put(
	ctx context.Context,
	organizationID uuid.UUID,
	integrationID uuid.UUID,
	bucket string,
	key string,
	contentType string,
	reader io.Reader,
	size int64,
) error {
	client, err := s.client(ctx, organizationID, integrationID, bucket)
	if err != nil {
		return err
	}
	return client.Put(ctx, key, contentType, reader, size)
}

func (s *ObjectStorage) PlaybackURL(
	ctx context.Context,
	organizationID uuid.UUID,
	integrationID uuid.UUID,
	bucket string,
	key string,
) (string, time.Time, error) {
	client, err := s.client(ctx, organizationID, integrationID, bucket)
	if err != nil {
		return "", time.Time{}, err
	}
	return client.PlaybackURL(ctx, key)
}

func (s *ObjectStorage) Delete(
	ctx context.Context,
	organizationID uuid.UUID,
	integrationID uuid.UUID,
	bucket string,
	key string,
) error {
	client, err := s.client(ctx, organizationID, integrationID, bucket)
	if err != nil {
		return err
	}
	return client.Delete(ctx, key)
}

func (s *ObjectStorage) client(
	ctx context.Context,
	organizationID uuid.UUID,
	integrationID uuid.UUID,
	bucket string,
) (*s3integration.Client, error) {
	if organizationID == uuid.Nil || integrationID == uuid.Nil || bucket == "" {
		return nil, fmt.Errorf("customer S3 integration, organization and bucket are required")
	}
	resolved, err := s.integrations.ResolveByID(ctx, organizationID, integrationID)
	if err != nil {
		return nil, fmt.Errorf("resolve pinned recording storage: %w", err)
	}
	if err := validateResolvedDestination(resolved, organizationID); err != nil {
		return nil, err
	}
	if resolved.ID != integrationID || resolved.Bucket != bucket {
		return nil, fmt.Errorf("recording storage destination does not match pinned integration")
	}
	client, err := s3integration.New(ctx, s3integration.Config{
		Endpoint:     resolved.EndpointURL,
		Region:       resolved.Region,
		Bucket:       resolved.Bucket,
		AccessKey:    resolved.AccessKeyID,
		SecretKey:    resolved.SecretAccessKey,
		UsePathStyle: resolved.UsePathStyle,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize customer S3 integration: %w", err)
	}
	return client, nil
}

func validateResolvedDestination(resolved ResolvedIntegration, organizationID uuid.UUID) error {
	if resolved.ID == uuid.Nil ||
		resolved.OrganizationID != organizationID ||
		resolved.Provider != ProviderS3 ||
		resolved.Purpose != PurposeRecordings ||
		resolved.Bucket == "" {
		return fmt.Errorf("recording storage integration scope or provider is invalid")
	}
	return nil
}
