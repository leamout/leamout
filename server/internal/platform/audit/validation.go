package audit

import (
	"github.com/leamout/leamout/server/pkg/listquery"
)

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
