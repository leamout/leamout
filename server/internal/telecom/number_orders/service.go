package number_orders

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

var (
	e164Pattern    = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
	countryPattern = regexp.MustCompile(`^[A-Z]{2}$`)
)

type numberOrderRepository interface {
	Create(context.Context, CreateInput) (sqlc.NumberOrder, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (sqlc.NumberOrder, error)
	List(context.Context, uuid.UUID) ([]sqlc.NumberOrder, error)
	MarkPurchasing(context.Context, uuid.UUID) (sqlc.NumberOrder, error)
	MarkPurchased(context.Context, uuid.UUID, string) (sqlc.NumberOrder, error)
	MarkPersisting(context.Context, uuid.UUID, *string) (sqlc.NumberOrder, error)
	MarkConfiguring(context.Context, uuid.UUID, *uuid.UUID) (sqlc.NumberOrder, error)
	MarkCompleted(context.Context, uuid.UUID) (sqlc.NumberOrder, error)
	MarkFailed(context.Context, uuid.UUID, Status, Failure) (sqlc.NumberOrder, error)
	ListRecoverable(context.Context, int32) ([]sqlc.NumberOrder, error)
}

type Service struct {
	repo numberOrderRepository
}

func NewService(repo numberOrderRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (sqlc.NumberOrder, error) {
	if input.OrganizationID == uuid.Nil {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("organization_id is required")
	}
	if input.ProviderID == uuid.Nil {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("provider_id is required")
	}

	input.ProviderInventoryID = strings.TrimSpace(input.ProviderInventoryID)
	if input.ProviderInventoryID == "" {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("provider inventory id is required")
	}
	input.ProviderProductID = strings.TrimSpace(input.ProviderProductID)
	if input.ProviderProductID == "" {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("provider product id is required")
	}

	input.Number = strings.TrimSpace(input.Number)
	if !e164Pattern.MatchString(input.Number) {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("number must be in E.164 format")
	}
	input.CountryCode = strings.ToUpper(strings.TrimSpace(input.CountryCode))
	if !countryPattern.MatchString(input.CountryCode) {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("country_code must be a two-letter ISO country code")
	}

	order, err := s.repo.Create(ctx, input)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.NumberOrder{}, apperror.NewNotFound("active organization or provider not found")
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.NumberOrder{}, apperror.NewConflict("number selection already has an order")
		}
		return sqlc.NumberOrder{}, apperror.NewInternal("create number order", err)
	}
	return order, nil
}

func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (sqlc.NumberOrder, error) {
	if organizationID == uuid.Nil {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("organization_id is required")
	}
	if id == uuid.Nil {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("number_order_id is required")
	}
	order, err := s.repo.Get(ctx, organizationID, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.NumberOrder{}, apperror.NewNotFound("number order not found")
	}
	if err != nil {
		return sqlc.NumberOrder{}, apperror.NewInternal("get number order", err)
	}
	return order, nil
}

func (s *Service) List(ctx context.Context, organizationID uuid.UUID) ([]sqlc.NumberOrder, error) {
	if organizationID == uuid.Nil {
		return nil, apperror.NewBadRequest("organization_id is required")
	}
	orders, err := s.repo.List(ctx, organizationID)
	if err != nil {
		return nil, apperror.NewInternal("list number orders", err)
	}
	return orders, nil
}

func (s *Service) BeginPurchase(ctx context.Context, id uuid.UUID) (sqlc.NumberOrder, error) {
	return transition(s.repo.MarkPurchasing(ctx, id), "begin number purchase")
}

func (s *Service) MarkPurchased(ctx context.Context, id uuid.UUID, providerOrderID string) (sqlc.NumberOrder, error) {
	providerOrderID = strings.TrimSpace(providerOrderID)
	if providerOrderID == "" {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("provider order id is required")
	}
	return transition(s.repo.MarkPurchased(ctx, id, providerOrderID), "mark number order purchased")
}

func (s *Service) BeginPersisting(ctx context.Context, id uuid.UUID, providerResourceID *string) (sqlc.NumberOrder, error) {
	if providerResourceID != nil {
		value := strings.TrimSpace(*providerResourceID)
		if value == "" {
			return sqlc.NumberOrder{}, apperror.NewBadRequest("provider resource id must not be blank")
		}
		providerResourceID = &value
	}
	return transition(s.repo.MarkPersisting(ctx, id, providerResourceID), "begin number persistence")
}

func (s *Service) BeginConfiguring(ctx context.Context, id uuid.UUID, phoneNumberID *uuid.UUID) (sqlc.NumberOrder, error) {
	if phoneNumberID != nil && *phoneNumberID == uuid.Nil {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("phone number id must be valid")
	}
	return transition(s.repo.MarkConfiguring(ctx, id, phoneNumberID), "begin number routing configuration")
}

func (s *Service) Complete(ctx context.Context, id uuid.UUID) (sqlc.NumberOrder, error) {
	return transition(s.repo.MarkCompleted(ctx, id), "complete number order")
}

func (s *Service) Fail(ctx context.Context, id uuid.UUID, expected Status, failure Failure) (sqlc.NumberOrder, error) {
	if id == uuid.Nil {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("number_order_id is required")
	}
	if failure.Stage != StagePurchasing && failure.Stage != StagePersisting && failure.Stage != StageConfiguring {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("invalid failed stage")
	}
	if expected != Status(failure.Stage) {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("failed stage must match current processing status")
	}
	failure.Code = strings.TrimSpace(failure.Code)
	failure.Message = strings.TrimSpace(failure.Message)
	if failure.Message == "" {
		return sqlc.NumberOrder{}, apperror.NewBadRequest("failure message is required")
	}
	return transition(s.repo.MarkFailed(ctx, id, expected, failure), "fail number order")
}

func (s *Service) ListRecoverable(ctx context.Context, limit int32) ([]sqlc.NumberOrder, error) {
	if limit <= 0 || limit > 500 {
		return nil, apperror.NewBadRequest("recoverable order limit must be between 1 and 500")
	}
	orders, err := s.repo.ListRecoverable(ctx, limit)
	if err != nil {
		return nil, apperror.NewInternal("list recoverable number orders", err)
	}
	return orders, nil
}

func transition(order sqlc.NumberOrder, err error, action string) (sqlc.NumberOrder, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.NumberOrder{}, apperror.NewConflict("number order is not in a valid state to " + action)
	}
	if err != nil {
		return sqlc.NumberOrder{}, apperror.NewInternal(action, err)
	}
	return order, nil
}
