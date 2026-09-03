package session

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/internal/security/token"
)

const sessionTokenBytes = 32

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(
	ctx context.Context,
	userID uuid.UUID,
	ipAddress *string,
	userAgent *string,
) (string, sqlc.Session, error) {
	if err := validateUserID(userID); err != nil {
		return "", sqlc.Session{}, err
	}

	value, err := token.Generate(sessionTokenBytes)
	if err != nil {
		return "", sqlc.Session{}, err
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	session, err := s.repo.Create(
		ctx,
		sqlc.CreateSessionParams{
			UserID:    userID,
			TokenHash: token.Hash(value),
			IpAddress: parseIP(ipAddress),
			UserAgent: userAgent,
			ExpiresAt: pgconv.NullableTimestamptz(&expiresAt),
		},
	)
	if err != nil {
		return "", sqlc.Session{}, err
	}

	return value, session, nil
}

func (s *Service) Get(
	ctx context.Context,
	value string,
) (sqlc.Session, error) {
	if err := validateToken(value); err != nil {
		return sqlc.Session{}, ErrInvalidSession
	}

	session, err := s.repo.GetByTokenHash(
		ctx,
		token.Hash(value),
	)
	if err != nil {
		return sqlc.Session{}, ErrInvalidSession
	}

	if err := validateSessionExpiry(session); err != nil {
		return sqlc.Session{}, err
	}

	return session, nil
}

// ResolveSession resolves a session token for the authentication layer.
func (s *Service) ResolveSession(
	ctx context.Context,
	value string,
) (authn.Session, error) {
	session, err := s.Get(ctx, value)
	if err != nil {
		return authn.Session{}, err
	}

	return authn.Session{
		ID:     session.ID,
		UserID: session.UserID,
	}, nil
}

func (s *Service) List(
	ctx context.Context,
	userID uuid.UUID,
) ([]sqlc.Session, error) {
	if err := validateUserID(userID); err != nil {
		return nil, err
	}

	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) Revoke(
	ctx context.Context,
	sessionID uuid.UUID,
	userID uuid.UUID,
) error {
	if err := validateSessionID(sessionID); err != nil {
		return err
	}

	if err := validateUserID(userID); err != nil {
		return err
	}

	return s.repo.Revoke(
		ctx,
		sqlc.RevokeSessionParams{
			ID:     sessionID,
			UserID: userID,
		},
	)
}

func (s *Service) RevokeAll(
	ctx context.Context,
	userID uuid.UUID,
) error {
	if err := validateUserID(userID); err != nil {
		return err
	}

	return s.repo.RevokeUserSessions(ctx, userID)
}
