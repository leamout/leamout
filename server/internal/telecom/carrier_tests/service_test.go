package carrier_tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/modules/audit"
	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/routing"
)

type fakeStore struct {
	created  int
	finished string
	route    uuid.UUID
}

func (f *fakeStore) Create(_ context.Context, org, connection uuid.UUID, actor audit.Actor, req Request) (Result, error) {
	f.created++
	return Result{ID: uuid.New(), OrganizationID: org, CarrierConnectionID: connection, TrunkID: req.TrunkID, ActorType: actor.Type, ActorID: actor.ID}, nil
}
func (f *fakeStore) AttributeRoute(_ context.Context, _ uuid.UUID, route routing.OutboundDecision) error {
	f.route = route.EndpointID
	return nil
}
func (f *fakeStore) Finish(_ context.Context, id uuid.UUID, status string, callID, code, message *string, answered bool) (Result, error) {
	f.finished = status
	return Result{ID: id, Status: status, SIPCallID: callID, ResponseCode: code}, nil
}
func (f *fakeStore) List(context.Context, uuid.UUID, uuid.UUID, int32, int32) ([]Result, error) {
	return nil, nil
}

type fakeRoutes struct {
	decision routing.OutboundDecision
	err      error
}

func (f fakeRoutes) ResolveOutbound(context.Context, routing.OutboundRequest) (routing.OutboundDecision, error) {
	return f.decision, f.err
}

type fakeCalls struct {
	originated int
	hungup     int
	err        error
}

func (f *fakeCalls) Originate(context.Context, calls.OriginateRequest) (string, error) {
	f.originated++
	return "test-channel", f.err
}
func (f *fakeCalls) Hangup(context.Context, string) error { f.hungup++; return f.err }

type fakeLimiter struct {
	allowed bool
	calls   int
	err     error
}

func (f *fakeLimiter) AllowFixedWindow(context.Context, string, int64, time.Duration) (bool, error) {
	f.calls++
	return f.allowed, f.err
}

func actorContext() context.Context {
	return authn.WithPrincipal(context.Background(), authn.Principal{Subject: authn.Subject{ID: uuid.New(), Type: authn.SubjectUser}, Credential: authn.Credential{Type: authn.CredentialSession}})
}

func TestRunRejectsDestinationOutsideExplicitAllowlist(t *testing.T) {
	store, limiter := &fakeStore{}, &fakeLimiter{allowed: true}
	service, err := NewService(store, fakeRoutes{}, &fakeCalls{}, limiter, []string{"+14155550100"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Run(actorContext(), uuid.New(), uuid.New(), Request{TrunkID: uuid.New(), From: "+14155550101", To: "+14155550102"})
	if err == nil || store.created != 0 || limiter.calls != 0 {
		t.Fatalf("unsafe request was processed: err=%v created=%d limited=%d", err, store.created, limiter.calls)
	}
}

func TestRunAppliesDedicatedRateLimitBeforePersistence(t *testing.T) {
	store := &fakeStore{}
	service, err := NewService(store, fakeRoutes{}, &fakeCalls{}, &fakeLimiter{allowed: false}, []string{"+14155550100"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Run(actorContext(), uuid.New(), uuid.New(), Request{TrunkID: uuid.New(), From: "+14155550101", To: "+14155550100"})
	if err == nil || store.created != 0 {
		t.Fatalf("rate-limited call was persisted: err=%v", err)
	}
}

func TestRunPersistsRouteAndCompletedResult(t *testing.T) {
	organizationID, connectionID, trunkID, endpointID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	store, controller := &fakeStore{}, &fakeCalls{}
	service, err := NewService(store, fakeRoutes{decision: routing.OutboundDecision{CarrierConnectionID: connectionID, TrunkID: trunkID, EndpointID: endpointID, Host: "carrier.test", Port: 5060, Transport: "udp", From: "+14155550101", To: "+14155550100"}}, controller, &fakeLimiter{allowed: true}, []string{"+14155550100"})
	if err != nil {
		t.Fatal(err)
	}
	service.maximumDuration = time.Millisecond
	result, err := service.Run(actorContext(), organizationID, connectionID, Request{TrunkID: trunkID, From: "+14155550101", To: "+14155550100"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || store.route != endpointID || controller.originated != 1 || controller.hungup != 1 {
		t.Fatalf("result=%+v route=%s controller=%+v", result, store.route, controller)
	}
}

func TestRunPersistsFailedOrigination(t *testing.T) {
	connectionID, trunkID := uuid.New(), uuid.New()
	store := &fakeStore{}
	service, err := NewService(store, fakeRoutes{decision: routing.OutboundDecision{CarrierConnectionID: connectionID, TrunkID: trunkID, EndpointID: uuid.New(), Host: "carrier.test", Port: 5060, Transport: "udp", From: "+14155550101", To: "+14155550100"}}, &fakeCalls{err: errors.New("carrier rejected")}, &fakeLimiter{allowed: true}, []string{"+14155550100"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Run(actorContext(), uuid.New(), connectionID, Request{TrunkID: trunkID, From: "+14155550101", To: "+14155550100"})
	if err != nil || result.Status != "failed" || store.finished != "failed" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
