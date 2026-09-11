package payments

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/commercial/purchase"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

var (
	ErrPaymentNotFound = apperror.NewNotFound("payment not found")
	ErrPaymentConflict = apperror.NewConflict("payment already exists")
	ErrInvalidPayment  = apperror.NewBadRequest("invalid payment")
)

type Repository struct {
	db       *pgxpool.Pool
	queries  *sqlc.Queries
	purchase *purchase.Service
}

func NewRepository(db *pgxpool.Pool, purchaseService *purchase.Service) *Repository {
	return &Repository{db: db, queries: sqlc.New(db), purchase: purchaseService}
}

func (r *Repository) Create(ctx context.Context, organizationID uuid.UUID, provider string, input CreateInput) (Payment, error) {
	status := string(input.Status)
	row, err := r.queries.CreatePayment(ctx, sqlc.CreatePaymentParams{
		ProviderPaymentID: input.ProviderID,
		AmountMinor:       input.AmountMinor,
		Currency:          input.Currency,
		Status:            &status,
		Metadata:          input.Metadata,
		CheckoutID:        input.CheckoutID,
		OrganizationID:    organizationID,
		Provider:          provider,
	})
	if err != nil {
		return Payment{}, mapWriteError(err)
	}
	return paymentFromRow(row), nil
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (Payment, error) {
	row, err := r.queries.GetPayment(ctx, sqlc.GetPaymentParams{
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Payment{}, mapReadError(err)
	}
	return paymentFromRow(row), nil
}

func (r *Repository) GetByCheckout(ctx context.Context, organizationID, checkoutID uuid.UUID) (Payment, error) {
	row, err := r.queries.GetPaymentByCheckout(ctx, sqlc.GetPaymentByCheckoutParams{
		OrganizationID: organizationID,
		CheckoutID:     checkoutID,
	})
	if err != nil {
		return Payment{}, mapReadError(err)
	}
	return paymentFromRow(row), nil
}

func (r *Repository) SetProviderID(ctx context.Context, organizationID, id uuid.UUID, providerID string, status Status) (Payment, error) {
	row, err := r.queries.SetPaymentProviderID(ctx, sqlc.SetPaymentProviderIDParams{
		ProviderPaymentID: &providerID,
		Status:            string(status),
		OrganizationID:    organizationID,
		ID:                id,
	})
	if err != nil {
		return Payment{}, mapWriteError(err)
	}
	return paymentFromRow(row), nil
}

func (r *Repository) UpdateStatus(ctx context.Context, organizationID, id uuid.UUID, status Status, paidAt *time.Time) (Payment, error) {
	row, err := r.queries.UpdatePaymentStatus(ctx, sqlc.UpdatePaymentStatusParams{
		Status:         string(status),
		PaidAt:         pgconv.NullableTimestamptz(paidAt),
		OrganizationID: organizationID,
		ID:             id,
	})
	if err != nil {
		return Payment{}, mapWriteError(err)
	}
	return paymentFromRow(row), nil
}

func paymentFromRow(row sqlc.Payment) Payment {
	return Payment{
		ID:             row.ID,
		CheckoutID:     row.CheckoutID,
		OrganizationID: row.OrganizationID,
		Provider:       row.Provider,
		ProviderID:     row.ProviderPaymentID,
		Status:         Status(row.Status),
		AmountMinor:    row.AmountMinor,
		Currency:       row.Currency,
		PaidAt:         pgconv.TimestamptzToTimePtr(row.PaidAt),
		Metadata:       row.Metadata,
		CreatedAt:      pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:      pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func mapReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPaymentNotFound
	}
	return err
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrPaymentConflict
		}
		if pgErr.Code == "23503" || pgErr.Code == "23514" {
			return ErrInvalidPayment
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidPayment
	}
	return err
}
