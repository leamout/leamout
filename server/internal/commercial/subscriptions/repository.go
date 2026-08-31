package subscriptions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository persists subscription state within an organization boundary.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (Subscription, error) {
	row, err := r.queries.GetSubscription(ctx, sqlc.GetSubscriptionParams{
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Subscription{}, mapReadError(err)
	}
	return subscriptionFromRow(row), nil
}

func (r *Repository) Current(ctx context.Context, organizationID uuid.UUID) (Subscription, error) {
	row, err := r.queries.GetCurrentSubscription(ctx, organizationID)
	if err != nil {
		return Subscription{}, mapReadError(err)
	}
	return subscriptionFromRow(row), nil
}

func (r *Repository) List(ctx context.Context, organizationID uuid.UUID) ([]Subscription, error) {
	rows, err := r.queries.ListSubscriptionsByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	result := make([]Subscription, 0, len(rows))
	for _, row := range rows {
		result = append(result, subscriptionFromRow(row))
	}
	return result, nil
}

func (r *Repository) Create(ctx context.Context, organizationID uuid.UUID, input CreateInput) (Subscription, error) {
	status := (*string)(nil)
	if input.Status != nil {
		value := string(*input.Status)
		status = &value
	}
	var provider, providerID *string
	if input.Provider != nil {
		provider = &input.Provider.Provider
		providerID = &input.Provider.SubscriptionID
	}
	row, err := r.queries.CreateSubscription(ctx, sqlc.CreateSubscriptionParams{
		Status:                 status,
		StartsAt:               pgconv.NullableTimestamptz(input.StartsAt),
		RenewsAt:               pgconv.NullableTimestamptz(input.RenewsAt),
		EndsAt:                 pgconv.NullableTimestamptz(input.EndsAt),
		BillingProvider:        provider,
		ProviderSubscriptionID: providerID,
		PlanID:                 input.PlanID,
		OrganizationID:         organizationID,
	})
	if err != nil {
		return Subscription{}, mapWriteError(err)
	}
	return subscriptionFromRow(row), nil
}

func (r *Repository) UpdatePlan(ctx context.Context, organizationID, id, planID uuid.UUID) (Subscription, error) {
	row, err := r.queries.UpdateSubscription(ctx, sqlc.UpdateSubscriptionParams{
		PlanID:         &planID,
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Subscription{}, mapWriteError(err)
	}
	return subscriptionFromUpdateRow(row), nil
}

func (r *Repository) UpdatePeriod(ctx context.Context, organizationID, id uuid.UUID, input PeriodUpdate) (Subscription, error) {
	row, err := r.queries.UpdateSubscription(ctx, sqlc.UpdateSubscriptionParams{
		RenewsAt:       pgconv.NullableTimestamptz(input.RenewsAt),
		EndsAt:         pgconv.NullableTimestamptz(input.EndsAt),
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Subscription{}, mapWriteError(err)
	}
	return subscriptionFromUpdateRow(row), nil
}

func (r *Repository) UpdateStatus(ctx context.Context, organizationID, id uuid.UUID, status Status) (Subscription, error) {
	value := string(status)
	row, err := r.queries.UpdateSubscription(ctx, sqlc.UpdateSubscriptionParams{
		Status:         &value,
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Subscription{}, mapWriteError(err)
	}
	return subscriptionFromUpdateRow(row), nil
}

func (r *Repository) SetProvider(ctx context.Context, organizationID, id uuid.UUID, reference ProviderReference) (Subscription, error) {
	provider := reference.Provider
	providerID := reference.SubscriptionID
	row, err := r.queries.SetSubscriptionProvider(ctx, sqlc.SetSubscriptionProviderParams{
		BillingProvider:        &provider,
		ProviderSubscriptionID: &providerID,
		OrganizationID:         organizationID,
		ID:                     id,
	})
	if err != nil {
		return Subscription{}, mapWriteError(err)
	}
	return subscriptionFromProviderRow(row), nil
}

func (r *Repository) GetByProvider(ctx context.Context, reference ProviderReference) (Subscription, error) {
	provider := reference.Provider
	providerID := reference.SubscriptionID
	row, err := r.queries.GetSubscriptionByProviderID(ctx, sqlc.GetSubscriptionByProviderIDParams{
		BillingProvider:        &provider,
		ProviderSubscriptionID: &providerID,
	})
	if err != nil {
		return Subscription{}, mapReadError(err)
	}
	return subscriptionFromRow(row), nil
}

func subscriptionFromRow(row sqlc.Subscription) Subscription {
	return Subscription{
		ID:                     row.ID,
		OrganizationID:         row.OrganizationID,
		PlanID:                 row.PlanID,
		Status:                 Status(row.Status),
		StartsAt:               pgconv.TimestamptzToTime(row.StartsAt),
		RenewsAt:               pgconv.TimestamptzToTimePtr(row.RenewsAt),
		EndsAt:                 pgconv.TimestamptzToTimePtr(row.EndsAt),
		BillingProvider:        row.BillingProvider,
		ProviderSubscriptionID: row.ProviderSubscriptionID,
		CreatedAt:              pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:              pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func subscriptionFromUpdateRow(row sqlc.UpdateSubscriptionRow) Subscription {
	return Subscription{
		ID:                     row.ID_2,
		OrganizationID:         row.OrganizationID,
		PlanID:                 row.PlanID,
		Status:                 Status(row.Status_2),
		StartsAt:               pgconv.TimestamptzToTime(row.StartsAt),
		RenewsAt:               pgconv.TimestamptzToTimePtr(row.RenewsAt),
		EndsAt:                 pgconv.TimestamptzToTimePtr(row.EndsAt),
		BillingProvider:        row.BillingProvider,
		ProviderSubscriptionID: row.ProviderSubscriptionID,
		CreatedAt:              pgconv.TimestamptzToTime(row.CreatedAt_2),
		UpdatedAt:              pgconv.TimestamptzToTime(row.UpdatedAt_2),
	}
}

func subscriptionFromProviderRow(row sqlc.SetSubscriptionProviderRow) Subscription {
	return Subscription{
		ID:                     row.ID_2,
		OrganizationID:         row.OrganizationID,
		PlanID:                 row.PlanID,
		Status:                 Status(row.Status_2),
		StartsAt:               pgconv.TimestamptzToTime(row.StartsAt),
		RenewsAt:               pgconv.TimestamptzToTimePtr(row.RenewsAt),
		EndsAt:                 pgconv.TimestamptzToTimePtr(row.EndsAt),
		BillingProvider:        row.BillingProvider,
		ProviderSubscriptionID: row.ProviderSubscriptionID,
		CreatedAt:              pgconv.TimestamptzToTime(row.CreatedAt_2),
		UpdatedAt:              pgconv.TimestamptzToTime(row.UpdatedAt_2),
	}
}

func mapReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrSubscriptionNotFound
	}
	return err
}

func mapWriteError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrOrganizationUnavailable
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrProviderConflict
		case "23514":
			return ErrInvalidPeriod
		}
	}
	return err
}
