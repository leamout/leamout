package managednumbers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/integrations/carriers/didww"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) BeginOrder(ctx context.Context, in OrderRequest) (Operation, error) {
	payload, err := json.Marshal(in)
	if err != nil {
		return Operation{}, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Operation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status
		FROM number_orders
		WHERE id=$1
		  AND organization_id=$2
		  AND provider_id=$3
		  AND provider_inventory_id=$4
		  AND provider_product_id=$5
		  AND number=$6
		  AND country_code=$7
		FOR UPDATE
	`, in.NumberOrderID, in.OrganizationID, in.ProviderID, strings.TrimSpace(in.AvailableDIDID), strings.TrimSpace(in.SKUID), strings.TrimSpace(in.Number), strings.ToUpper(strings.TrimSpace(in.CountryCode))).Scan(&status)
	if err != nil {
		return Operation{}, err
	}
	if status == "failed" {
		return Operation{}, fmt.Errorf("number order is terminally failed")
	}
	if status != "pending" && status != "processing" && status != "completed" {
		return Operation{}, fmt.Errorf("number order has invalid status %q", status)
	}

	var op Operation
	err = tx.QueryRow(ctx, `
		INSERT INTO provider_operations (
			organization_id, carrier_provider_id, operation_type,
			number_order_id, idempotency_key, request
		)
		VALUES ($1,$2,'number_order',$3,$4,$5)
		ON CONFLICT (organization_id,carrier_provider_id,idempotency_key)
		DO UPDATE SET idempotency_key=provider_operations.idempotency_key
		WHERE provider_operations.operation_type=EXCLUDED.operation_type
		  AND provider_operations.number_order_id=EXCLUDED.number_order_id
		  AND provider_operations.request=EXCLUDED.request
		RETURNING id,state,phone_number_id
	`, in.OrganizationID, in.ProviderID, in.NumberOrderID, in.IdempotencyKey, payload).Scan(&op.ID, &op.State, &op.PhoneNumberID)
	if err != nil {
		return Operation{}, err
	}

	if status == "pending" {
		tag, err := tx.Exec(ctx, `UPDATE number_orders SET status='processing' WHERE id=$1 AND organization_id=$2 AND status='pending'`, in.NumberOrderID, in.OrganizationID)
		if err != nil {
			return Operation{}, err
		}
		if tag.RowsAffected() != 1 {
			return Operation{}, pgx.ErrNoRows
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Operation{}, err
	}
	return op, nil
}

func (r *Repository) BeginRelease(ctx context.Context, in ReleaseRequest) (Operation, error) {
	payload, err := json.Marshal(in)
	if err != nil {
		return Operation{}, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Operation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var op Operation
	err = tx.QueryRow(ctx, `
		INSERT INTO provider_operations (
			organization_id, carrier_provider_id, operation_type,
			idempotency_key, request, phone_number_id, provider_resource_id
		)
		SELECT $1,$2,'number_release',$3,$4,pn.id,$6
		FROM phone_numbers AS pn
		WHERE pn.id=$5
		  AND pn.organization_id=$1
		  AND pn.provider_id=$2
		  AND pn.provider_resource_id=$6
		  AND pn.provisioning_mode='managed'
		  AND pn.status IN ('active','disabled')
		ON CONFLICT (organization_id,carrier_provider_id,idempotency_key)
		DO UPDATE SET idempotency_key=provider_operations.idempotency_key
		WHERE provider_operations.operation_type=EXCLUDED.operation_type
		  AND provider_operations.phone_number_id=EXCLUDED.phone_number_id
		  AND provider_operations.provider_resource_id=EXCLUDED.provider_resource_id
		  AND provider_operations.request=EXCLUDED.request
		RETURNING id,state,phone_number_id
	`, in.OrganizationID, in.ProviderID, in.IdempotencyKey, payload, in.PhoneNumberID, in.ProviderResourceID).Scan(&op.ID, &op.State, &op.PhoneNumberID)
	if err != nil {
		return Operation{}, err
	}
	if _, err = tx.Exec(ctx, `
		UPDATE phone_numbers
		SET status='disabled', voice_enabled=false, sms_enabled=false, updated_at=now()
		WHERE id=$1 AND organization_id=$2 AND provisioning_mode='managed' AND status IN ('active','disabled')
	`, in.PhoneNumberID, in.OrganizationID); err != nil {
		return Operation{}, err
	}
	return op, tx.Commit(ctx)
}

func (r *Repository) ProviderAccepted(ctx context.Context, id uuid.UUID, providerID string, response any) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE provider_operations
		SET state='provider_accepted', provider_operation_id=$2, response=$3,
		    attempts=attempts+1, last_error=NULL, next_attempt_at=now()
		WHERE id=$1 AND operation_type='number_order' AND state IN ('pending','provider_accepted')
	`, id, providerID, payload)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) RecordAttemptFailure(ctx context.Context, id uuid.UUID, cause error) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE provider_operations
		SET attempts=attempts+1, last_error=$2, next_attempt_at=now()+interval '5 minutes'
		WHERE id=$1 AND state IN ('pending','provider_accepted')
	`, id, cause.Error())
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) CompleteOrder(ctx context.Context, operationID uuid.UUID, did didww.DID, in OrderRequest) (uuid.UUID, error) {
	response, err := json.Marshal(did)
	if err != nil {
		return uuid.Nil, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var phoneNumberID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO phone_numbers (
			organization_id,number,country_code,provisioning_mode,
			carrier_connection_id,provider_id,provider_resource_id,
			voice_enabled,sms_enabled,status
		)
		SELECT $1,$2,$3,'managed',$4,$5,$6,true,false,'active'
		FROM carrier_connections AS cc
		WHERE cc.id=$4 AND cc.scope='platform' AND cc.provider_id=$5 AND cc.status='active'
		ON CONFLICT (provider_id,provider_resource_id)
			WHERE provider_id IS NOT NULL AND provider_resource_id IS NOT NULL
		DO UPDATE SET updated_at=phone_numbers.updated_at
		WHERE phone_numbers.organization_id=EXCLUDED.organization_id AND phone_numbers.status<>'released'
		RETURNING id
	`, in.OrganizationID, "+"+strings.TrimPrefix(did.Number, "+"), strings.ToUpper(in.CountryCode), in.IngressConnectionID, in.ProviderID, did.ID).Scan(&phoneNumberID)
	if err != nil {
		return uuid.Nil, err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE provider_operations
		SET state='succeeded', provider_resource_id=$2, phone_number_id=$3,
		    response=$4, completed_at=now(), last_error=NULL, next_attempt_at=NULL
		WHERE id=$1 AND organization_id=$5 AND carrier_provider_id=$6
		  AND operation_type='number_order' AND number_order_id=$7 AND state='provider_accepted'
	`, operationID, did.ID, phoneNumberID, response, in.OrganizationID, in.ProviderID, in.NumberOrderID)
	if err != nil {
		return uuid.Nil, err
	}
	if tag.RowsAffected() != 1 {
		return uuid.Nil, pgx.ErrNoRows
	}

	tag, err = tx.Exec(ctx, `
		UPDATE number_orders
		SET status='completed', phone_number_id=$2, error_code=NULL, error_message=NULL
		WHERE id=$1 AND organization_id=$3 AND provider_id=$4 AND status='processing'
	`, in.NumberOrderID, phoneNumberID, in.OrganizationID, in.ProviderID)
	if err != nil {
		return uuid.Nil, err
	}
	if tag.RowsAffected() != 1 {
		return uuid.Nil, pgx.ErrNoRows
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return phoneNumberID, nil
}

func (r *Repository) CompleteRelease(ctx context.Context, operationID, numberID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE phone_numbers
		SET status='released', carrier_connection_id=NULL,
		    voice_enabled=false, sms_enabled=false, updated_at=now()
		WHERE id=$1 AND provisioning_mode='managed' AND status IN ('active','disabled')
	`, numberID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}

	tag, err = tx.Exec(ctx, `
		UPDATE provider_operations
		SET state='succeeded', completed_at=now(), last_error=NULL, next_attempt_at=NULL
		WHERE id=$1 AND phone_number_id=$2 AND operation_type='number_release'
		  AND state IN ('pending','provider_accepted')
	`, operationID, numberID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return tx.Commit(ctx)
}
