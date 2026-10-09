package audit

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/listquery"
)

type ListRequest struct {
	Action         *string
	ActorType      *string
	ActorID        *uuid.UUID
	TargetType     *string
	TargetID       *uuid.UUID
	OccurredFrom   *time.Time
	OccurredBefore *time.Time
	Offset         int32
	Limit          int32
}

func parseFilters(r *http.Request) (ListRequest, error) {
	p := listquery.Parser{Values: r.URL.Query()}
	req := ListRequest{
		Action:         p.Text("action"),
		ActorType:      p.Text("actor_type"),
		ActorID:        p.UUID("actor_id"),
		TargetType:     p.Text("target_type"),
		TargetID:       p.UUID("target_id"),
		OccurredFrom:   p.Time("occurred_from"),
		OccurredBefore: p.Time("occurred_before"),
	}
	return req, p.Err
}

func validateFilters(req ListRequest) error {
	if err := listquery.Enum("actor_type", req.ActorType, "user", "organization_token"); err != nil {
		return err
	}
	if err := listquery.ID("actor_id", req.ActorID); err != nil {
		return err
	}
	if err := listquery.ID("target_id", req.TargetID); err != nil {
		return err
	}
	return listquery.Range(req.OccurredFrom, req.OccurredBefore)
}
