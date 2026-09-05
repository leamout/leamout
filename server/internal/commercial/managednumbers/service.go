package managednumbers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/integrations/carriers/didww"
)

var ErrOperationIncomplete = errors.New("managed number operation is awaiting provider state")

type Provider interface {
	SearchNumbers(context.Context, didww.SearchNumbersRequest) ([]didww.AvailableNumber, error)
	OrderNumber(context.Context, didww.OrderNumberRequest) (didww.Order, error)
	FindNumber(context.Context, string) (didww.DID, error)
	ConfigureRouting(context.Context, string, string) (didww.DID, error)
	ReleaseNumber(context.Context, string) error
}

type Store interface {
	BeginOrder(context.Context, OrderRequest) (Operation, error)
	ProviderAccepted(context.Context, uuid.UUID, string, any) error
	CompleteOrder(context.Context, uuid.UUID, didww.DID, OrderRequest) (uuid.UUID, error)
	BeginRelease(context.Context, ReleaseRequest) (Operation, error)
	CompleteRelease(context.Context, uuid.UUID, uuid.UUID) error
	Fail(context.Context, uuid.UUID, error) error
}

type Operation struct {
	ID            uuid.UUID
	State         string
	PhoneNumberID *uuid.UUID
}
type OrderRequest struct {
	OrganizationID, ProviderID, IngressConnectionID                            uuid.UUID
	IdempotencyKey, AvailableDIDID, SKUID, Number, CountryCode, VoiceInTrunkID string
}
type ReleaseRequest struct {
	OrganizationID, ProviderID, PhoneNumberID uuid.UUID
	IdempotencyKey, ProviderResourceID        string
}

type Drift struct{ ProviderMissing, ProviderInactive, RoutingRepaired bool }

type Service struct {
	provider Provider
	store    Store
}

func NewService(provider Provider, store Store) (*Service, error) {
	if provider == nil || store == nil {
		return nil, fmt.Errorf("managed number provider and store are required")
	}
	return &Service{provider: provider, store: store}, nil
}

func (s *Service) Search(ctx context.Context, request didww.SearchNumbersRequest) ([]didww.AvailableNumber, error) {
	request.Feature = "voice_in"
	return s.provider.SearchNumbers(ctx, request)
}

// Order durably records intent before contacting DIDWW. The operation ID is
// sent as DIDWW's external reference so a timed-out response can be reconciled.
func (s *Service) Order(ctx context.Context, request OrderRequest) (Operation, error) {
	if err := validateOrder(request); err != nil {
		return Operation{}, err
	}
	operation, err := s.store.BeginOrder(ctx, request)
	if err != nil {
		return Operation{}, err
	}
	if operation.State == "succeeded" {
		return operation, nil
	}
	if operation.State == "provider_accepted" {
		return operation, ErrOperationIncomplete
	}
	order, err := s.provider.OrderNumber(ctx, didww.OrderNumberRequest{AvailableDIDID: request.AvailableDIDID, SKUID: request.SKUID, ExternalReferenceID: operation.ID.String()})
	if err != nil {
		_ = s.store.Fail(context.WithoutCancel(ctx), operation.ID, err)
		return operation, err
	}
	if err = s.store.ProviderAccepted(ctx, operation.ID, order.ID, order); err != nil {
		return operation, err
	}
	operation.State = "provider_accepted"
	return operation, ErrOperationIncomplete
}

// ReconcileOrder resolves the ordered DID, configures provider ingress, and
// atomically creates the active tenant-owned managed number.
func (s *Service) ReconcileOrder(ctx context.Context, operationID uuid.UUID, request OrderRequest) (uuid.UUID, error) {
	did, err := s.provider.FindNumber(ctx, request.Number)
	if err != nil {
		return uuid.Nil, err
	}
	if did.Terminated || did.PendingRemoval {
		return uuid.Nil, fmt.Errorf("DIDWW number is not active")
	}
	if strings.TrimPrefix(strings.TrimSpace(did.Number), "+") != strings.TrimPrefix(strings.TrimSpace(request.Number), "+") {
		return uuid.Nil, fmt.Errorf("DIDWW order resolved to an unexpected number")
	}
	did, err = s.provider.ConfigureRouting(ctx, did.ID, request.VoiceInTrunkID)
	if err != nil {
		return uuid.Nil, err
	}
	if did.VoiceInTrunkID != request.VoiceInTrunkID {
		return uuid.Nil, fmt.Errorf("DIDWW routing reconciliation failed")
	}
	return s.store.CompleteOrder(ctx, operationID, did, request)
}

func (s *Service) Release(ctx context.Context, request ReleaseRequest) error {
	if request.OrganizationID == uuid.Nil || request.ProviderID == uuid.Nil || request.PhoneNumberID == uuid.Nil || strings.TrimSpace(request.ProviderResourceID) == "" || strings.TrimSpace(request.IdempotencyKey) == "" {
		return fmt.Errorf("managed release identity is required")
	}
	op, err := s.store.BeginRelease(ctx, request)
	if err != nil {
		return err
	}
	if op.State == "succeeded" {
		return nil
	}
	if err = s.provider.ReleaseNumber(ctx, request.ProviderResourceID); err != nil && !isProviderNotFound(err) {
		_ = s.store.Fail(context.WithoutCancel(ctx), op.ID, err)
		return err
	}
	return s.store.CompleteRelease(ctx, op.ID, request.PhoneNumberID)
}

func isProviderNotFound(err error) bool {
	var apiError *didww.APIError
	return errors.As(err, &apiError) && apiError.StatusCode == 404
}

func (s *Service) ReconcileActive(ctx context.Context, number, expectedVoiceInTrunkID string) (Drift, error) {
	if strings.TrimSpace(number) == "" || strings.TrimSpace(expectedVoiceInTrunkID) == "" {
		return Drift{}, fmt.Errorf("number and expected Voice IN trunk are required")
	}
	did, err := s.provider.FindNumber(ctx, number)
	if err != nil {
		return Drift{ProviderMissing: true}, err
	}
	drift := Drift{ProviderInactive: did.Terminated || did.PendingRemoval || did.Blocked}
	if did.VoiceInTrunkID != expectedVoiceInTrunkID && !drift.ProviderInactive {
		if _, err = s.provider.ConfigureRouting(ctx, did.ID, expectedVoiceInTrunkID); err != nil {
			return drift, err
		}
		drift.RoutingRepaired = true
	}
	return drift, nil
}

func validateOrder(r OrderRequest) error {
	if r.OrganizationID == uuid.Nil || r.ProviderID == uuid.Nil || r.IngressConnectionID == uuid.Nil {
		return fmt.Errorf("organization, provider, and ingress connection are required")
	}
	for name, value := range map[string]string{"idempotency key": r.IdempotencyKey, "available DID": r.AvailableDIDID, "SKU": r.SKUID, "number": r.Number, "country": r.CountryCode, "Voice IN trunk": r.VoiceInTrunkID} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}
