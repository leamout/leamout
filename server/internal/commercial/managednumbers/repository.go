package managednumbers

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/integrations/carriers/didww"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) BeginOrder(ctx context.Context, in OrderRequest) (Operation, error) {
	payload, _ := json.Marshal(in)
	return r.begin(ctx, in.OrganizationID, in.ProviderID, "number_order", in.IdempotencyKey, payload, nil)
}
func (r *Repository) BeginRelease(ctx context.Context, in ReleaseRequest) (Operation, error) {
	payload, _ := json.Marshal(in)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Operation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var op Operation
	err = tx.QueryRow(ctx, `INSERT INTO provider_operations (organization_id,carrier_provider_id,operation_type,idempotency_key,request,phone_number_id) SELECT $1,$2,'number_release',$3,$4,pn.id FROM phone_numbers pn WHERE pn.id=$5 AND pn.organization_id=$1 AND pn.provider_id=$2 AND pn.provider_resource_id=$6 AND pn.provisioning_mode='managed' AND pn.status IN ('active','disabled','releasing') ON CONFLICT (carrier_provider_id,idempotency_key) DO UPDATE SET idempotency_key=provider_operations.idempotency_key WHERE provider_operations.organization_id=EXCLUDED.organization_id AND provider_operations.operation_type=EXCLUDED.operation_type RETURNING id,state,phone_number_id`, in.OrganizationID, in.ProviderID, in.IdempotencyKey, payload, in.PhoneNumberID, in.ProviderResourceID).Scan(&op.ID, &op.State, &op.PhoneNumberID)
	if err != nil {
		return Operation{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE phone_numbers SET status='releasing',voice_enabled=false,sms_enabled=false,updated_at=now() WHERE id=$1 AND organization_id=$2 AND provisioning_mode='managed' AND status IN ('active','disabled')`, in.PhoneNumberID, in.OrganizationID); err != nil {
		return Operation{}, err
	}
	return op, tx.Commit(ctx)
}
func (r *Repository) begin(ctx context.Context, org, provider uuid.UUID, kind, key string, payload []byte, numberID *uuid.UUID) (Operation, error) {
	var op Operation
	err := r.db.QueryRow(ctx, `INSERT INTO provider_operations (organization_id,carrier_provider_id,operation_type,idempotency_key,request,phone_number_id) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (carrier_provider_id,idempotency_key) DO UPDATE SET idempotency_key=provider_operations.idempotency_key WHERE provider_operations.organization_id=EXCLUDED.organization_id AND provider_operations.operation_type=EXCLUDED.operation_type RETURNING id,state,phone_number_id`, org, provider, kind, key, payload, numberID).Scan(&op.ID, &op.State, &op.PhoneNumberID)
	return op, err
}
func (r *Repository) ProviderAccepted(ctx context.Context, id uuid.UUID, providerID string, response any) error {
	payload, _ := json.Marshal(response)
	tag, err := r.db.Exec(ctx, `UPDATE provider_operations SET state='provider_accepted',provider_operation_id=$2,response=$3,attempts=attempts+1,last_error=NULL,next_attempt_at=now() WHERE id=$1 AND state IN ('pending','failed','provider_accepted')`, id, providerID, payload)
	if err == nil && tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return err
}
func (r *Repository) Fail(ctx context.Context, id uuid.UUID, cause error) error {
	_, err := r.db.Exec(ctx, `UPDATE provider_operations SET state='failed',attempts=attempts+1,last_error=$2,next_attempt_at=now()+interval '5 minutes' WHERE id=$1 AND state<>'succeeded'`, id, cause.Error())
	return err
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
	var id uuid.UUID
	err = tx.QueryRow(ctx, `INSERT INTO phone_numbers (organization_id,number,country_code,provisioning_mode,carrier_connection_id,provider_id,provider_resource_id,voice_enabled,sms_enabled,status) SELECT $1,$2,$3,'managed',$4,$5,$6,true,false,'active' FROM carrier_connections cc WHERE cc.id=$4 AND cc.scope='platform' AND cc.provider_id=$5 AND cc.status='active' ON CONFLICT (provider_id,provider_resource_id) DO UPDATE SET updated_at=phone_numbers.updated_at WHERE phone_numbers.organization_id=EXCLUDED.organization_id AND phone_numbers.status<>'released' RETURNING id`, in.OrganizationID, "+"+strings.TrimPrefix(did.Number, "+"), strings.ToUpper(in.CountryCode), in.IngressConnectionID, in.ProviderID, did.ID).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	tag, err := tx.Exec(ctx, `UPDATE provider_operations SET state='succeeded',provider_resource_id=$2,phone_number_id=$3,response=$4,completed_at=now(),last_error=NULL,next_attempt_at=NULL WHERE id=$1 AND organization_id=$5 AND carrier_provider_id=$6`, operationID, did.ID, id, response, in.OrganizationID, in.ProviderID)
	if err != nil || tag.RowsAffected() != 1 {
		if err == nil {
			err = pgx.ErrNoRows
		}
		return uuid.Nil, err
	}
	return id, tx.Commit(ctx)
}
func (r *Repository) CompleteRelease(ctx context.Context, operationID, numberID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE phone_numbers SET status='released',carrier_connection_id=NULL,voice_enabled=false,sms_enabled=false,updated_at=now() WHERE id=$1 AND provisioning_mode='managed' AND status IN ('active','disabled','releasing')`, numberID)
	if err != nil || tag.RowsAffected() != 1 {
		if err == nil {
			err = pgx.ErrNoRows
		}
		return err
	}
	tag, err = tx.Exec(ctx, `UPDATE provider_operations SET state='succeeded',completed_at=now(),last_error=NULL,next_attempt_at=NULL WHERE id=$1 AND phone_number_id=$2`, operationID, numberID)
	if err != nil || tag.RowsAffected() != 1 {
		if err == nil {
			err = pgx.ErrNoRows
		}
		return err
	}
	return tx.Commit(ctx)
}
