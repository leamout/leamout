package wholesale

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeStore struct {
	result Result
	err    error
	cdr    CDR
}

func (f *fakeStore) Reconcile(_ context.Context, cdr CDR) (Result, error) {
	f.cdr = cdr
	return f.result, f.err
}

func validCDR() CDR {
	return CDR{
		Provider: " COMMPEAK ", ProviderRecordID: " cdr-1 ", Direction: " TERMINATION ", SIPCallID: " call-1 ",
		StartedAt: time.Now().UTC(), DurationSeconds: 12, Currency: " usd ", CostMicros: 2500,
		Raw: map[string]any{"id": "cdr-1"},
	}
}

func TestServiceNormalizesAndReconcilesTerminationCDR(t *testing.T) {
	store := &fakeStore{result: Result{CallID: uuid.New()}}
	result, err := NewService(store).Reconcile(t.Context(), validCDR())
	if err != nil {
		t.Fatal(err)
	}
	if result.CallID != store.result.CallID {
		t.Fatalf("result = %+v", result)
	}
	if store.cdr.Provider != "commpeak" || store.cdr.ProviderRecordID != "cdr-1" || store.cdr.Direction != "termination" ||
		store.cdr.SIPCallID != "call-1" || store.cdr.Currency != "USD" {
		t.Fatalf("CDR was not normalized: %+v", store.cdr)
	}
}

func TestServiceRejectsOriginationCDR(t *testing.T) {
	cdr := validCDR()
	cdr.Direction = "origination"
	_, err := NewService(&fakeStore{}).Reconcile(t.Context(), cdr)
	if err == nil {
		t.Fatal("origination CDR was accepted as managed termination cost")
	}
}

func TestServicePropagatesReconciliationFailure(t *testing.T) {
	want := errors.New("database unavailable")
	_, err := NewService(&fakeStore{err: want}).Reconcile(t.Context(), validCDR())
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
