package deployments

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeActivationStore struct {
	max      int32
	existing *Deployment
	active   int64
	created  Deployment
	steps    []string
}

func (f *fakeActivationStore) lockLicense(context.Context, ActivateRequest) (int32, error) {
	f.steps = append(f.steps, "lock")
	return f.max, nil
}

func (f *fakeActivationStore) get(context.Context, ActivateRequest) (Deployment, error) {
	f.steps = append(f.steps, "get")
	if f.existing == nil {
		return Deployment{}, pgx.ErrNoRows
	}
	return *f.existing, nil
}

func (f *fakeActivationStore) countActive(context.Context, ActivateRequest) (int64, error) {
	f.steps = append(f.steps, "count")
	return f.active, nil
}

func (f *fakeActivationStore) insert(context.Context, ActivateRequest) (Deployment, error) {
	f.steps = append(f.steps, "insert")
	return f.created, nil
}

func TestActivateLocksBeforeCapacityCheckAndInsert(t *testing.T) {
	created := Deployment{ID: uuid.New(), Status: StatusActive}
	store := &fakeActivationStore{max: 2, active: 1, created: created}
	got, err := activate(context.Background(), store, validActivateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != created.ID {
		t.Fatalf("created deployment = %s, want %s", got.ID, created.ID)
	}
	want := []string{"lock", "get", "count", "insert"}
	if len(store.steps) != len(want) {
		t.Fatalf("steps = %v, want %v", store.steps, want)
	}
	for index := range want {
		if store.steps[index] != want[index] {
			t.Fatalf("steps = %v, want %v", store.steps, want)
		}
	}
}

func TestActivateIsIdempotentForActiveDeployment(t *testing.T) {
	existing := Deployment{ID: uuid.New(), Status: StatusActive}
	store := &fakeActivationStore{max: 1, existing: &existing}
	got, err := activate(context.Background(), store, validActivateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != existing.ID || len(store.steps) != 2 {
		t.Fatalf("unexpected idempotent activation: deployment=%+v steps=%v", got, store.steps)
	}
}

func TestActivateRejectsDeactivatedIdentityAndExhaustedCapacity(t *testing.T) {
	deactivated := Deployment{ID: uuid.New(), Status: StatusDeactivated}
	store := &fakeActivationStore{max: 1, existing: &deactivated}
	if _, err := activate(context.Background(), store, validActivateRequest()); !errors.Is(err, ErrDeploymentDeactivated) {
		t.Fatalf("expected deactivated error, got %v", err)
	}

	store = &fakeActivationStore{max: 1, active: 1}
	if _, err := activate(context.Background(), store, validActivateRequest()); !errors.Is(err, ErrDeploymentLimitReached) {
		t.Fatalf("expected capacity error, got %v", err)
	}
	if len(store.steps) != 3 {
		t.Fatalf("capacity rejection attempted insertion: %v", store.steps)
	}
}

func validActivateRequest() ActivateRequest {
	return ActivateRequest{
		OrganizationID: uuid.New(), LicenseID: uuid.New(), DeploymentID: "customer-prod", At: time.Now(),
	}
}
