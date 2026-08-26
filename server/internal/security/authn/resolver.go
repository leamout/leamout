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

type APIKeyResolver interface {
	ResolveAPIKey(
		ctx context.Context,
		key string,
	) (APIKey, error)
}

type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type APIKey struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type Resolver struct {
	sessions SessionResolver
	apiKeys  APIKeyResolver
}

func NewResolver(
	sessions SessionResolver,
	apiKeys APIKeyResolver,
) *Resolver {
	return &Resolver{
		sessions: sessions,
		apiKeys:  apiKeys,
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

	case CredentialAPIKey:
		return r.resolveAPIKey(ctx, input.Value)

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

func (r *Resolver) resolveAPIKey(
	ctx context.Context,
	key string,
) (Principal, error) {
	apiKey, err := r.apiKeys.ResolveAPIKey(
		ctx,
		key,
	)
	if err != nil {
		return Principal{}, ErrInvalidCredential
	}

	return Principal{
		Subject: Subject{
			ID:   apiKey.UserID,
			Type: SubjectUser,
		},
		Credential: Credential{
			ID:   apiKey.ID,
			Type: CredentialAPIKey,
		},
		Assurance: AssuranceUnknown,
	}, nil
}
