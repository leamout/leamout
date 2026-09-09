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

type Repository struct{ queries *sqlc.Queries }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{queries: sqlc.New(db)} }

func (r *Repository) Create(ctx context.Context, organizationID uuid.UUID, input CreateInput) (Order, error) {
	if err := validateCreate(input, time.Now()); err != nil {
		return Order{}, err
	}
	row, err := r.queries.CreateCheckoutOrder(ctx, sqlc.CreateCheckoutOrderParams{
		OrganizationID: organizationID, WalletID: input.WalletID, PriceID: input.PriceID,
		InvoiceID: input.InvoiceID, OrderType: string(input.Type), Provider: string(input.Provider),
		PaymentMethod: string(input.PaymentMethod), Reference: input.Reference, Amount: input.AmountMinor,
		Currency: input.Currency, ExpiresAt: pgconv.NullableTimestamptz(&input.ExpiresAt), Metadata: input.Metadata,
	})
	if err != nil {
		return Order{}, mapWriteError(err)
	}
	return orderFromRow(row), nil
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (Order, error) {
	row, err := r.queries.GetCheckoutOrder(ctx, sqlc.GetCheckoutOrderParams{OrganizationID: organizationID, ID: id})
	if err != nil {
		return Order{}, mapReadError(err)
	}
	return orderFromRow(row), nil
}

func (r *Repository) GetByReference(ctx context.Context, reference string) (Order, error) {
	row, err := r.queries.GetCheckoutOrderByReference(ctx, reference)
	if err != nil {
		return Order{}, mapReadError(err)
	}
	return orderFromRow(row), nil
}

func (r *Repository) Transition(ctx context.Context, organizationID, id uuid.UUID, transition Transition) (Order, error) {
	if err := validateTransition(transition); err != nil {
		return Order{}, err
	}
	row, err := r.queries.CompareAndSetCheckoutOrderState(ctx, sqlc.CompareAndSetCheckoutOrderStateParams{
		Status: string(transition.Status), NextAction: string(transition.NextAction),
		ProviderMessage: transition.ProviderMessage, CompletedAt: pgconv.NullableTimestamptz(transition.CompletedAt),
		OrganizationID: organizationID, ID: id, ExpectedStatus: string(transition.Expected),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrInvalidTransition
		}
		return Order{}, mapWriteError(err)
	}
	return orderFromRow(row), nil
}

func (r *Repository) Expire(ctx context.Context) ([]Order, error) {
	rows, err := r.queries.ExpireCheckoutOrders(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Order, 0, len(rows))
	for _, row := range rows {
		result = append(result, orderFromRow(row))
	}
	return result, nil
}

func orderFromRow(row sqlc.CheckoutOrder) Order {
	return Order{ID: row.ID, OrganizationID: row.OrganizationID, WalletID: row.WalletID, PriceID: row.PriceID, InvoiceID: row.InvoiceID, Type: OrderType(row.OrderType), Provider: Provider(row.Provider), PaymentMethod: PaymentMethod(row.PaymentMethod), Reference: row.Reference, AmountMinor: row.Amount, Currency: row.Currency, Status: Status(row.Status), NextAction: NextAction(row.NextAction), ProviderMessage: row.ProviderMessage, ExpiresAt: pgconv.TimestamptzToTime(row.ExpiresAt), CompletedAt: pgconv.TimestamptzToTimePtr(row.CompletedAt), Metadata: row.Metadata, CreatedAt: pgconv.TimestamptzToTime(row.CreatedAt), UpdatedAt: pgconv.TimestamptzToTime(row.UpdatedAt)}
}

func mapReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrOrderNotFound
	}
	return err
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.ConstraintName == "checkout_orders_reference_key" {
			return ErrReferenceConflict
		}
		if pgErr.Code == "23514" || pgErr.Code == "23503" {
			return ErrInvalidOrder
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidOrder
	}
	return err
}
