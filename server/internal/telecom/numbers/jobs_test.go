package numbers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type providerOperationJobRepositoryStub struct {
	operations []sqlc.ProviderOperation
	locked     bool
}

func (r *providerOperationJobRepositoryStub) ListProviderOperationsReady(context.Context, int32) ([]sqlc.ProviderOperation, error) {
	return r.operations, nil
}
func (r *providerOperationJobRepositoryStub) TryProviderOperationLock(context.Context, uuid.UUID) (func(), bool, error) {
	if !r.locked {
		return func() {}, false, nil
	}
	return func() {}, true, nil
}

type providerOperationExecutorStub struct{ operations []sqlc.ProviderOperation }

func (e *providerOperationExecutorStub) ExecuteProviderOperation(_ context.Context, operation sqlc.ProviderOperation) error {
	e.operations = append(e.operations, operation)
	return nil
}

func TestProviderOperationJobUsesExecutorBoundary(t *testing.T) {
	operation := sqlc.ProviderOperation{ID: uuid.New(), OperationType: "number_provision"}
	repo := &providerOperationJobRepositoryStub{operations: []sqlc.ProviderOperation{operation}, locked: true}
	executor := &providerOperationExecutorStub{}
	job, err := NewProviderOperationJob(repo, executor, DefaultProviderOperationJobConfig())
	if err != nil {
		t.Fatal(err)
	}
	if err := job.Execute(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(executor.operations) != 1 || executor.operations[0].ID != operation.ID {
		t.Fatalf("executor operations = %+v; want operation %s", executor.operations, operation.ID)
	}
}

func TestProviderOperationJobSkipsUnlockedOperations(t *testing.T) {
	operation := sqlc.ProviderOperation{ID: uuid.New(), OperationType: "number_provision"}
	repo := &providerOperationJobRepositoryStub{operations: []sqlc.ProviderOperation{operation}}
	executor := &providerOperationExecutorStub{}
	job, err := NewProviderOperationJob(repo, executor, DefaultProviderOperationJobConfig())
	if err != nil {
		t.Fatal(err)
	}
	if err := job.Execute(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(executor.operations) != 0 {
		t.Fatalf("executor called for unlocked operation: %+v", executor.operations)
	}
}
