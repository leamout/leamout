package carriers

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

func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, req CreateRequest) (Response, error) {
	if err := validateID(organizationID, "organization_id"); err != nil {
		return Response{}, err
	}
	if err := validateID(req.ProviderID, "provider_id"); err != nil {
		return Response{}, err
	}
	name, err := normalizeName(req.Name)
	if err != nil {
		return Response{}, err
	}
	if req.Status != nil {
		value, err := normalizeStatus(*req.Status)
		if err != nil {
			return Response{}, err
		}
		req.Status = &value
	}
	if err := validateLimits(req.MaxCPS, req.MaxConcurrentCalls, req.MaxDailyMinutes); err != nil {
		return Response{}, err
	}
	codecs, err := normalizeCodecs(req.Codecs)
	if err != nil {
		return Response{}, err
	}
	item, err := s.repo.Create(ctx, sqlc.CreateCarrierConnectionParams{OrganizationID: organizationID, ProviderID: req.ProviderID, Name: name, Status: req.Status, InboundEnabled: req.InboundEnabled, MaxCps: req.MaxCPS, MaxConcurrentCalls: req.MaxConcurrentCalls, MaxDailyMinutes: req.MaxDailyMinutes, Codecs: codecs, SupportsVideo: req.SupportsVideo, SupportsFax: req.SupportsFax})
	if err != nil {
		return Response{}, writeError(err, "carrier connection", "provider not found")
	}
	return response(item), nil
}

func (s *Service) List(ctx context.Context, organizationID uuid.UUID) ([]Response, error) {
	if err := validateID(organizationID, "organization_id"); err != nil {
		return nil, err
	}
	items, err := s.repo.List(ctx, organizationID)
	if err != nil {
		return nil, apperror.NewInternal("list carrier connections", err)
	}
	result := make([]Response, 0, len(items))
	for _, item := range items {
		result = append(result, listResponse(item))
	}
	return result, nil
}

func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (Response, error) {
	if err := validateID(organizationID, "organization_id"); err != nil {
		return Response{}, err
	}
	if err := validateID(id, "carrier connection id"); err != nil {
		return Response{}, err
	}
	item, err := s.repo.Get(ctx, organizationID, id)
	if err != nil {
		return Response{}, readError(err, "carrier connection not found")
	}
	return getResponse(item), nil
}

func (s *Service) Update(ctx context.Context, organizationID, id uuid.UUID, req UpdateRequest) (Response, error) {
	if _, err := s.Get(ctx, organizationID, id); err != nil {
		return Response{}, err
	}
	if req.Name == nil && req.Status == nil && req.InboundEnabled == nil && req.MaxCPS == nil && req.MaxConcurrentCalls == nil && req.MaxDailyMinutes == nil && req.Codecs == nil && req.SupportsVideo == nil && req.SupportsFax == nil {
		return Response{}, apperror.NewBadRequest("at least one field is required")
	}
	if req.Name != nil {
		value, err := normalizeName(*req.Name)
		if err != nil {
			return Response{}, err
		}
		req.Name = &value
	}
	if req.Status != nil {
		value, err := normalizeStatus(*req.Status)
		if err != nil {
			return Response{}, err
		}
		req.Status = &value
	}
	if err := validateLimits(req.MaxCPS, req.MaxConcurrentCalls, req.MaxDailyMinutes); err != nil {
		return Response{}, err
	}
	codecs, err := normalizeCodecs(req.Codecs)
	if err != nil {
		return Response{}, err
	}
	item, err := s.repo.Update(ctx, sqlc.UpdateCarrierConnectionParams{Name: req.Name, Status: req.Status, InboundEnabled: req.InboundEnabled, MaxCps: req.MaxCPS, MaxConcurrentCalls: req.MaxConcurrentCalls, MaxDailyMinutes: req.MaxDailyMinutes, Codecs: codecs, SupportsVideo: req.SupportsVideo, SupportsFax: req.SupportsFax, ID: id, OrganizationID: organizationID})
	if err != nil {
		return Response{}, writeError(err, "carrier connection", "carrier connection not found")
	}
	return response(item), nil
}

func (s *Service) Delete(ctx context.Context, organizationID, id uuid.UUID) error {
	if _, err := s.Get(ctx, organizationID, id); err != nil {
		return err
	}
	if err := s.repo.Disable(ctx, organizationID, id); err != nil {
		return apperror.NewInternal("disable carrier connection", err)
	}
	return nil
}

func (s *Service) CreateSourceIP(ctx context.Context, organizationID, connectionID uuid.UUID, req SourceIPCreateRequest) (SourceIPResponse, error) {
	if _, err := s.Get(ctx, organizationID, connectionID); err != nil {
		return SourceIPResponse{}, err
	}
	cidr, err := normalizeCIDR(req.CIDR)
	if err != nil {
		return SourceIPResponse{}, err
	}
	item, err := s.repo.CreateSourceIP(ctx, organizationID, connectionID, cidr)
	if err != nil {
		return SourceIPResponse{}, writeError(err, "source IP", "carrier connection not found")
	}
	return sourceIPResponse(item), nil
}

func (s *Service) ListSourceIPs(ctx context.Context, organizationID, connectionID uuid.UUID) ([]SourceIPResponse, error) {
	if _, err := s.Get(ctx, organizationID, connectionID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListSourceIPs(ctx, organizationID, connectionID)
	if err != nil {
		return nil, apperror.NewInternal("list carrier connection source IPs", err)
	}
	result := make([]SourceIPResponse, 0, len(items))
	for _, item := range items {
		result = append(result, sourceIPResponse(item))
	}
	return result, nil
}

func (s *Service) DeleteSourceIP(ctx context.Context, organizationID, connectionID, id uuid.UUID) error {
	if err := validateID(id, "source IP id"); err != nil {
		return err
	}
	items, err := s.ListSourceIPs(ctx, organizationID, connectionID)
	if err != nil {
		return err
	}
	found := false
	for _, item := range items {
		if item.ID == id {
			found = true
			break
		}
	}
	if !found {
		return apperror.NewNotFound("source IP not found")
	}
	if err := s.repo.DeleteSourceIP(ctx, organizationID, connectionID, id); err != nil {
		return apperror.NewInternal("delete source IP", err)
	}
	return nil
}

func (s *Service) ListProviders(ctx context.Context) ([]ProviderResponse, error) {
	providers, err := s.repo.ListProviders(ctx)
	if err != nil {
		return nil, apperror.NewInternal("list carrier providers", err)
	}

	result := make([]ProviderResponse, 0, len(providers))
	for _, provider := range providers {
		result = append(result, providerResponse(provider))
	}

	return result, nil
}

func (s *Service) GetProvider(ctx context.Context, id uuid.UUID) (ProviderResponse, error) {
	if err := validateID(id, "carrier provider id"); err != nil {
		return ProviderResponse{}, err
	}

	provider, err := s.repo.GetProvider(ctx, id)
	if err != nil {
		return ProviderResponse{}, readError(err, "carrier provider not found")
	}

	return providerResponse(provider), nil
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
