package number_orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

type numberOrderRepository interface {
	Create(context.Context, uuid.UUID, string) (sqlc.NumberOrder, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (sqlc.NumberOrder, error)
	MarkProviderOperationAccepted(context.Context, sqlc.ProviderOperation, string, []byte) error
	MarkNumberOrderProcessing(context.Context, sqlc.ProviderOperation) error
	RecordProviderOperationFailure(context.Context, uuid.UUID, error) error
	FailProviderOperation(context.Context, sqlc.ProviderOperation, error) error
	CompleteProviderOperation(context.Context, sqlc.ProviderOperation, ProviderOperationRequest, string, []byte) error
}

type ManagedNumberProvider interface {
	FindNumberOrder(context.Context, string) (ProviderOrder, bool, error)
	CreateNumberOrder(context.Context, ProviderOrderRequest) (ProviderOrder, error)
	FindManagedNumber(context.Context, string) (ProviderNumber, error)
	ConfigureNumberRouting(context.Context, string, string) (ProviderNumber, error)
}

type Service struct {
	repo      numberOrderRepository
	providers map[string]ManagedNumberProvider
}

func NewService(repo numberOrderRepository) *Service {
	return &Service{repo: repo, providers: make(map[string]ManagedNumberProvider)}
}

func (s *Service) SetManagedProvider(slug string, provider ManagedNumberProvider) {
	slug = strings.TrimSpace(slug)
	if slug == "" || provider == nil {
		return
	}
	s.providers[slug] = provider
}

func (s *Service) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	req CreateRequest,
) (Response, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return Response{}, err
	}
	selectionID, err := normalizeSelectionID(req.SelectionID)
	if err != nil {
		return Response{}, err
	}

	order, err := s.repo.Create(ctx, organizationID, selectionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSelectionNotFound):
			return Response{}, apperror.NewNotFound(err.Error())
		case errors.Is(err, ErrSelectionUnavailable):
			return Response{}, apperror.NewConflict(err.Error())
		case errors.Is(err, ErrProviderRoutingUnavailable):
			return Response{}, apperror.NewServiceUnavailable(err.Error(), err)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Response{}, apperror.NewConflict("number is already being acquired")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, apperror.NewNotFound("active organization or managed provider not found")
		}
		return Response{}, apperror.NewInternal("create number order", err)
	}
	return response(order), nil
}

func (s *Service) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (Response, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return Response{}, err
	}
	if err := validateOrderID(id); err != nil {
		return Response{}, err
	}

	order, err := s.repo.Get(ctx, organizationID, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Response{}, apperror.NewNotFound("number order not found")
	}
	if err != nil {
		return Response{}, apperror.NewInternal("get number order", err)
	}
	return response(order), nil
}

