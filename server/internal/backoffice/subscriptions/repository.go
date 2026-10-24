package subscriptions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/leamout/leamout/server/internal/database/sqlc"
)

// Repository owns explicit cross-organization operator read projections.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context, filter Filter) (List, error) {
	organizationID, err := optionalID(filter.OrganizationID)
	if err != nil {
		return List{}, err
	}
	rows, err := r.queries.ListBackofficeSubscriptions(ctx, sqlc.ListBackofficeSubscriptionsParams{
		Search:         filter.Search,
		OrganizationID: organizationID,
		PageLimit:      pageSize + 1,
		PageOffset:     (filter.Page - 1) * pageSize,
		Status:         filter.Status,
		Plan:           filter.Plan,
		Cancellation:   filter.Cancellation,
	})
	if err != nil {
		return List{}, err
	}
	result := List{
		Filter:  filter,
		HasNext: len(rows) > pageSize,
	}
	if result.HasNext {
		rows = rows[:pageSize]
	}
	for _, row := range rows {
		result.Subscriptions = append(result.Subscriptions, Subscription{
			ID:                 row.ID,
			OrganizationID:     row.OrganizationID,
			OrganizationName:   row.OrganizationName,
			OrganizationStatus: row.OrganizationStatus,
			PlanID:             row.PlanID,
			PlanName:           row.PlanName,
			PlanCode:           row.PlanCode,
			Status:             row.Status,
			PeriodEnd:          formatTime(row.CurrentPeriodEnd),
			CancelAtPeriodEnd:  row.CancelAtPeriodEnd,
			CreatedAt:          formatTime(row.CreatedAt),
		})
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeSubscription(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	result := Detail{
		Subscription: Subscription{
			ID:                 row.ID,
			OrganizationID:     row.OrganizationID,
			OrganizationName:   row.OrganizationName,
			OrganizationStatus: row.OrganizationStatus,
			PlanID:             row.PlanID,
			PlanName:           row.PlanName,
			PlanCode:           row.PlanCode,
			Status:             row.Status,
			PeriodEnd:          formatTime(row.CurrentPeriodEnd),
			CancelAtPeriodEnd:  row.CancelAtPeriodEnd,
			CreatedAt:          formatTime(row.CreatedAt),
		},
		PeriodStart:            formatTime(row.CurrentPeriodStart),
		TrialEndsAt:            formatTime(row.TrialEndsAt),
		PastDueAt:              formatTime(row.PastDueAt),
		GraceEndsAt:            formatTime(row.GraceEndsAt),
		CanceledAt:             formatTime(row.CanceledAt),
		UpdatedAt:              formatTime(row.UpdatedAt),
		ProviderEventCreatedAt: formatTime(row.ProviderEventCreatedAt),
		Provider:               row.Provider,
		ProviderCustomerID:     row.ProviderCustomerID,
		ProviderSubscriptionID: row.ProviderSubscriptionID,
	}
	return result, nil
}

func formatTime(value pgtype.Timestamptz) string {
	if !value.Valid {
		return "—"
	}
	return value.Time.UTC().Format("2006-01-02 15:04:05 UTC")
}

func optionalID(value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
