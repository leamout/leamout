package checkout

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) Create(ctx context.Context, organizationID uuid.UUID, input CreateInput) (Checkout, error) {
	row, err := r.queries.CreateCheckout(ctx, sqlc.CreateCheckoutParams{
		OrganizationID: organizationID,
		WalletID:       input.WalletID,
		PriceID:        input.PriceID,
		CheckoutType:   string(input.Type),
		Reference:      input.Reference,
		AmountMinor:    input.AmountMinor,
		Currency:       input.Currency,
		ExpiresAt:      pgconv.NullableTimestamptz(&input.ExpiresAt),
		Metadata:       input.Metadata,
	})
	if err != nil {
		return Checkout{}, mapWriteError(err)
	}

	return checkoutFromRow(row), nil
}

func (r *Repository) StartPayment(
	ctx context.Context,
	organizationID, id uuid.UUID,
	input StartPayment,
) (Checkout, error) {
	row, err := r.queries.StartCheckoutPayment(ctx, sqlc.StartCheckoutPaymentParams{
		Provider:       string(input.Provider),
		PaymentMethod:  string(input.PaymentMethod),
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Checkout{}, ErrInvalidTransition
		}
		return Checkout{}, mapWriteError(err)
	}
	return checkoutFromRow(row), nil
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (Checkout, error) {
	row, err := r.queries.GetCheckout(ctx, sqlc.GetCheckoutParams{
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Checkout{}, mapReadError(err)
	}

	return checkoutFromRow(row), nil
}

func (r *Repository) GetByReference(ctx context.Context, reference string) (Checkout, error) {
	row, err := r.queries.GetCheckoutByReference(ctx, reference)
	if err != nil {
		return Checkout{}, mapReadError(err)
	}

	return checkoutFromRow(row), nil
}

func (r *Repository) Transition(ctx context.Context, organizationID, id uuid.UUID, transition Transition) (Checkout, error) {
	row, err := r.queries.CompareAndSetCheckoutState(ctx, sqlc.CompareAndSetCheckoutStateParams{
		Status:          string(transition.Status),
		NextAction:      string(transition.NextAction),
		ProviderMessage: transition.ProviderMessage,
		CompletedAt:     pgconv.NullableTimestamptz(transition.CompletedAt),
		OrganizationID:  organizationID,
		ID:              id,
		ExpectedStatus:  string(transition.Expected),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Checkout{}, ErrInvalidTransition
		}
		return Checkout{}, mapWriteError(err)
	}

	return checkoutFromRow(row), nil
}

func (r *Repository) ClaimRefresh(ctx context.Context, organizationID, id uuid.UUID, refreshBefore time.Time) (Checkout, error) {
	row, err := r.queries.ClaimCheckoutRefresh(ctx, sqlc.ClaimCheckoutRefreshParams{
		OrganizationID: organizationID,
		ID:             id,
		RefreshBefore:  pgconv.NullableTimestamptz(&refreshBefore),
	})
	if err != nil {
		return Checkout{}, mapReadError(err)
	}

	return checkoutFromRow(row), nil
}

func (r *Repository) Expire(ctx context.Context) ([]Checkout, error) {
	rows, err := r.queries.ExpireCheckouts(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Checkout, 0, len(rows))
	for _, row := range rows {
		result = append(result, checkoutFromRow(row))
	}

	return result, nil
}

func checkoutFromRow(row sqlc.Checkout) Checkout {
	return Checkout{
		ID:              row.ID,
		OrganizationID:  row.OrganizationID,
		WalletID:        row.WalletID,
		PriceID:         row.PriceID,
		Type:            Type(row.CheckoutType),
		Provider:        Provider(nullableString(row.Provider)),
		PaymentMethod:   PaymentMethod(nullableString(row.PaymentMethod)),
		Reference:       row.Reference,
		AmountMinor:     row.AmountMinor,
		Currency:        row.Currency,
		Status:          Status(row.Status),
		NextAction:      NextAction(row.NextAction),
		ProviderMessage: row.ProviderMessage,
		ExpiresAt:       pgconv.TimestamptzToTime(row.ExpiresAt),
		CompletedAt:     pgconv.TimestamptzToTimePtr(row.CompletedAt),
		Metadata:        row.Metadata,
		CreatedAt:       pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func nullableString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func mapReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCheckoutNotFound
	}
	return err
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.ConstraintName == "checkouts_reference_key" {
			return ErrReferenceConflict
		}
		if pgErr.Code == "23514" || pgErr.Code == "23503" {
			return ErrInvalidCheckout
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidCheckout
	}
	return err
}
