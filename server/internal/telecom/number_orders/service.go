package number_orders

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

type numberOrderRepository interface {
	Create(context.Context, uuid.UUID, string) (sqlc.NumberOrder, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (sqlc.NumberOrder, error)
}

type Service struct {
	repo numberOrderRepository
}

func NewService(repo numberOrderRepository) *Service {
	return &Service{repo: repo}
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
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Response{}, apperror.NewConflict("number is already being acquired")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, apperror.NewNotFound("active organization or provider not found")
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
