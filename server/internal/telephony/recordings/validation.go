package recordings

import (
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/listquery"
)

func validateOrganizationID(id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.NewBadRequest("organization context required")
	}
	return nil
}
func validatePagination(offset, limit int32) error {
	if offset < 0 {
		return apperror.NewBadRequest("offset cannot be negative")
	}
	if limit < 1 || limit > 100 {
		return apperror.NewBadRequest("limit must be between 1 and 100")
	}
	return nil
}

func validateFilters(req ListRequest) error {
	if err := listquery.Enum("status", req.Status, "recording", "uploading", "completed", "failed"); err != nil {
		return err
	}
	if err := listquery.ID("call_id", req.CallID); err != nil {
		return err
	}
	return listquery.Range(req.CreatedFrom, req.CreatedBefore)
}


func recordingStoragePrefix(organizationID uuid.UUID) (string, error) {
	if organizationID == uuid.Nil {
		return "", fmt.Errorf("recording organization id is required")
	}
	return "organizations/" + organizationID.String() + "/recordings/", nil
}

func validateRecordingObjectKey(organizationID uuid.UUID, key string) error {
	prefix, err := recordingStoragePrefix(organizationID)
	if err != nil {
		return err
	}
	if key == "" || path.IsAbs(key) || path.Clean(key) != key {
		return fmt.Errorf("recording storage key is invalid")
	}
	if !strings.HasPrefix(key, prefix) || key == prefix {
		return fmt.Errorf("recording storage key is outside organization prefix")
	}
	return nil
}

func recordingStorageDestination(recording sqlc.Recording) (uuid.UUID, string, string, error) {
	if recording.StorageIntegrationID == nil ||
		*recording.StorageIntegrationID == uuid.Nil ||
		recording.StorageProvider == nil ||
		*recording.StorageProvider != "s3" ||
		recording.StorageBucket == nil ||
		*recording.StorageBucket == "" ||
		recording.StorageKey == nil {
		return uuid.Nil, "", "", fmt.Errorf("recording does not have a pinned customer S3 destination")
	}
	if err := validateRecordingObjectKey(recording.OrganizationID, *recording.StorageKey); err != nil {
		return uuid.Nil, "", "", err
	}
	return *recording.StorageIntegrationID, *recording.StorageBucket, *recording.StorageKey, nil
}
