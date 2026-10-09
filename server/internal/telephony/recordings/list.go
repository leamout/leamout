package recordings

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/listquery"
)

type ListRequest struct {
	Status        *string
	CallID        *uuid.UUID
	CreatedFrom   *time.Time
	CreatedBefore *time.Time
	Offset        int32
	Limit         int32
}

func parseFilters(r *http.Request) (ListRequest, error) {
	p := listquery.Parser{Values: r.URL.Query()}
	req := ListRequest{
		Status:        p.Text("status"),
		CallID:        p.UUID("call_id"),
		CreatedFrom:   p.Time("created_from"),
		CreatedBefore: p.Time("created_before"),
	}
	return req, p.Err
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
