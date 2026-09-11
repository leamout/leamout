// Package tenantdb establishes transaction-local PostgreSQL tenant identity.
// Cloud repositories use it before executing tenant-owned queries so row-level
// security policies can rely on a value that cannot leak through the pool.
package tenantdb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const bindOrganizationSQL = `SELECT set_config('leamout.organization_id', $1, true)`

type executor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

type beginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

// BindOrganization sets the tenant identity for the current transaction. The
// third set_config argument is true, making the setting transaction-local and
// preventing tenant identity from surviving a commit or rollback on a pooled
// connection.
func BindOrganization(ctx context.Context, tx executor, organizationID uuid.UUID) error {
	if tx == nil {
		return fmt.Errorf("tenant database executor is required")
	}
	if organizationID == uuid.Nil {
		return fmt.Errorf("tenant organization ID is required")
	}
	if _, err := tx.Exec(ctx, bindOrganizationSQL, organizationID.String()); err != nil {
		return fmt.Errorf("bind tenant organization: %w", err)
	}
	return nil
}

// WithinOrganization executes fn in a transaction bound to organizationID.
// It never runs fn if tenant binding fails and always rolls back uncommitted
// work as a safety net.
func WithinOrganization(ctx context.Context, db beginner, organizationID uuid.UUID, fn func(pgx.Tx) error) error {
	if db == nil {
		return fmt.Errorf("tenant database is required")
	}
	if fn == nil {
		return fmt.Errorf("tenant transaction function is required")
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tenant transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := BindOrganization(ctx, tx, organizationID); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tenant transaction: %w", err)
	}
	return nil
}
