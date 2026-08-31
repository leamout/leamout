package audit

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository {
	if db == nil {
		panic("audit: database is required")
	}
	return &Repository{db: db}
}

func Insert(ctx context.Context, tx pgx.Tx, event Event) error {
	if tx == nil {
		return fmt.Errorf("audit transaction is required")
	}
	_, err := tx.Exec(ctx, `INSERT INTO audit_events (organization_id, actor_type, actor_id, action, target_type, target_id, metadata) VALUES ($1,$2,$3,$4,$5,$6,$7)`, event.OrganizationID, event.ActorType, event.ActorID, event.Action, event.TargetType, event.TargetID, event.Metadata)
	return err
}

func (r *Repository) List(ctx context.Context, organizationID uuid.UUID, limit, offset int32) ([]Event, error) {
	rows, err := r.db.Query(ctx, `SELECT id, organization_id, actor_type, actor_id, action, target_type, target_id, metadata, occurred_at FROM audit_events WHERE organization_id=$1 ORDER BY occurred_at DESC, id DESC LIMIT $2 OFFSET $3`, organizationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Event, 0)
	for rows.Next() {
		var item Event
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.ActorType, &item.ActorID, &item.Action, &item.TargetType, &item.TargetID, &item.Metadata, &item.OccurredAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
