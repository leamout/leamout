package number_orders

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type executionRouterStub struct {
	operations []sqlc.ProviderOperation
}

func (s *executionRouterStub) ExecuteProviderOperation(
	_ context.Context,
	operation sqlc.ProviderOperation,
) error {
	s.operations = append(s.operations, operation)
	return nil
}

func TestExecutionRouterDispatchesByTarget(t *testing.T) {
	direct := &executionRouterStub{}
	transit := &executionRouterStub{}
	router := NewExecutionRouter(direct, transit)

	directOperation := sqlc.ProviderOperation{ID: uuid.New(), ExecutionTarget: "direct"}
	if err := router.ExecuteProviderOperation(context.Background(), directOperation); err != nil {
		t.Fatal(err)
	}
	transitOperation := sqlc.ProviderOperation{ID: uuid.New(), ExecutionTarget: "transit"}
	if err := router.ExecuteProviderOperation(context.Background(), transitOperation); err != nil {
		t.Fatal(err)
	}

	if len(direct.operations) != 1 || direct.operations[0].ID != directOperation.ID {
		t.Fatalf("direct operations = %#v", direct.operations)
	}
	if len(transit.operations) != 1 || transit.operations[0].ID != transitOperation.ID {
		t.Fatalf("transit operations = %#v", transit.operations)
	}
}

func TestExecutionRouterRejectsUnavailableAndUnknownTargets(t *testing.T) {
	direct := &executionRouterStub{}
	router := NewExecutionRouter(direct, nil)

	err := router.ExecuteProviderOperation(
		context.Background(),
		sqlc.ProviderOperation{ExecutionTarget: "transit"},
	)
	if err == nil || !strings.Contains(err.Error(), "transit managed carrier executor is not configured") {
		t.Fatalf("error = %v", err)
	}

	err = router.ExecuteProviderOperation(
		context.Background(),
		sqlc.ProviderOperation{ExecutionTarget: "other"},
	)
	if err == nil || !strings.Contains(err.Error(), "unsupported provider operation execution target") {
		t.Fatalf("error = %v", err)
	}
}
