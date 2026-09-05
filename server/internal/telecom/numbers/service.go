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

type numberRepository interface {
	CreateBYOC(context.Context, uuid.UUID, BYOCCreateRequest) (sqlc.PhoneNumber, error)
	CreateManaged(context.Context, uuid.UUID, ManagedCreateRequest) (sqlc.PhoneNumber, error)
	List(context.Context, uuid.UUID) ([]sqlc.PhoneNumber, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (sqlc.PhoneNumber, error)
	GetForRelease(context.Context, uuid.UUID, uuid.UUID) (sqlc.PhoneNumber, error)
	Update(context.Context, uuid.UUID, uuid.UUID, UpdateRequest) (sqlc.PhoneNumber, error)
	ReleaseBYOC(context.Context, uuid.UUID, uuid.UUID) (sqlc.PhoneNumber, error)
	SetCarrierConnection(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, audit.Event) (sqlc.PhoneNumber, error)
	SaveManagedSelection(context.Context, uuid.UUID, ManagedNumberCandidate) (string, error)
}

type ManagedNumberInventory interface {
	SearchAvailable(context.Context, AvailableSearchRequest) ([]ManagedNumberCandidate, error)
}

type Service struct {
	repo             numberRepository
	managedInventory ManagedNumberInventory
}

func NewService(repo numberRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetManagedAcquisition(inventory ManagedNumberInventory) {
	s.managedInventory = inventory
}

func (s *Service) SearchAvailable(
	ctx context.Context,
	organizationID uuid.UUID,
	req AvailableSearchRequest,
) ([]AvailableNumberResponse, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return nil, err
	}
	if s.managedInventory == nil {
		return nil, apperror.NewServiceUnavailable("managed number inventory is not configured", nil)
	}

	countryCode, err := normalizeCountryCode(req.CountryCode)
	if err != nil {
		return nil, err
	}
	contains, err := normalizeNumberContains(req.Contains)
	if err != nil {
		return nil, err
	}
	req.CountryCode = countryCode
	req.Contains = contains

	candidates, err := s.managedInventory.SearchAvailable(ctx, req)
	if err != nil {
		return nil, apperror.NewServiceUnavailable("search managed number inventory", err)
	}

	result := make([]AvailableNumberResponse, 0, len(candidates))
	for _, candidate := range candidates {
		candidate.Provider = strings.TrimSpace(candidate.Provider)
		candidate.ProviderInventoryID = strings.TrimSpace(candidate.ProviderInventoryID)
		candidate.ProviderProductID = strings.TrimSpace(candidate.ProviderProductID)
		if candidate.Provider == "" || candidate.ProviderInventoryID == "" || candidate.ProviderProductID == "" {
			return nil, apperror.NewServiceUnavailable("managed number provider returned incomplete purchase metadata", nil)
		}
		candidate.Number, err = normalizeNumber(candidate.Number)
		if err != nil {
			return nil, apperror.NewServiceUnavailable("managed number provider returned an invalid number", err)
		}
		candidate.CountryCode, err = normalizeCountryCode(candidate.CountryCode)
		if err != nil {
			return nil, apperror.NewServiceUnavailable("managed number provider returned an invalid country", err)
		}
		if candidate.ChannelsIncludedCount <= 0 {
			return nil, apperror.NewServiceUnavailable("managed number provider returned a number without included voice capacity", nil)
		}

		selectionID, err := s.repo.SaveManagedSelection(ctx, organizationID, candidate)
		if err != nil {
			return nil, apperror.NewServiceUnavailable("store managed number selection", err)
		}
		result = append(result, AvailableNumberResponse{
			SelectionID:  selectionID,
			Number:       candidate.Number,
			CountryCode:  candidate.CountryCode,
			VoiceEnabled: true,
		})
	}

	return result, nil
}

func (s *Service) CreateBYOC(
	ctx context.Context,
	organizationID uuid.UUID,
	req BYOCCreateRequest,
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

	result, err := s.repo.CreateBYOC(ctx, organizationID, req)
	return result, writeError(err, "create BYOC phone number")
}

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
		return sqlc.PhoneNumber{}, apperror.NewNotFound("active organization, provider, or matching platform carrier connection not found")
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

	if req.VoiceEnabled == nil && req.SMSEnabled == nil {
		return sqlc.PhoneNumber{}, apperror.NewBadRequest("at least one field is required")
	}

	result, err := s.repo.Update(ctx, organizationID, id, req)
	return result, writeError(err, "phone number not found")
}

func (s *Service) ReleaseBYOC(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) error {
	if err := validIDs(organizationID, id); err != nil {
		return err
	}

	current, err := s.repo.GetForRelease(ctx, organizationID, id)
	if err != nil {
		return readError(err, "phone number not found")
	}
	if current.ProvisioningMode != string(ProvisioningModeBYOC) {
		return apperror.NewConflict("managed phone numbers must be released through the managed provisioning workflow")
	}

	_, err = s.repo.ReleaseBYOC(ctx, organizationID, id)
	return writeError(err, "release BYOC phone number")
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
	if current.ProvisioningMode != string(ProvisioningModeBYOC) {
		return sqlc.PhoneNumber{}, apperror.NewConflict("managed phone number carrier connections are platform-managed")
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
