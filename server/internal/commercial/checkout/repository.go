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
	if err := validateCreate(input, time.Now()); err != nil {
		return Checkout{}, err
	}

	row, err := r.queries.CreateCheckoutOrder(ctx, sqlc.CreateCheckoutOrderParams{
		OrganizationID: organizationID,
		WalletID:       input.WalletID,
		PriceID:        input.PriceID,
		CheckoutType:   string(input.Type),
		Provider:       string(input.Provider),
		PaymentMethod:  string(input.PaymentMethod),
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

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (Checkout, error) {
	row, err := r.queries.GetCheckoutOrder(ctx, sqlc.GetCheckoutOrderParams{
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Checkout{}, mapReadError(err)
	}

	return checkoutFromRow(row), nil
}

func (r *Repository) GetByReference(ctx context.Context, reference string) (Checkout, error) {
	row, err := r.queries.GetCheckoutOrderByReference(ctx, reference)
	if err != nil {
		return Checkout{}, mapReadError(err)
	}

	return checkoutFromRow(row), nil
}

func (r *Repository) Transition(ctx context.Context, organizationID, id uuid.UUID, transition Transition) (Checkout, error) {
	if err := validateTransition(transition); err != nil {
		return Checkout{}, err
	}

	row, err := r.queries.CompareAndSetCheckoutOrderState(ctx, sqlc.CompareAndSetCheckoutOrderStateParams{
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
	row, err := r.queries.ClaimCheckoutOrderRefresh(ctx, sqlc.ClaimCheckoutOrderRefreshParams{
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
	rows, err := r.queries.ExpireCheckoutOrders(ctx)
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
		Provider:        Provider(row.Provider),
		PaymentMethod:   PaymentMethod(row.PaymentMethod),
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
