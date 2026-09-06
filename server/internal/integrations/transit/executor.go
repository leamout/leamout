package transit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/database/sqlc"
)

type numberOrderExecutionRepository interface {
	MarkProviderOperationAccepted(context.Context, sqlc.ProviderOperation, string, []byte) error
	MarkNumberOrderProcessing(context.Context, sqlc.ProviderOperation) error
	RecordProviderOperationFailure(context.Context, uuid.UUID, error) error
	FailProviderOperation(context.Context, sqlc.ProviderOperation, error) error
	CompleteTransitProviderOperation(
		context.Context,
		sqlc.ProviderOperation,
		string,
		string,
		string,
		string,
		[]byte,
	) error
}

type NumberOrderExecutor struct {
	client *Client
	repo   numberOrderExecutionRepository
}

func NewNumberOrderExecutor(
	client *Client,
	repo numberOrderExecutionRepository,
) (*NumberOrderExecutor, error) {
	if client == nil {
		return nil, fmt.Errorf("transit client is required")
	}
	if repo == nil {
		return nil, fmt.Errorf("number order repository is required")
	}
	return &NumberOrderExecutor{client: client, repo: repo}, nil
}

type numberOrderOperationRequest struct {
	SelectionID string `json:"selection_id"`
	Number      string `json:"number"`
	CountryCode string `json:"country_code"`
}

func (e *NumberOrderExecutor) ExecuteProviderOperation(
	ctx context.Context,
	operation sqlc.ProviderOperation,
) error {
	if operation.OperationType != "number_order" || operation.ExecutionTarget != "transit" {
		return nil
	}
	if operation.NumberOrderID == nil {
		return e.failOperation(ctx, operation, fmt.Errorf("transit number order operation is missing number_order_id"))
	}

	var request numberOrderOperationRequest
	if err := json.Unmarshal(operation.Request, &request); err != nil {
		return e.failOperation(ctx, operation, fmt.Errorf("decode transit number order request: %w", err))
	}
	request.SelectionID = strings.TrimSpace(request.SelectionID)
	request.Number = strings.TrimSpace(request.Number)
	request.CountryCode = strings.ToUpper(strings.TrimSpace(request.CountryCode))
	if request.SelectionID == "" || request.Number == "" || request.CountryCode == "" {
		return e.failOperation(ctx, operation, fmt.Errorf("transit number order request is incomplete"))
	}

	result, err := e.client.ExecuteNumberOrder(ctx, ExecuteNumberOrderRequest{
		OperationID: operation.ID,
		SelectionID: request.SelectionID,
	})
	if err != nil {
		return e.handleTransitError(ctx, operation, fmt.Errorf("execute transit number order: %w", err))
	}

	response, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("encode transit number order response: %w", err)
	}

	state := strings.ToLower(strings.TrimSpace(result.State))
	switch state {
	case "pending", "accepted", "processing", "provider_accepted":
		return e.acceptOperation(ctx, operation, response)
	case "completed":
		if strings.TrimSpace(result.ManagedResourceID) == "" {
			return e.failOperation(ctx, operation, fmt.Errorf("transit completed number order without managed_resource_id"))
		}
		if strings.TrimSpace(result.Number) != request.Number {
			return e.failOperation(ctx, operation, fmt.Errorf("transit returned unexpected purchased number %q", result.Number))
		}
		if strings.ToUpper(strings.TrimSpace(result.CountryCode)) != request.CountryCode {
			return e.failOperation(ctx, operation, fmt.Errorf("transit returned unexpected country code %q", result.CountryCode))
		}
		if err := e.acceptOperation(ctx, operation, response); err != nil {
			return err
		}
		if err := e.repo.CompleteTransitProviderOperation(
			ctx,
			operation,
			request.SelectionID,
			request.Number,
			request.CountryCode,
			strings.TrimSpace(result.ManagedResourceID),
			response,
		); err != nil {
			return fmt.Errorf("complete transit managed number operation: %w", err)
		}
		return nil
	case "failed", "canceled", "cancelled":
		message := strings.TrimSpace(result.ErrorMessage)
		if message == "" {
			message = "transit number order failed"
		}
		if code := strings.TrimSpace(result.ErrorCode); code != "" {
			message = code + ": " + message
		}
		return e.failOperation(ctx, operation, errors.New(message))
	default:
		err := fmt.Errorf("transit number order returned unknown state %q", result.State)
		if recordErr := e.repo.RecordProviderOperationFailure(ctx, operation.ID, err); recordErr != nil {
			return fmt.Errorf("record transit order state failure: %w", recordErr)
		}
		return nil
	}
}

func (e *NumberOrderExecutor) acceptOperation(
	ctx context.Context,
	operation sqlc.ProviderOperation,
	response []byte,
) error {
	if err := e.repo.MarkProviderOperationAccepted(ctx, operation, operation.ID.String(), response); err != nil {
		return fmt.Errorf("persist accepted transit number order: %w", err)
	}
	if err := e.repo.MarkNumberOrderProcessing(ctx, operation); err != nil {
		return fmt.Errorf("mark transit number order processing: %w", err)
	}
	return nil
}

func (e *NumberOrderExecutor) handleTransitError(
	ctx context.Context,
	operation sqlc.ProviderOperation,
	err error,
) error {
	var classified interface{ Retryable() bool }
	if errors.As(err, &classified) && !classified.Retryable() {
		return e.failOperation(ctx, operation, err)
	}
	if recordErr := e.repo.RecordProviderOperationFailure(ctx, operation.ID, err); recordErr != nil {
		return fmt.Errorf("record transit provider operation failure: %w", recordErr)
	}
	return nil
}

func (e *NumberOrderExecutor) failOperation(
	ctx context.Context,
	operation sqlc.ProviderOperation,
	err error,
) error {
	if failErr := e.repo.FailProviderOperation(ctx, operation, err); failErr != nil {
		return fmt.Errorf("fail transit provider operation: %w", failErr)
	}
	return nil
}
