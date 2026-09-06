package number_orders

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/leamout/leamout/internal/database/sqlc"
)

func (r *Repository) CompleteTransitProviderOperation(
	ctx context.Context,
	operation sqlc.ProviderOperation,
	selectionID string,
	number string,
	countryCode string,
	managedResourceID string,
	response []byte,
) error {
	if operation.NumberOrderID == nil {
		return fmt.Errorf("transit number order provider operation is missing number_order_id")
	}
	if operation.ExecutionTarget != "transit" || operation.CarrierProviderID != nil {
		return fmt.Errorf("transit number order provider operation has invalid execution target")
	}
	if managedResourceID == "" {
		return fmt.Errorf("transit managed resource id is required")
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
		SelectionID:    selectionID,
		Number:         number,
		CountryCode:    countryCode,
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

	var phoneNumberID uuid.UUID
	err = tx.QueryRow(ctx, `
INSERT INTO phone_numbers (
    organization_id,
    number,
    country_code,
    provisioning_mode,
    carrier_connection_id,
    provider_id,
    provider_resource_id,
    voice_enabled,
    sms_enabled,
    status
)
SELECT
    $1,
    $2,
    $3,
    'managed',
    NULL,
    NULL,
    NULL,
    true,
    false,
    'active'
FROM organizations AS o
WHERE o.id = $1
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ON CONFLICT (number)
    WHERE status <> 'released'
DO UPDATE SET updated_at = phone_numbers.updated_at
WHERE phone_numbers.organization_id = EXCLUDED.organization_id
  AND phone_numbers.provisioning_mode = 'managed'
  AND phone_numbers.provider_id IS NULL
  AND phone_numbers.provider_resource_id IS NULL
  AND phone_numbers.status = 'active'
RETURNING id`,
		operation.OrganizationID,
		number,
		countryCode,
	).Scan(&phoneNumberID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("transit managed number conflicts with an existing phone number: %w", err)
		}
		return err
	}

	providerResourceID := managedResourceID
	id := operation.NumberOrderID
	phoneID := phoneNumberID
	if _, err := queries.MarkNumberOrderProviderOperationSucceeded(ctx, sqlc.MarkNumberOrderProviderOperationSucceededParams{
		ProviderResourceID: &providerResourceID,
		PhoneNumberID:      &phoneID,
		Response:           response,
		ID:                 operation.ID,
		OrganizationID:     operation.OrganizationID,
		NumberOrderID:      id,
	}); err != nil {
		return err
	}
	if _, err := queries.MarkNumberOrderCompleted(ctx, sqlc.MarkNumberOrderCompletedParams{
		PhoneNumberID:  &phoneID,
		ID:             order.ID,
		OrganizationID: order.OrganizationID,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
