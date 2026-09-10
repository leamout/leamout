package orders

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (Order, error) {
	row, err := r.queries.GetOrder(ctx, sqlc.GetOrderParams{
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, err
	}
	return orderFromRow(row), nil
}

func (r *Repository) List(ctx context.Context, organizationID uuid.UUID) ([]Order, error) {
	rows, err := r.queries.ListOrdersByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	result := make([]Order, 0, len(rows))
	for _, row := range rows {
		result = append(result, orderFromRow(row))
	}
	return result, nil
}

func orderFromRow(row sqlc.Order) Order {
	return Order{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		CheckoutID:     row.CheckoutID,
		PaymentID:      row.PaymentID,
		WalletID:       row.WalletID,
		PriceID:        row.PriceID,
		Type:           Type(row.OrderType),
		AmountMinor:    row.AmountMinor,
		Currency:       row.Currency,
		CompletedAt:    pgconv.TimestamptzToTime(row.CompletedAt),
		Metadata:       row.Metadata,
		CreatedAt:      pgconv.TimestamptzToTime(row.CreatedAt),
	}
}
