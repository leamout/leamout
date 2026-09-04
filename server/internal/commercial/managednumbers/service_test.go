package managednumbers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/integrations/carriers/didww"
)

type providerStub struct {
	ordered    didww.OrderNumberRequest
	did        didww.DID
	released   string
	releaseErr error
}

func (p *providerStub) SearchNumbers(context.Context, didww.SearchNumbersRequest) ([]didww.AvailableNumber, error) {
	return []didww.AvailableNumber{{ID: "available", Number: "12124727600"}}, nil
}
func (p *providerStub) OrderNumber(_ context.Context, r didww.OrderNumberRequest) (didww.Order, error) {
	p.ordered = r
	return didww.Order{ID: "order-1"}, nil
}
func (p *providerStub) FindNumber(context.Context, string) (didww.DID, error) { return p.did, nil }
func (p *providerStub) ConfigureRouting(_ context.Context, id, trunk string) (didww.DID, error) {
	p.did.ID = id
	p.did.VoiceInTrunkID = trunk
	return p.did, nil
}
func (p *providerStub) ReleaseNumber(_ context.Context, id string) error {
	p.released = id
	return p.releaseErr
}

type storeStub struct {
	op        Operation
	accepted  string
	completed bool
}

func (s *storeStub) BeginOrder(context.Context, OrderRequest) (Operation, error) { return s.op, nil }
func (s *storeStub) ProviderAccepted(_ context.Context, _ uuid.UUID, id string, _ any) error {
	s.accepted = id
	return nil
}
func (s *storeStub) CompleteOrder(context.Context, uuid.UUID, didww.DID, OrderRequest) (uuid.UUID, error) {
	s.completed = true
	return uuid.New(), nil
}
func (s *storeStub) BeginRelease(context.Context, ReleaseRequest) (Operation, error) {
	return s.op, nil
}
func (s *storeStub) CompleteRelease(context.Context, uuid.UUID, uuid.UUID) error {
	s.completed = true
	return nil
}
func (s *storeStub) Fail(context.Context, uuid.UUID, error) error { return nil }

func TestOrderPersistsIntentAndUsesOperationAsProviderReference(t *testing.T) {
	opID := uuid.New()
	provider := &providerStub{}
	store := &storeStub{op: Operation{ID: opID, State: "pending"}}
	service, _ := NewService(provider, store)
	request := validOrder()
	op, err := service.Order(context.Background(), request)
	if !errors.Is(err, ErrOperationIncomplete) {
		t.Fatalf("got %v", err)
	}
	if op.State != "provider_accepted" || provider.ordered.ExternalReferenceID != opID.String() || store.accepted != "order-1" {
		t.Fatalf("op=%+v provider=%+v store=%+v", op, provider, store)
	}
}
func TestReconcileConfiguresRoutingBeforeCreatingNumber(t *testing.T) {
	provider := &providerStub{did: didww.DID{ID: "did-1", Number: "12124727600"}}
	store := &storeStub{}
	service, _ := NewService(provider, store)
	if _, err := service.ReconcileOrder(context.Background(), uuid.New(), validOrder()); err != nil {
		t.Fatal(err)
	}
	if !store.completed || provider.did.VoiceInTrunkID != "voice-in-1" {
		t.Fatal("order completed before routing reconciliation")
	}
}
func TestReleaseCompletesProviderBeforeLocalRelease(t *testing.T) {
	provider := &providerStub{}
	store := &storeStub{op: Operation{ID: uuid.New(), State: "pending"}}
	service, _ := NewService(provider, store)
	request := ReleaseRequest{OrganizationID: uuid.New(), ProviderID: uuid.New(), PhoneNumberID: uuid.New(), IdempotencyKey: "release-1", ProviderResourceID: "did-1"}
	if err := service.Release(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if provider.released != "did-1" || !store.completed {
		t.Fatal("release was not reconciled")
	}
}
func TestReleaseTreatsAlreadyMissingProviderDIDAsIdempotent(t *testing.T) {
	provider := &providerStub{releaseErr: &didww.APIError{StatusCode: 404}}
	store := &storeStub{op: Operation{ID: uuid.New(), State: "failed"}}
	service, _ := NewService(provider, store)
	request := ReleaseRequest{OrganizationID: uuid.New(), ProviderID: uuid.New(), PhoneNumberID: uuid.New(), IdempotencyKey: "release-1", ProviderResourceID: "did-1"}
	if err := service.Release(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if !store.completed {
		t.Fatal("already released DID did not complete local reconciliation")
	}
}
func validOrder() OrderRequest {
	return OrderRequest{OrganizationID: uuid.New(), ProviderID: uuid.New(), IngressConnectionID: uuid.New(), IdempotencyKey: "order-1", AvailableDIDID: "available-1", SKUID: "sku-1", Number: "+12124727600", CountryCode: "US", VoiceInTrunkID: "voice-in-1"}
}
