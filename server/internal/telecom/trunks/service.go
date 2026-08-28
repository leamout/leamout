package trunks

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, req CreateRequest) (sqlc.Trunk, error) {
	if err := validateID(organizationID, "organization_id"); err != nil {
		return sqlc.Trunk{}, err
	}
	if err := validateID(req.CarrierConnectionID, "carrier_connection_id"); err != nil {
		return sqlc.Trunk{}, err
	}
	name, err := normalizeName(req.Name)
	if err != nil {
		return sqlc.Trunk{}, err
	}
	if req.Direction != nil {
		value, e := normalizeChoice(*req.Direction, directions, "direction")
		if e != nil {
			return sqlc.Trunk{}, e
		}
		req.Direction = &value
	}
	if req.Status != nil {
		value, e := normalizeChoice(*req.Status, statuses, "status")
		if e != nil {
			return sqlc.Trunk{}, e
		}
		req.Status = &value
	}
	result, err := s.repo.Create(ctx, sqlc.CreateTrunkParams{OrganizationID: organizationID, CarrierConnectionID: req.CarrierConnectionID, Name: name, Direction: req.Direction, Status: req.Status})
	return result, writeError(err, "trunk", "carrier connection not found")
}
func (s *Service) List(ctx context.Context, organizationID uuid.UUID) ([]sqlc.Trunk, error) {
	if err := validateID(organizationID, "organization_id"); err != nil {
		return nil, err
	}
	items, err := s.repo.List(ctx, organizationID)
	if err != nil {
		return nil, apperror.NewInternal("list trunks", err)
	}
	return items, nil
}
func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Trunk, error) {
	if err := validateID(organizationID, "organization_id"); err != nil {
		return sqlc.Trunk{}, err
	}
	if err := validateID(id, "trunk id"); err != nil {
		return sqlc.Trunk{}, err
	}
	item, err := s.repo.Get(ctx, organizationID, id)
	return item, readError(err, "trunk not found")
}
func (s *Service) Update(ctx context.Context, organizationID, id uuid.UUID, req UpdateRequest) (sqlc.Trunk, error) {
	if _, err := s.Get(ctx, organizationID, id); err != nil {
		return sqlc.Trunk{}, err
	}
	if req.Name == nil && req.Direction == nil && req.Status == nil {
		return sqlc.Trunk{}, apperror.NewBadRequest("at least one field is required")
	}
	if req.Name != nil {
		value, err := normalizeName(*req.Name)
		if err != nil {
			return sqlc.Trunk{}, err
		}
		req.Name = &value
	}
	if req.Direction != nil {
		value, err := normalizeChoice(*req.Direction, directions, "direction")
		if err != nil {
			return sqlc.Trunk{}, err
		}
		req.Direction = &value
	}
	if req.Status != nil {
		value, err := normalizeChoice(*req.Status, statuses, "status")
		if err != nil {
			return sqlc.Trunk{}, err
		}
		req.Status = &value
	}
	item, err := s.repo.Update(ctx, sqlc.UpdateTrunkParams{Name: req.Name, Direction: req.Direction, Status: req.Status, ID: id, OrganizationID: organizationID})
	return item, writeError(err, "trunk", "trunk not found")
}
func (s *Service) Delete(ctx context.Context, organizationID, id uuid.UUID) error {
	if _, err := s.Get(ctx, organizationID, id); err != nil {
		return err
	}
	if err := s.repo.Disable(ctx, organizationID, id); err != nil {
		return readError(err, "disable trunk")
	}
	return nil
}

