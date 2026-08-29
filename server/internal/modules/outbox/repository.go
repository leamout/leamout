package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return NewRepository(r.queries.WithTx(tx))
}

func (r *Repository) Insert(ctx context.Context, event Event) (sqlc.OutboxEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return sqlc.OutboxEvent{}, fmt.Errorf("marshal outbox payload: %w", err)
	}

	headers, err := json.Marshal(event.Headers)
	if err != nil {
		return sqlc.OutboxEvent{}, fmt.Errorf("marshal outbox headers: %w", err)
	}

	return r.queries.InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
		Subject:       event.Subject,
		AggregateType: event.AggregateType,
		AggregateID:   event.AggregateID,
		Payload:       payload,
		Headers:       headers,
		AvailableAt:   event.AvailableAt,
	})
}
