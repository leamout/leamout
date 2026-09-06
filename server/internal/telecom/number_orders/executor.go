package number_orders

import (
	"context"
	"fmt"
	"strings"

	"github.com/leamout/leamout/internal/database/sqlc"
)

// ExecutionRouter dispatches durable managed-carrier operations to the executor
// selected when the operation was persisted. Hosting mode does not participate
// in this decision.
type ExecutionRouter struct {
	direct  ProviderOperationExecutor
	transit ProviderOperationExecutor
}

func NewExecutionRouter(
	direct ProviderOperationExecutor,
	transit ProviderOperationExecutor,
) *ExecutionRouter {
	return &ExecutionRouter{direct: direct, transit: transit}
}

func (r *ExecutionRouter) ExecuteProviderOperation(
	ctx context.Context,
	operation sqlc.ProviderOperation,
) error {
	target := strings.ToLower(strings.TrimSpace(operation.ExecutionTarget))
	switch target {
	case "direct":
		if r == nil || r.direct == nil {
			return fmt.Errorf("direct managed carrier executor is not configured")
		}
		return r.direct.ExecuteProviderOperation(ctx, operation)
	case "transit":
		if r == nil || r.transit == nil {
			return fmt.Errorf("transit managed carrier executor is not configured")
		}
		return r.transit.ExecuteProviderOperation(ctx, operation)
	default:
		return fmt.Errorf("unsupported provider operation execution target %q", operation.ExecutionTarget)
	}
}
