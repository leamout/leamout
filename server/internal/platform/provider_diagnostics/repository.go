package provider_diagnostics

import (
	"context"
	"time"

	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) Snapshot(ctx context.Context, limit int32) (Snapshot, error) {
	summary, err := r.queries.GetProviderOperationDiagnosticsSummary(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	operationRows, err := r.queries.ListProviderOperationDiagnostics(ctx, limit)
	if err != nil {
		return Snapshot{}, err
	}
	pollRows, err := r.queries.ListProviderCDRPollDiagnostics(ctx)
	if err != nil {
		return Snapshot{}, err
	}

	operations := make([]OperationDiagnostic, 0, len(operationRows))
	for _, row := range operationRows {
		operations = append(operations, OperationDiagnostic{
			ID:            row.ID.String(),
			Provider:      row.Provider,
			OperationType: row.OperationType,
			State:         row.State,
			Attempts:      row.Attempts,
			LastError:     row.LastError,
			NextAttemptAt: optionalTimestamp(row.NextAttemptAt.Valid, row.NextAttemptAt.Time),
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
		})
	}

	polling := make([]CDRPollDiagnostic, 0, len(pollRows))
	for _, row := range pollRows {
		polling = append(polling, CDRPollDiagnostic{
			Provider:           row.Provider,
			Direction:          row.Direction,
			WindowDate:         row.WindowDate.Time,
			Page:               row.Page,
			AttemptCount:       row.AttemptCount,
			NextAttemptAt:      row.NextAttemptAt.Time,
			LastError:          row.LastError,
			LastSuccessAt:      optionalTimestamp(row.LastSuccessAt.Valid, row.LastSuccessAt.Time),
			LastPageReceivedAt: optionalTimestamp(row.LastPageReceivedAt.Valid, row.LastPageReceivedAt.Time),
		})
	}

	return Snapshot{
		Operations: OperationSummary{
			PendingCount:    summary.PendingCount,
			AcceptedCount:   summary.AcceptedCount,
			FailedCount:     summary.FailedCount,
			RetryReadyCount: summary.RetryReadyCount,
		},
		Problematic: operations,
		CDRPolling:  polling,
	}, nil
}

func optionalTimestamp(valid bool, value time.Time) *time.Time {
	if !valid {
		return nil
	}
	return &value
}
