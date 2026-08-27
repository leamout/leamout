package credentials

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

var (
	ErrInvalidInput = errors.New("invalid credential input")
	ErrInvalidToken = errors.New("invalid organization token")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// Create creates an organization credential and returns the plaintext token.
// The token is intentionally returned only from this operation and must not be
// persisted or logged by callers.
func (s *Service) Create(ctx context.Context, input CreateInput) (CreatedCredential, error) {
	if err := ValidateCreate(input); err != nil {
		return CreatedCredential{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	token, prefix, hash, err := GenerateToken()
	if err != nil {
		return CreatedCredential{}, err
	}

	row, err := s.repository.Create(ctx, input, hash, prefix)
	if err != nil {
		return CreatedCredential{}, err
	}

	credential, err := organizationTokenToCredential(row)
	if err != nil {
		return CreatedCredential{}, err
	}

	return CreatedCredential{Credential: credential, Token: token}, nil
}

func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (Credential, error) {
	if organizationID == uuid.Nil || id == uuid.Nil {
		return Credential{}, fmt.Errorf("%w: organization and credential ids are required", ErrInvalidInput)
	}

	row, err := s.repository.GetByID(ctx, organizationID, id)
	if err != nil {
		return Credential{}, err
	}
	return organizationTokenToCredential(row)
}

func (s *Service) Authenticate(ctx context.Context, token string) (Credential, error) {
	if token == "" {
		return Credential{}, ErrInvalidToken
	}

	prefix, err := ParseTokenPrefix(token)
	if err != nil {
		return Credential{}, ErrInvalidToken
	}

	// Prefix parsing provides a cheap format check. The database lookup is by
	// hash because the prefix is not secret and is not a credential verifier.
	_ = prefix
	row, err := s.repository.GetByTokenHash(ctx, HashToken(token))
	if err != nil {
		return Credential{}, ErrInvalidToken
	}

	return organizationTokenToCredential(row)
}

func (s *Service) Update(ctx context.Context, input UpdateInput, actorID uuid.UUID) (Credential, error) {
	if err := ValidateUpdate(input); err != nil {
		return Credential{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if actorID == uuid.Nil {
		return Credential{}, fmt.Errorf("%w: actor id is required", ErrInvalidInput)
	}

	row, err := s.repository.Update(ctx, input, actorID)
	if err != nil {
		return Credential{}, err
	}
	return organizationTokenToCredential(row)
}

func (s *Service) Disable(ctx context.Context, organizationID, id, actorID uuid.UUID) error {
	if organizationID == uuid.Nil || id == uuid.Nil || actorID == uuid.Nil {
		return fmt.Errorf("%w: organization, credential, and actor ids are required", ErrInvalidInput)
	}
	return s.repository.Disable(ctx, organizationID, id, actorID)
}

func (s *Service) Touch(ctx context.Context, id uuid.UUID, ip netip.Addr) error {
	if id == uuid.Nil || !ip.IsValid() {
		return fmt.Errorf("%w: credential id and valid ip are required", ErrInvalidInput)
	}
	return s.repository.Touch(ctx, id, ip)
}

func organizationTokenToCredential(row sqlc.OrganizationToken) (Credential, error) {
	var scopes []string
	if len(row.Scopes) > 0 {
		if err := unmarshalScopes(row.Scopes, &scopes); err != nil {
			return Credential{}, err
		}
	}

	return Credential{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		CreatedBy:      row.CreatedBy,
		Name:           row.Name,
		Description:    row.Description,
		TokenPrefix:    row.TokenPrefix,
		Scopes:         scopes,
		LastUsedAt:     nil,
		ExpiresAt:      nil,
		DisabledAt:     nil,
	}, nil
}
