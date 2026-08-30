package carriers

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

type Service struct {
	repo   *Repository
	cipher interface{ Encrypt(string) (string, error) }
}

func NewService(repo *Repository, cipher interface{ Encrypt(string) (string, error) }) *Service {
	return &Service{repo: repo, cipher: cipher}
}

func (s *Service) SetOutboundAuth(ctx context.Context, org, id uuid.UUID, req DigestAuthRequest) (Response, error) {
	if _, err := s.Get(ctx, org, id); err != nil {
		return Response{}, err
	}
	if strings.ToLower(strings.TrimSpace(req.Method)) != "digest" {
		return Response{}, apperror.NewBadRequest("outbound auth method must be digest")
	}
	username, err := requireAuthValue(req.Username, "username")
	if err != nil {
		return Response{}, err
	}
	secret, err := requireAuthValue(req.Secret, "secret")
	if err != nil {
		return Response{}, err
	}
	ciphertext, err := s.cipher.Encrypt(secret)
	if err != nil {
		return Response{}, apperror.NewInternal("encrypt outbound carrier credential", err)
	}
	realm, err := requireAuthValue(req.Realm, "realm")
	if err != nil {
		return Response{}, err
	}
	if err := s.repo.SetDigestAuth(ctx, org, id, "outbound", username, realm, ciphertext, digestHA1(username, realm, secret)); err != nil {
		return Response{}, digestWriteError(err, "set outbound carrier auth")
	}
	return s.Get(ctx, org, id)
}

func (s *Service) ClearOutboundAuth(ctx context.Context, org, id uuid.UUID) error {
	if _, err := s.Get(ctx, org, id); err != nil {
		return err
	}
	return writeAuthError(s.repo.ClearAuth(ctx, org, id, "outbound"), "clear outbound carrier auth")
}

func (s *Service) SetInboundAuth(ctx context.Context, org, id uuid.UUID, req InboundAuthRequest) (Response, error) {
	if _, err := s.Get(ctx, org, id); err != nil {
		return Response{}, err
	}
	switch strings.ToLower(strings.TrimSpace(req.Method)) {
	case "ip":
		if err := s.repo.SetInboundIPAuth(ctx, org, id); err != nil {
			return Response{}, apperror.NewInternal("set inbound carrier IP auth", err)
		}
	case "digest":
		username, err := requireAuthValue(req.Username, "username")
		if err != nil {
			return Response{}, err
		}
		secret, err := requireAuthValue(req.Secret, "secret")
		if err != nil {
			return Response{}, err
		}
		ciphertext, err := s.cipher.Encrypt(secret)
		if err != nil {
			return Response{}, apperror.NewInternal("encrypt inbound carrier credential", err)
		}
		realm, err := requireAuthValue(req.Realm, "realm")
		if err != nil {
			return Response{}, err
		}
		if err := s.repo.SetDigestAuth(ctx, org, id, "inbound", username, realm, ciphertext, digestHA1(username, realm, secret)); err != nil {
			return Response{}, digestWriteError(err, "set inbound carrier auth")
		}
	default:
		return Response{}, apperror.NewBadRequest("inbound auth method must be digest or ip")
	}
	return s.Get(ctx, org, id)
}

func (s *Service) ClearInboundAuth(ctx context.Context, org, id uuid.UUID) error {
	if _, err := s.Get(ctx, org, id); err != nil {
		return err
	}
	return writeAuthError(s.repo.ClearAuth(ctx, org, id, "inbound"), "clear inbound carrier auth")
}

func requireAuthValue(value, name string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", apperror.NewBadRequest(name + " is required")
	}
	return value, nil
}
func writeAuthError(err error, message string) error {
	if err != nil {
		return apperror.NewInternal(message, err)
	}
	return nil
}

func digestHA1(username, realm, secret string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(username+":"+realm+":"+secret)))
}

func digestWriteError(err error, message string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperror.NewConflict("digest username and realm are already in use")
	}
	return apperror.NewInternal(message, err)
}

func (s *Service) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	req CreateRequest,
) (Response, error) {
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
	item, err := s.repo.Create(ctx, sqlc.CreateCarrierConnectionParams{
		OrganizationID:     organizationID,
		ProviderID:         req.ProviderID,
		Name:               name,
		Status:             req.Status,
		InboundEnabled:     req.InboundEnabled,
		MaxCps:             req.MaxCPS,
		MaxConcurrentCalls: req.MaxConcurrentCalls,
		MaxDailyMinutes:    req.MaxDailyMinutes,
		Codecs:             codecs,
		SupportsVideo:      req.SupportsVideo,
		SupportsFax:        req.SupportsFax,
	})
	if err != nil {
		return Response{}, writeError(err, "carrier connection", "provider not found")
	}
	return response(item), nil
}

func (s *Service) List(
	ctx context.Context,
	organizationID uuid.UUID,
) ([]Response, error) {
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

func (s *Service) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (Response, error) {
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

func (s *Service) Update(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
	req UpdateRequest,
) (Response, error) {
	if _, err := s.Get(ctx, organizationID, id); err != nil {
		return Response{}, err
	}
	if req.Name == nil &&
		req.Status == nil &&
		req.InboundEnabled == nil &&
		req.MaxCPS == nil &&
		req.MaxConcurrentCalls == nil &&
		req.MaxDailyMinutes == nil &&
		req.Codecs == nil &&
		req.SupportsVideo == nil &&
		req.SupportsFax == nil {
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
	item, err := s.repo.Update(ctx, sqlc.UpdateCarrierConnectionParams{
		Name:               req.Name,
		Status:             req.Status,
		InboundEnabled:     req.InboundEnabled,
		MaxCps:             req.MaxCPS,
		MaxConcurrentCalls: req.MaxConcurrentCalls,
		MaxDailyMinutes:    req.MaxDailyMinutes,
		Codecs:             codecs,
		SupportsVideo:      req.SupportsVideo,
		SupportsFax:        req.SupportsFax,
		ID:                 id,
		OrganizationID:     organizationID,
	})
	if err != nil {
		return Response{}, writeError(err, "carrier connection", "carrier connection not found")
	}
	return response(item), nil
}

func (s *Service) Delete(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) error {
	if _, err := s.Get(ctx, organizationID, id); err != nil {
		return err
	}
	if err := s.repo.Disable(ctx, organizationID, id); err != nil {
		return apperror.NewInternal("disable carrier connection", err)
	}
	return nil
}

func (s *Service) CreateSourceIP(
	ctx context.Context,
	organizationID uuid.UUID,
	connectionID uuid.UUID,
	req SourceIPCreateRequest,
) (SourceIPResponse, error) {
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

func (s *Service) ListSourceIPs(
	ctx context.Context,
	organizationID uuid.UUID,
	connectionID uuid.UUID,
) ([]SourceIPResponse, error) {
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

func (s *Service) DeleteSourceIP(
	ctx context.Context,
	organizationID uuid.UUID,
	connectionID uuid.UUID,
	id uuid.UUID,
) error {
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
