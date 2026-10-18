package recordings

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	platformstorage "github.com/leamout/leamout/server/internal/platform/storage"
)

type fakeObjectStore struct {
	putKey, deletedKey string
	putBody            []byte
	putErr             error
}

func (f *fakeObjectStore) Put(_ context.Context, key, _ string, reader io.Reader, _ int64) error {
	f.putKey = key
	f.putBody, _ = io.ReadAll(reader)
	return f.putErr
}
func (f *fakeObjectStore) PlaybackURL(context.Context, string) (string, time.Time, error) {
	return "https://storage.example/recording?signature=test", time.Now().Add(time.Minute), nil
}
func (f *fakeObjectStore) Delete(_ context.Context, key string) error { f.deletedKey = key; return nil }
func (f *fakeObjectStore) Bucket() string                             { return "recordings" }

// newTestBYOSStorage emulates tenant-scoped S3 integration lookups without
// starting a platform-managed object store.
func newTestBYOSStorage(store *fakeObjectStore, organizationID uuid.UUID) (*ObjectStorage, uuid.UUID) {
	integrationID := uuid.New()
	resolved := platformstorage.ResolvedIntegration{
		Integration: platformstorage.Integration{
			ID:             integrationID,
			OrganizationID: organizationID,
			Provider:       platformstorage.ProviderS3,
			Purpose:        platformstorage.PurposeRecordings,
			Bucket:         store.Bucket(),
			Status:         platformstorage.StatusActive,
		},
	}
	return &ObjectStorage{
		resolveActive: func(_ context.Context, orgID uuid.UUID) (platformstorage.ResolvedIntegration, error) {
			if orgID != organizationID {
				return platformstorage.ResolvedIntegration{}, fmt.Errorf("wrong organization")
			}
			return resolved, nil
		},
		resolveByID: func(_ context.Context, orgID, id uuid.UUID) (platformstorage.ResolvedIntegration, error) {
			if orgID != organizationID || id != integrationID {
				return platformstorage.ResolvedIntegration{}, fmt.Errorf("wrong integration scope")
			}
			return resolved, nil
		},
		newClient: func(context.Context, platformstorage.ResolvedIntegration) (ObjectStore, error) {
			return store, nil
		},
	}, integrationID
}

func TestObjectStoragePlaybackAndDelete(t *testing.T) {
	client := &fakeObjectStore{}
	organizationID := uuid.New()
	storage, integrationID := newTestBYOSStorage(client, organizationID)
	key := "organizations/" + organizationID.String() + "/recordings/2026/10/02/id.wav"
	provider, bucket := "s3", "recordings"
	recording := sqlc.Recording{
		OrganizationID:       organizationID,
		StorageIntegrationID: &integrationID,
		StorageKey:           &key,
		StorageProvider:      &provider,
		StorageBucket:        &bucket,
	}
	url, expires, err := storage.PlaybackURL(context.Background(), recording)
	if err != nil {
		t.Fatalf("PlaybackURL() error = %v", err)
	}
	if url == "" || expires.IsZero() {
		t.Fatal("PlaybackURL() returned invalid result")
	}
	if err := storage.Delete(context.Background(), recording); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if client.deletedKey != key {
		t.Fatalf("deleted key = %q", client.deletedKey)
	}
}

func TestObjectStorageRejectsLegacyManagedRecording(t *testing.T) {
	organizationID := uuid.New()
	storage, _ := newTestBYOSStorage(&fakeObjectStore{}, organizationID)
	key := "organizations/" + organizationID.String() + "/recordings/old.wav"
	provider, bucket := "s3", "recordings"
	recording := sqlc.Recording{
		OrganizationID:  organizationID,
		StorageKey:      &key,
		StorageProvider: &provider,
		StorageBucket:   &bucket,
	}
	if _, _, err := storage.PlaybackURL(context.Background(), recording); err == nil {
		t.Fatal("legacy managed recording was allowed to use BYOS")
	}
	if err := storage.Delete(context.Background(), recording); err == nil {
		t.Fatal("legacy managed recording was allowed to delete from BYOS")
	}
}

func TestObjectStorageRejectsMissingBYOS(t *testing.T) {
	organizationID := uuid.New()
	storage := &ObjectStorage{
		resolveActive: func(context.Context, uuid.UUID) (platformstorage.ResolvedIntegration, error) {
			return platformstorage.ResolvedIntegration{}, pgx.ErrNoRows
		},
	}
	_, _, _, err := storage.ResolveUpload(context.Background(), sqlc.Recording{OrganizationID: organizationID})
	if err == nil || !strings.Contains(err.Error(), "active BYOS storage") {
		t.Fatalf("missing BYOS = %v", err)
	}
}

func TestObjectStorageRejectsNonS3Recording(t *testing.T) {
	storage, _ := newTestBYOSStorage(&fakeObjectStore{}, uuid.New())
	if _, _, err := storage.PlaybackURL(context.Background(), sqlc.Recording{}); err == nil {
		t.Fatal("expected playback error")
	}
	if err := storage.Delete(context.Background(), sqlc.Recording{}); err == nil {
		t.Fatal("expected delete error")
	}
}
