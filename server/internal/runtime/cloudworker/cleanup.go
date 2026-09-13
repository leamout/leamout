package worker

import (
	"time"

	"github.com/leamout/leamout/internal/modules/idempotency"
)

func newIdempotencyCleanup(deps *dependencies) (*idempotency.CleanupJob, error) {
	job, err := idempotency.NewCleanupJob(
		idempotency.NewRepository(deps.queries),
		time.Hour,
	)
	if err != nil {
		return nil, wrapWorkerError("initialize idempotency cleanup job", err)
	}
	return job, nil
}
