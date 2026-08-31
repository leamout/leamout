package subscriptions

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapWriteErrorMapsNamedUniqueConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		constraint string
		want       error
	}{
		{
			name:       "current subscription",
			constraint: "uq_subscriptions_current_organization",
			want:       ErrCurrentSubscriptionExists,
		},
		{
			name:       "provider reference",
			constraint: "uq_subscriptions_provider_subscription",
			want:       ErrProviderConflict,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := mapWriteError(&pgconn.PgError{
				Code:           "23505",
				ConstraintName: test.constraint,
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("mapWriteError() error = %v, want %v", err, test.want)
			}
		})
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
