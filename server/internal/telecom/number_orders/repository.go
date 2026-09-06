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

	carrierConnectionID, providerRoutingResourceID, err := resolveProviderRoutingTarget(ctx, queries, provider.ID)
	if err != nil {
		return sqlc.NumberOrder{}, err
	}

	order, err := queries.CreateNumberOrder(ctx, sqlc.CreateNumberOrderParams{
		OrganizationID: organizationID,
		SelectionID:    selectionID,
		Number:         selection.Number,
		CountryCode:    selection.CountryCode,
	})
	if err != nil {
		return sqlc.NumberOrder{}, err
	}

	request, err := json.Marshal(ProviderOperationRequest{
		SelectionID:               selectionID,
		Provider:                  selection.Provider,
		ProviderInventoryID:       selection.ProviderInventoryID,
		ProviderProductID:         selection.ProviderProductID,
		Number:                    selection.Number,
		CountryCode:               selection.CountryCode,
		CarrierConnectionID:       carrierConnectionID,
		ProviderRoutingResourceID: providerRoutingResourceID,
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

func resolveProviderRoutingTarget(
	ctx context.Context,
	queries *sqlc.Queries,
	providerID uuid.UUID,
) (uuid.UUID, string, error) {
	targets, err := queries.ListProviderRoutingTargets(ctx, providerID)
	if err != nil {
		return uuid.Nil, "", err
	}
	if len(targets) != 1 {
		return uuid.Nil, "", ErrProviderRoutingUnavailable
	}
	return targets[0].CarrierConnectionID, targets[0].ProviderResourceID, nil
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

func (r *Repository) ListProviderOperationsReady(ctx context.Context, limit int32) ([]sqlc.ProviderOperation, error) {
	return r.queries.ListProviderOperationsReadyForRetry(ctx, limit)
}

func (r *Repository) TryProviderOperationLock(ctx context.Context, id uuid.UUID) (func(), bool, error) {
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, false, err
	}
	queries := sqlc.New(conn)
	locked, err := queries.TryProviderOperationAdvisoryLock(ctx, id.String())
	if err != nil {
		conn.Release()
		return nil, false, err
	}
	if !locked {
		conn.Release()
		return nil, false, nil
	}
	release := func() {
		_ = queries.ReleaseProviderOperationAdvisoryLock(context.Background(), id.String())
		conn.Release()
	}
	return release, true, nil
}

func (r *Repository) MarkProviderOperationAccepted(
	ctx context.Context,
	operation sqlc.ProviderOperation,
	providerOperationID string,
	response []byte,
) error {
	_, err := r.queries.MarkProviderOperationAccepted(ctx, sqlc.MarkProviderOperationAcceptedParams{
		ProviderOperationID: &providerOperationID,
		Response:            response,
		ID:                  operation.ID,
	})
	if errors.Is(err, pgx.ErrNoRows) && operation.State == "provider_accepted" {
		return nil
	}
	return err
}

func (r *Repository) MarkNumberOrderProcessing(ctx context.Context, operation sqlc.ProviderOperation) error {
	if operation.NumberOrderID == nil {
		return fmt.Errorf("number order provider operation is missing number_order_id")
	}
	_, err := r.queries.MarkNumberOrderProcessing(ctx, sqlc.MarkNumberOrderProcessingParams{
		ID:             *operation.NumberOrderID,
		OrganizationID: operation.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

func (r *Repository) RecordProviderOperationFailure(ctx context.Context, id uuid.UUID, err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	_, recordErr := r.queries.RecordProviderOperationAttemptFailure(ctx, sqlc.RecordProviderOperationAttemptFailureParams{
		LastError: &message,
		ID:        id,
	})
	return recordErr
}

func (r *Repository) FailProviderOperation(ctx context.Context, operation sqlc.ProviderOperation, cause error) error {
	if operation.NumberOrderID == nil {
		return fmt.Errorf("number order provider operation is missing number_order_id")
	}
	message := "provider number order failed"
	if cause != nil {
		message = cause.Error()
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	if err := queries.MarkProviderOperationFailed(ctx, sqlc.MarkProviderOperationFailedParams{
		LastError: &message,
		ID:        operation.ID,
	}); err != nil {
		return err
	}
	if err := queries.MarkNumberOrderFailed(ctx, sqlc.MarkNumberOrderFailedParams{
		ErrorMessage:   &message,
		ID:             *operation.NumberOrderID,
		OrganizationID: operation.OrganizationID,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) CompleteProviderOperation(
	ctx context.Context,
	operation sqlc.ProviderOperation,
	request ProviderOperationRequest,
	providerResourceID string,
	response []byte,
) error {
	if operation.NumberOrderID == nil {
		return fmt.Errorf("number order provider operation is missing number_order_id")
	}
	if operation.CarrierProviderID == uuid.Nil {
		return fmt.Errorf("number order provider operation is missing carrier_provider_id")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	order, err := queries.LockNumberOrderForProviderOperation(ctx, sqlc.LockNumberOrderForProviderOperationParams{
		ID:             *operation.NumberOrderID,
		OrganizationID: operation.OrganizationID,
		SelectionID:    request.SelectionID,
		Number:         request.Number,
		CountryCode:    request.CountryCode,
	})
	if err != nil {
		return err
	}
	if order.Status == "completed" {
		return tx.Commit(ctx)
	}
	if order.Status == "pending" {
		order, err = queries.MarkNumberOrderProcessing(ctx, sqlc.MarkNumberOrderProcessingParams{
			ID:             order.ID,
			OrganizationID: order.OrganizationID,
		})
		if err != nil {
			return err
		}
	}

	phoneNumber, err := queries.EnsureManagedPhoneNumberForProviderOperation(ctx, sqlc.EnsureManagedPhoneNumberForProviderOperationParams{
		OrganizationID:      operation.OrganizationID,
		Number:              request.Number,
		CountryCode:         request.CountryCode,
		ProviderID:          operation.CarrierProviderID,
		ProviderResourceID:  &providerResourceID,
		CarrierConnectionID: request.CarrierConnectionID,
	})
	if err != nil {
		return err
	}

	phoneNumberID := phoneNumber.ID
	if _, err := queries.MarkNumberOrderProviderOperationSucceeded(ctx, sqlc.MarkNumberOrderProviderOperationSucceededParams{
		ProviderResourceID: &providerResourceID,
		PhoneNumberID:      &phoneNumberID,
		Response:           response,
		ID:                 operation.ID,
		OrganizationID:     operation.OrganizationID,
		CarrierProviderID:  operation.CarrierProviderID,
		NumberOrderID:      operation.NumberOrderID,
	}); err != nil {
		return err
	}
	if _, err := queries.MarkNumberOrderCompleted(ctx, sqlc.MarkNumberOrderCompletedParams{
		PhoneNumberID:  &phoneNumberID,
		ID:             order.ID,
		OrganizationID: order.OrganizationID,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
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
