package tenantdb

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type recordingExecutor struct {
	query string
	args  []any
	err   error
}

func (e *recordingExecutor) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	e.query = query
	e.args = args
	return pgconn.CommandTag{}, e.err
}

func TestBindOrganizationUsesTransactionLocalParameterizedSetting(t *testing.T) {
	organizationID := uuid.New()
	executor := &recordingExecutor{}
	if err := BindOrganization(t.Context(), executor, organizationID); err != nil {
		t.Fatalf("BindOrganization() error = %v", err)
	}
	if executor.query != bindOrganizationSQL {
		t.Fatalf("query = %q", executor.query)
	}
	if len(executor.args) != 1 || executor.args[0] != organizationID.String() {
		t.Fatalf("args = %#v", executor.args)
	}
}

func TestBindOrganizationFailsClosed(t *testing.T) {
	if err := BindOrganization(t.Context(), &recordingExecutor{}, uuid.Nil); err == nil {
		t.Fatal("BindOrganization() accepted a nil organization")
	}
	want := errors.New("database unavailable")
	if err := BindOrganization(t.Context(), &recordingExecutor{err: want}, uuid.New()); !errors.Is(err, want) {
		t.Fatalf("BindOrganization() error = %v, want wrapped %v", err, want)
	}
}
