package number_orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	redisv9 "github.com/redis/go-redis/v9"

	"github.com/leamout/leamout/internal/database/sqlc"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/telecom/numbers"
)

const managedNumberSelectionTTL = 10 * time.Minute

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
	redis   *redisintegration.Client
}

func NewRepository(db *pgxpool.Pool, redis *redisintegration.Client) *Repository {
	return &Repository{db: db, queries: sqlc.New(db), redis: redis}
}

func (r *Repository) SaveManagedSelection(
	ctx context.Context,
	organizationID uuid.UUID,
	candidate numbers.ManagedNumberCandidate,
) (string, error) {
	if r == nil || r.redis == nil {
		return "", fmt.Errorf("managed number selection store is unavailable")
	}
	if organizationID == uuid.Nil {
		return "", fmt.Errorf("organization id is required")
	}

	selectionID := "sel_" + uuid.NewString()
	if err := r.redis.SetJSON(ctx, selectionKey(organizationID, selectionID), candidate, managedNumberSelectionTTL); err != nil {
		return "", fmt.Errorf("store managed number selection: %w", err)
	}
	return selectionID, nil
}

func (r *Repository) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	selectionID string,
) (sqlc.NumberOrder, error) {
	selection, err := r.loadSelection(ctx, organizationID, selectionID)
	if err != nil {
		return sqlc.NumberOrder{}, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.NumberOrder{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	provider, err := queries.GetCarrierProviderBySlug(ctx, selection.Provider)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.NumberOrder{}, ErrSelectionUnavailable
	}
	if err != nil {
		return sqlc.NumberOrder{}, err
	}

	order, err := queries.CreateNumberOrder(ctx, sqlc.CreateNumberOrderParams{
		OrganizationID:      organizationID,
		ProviderInventoryID: selection.ProviderInventoryID,
		ProviderProductID:   selection.ProviderProductID,
		Number:              selection.Number,
		CountryCode:         selection.CountryCode,
		ProviderID:          provider.ID,
	})
	if err != nil {
		return sqlc.NumberOrder{}, err
	}

	request, err := json.Marshal(map[string]any{
		"available_did_id": selection.ProviderInventoryID,
		"sku_id":           selection.ProviderProductID,
		"number":           selection.Number,
		"country_code":     selection.CountryCode,
	})
	if err != nil {
		return sqlc.NumberOrder{}, fmt.Errorf("encode provider operation request: %w", err)
	}
	orderID := order.ID
	if _, err := queries.CreateNumberOrderProviderOperation(ctx, sqlc.CreateNumberOrderProviderOperationParams{
		OrganizationID:    organizationID,
		CarrierProviderID: provider.ID,
		NumberOrderID:     &orderID,
		IdempotencyKey:    "number-order:" + order.ID.String(),
		Request:           request,
	}); err != nil {
		return sqlc.NumberOrder{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.NumberOrder{}, err
	}
	_ = r.redis.Delete(ctx, selectionKey(organizationID, selectionID))
	return order, nil
}

func (r *Repository) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.NumberOrder, error) {
	return r.queries.GetNumberOrderByID(ctx, sqlc.GetNumberOrderByIDParams{
		ID:             id,
		OrganizationID: organizationID,
	})
}

func (r *Repository) loadSelection(
	ctx context.Context,
	organizationID uuid.UUID,
	selectionID string,
) (numbers.ManagedNumberCandidate, error) {
	if r == nil || r.redis == nil {
		return numbers.ManagedNumberCandidate{}, fmt.Errorf("managed number selection store is unavailable")
	}
	var selection numbers.ManagedNumberCandidate
	if err := r.redis.GetJSON(ctx, selectionKey(organizationID, selectionID), &selection); err != nil {
		if errors.Is(err, redisv9.Nil) {
			return numbers.ManagedNumberCandidate{}, ErrSelectionNotFound
		}
		return numbers.ManagedNumberCandidate{}, fmt.Errorf("load managed number selection: %w", err)
	}
	return selection, nil
}

func selectionKey(organizationID uuid.UUID, selectionID string) string {
	return "telecom:numbers:selection:" + organizationID.String() + ":" + selectionID
}
