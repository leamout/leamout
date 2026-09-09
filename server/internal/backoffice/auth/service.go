package auth

import (
	"context"
	"errors"

	"github.com/leamout/leamout/internal/database/sqlc"
	identityauth "github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/security/authn"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthenticated    = errors.New("backoffice authentication required")
	ErrForbidden          = errors.New("backoffice access forbidden")
)

type Service struct {
	auth       *identityauth.Service
	sessions   *session.Service
	repository *Repository
}

func NewService(
	authService *identityauth.Service,
	sessionService *session.Service,
	repository *Repository,
) *Service {
	return &Service{
		auth:       authService,
		sessions:   sessionService,
		repository: repository,
	}
}

func (s *Service) LoginWithPassword(
	ctx context.Context,
	email string,
	password string,
	ipAddress *string,
	userAgent *string,
) (string, sqlc.Session, error) {
	transaction, err := s.auth.Start(ctx, email)
	if err != nil {
		return "", sqlc.Session{}, ErrInvalidCredentials
	}

	user, err := s.auth.LoginWithPassword(ctx, transaction.ID, password)
	if err != nil {
		return "", sqlc.Session{}, ErrInvalidCredentials
	}
	if !user.IsPlatformAdmin {
		return "", sqlc.Session{}, ErrForbidden
	}

	return s.sessions.CreateForAudience(
		ctx,
		user.ID,
		ipAddress,
		userAgent,
		session.AudienceBackoffice,
	)
}

func (s *Service) Authenticate(
	ctx context.Context,
	token string,
) (authn.Principal, error) {
	backofficeSession, err := s.sessions.GetForAudience(
		ctx,
		token,
		session.AudienceBackoffice,
	)
	if err != nil {
		return authn.Principal{}, ErrUnauthenticated
	}

	user, err := s.repository.GetUser(ctx, backofficeSession.UserID)
	if err != nil {
		return authn.Principal{}, ErrUnauthenticated
	}
	if !user.IsPlatformAdmin {
		return authn.Principal{}, ErrForbidden
	}

	return authn.Principal{
		Subject: authn.Subject{
			ID:   user.ID,
			Type: authn.SubjectUser,
		},
		Credential: authn.Credential{
			ID:   backofficeSession.ID,
			Type: authn.CredentialSession,
		},
		Assurance: authn.AssuranceUnknown,
	}, nil
}

func (s *Service) Logout(
	ctx context.Context,
	principal authn.Principal,
) error {
	userID, ok := principal.UserID()
	if !ok {
		return ErrUnauthenticated
	}
	if principal.Credential.Type != authn.CredentialSession {
		return ErrUnauthenticated
	}

	return s.sessions.Revoke(ctx, principal.Credential.ID, userID)
}