func (s *Service) ExecuteProviderOperation(ctx context.Context, operation sqlc.ProviderOperation) error {
	if operation.OperationType != "number_order" || operation.NumberOrderID == nil {
		return nil
	}
	if operation.ExecutionTarget != "direct" {
		return nil
	}
	if operation.CarrierProviderID == nil {
		return s.failOperation(ctx, operation, fmt.Errorf("direct provider operation is missing carrier_provider_id"))
	}

	var request ProviderOperationRequest
	if err := json.Unmarshal(operation.Request, &request); err != nil {
		return s.failOperation(ctx, operation, fmt.Errorf("decode provider operation request: %w", err))
	}
	if err := validateProviderOperationRequest(request); err != nil {
		return s.failOperation(ctx, operation, err)
	}

	provider := s.providers[request.Provider]
	if provider == nil {
		err := fmt.Errorf("managed number provider %q is not configured", request.Provider)
		if recordErr := s.repo.RecordProviderOperationFailure(ctx, operation.ID, err); recordErr != nil {
			return fmt.Errorf("record provider configuration failure: %w", recordErr)
		}
		return nil
	}

	externalReferenceID := operation.ID.String()
	providerOrder, found, err := provider.FindNumberOrder(ctx, externalReferenceID)
	if err != nil {
		return s.handleProviderError(ctx, operation, fmt.Errorf("reconcile provider number order: %w", err))
	}
	if !found {
		providerOrder, err = provider.CreateNumberOrder(ctx, ProviderOrderRequest{
			ProviderInventoryID: request.ProviderInventoryID,
			ProviderProductID:   request.ProviderProductID,
			ExternalReferenceID: externalReferenceID,
		})
		if err != nil {
			return s.handleProviderError(ctx, operation, fmt.Errorf("create provider number order: %w", err))
		}
	}

	providerResponse, err := json.Marshal(providerOrder)
	if err != nil {
		return fmt.Errorf("encode provider number order response: %w", err)
	}
	if err := s.repo.MarkProviderOperationAccepted(ctx, operation, providerOrder.ID, providerResponse); err != nil {
		return fmt.Errorf("persist accepted provider number order: %w", err)
	}
	if err := s.repo.MarkNumberOrderProcessing(ctx, operation); err != nil {
		return fmt.Errorf("mark number order processing: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(providerOrder.Status)) {
	case "pending":
		return nil
	case "canceled", "cancelled", "failed":
		return s.failOperation(ctx, operation, fmt.Errorf("provider number order ended with status %q", providerOrder.Status))
	case "completed":
		return s.completeProviderOperation(ctx, operation, request, provider, providerResponse)
	default:
		err := fmt.Errorf("provider number order returned unknown status %q", providerOrder.Status)
		if recordErr := s.repo.RecordProviderOperationFailure(ctx, operation.ID, err); recordErr != nil {
			return fmt.Errorf("record unknown provider order status: %w", recordErr)
		}
		return nil
	}
}

func (s *Service) completeProviderOperation(
	ctx context.Context,
	operation sqlc.ProviderOperation,
	request ProviderOperationRequest,
	provider ManagedNumberProvider,
	providerResponse []byte,
) error {
	providerNumber, err := provider.FindManagedNumber(ctx, request.Number)
	if err != nil {
		return s.handleProviderError(ctx, operation, fmt.Errorf("resolve purchased provider number: %w", err))
	}
	if strings.TrimSpace(providerNumber.ID) == "" {
		return s.handleProviderError(ctx, operation, fmt.Errorf("provider returned purchased number without resource id"))
	}
	if strings.TrimSpace(providerNumber.Number) != request.Number {
		return s.failOperation(ctx, operation, fmt.Errorf("provider returned unexpected purchased number %q", providerNumber.Number))
	}

	if strings.TrimSpace(providerNumber.RoutingResourceID) != request.ProviderRoutingResourceID {
		providerNumber, err = provider.ConfigureNumberRouting(
			ctx,
			providerNumber.ID,
			request.ProviderRoutingResourceID,
		)
		if err != nil {
			return s.handleProviderError(ctx, operation, fmt.Errorf("configure provider number routing: %w", err))
		}
		if strings.TrimSpace(providerNumber.RoutingResourceID) != request.ProviderRoutingResourceID {
			return s.handleProviderError(ctx, operation, fmt.Errorf("provider number routing did not converge to expected resource"))
		}
	}

	if err := s.repo.CompleteProviderOperation(
		ctx,
		operation,
		request,
		providerNumber.ID,
		providerResponse,
	); err != nil {
		return fmt.Errorf("complete managed number provider operation: %w", err)
	}
	return nil
}

func (s *Service) handleProviderError(ctx context.Context, operation sqlc.ProviderOperation, err error) error {
	var classified interface{ Retryable() bool }
	if errors.As(err, &classified) && !classified.Retryable() {
		return s.failOperation(ctx, operation, err)
	}
	if recordErr := s.repo.RecordProviderOperationFailure(ctx, operation.ID, err); recordErr != nil {
		return fmt.Errorf("record provider operation failure: %w", recordErr)
	}
	return nil
}

func (s *Service) failOperation(ctx context.Context, operation sqlc.ProviderOperation, err error) error {
	if failErr := s.repo.FailProviderOperation(ctx, operation, err); failErr != nil {
		return fmt.Errorf("fail provider operation: %w", failErr)
	}
	return nil
}

func validateProviderOperationRequest(request ProviderOperationRequest) error {
	if strings.TrimSpace(request.SelectionID) == "" ||
		strings.TrimSpace(request.Provider) == "" ||
		strings.TrimSpace(request.ProviderInventoryID) == "" ||
		strings.TrimSpace(request.ProviderProductID) == "" ||
		strings.TrimSpace(request.Number) == "" ||
		strings.TrimSpace(request.CountryCode) == "" ||
		request.CarrierConnectionID == uuid.Nil ||
		strings.TrimSpace(request.ProviderRoutingResourceID) == "" {
		return fmt.Errorf("provider operation request is incomplete")
	}
	return nil
}

// ExecutionRouter is service-layer orchestration for durable managed-carrier
// operations. Persistence records the execution target; this router selects the
// corresponding execution service without coupling the background job to a
// specific carrier or transport.
type ExecutionRouter struct {
	direct  ProviderOperationExecutor
	transit ProviderOperationExecutor
}

func NewExecutionRouter(
	direct ProviderOperationExecutor,
	transit ProviderOperationExecutor,
) *ExecutionRouter {
	return &ExecutionRouter{direct: direct, transit: transit}
}

func (r *ExecutionRouter) ExecuteProviderOperation(
	ctx context.Context,
	operation sqlc.ProviderOperation,
) error {
	target := strings.ToLower(strings.TrimSpace(operation.ExecutionTarget))
	switch target {
	case "direct":
		if r == nil || r.direct == nil {
			return fmt.Errorf("direct managed carrier executor is not configured")
		}
		return r.direct.ExecuteProviderOperation(ctx, operation)
	case "transit":
		if r == nil || r.transit == nil {
			return fmt.Errorf("transit managed carrier executor is not configured")
		}
		return r.transit.ExecuteProviderOperation(ctx, operation)
	default:
		return fmt.Errorf("unsupported provider operation execution target %q", operation.ExecutionTarget)
	}
}
