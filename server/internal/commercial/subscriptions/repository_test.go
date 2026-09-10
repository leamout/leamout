package subscriptions

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapWriteErrorMapsNamedUniqueConflicts(t *testing.T) {
	t.Parallel()

	err := mapWriteError(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: "uq_subscriptions_current_organization",
	})
	if !errors.Is(err, ErrCurrentSubscriptionExists) {
		t.Fatalf("mapWriteError() error = %v, want %v", err, ErrCurrentSubscriptionExists)
	}
}

func TestMapWriteErrorPreservesUnknownUniqueViolation(t *testing.T) {
	t.Parallel()

	pgErr := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "some_future_unique_constraint",
	}
	if got := mapWriteError(pgErr); !errors.Is(got, pgErr) {
		t.Fatalf("mapWriteError() error = %v, want original error", got)
	}
}
