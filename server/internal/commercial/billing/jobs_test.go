package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type targetRepositoryStub struct {
	calls []uuid.UUID
	pages [][]ReconciliationTarget
	err   error
}

func (r *targetRepositoryStub) ListReconciliationTargets(_ context.Context, cursor uuid.UUID) ([]ReconciliationTarget, error) {
	r.calls = append(r.calls, cursor)
	if r.err != nil {
		return nil, r.err
	}
	if len(r.pages) == 0 {
		return nil, nil
	}
	page := r.pages[0]
	r.pages = r.pages[1:]
	return page, nil
}

type reconcilerStub struct {
	targets []ReconciliationTarget
	fail    uuid.UUID
}

func (s *reconcilerStub) ReconcileSubscription(_ context.Context, target ReconciliationTarget) error {
	s.targets = append(s.targets, target)
	if target.OrganizationID == s.fail {
		return errors.New("provider unavailable")
	}
	return nil
}

func TestReconciliationPagesAndContinuesAfterProviderFailure(t *testing.T) {
	first := make([]ReconciliationTarget, 100)
	for i := range first {
		first[i] = ReconciliationTarget{OrganizationID: uuid.New(), SubscriptionID: "sub_test", CustomerID: "cus_test"}
	}
	last := ReconciliationTarget{OrganizationID: uuid.New(), SubscriptionID: "sub_last", CustomerID: "cus_last"}
	repo := &targetRepositoryStub{pages: [][]ReconciliationTarget{first, {last}}}
	service := &reconcilerStub{fail: first[0].OrganizationID}
	reports := 0
	job, err := NewReconciliationJob(repo, service, time.Minute, func(context.Context, error) { reports++ })
	if err != nil {
		t.Fatal(err)
	}
	if err := job.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(service.targets) != 101 || reports != 1 {
		t.Fatalf("targets=%d reports=%d", len(service.targets), reports)
	}
	if len(repo.calls) != 2 || repo.calls[0] != uuid.Nil || repo.calls[1] != first[99].OrganizationID {
		t.Fatalf("pagination cursors=%v", repo.calls)
	}
}
func TestReconciliationCancellationAndDatabaseFailure(t *testing.T) {
	repo := &targetRepositoryStub{err: errors.New("database unavailable")}
	service := &reconcilerStub{}
	job, err := NewReconciliationJob(repo, service, time.Hour, func(context.Context, error) {})
	if err != nil {
		t.Fatal(err)
	}
	if err := job.reconcile(context.Background()); err == nil {
		t.Fatal("database error swallowed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := job.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := NewReconciliationJob(repo, service, 0, func(context.Context, error) {}); err == nil {
		t.Fatal("invalid interval accepted")
	}
}
