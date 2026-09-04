package numbers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/modules/audit"
	"github.com/leamout/leamout/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	req CreateRequest,
) (sqlc.PhoneNumber, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return sqlc.PhoneNumber{}, err
	}

	var err error
	req.Number, err = normalizeNumber(req.Number)
	if err != nil {
		return sqlc.PhoneNumber{}, err
	}

	req.CountryCode, err = normalizeCountryCode(req.CountryCode)
	if err != nil {
		return sqlc.PhoneNumber{}, err
	}

	result, err := s.repo.Create(ctx, organizationID, req)
	return result, writeError(err, "create phone number")
}

// CreateManaged persists an already-provisioned upstream number. Only trusted
// provisioning orchestration should call this method; provider metadata is not
// accepted by the public phone-number creation route.
func (s *Service) CreateManaged(
	ctx context.Context,
	organizationID uuid.UUID,
	req ManagedCreateRequest,
) (sqlc.PhoneNumber, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return sqlc.PhoneNumber{}, err
	}
	if req.ProviderID == uuid.Nil {
		return sqlc.PhoneNumber{}, apperror.NewBadRequest("provider_id is required")
	}
	if req.CarrierConnectionID != nil && *req.CarrierConnectionID == uuid.Nil {
		return sqlc.PhoneNumber{}, apperror.NewBadRequest("carrier_connection_id must be valid when provided")
	}

	var err error
	req.Number, err = normalizeNumber(req.Number)
	if err != nil {
		return sqlc.PhoneNumber{}, err
	}
	req.CountryCode, err = normalizeCountryCode(req.CountryCode)
	if err != nil {
		return sqlc.PhoneNumber{}, err
	}
	req.ProviderResourceID = strings.TrimSpace(req.ProviderResourceID)
	if req.ProviderResourceID == "" {
		return sqlc.PhoneNumber{}, apperror.NewBadRequest("provider_resource_id is required")
	}

	result, err := s.repo.CreateManaged(ctx, organizationID, req)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.PhoneNumber{}, apperror.NewNotFound("active organization, provider, or platform carrier connection not found")
	}
	return result, writeError(err, "create managed phone number")
}

func (s *Service) List(
	ctx context.Context,
	organizationID uuid.UUID,
) ([]sqlc.PhoneNumber, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return nil, err
	}

	result, err := s.repo.List(ctx, organizationID)
	if err != nil {
		return nil, apperror.NewInternal("list phone numbers", err)
	}

	return result, nil
}

func (s *Service) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.PhoneNumber, error) {
	if err := validIDs(organizationID, id); err != nil {
		return sqlc.PhoneNumber{}, err
	}

	result, err := s.repo.Get(ctx, organizationID, id)
	return result, readError(err, "phone number not found")
}

func (s *Service) Update(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
	req UpdateRequest,
) (sqlc.PhoneNumber, error) {
	if err := validIDs(organizationID, id); err != nil {
		return sqlc.PhoneNumber{}, err
	}

	if req.CountryCode == nil && req.VoiceEnabled == nil && req.SMSEnabled == nil {
		return sqlc.PhoneNumber{}, apperror.NewBadRequest("at least one field is required")
	}

	if req.CountryCode != nil {
		countryCode, err := normalizeCountryCode(*req.CountryCode)
		if err != nil {
			return sqlc.PhoneNumber{}, err
		}
		req.CountryCode = &countryCode
	}

	result, err := s.repo.Update(ctx, organizationID, id, req)
	return result, writeError(err, "phone number not found")
}

func (s *Service) Delete(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) error {
	if _, err := s.Get(ctx, organizationID, id); err != nil {
		return err
	}

	return writeError(s.repo.Disable(ctx, organizationID, id), "disable phone number")
}

func (s *Service) SetCarrierConnection(ctx context.Context, organizationID, id uuid.UUID, req CarrierConnectionRequest) (sqlc.PhoneNumber, error) {
	if err := validIDs(organizationID, id); err != nil {
		return sqlc.PhoneNumber{}, err
	}
	if req.CarrierConnectionID == uuid.Nil {
		return sqlc.PhoneNumber{}, apperror.NewBadRequest("carrier_connection_id is required")
	}
	current, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return sqlc.PhoneNumber{}, err
	}
	actor, err := audit.ActorFromContext(ctx)
	if err != nil {
		return sqlc.PhoneNumber{}, apperror.NewInternal("attribute number assignment audit event", err)
	}
	action := "number.carrier_assigned"
	metadata := map[string]any{"carrier_connection_id": req.CarrierConnectionID}
	if current.CarrierConnectionID != nil {
		action = "number.carrier_reassigned"
		metadata["previous_carrier_connection_id"] = *current.CarrierConnectionID
	}
	event, err := audit.NewEvent(organizationID, actor, action, "phone_number", id, metadata)
	if err != nil {
		return sqlc.PhoneNumber{}, apperror.NewInternal("create number assignment audit event", err)
	}
	result, err := s.repo.SetCarrierConnection(ctx, organizationID, id, req.CarrierConnectionID, event)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.PhoneNumber{}, apperror.NewNotFound("phone number or active carrier connection not found")
	}
	return result, writeError(err, "assign carrier connection")
}

func validIDs(organizationID, id uuid.UUID) error {
	if err := validateOrganizationID(organizationID); err != nil {
		return err
	}

	return validateNumberID(id)
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

func writeError(err error, message string) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.NewConflict("phone number already exists")
	}

	return readError(err, message)
}