func (s *Service) CreateEndpoint(ctx context.Context, organizationID, trunkID uuid.UUID, req EndpointCreateRequest) (sqlc.TrunkEndpoint, error) {
	if _, err := s.Get(ctx, organizationID, trunkID); err != nil {
		return sqlc.TrunkEndpoint{}, err
	}
	host, err := normalizeHost(req.Host)
	if err != nil {
		return sqlc.TrunkEndpoint{}, err
	}
	if req.Port != nil {
		if err := validatePort(*req.Port); err != nil {
			return sqlc.TrunkEndpoint{}, err
		}
	}
	if req.Transport != nil {
		value, e := normalizeChoice(*req.Transport, transports, "transport")
		if e != nil {
			return sqlc.TrunkEndpoint{}, e
		}
		req.Transport = &value
	}
	if req.Direction != nil {
		value, e := normalizeChoice(*req.Direction, directions, "direction")
		if e != nil {
			return sqlc.TrunkEndpoint{}, e
		}
		req.Direction = &value
	}
	if req.Priority != nil {
		if err := validatePriority(*req.Priority); err != nil {
			return sqlc.TrunkEndpoint{}, err
		}
	}
	if req.Weight != nil {
		if err := validateWeight(*req.Weight); err != nil {
			return sqlc.TrunkEndpoint{}, err
		}
	}
	item, err := s.repo.CreateEndpoint(ctx, sqlc.CreateTrunkEndpointParams{OrganizationID: organizationID, TrunkID: trunkID, Host: host, Port: req.Port, Transport: req.Transport, Direction: req.Direction, Priority: req.Priority, Weight: req.Weight, Enabled: req.Enabled})
	return item, writeError(err, "trunk endpoint", "trunk not found")
}
func (s *Service) ListEndpoints(ctx context.Context, organizationID, trunkID uuid.UUID) ([]sqlc.TrunkEndpoint, error) {
	if _, err := s.Get(ctx, organizationID, trunkID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListEndpoints(ctx, organizationID, trunkID)
	if err != nil {
		return nil, apperror.NewInternal("list trunk endpoints", err)
	}
	return items, nil
}
func (s *Service) GetEndpoint(ctx context.Context, organizationID, trunkID, id uuid.UUID) (sqlc.TrunkEndpoint, error) {
	if _, err := s.Get(ctx, organizationID, trunkID); err != nil {
		return sqlc.TrunkEndpoint{}, err
	}
	if err := validateID(id, "endpoint id"); err != nil {
		return sqlc.TrunkEndpoint{}, err
	}
	item, err := s.repo.GetEndpoint(ctx, organizationID, trunkID, id)
	return item, readError(err, "trunk endpoint not found")
}
func (s *Service) UpdateEndpoint(ctx context.Context, organizationID, trunkID, id uuid.UUID, req EndpointUpdateRequest) (sqlc.TrunkEndpoint, error) {
	if _, err := s.GetEndpoint(ctx, organizationID, trunkID, id); err != nil {
		return sqlc.TrunkEndpoint{}, err
	}
	if req.Host == nil && req.Port == nil && req.Transport == nil && req.Direction == nil && req.Priority == nil && req.Weight == nil && req.Enabled == nil {
		return sqlc.TrunkEndpoint{}, apperror.NewBadRequest("at least one field is required")
	}
	if req.Host != nil {
		value, err := normalizeHost(*req.Host)
		if err != nil {
			return sqlc.TrunkEndpoint{}, err
		}
		req.Host = &value
	}
	if req.Port != nil {
		if err := validatePort(*req.Port); err != nil {
			return sqlc.TrunkEndpoint{}, err
		}
	}
	if req.Transport != nil {
		value, err := normalizeChoice(*req.Transport, transports, "transport")
		if err != nil {
			return sqlc.TrunkEndpoint{}, err
		}
		req.Transport = &value
	}
	if req.Direction != nil {
		value, err := normalizeChoice(*req.Direction, directions, "direction")
		if err != nil {
			return sqlc.TrunkEndpoint{}, err
		}
		req.Direction = &value
	}
	if req.Priority != nil {
		if err := validatePriority(*req.Priority); err != nil {
			return sqlc.TrunkEndpoint{}, err
		}
	}
	if req.Weight != nil {
		if err := validateWeight(*req.Weight); err != nil {
			return sqlc.TrunkEndpoint{}, err
		}
	}
	item, err := s.repo.UpdateEndpoint(ctx, sqlc.UpdateTrunkEndpointParams{Host: req.Host, Port: req.Port, Transport: req.Transport, Direction: req.Direction, Priority: req.Priority, Weight: req.Weight, Enabled: req.Enabled, ID: id, TrunkID: trunkID, OrganizationID: organizationID})
	return item, writeError(err, "trunk endpoint", "trunk endpoint not found")
}
func (s *Service) DeleteEndpoint(ctx context.Context, organizationID, trunkID, id uuid.UUID) error {
	if _, err := s.GetEndpoint(ctx, organizationID, trunkID, id); err != nil {
		return err
	}
	if err := s.repo.DeleteEndpoint(ctx, organizationID, trunkID, id); err != nil {
		return readError(err, "delete trunk endpoint")
	}
	return nil
}

func readError(err error, message string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NewNotFound(message)
	}
	return apperror.NewInternal(message, err)
}
func writeError(err error, resource, notFound string) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.NewConflict(resource + " already exists")
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NewNotFound(notFound)
	}
	return apperror.NewInternal("write "+resource, err)
}
