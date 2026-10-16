package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/integrations/keycloak"
	"github.com/leamout/leamout/server/internal/security/encryption"

	"github.com/google/uuid"

	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/security/authn"
	"github.com/leamout/leamout/server/internal/security/token"
)

const (
	sessionTokenBytes = 32
	sessionTTL        = 30 * 24 * time.Hour
)

type Service struct {
	repo   *Repository
	db     *pgxpool.Pool
	client *keycloak.Client
	cipher *encryption.Cipher
}

func NewService(repo *Repository, db *pgxpool.Pool, client *keycloak.Client, cipher *encryption.Cipher) *Service {
	return &Service{
		repo:   repo,
		db:     db,
		client: client,
		cipher: cipher,
	}
}

// Create stores a browser session backed exclusively by Keycloak tokens.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, subject string, identitySessionID string, tokens keycloak.Tokens) (string, sqlc.Session, error) {
	if err := validateUserID(userID); err != nil {
		return "", sqlc.Session{}, err
	}
	if subject == "" || identitySessionID == "" || !tokens.ValidUntil.After(time.Now()) {
		return "", sqlc.Session{}, ErrInvalidSession
	}
	value, err := token.Generate(sessionTokenBytes)
	if err != nil {
		return "", sqlc.Session{}, err
	}
	encrypted, err := s.encrypt(userID, tokens)
	if err != nil {
		return "", sqlc.Session{}, err
	}
	expiresAt := time.Now().Add(sessionTTL)
	created, err := s.repo.Create(ctx, sqlc.CreateSessionParams{
		UserID:            userID,
		TokenHash:         token.Hash(value),
		IdentitySubject:   subject,
		IdentitySessionID: identitySessionID,
		TokensCiphertext:  encrypted,
		Assurance:         authn.AssuranceUnknown.String(),
		ExpiresAt:         pgconv.NullableTimestamptz(&expiresAt),
	})
	return value, created, err
}

func (s *Service) encrypt(userID uuid.UUID, tokens keycloak.Tokens) (string, error) {
	data, err := json.Marshal(tokens)
	if err != nil {
		return "", fmt.Errorf("encode identity tokens: %w", err)
	}
	return s.cipher.EncryptForScope("identity-session:"+userID.String(), string(data))
}

func (s *Service) decrypt(stored sqlc.Session) (keycloak.Tokens, error) {
	data, err := s.cipher.DecryptForScope("identity-session:"+stored.UserID.String(), stored.TokensCiphertext)
	if err != nil {
		return keycloak.Tokens{}, ErrInvalidSession
	}
	var tokens keycloak.Tokens
	if err := json.Unmarshal([]byte(data), &tokens); err != nil {
		return keycloak.Tokens{}, ErrInvalidSession
	}
	return tokens, nil
}

// refresh serializes rotation across API replicas, then rechecks the stored token.
func (s *Service) refresh(ctx context.Context, stored sqlc.Session) error {
	tokens, err := s.decrypt(stored)
	if err != nil {
		return err
	}
	if time.Now().Add(10 * time.Second).Before(tokens.ValidUntil) {
		return nil
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin session refresh: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := sqlc.New(tx)
	stored, err = queries.LockIdentitySession(ctx, stored.ID)
	if err != nil {
		return ErrInvalidSession
	}
	tokens, err = s.decrypt(stored)
	if err != nil {
		return err
	}
	if !time.Now().Add(10 * time.Second).Before(tokens.ValidUntil) {
		tokens, err = s.client.Refresh(ctx, tokens, stored.IdentitySubject, stored.IdentitySessionID)
		if err != nil {
			return ErrInvalidSession
		}
		encrypted, err := s.encrypt(stored.UserID, tokens)
		if err != nil {
			return err
		}
		if err := queries.UpdateIdentitySessionTokens(ctx, sqlc.UpdateIdentitySessionTokensParams{ID: stored.ID, TokensCiphertext: encrypted}); err != nil {
			return fmt.Errorf("persist session refresh: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit session refresh: %w", err)
	}
	return nil
}

func (s *Service) Get(
	ctx context.Context,
	value string,
) (sqlc.Session, error) {
	if err := validateToken(value); err != nil {
		return sqlc.Session{}, ErrInvalidSession
	}

	session, err := s.repo.GetByTokenHash(ctx, token.Hash(value))
	if err != nil {
		return sqlc.Session{}, ErrInvalidSession
	}

	if err := validateSessionExpiry(session); err != nil {
		return sqlc.Session{}, err
	}

	return session, nil
}

func (s *Service) ResolveSession(
	ctx context.Context,
	value string,
) (authn.Session, error) {
	session, err := s.Get(ctx, value)
	if err != nil {
		return authn.Session{}, err
	}

	if err := s.refresh(ctx, session); err != nil {
		return authn.Session{}, err
	}

	return authn.Session{
		ID:        session.ID,
		UserID:    session.UserID,
		Assurance: parseAssurance(session.Assurance),
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

func parseAssurance(value string) authn.AssuranceLevel {
	switch value {
	case authn.AssurancePassword.String():
		return authn.AssurancePassword
	case authn.AssuranceOTP.String():
		return authn.AssuranceOTP
	case authn.AssuranceMFA.String():
		return authn.AssuranceMFA
	default:
		return authn.AssuranceUnknown
	}
}
