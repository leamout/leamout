package authn

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredential = errors.New("invalid credential")
)

type SessionResolver interface {
	ResolveSession(
		ctx context.Context,
		token string,
	) (Session, error)
}

type OrganizationTokenResolver interface {
	ResolveOrganizationToken(
		ctx context.Context,
		key string,
	) (OrganizationToken, error)
}

type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type OrganizationToken struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type Resolver struct {
	sessions           SessionResolver
	organizationTokens OrganizationTokenResolver
}

func NewResolver(
	sessions SessionResolver,
	organizationTokens OrganizationTokenResolver,
) *Resolver {
	return &Resolver{
		sessions:           sessions,
		organizationTokens: organizationTokens,
	}
}

func (r *Resolver) Resolve(
	ctx context.Context,
	input CredentialInput,
) (Principal, error) {
	if err := input.Validate(); err != nil {
		return Principal{}, err
	}

	switch input.Type {
	case CredentialSession:
		return r.resolveSession(ctx, input.Value)

	case CredentialOrganizationToken:
		return r.resolveOrganizationToken(ctx, input.Value)

	default:
		return Principal{}, ErrInvalidCredential
	}
}

func (r *Resolver) resolveSession(
	ctx context.Context,
	token string,
) (Principal, error) {
	session, err := r.sessions.ResolveSession(
		ctx,
		token,
	)
	if err != nil {
		return Principal{}, ErrInvalidCredential
	}

	return Principal{
		Subject: Subject{
			ID:   session.UserID,
			Type: SubjectUser,
		},
		Credential: Credential{
			ID:   session.ID,
			Type: CredentialSession,
		},
		Assurance: AssuranceUnknown,
	}, nil
}

func (r *Resolver) resolveOrganizationToken(
	ctx context.Context,
	key string,
) (Principal, error) {
	organizationToken, err := r.organizationTokens.ResolveOrganizationToken(
		ctx,
		key,
	)
	if err != nil {
		return Principal{}, ErrInvalidCredential
	}

	return Principal{
		Subject: Subject{
			ID:   organizationToken.UserID,
			Type: SubjectUser,
		},
		Credential: Credential{
			ID:   organizationToken.ID,
			Type: CredentialOrganizationToken,
		},
		Assurance: AssuranceUnknown,
	}, nil
}
