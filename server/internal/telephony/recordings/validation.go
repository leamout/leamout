package recordings

import (
	"github.com/google/uuid"
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
