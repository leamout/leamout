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
			return sqlc.NumberOrder{}, apperror.NewConflict("number selection already has an open order")
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
