package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/coffeyvidzro/monogo/internal/database/pgconv"
	"github.com/coffeyvidzro/monogo/internal/database/sqlc"
	"github.com/coffeyvidzro/monogo/internal/platform/email"
	"github.com/coffeyvidzro/monogo/internal/security/otp"
	"github.com/coffeyvidzro/monogo/internal/security/password"
	"github.com/coffeyvidzro/monogo/internal/security/token"
	"github.com/coffeyvidzro/monogo/pkg/apperror"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo   *Repository
	pool   *pgxpool.Pool
	emails *email.Service
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ConfigureEmail wires transactional delivery at application startup.
func (s *Service) ConfigureEmail(pool *pgxpool.Pool, emails *email.Service) {
	s.pool = pool
	s.emails = emails
}

func (s *Service) Start(ctx context.Context, email string) (sqlc.AuthTransaction, error) {
	email = normalizeEmail(email)
	if address, err := mail.ParseAddress(email); err != nil || address.Address != email {
		return sqlc.AuthTransaction{}, apperror.NewBadRequest("email is required")
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.AuthTransaction{}, err
	}
	var userID *uuid.UUID
	if err == nil {
		userID = &user.ID
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	transaction, err := s.repo.CreateAuthTransaction(ctx, sqlc.CreateAuthTransactionParams{
		UserID:     userID,
		Identifier: email,
		ExpiresAt:  pgconv.NullableTimestamptz(&expiresAt),
	})
	if err != nil {
		return sqlc.AuthTransaction{}, err
	}

	return transaction, nil
}

// LoginWithPassword authenticates a user using the password associated with
// the authentication transaction.
func (s *Service) LoginWithPassword(ctx context.Context, transactionID uuid.UUID, value string) (sqlc.User, error) {
	transaction, err := s.getValidTransaction(ctx, transactionID)
	if err != nil {
		return sqlc.User{}, err
	}

	if transaction.UserID == nil {
		return sqlc.User{}, apperror.NewUnauthorized("invalid credentials")
	}

	user, err := s.repo.GetUserByID(ctx, *transaction.UserID)
	if err != nil {
		return sqlc.User{}, apperror.NewUnauthorized("invalid credentials")
	}

	if user.DisabledAt.Valid {
		return sqlc.User{}, apperror.NewUnauthorized("account is disabled")
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return sqlc.User{}, apperror.NewBadRequest("password is not enrolled")
	}

	if !password.Verify(value, *user.PasswordHash) {
		return sqlc.User{}, apperror.NewUnauthorized("invalid credentials")
	}

	if _, err := s.repo.MarkAuthTransactionAuthenticated(ctx, transactionID); err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

func (s *Service) SetPassword(ctx context.Context, userID uuid.UUID, value string) (sqlc.User, error) {
	if userID == uuid.Nil {
		return sqlc.User{}, apperror.NewUnauthorized("authentication required")
	}

	hash, err := password.Hash(value)
	if err != nil {
		return sqlc.User{}, apperror.NewInternal("failed to hash password", err)
	}

	return s.repo.SetUserPassword(ctx, sqlc.SetUserPasswordParams{
		ID:           userID,
		PasswordHash: &hash,
	})
}

func (s *Service) SendOTP(ctx context.Context, transactionID uuid.UUID) (string, error) {
	if s.pool == nil || s.emails == nil {
		return "", apperror.NewInternal("email delivery is not configured", nil)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Serialize resend and verification for this authentication transaction.
	if _, err = tx.Exec(ctx, `SELECT id FROM auth_transactions WHERE id=$1 FOR UPDATE`, transactionID); err != nil {
		return "", err
	}
	queries := sqlc.New(tx)
	local := NewService(NewRepository(queries))
	transaction, err := local.getValidTransaction(ctx, transactionID)
	if err != nil {
		return "", err
	}
	// Serialize recipient quotas across different transactions.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, transaction.Identifier); err != nil {
		return "", err
	}
	var limited bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM email_deliveries WHERE recipient=$1 AND template='otp' AND created_at>now()-interval '60 seconds') OR (SELECT count(*) FROM email_deliveries WHERE recipient=$1 AND template='otp' AND created_at>now()-interval '1 hour')>=5`, transaction.Identifier).Scan(&limited)
	if err != nil {
		return "", err
	}
	if limited {
		return "", apperror.NewTooManyRequests("please wait before requesting another code")
	}
	if _, err = tx.Exec(ctx, `UPDATE auth_challenges SET consumed_at=now() WHERE auth_transaction_id=$1 AND purpose='email_otp' AND consumed_at IS NULL`, transactionID); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `UPDATE email_deliveries SET status='expired', encrypted_data=NULL, lock_token=NULL, locked_at=NULL WHERE challenge_id IN (SELECT id FROM auth_challenges WHERE auth_transaction_id=$1) AND status IN ('pending','sending')`, transactionID); err != nil {
		return "", err
	}
	code, err := otp.GenerateNumeric(6)
	if err != nil {
		return "", err
	}
	expiresAt := pgconv.TimestamptzToTime(transaction.ExpiresAt)
	challenge, err := queries.CreateAuthChallenge(ctx, sqlc.CreateAuthChallengeParams{
		Identifier: transaction.Identifier, SecretHash: token.Hash(code), ExpiresAt: pgconv.NullableTimestamptz(&expiresAt), Purpose: "email_otp", State: []byte(`{}`), AuthTransactionID: &transactionID, MaxAttempts: 5,
	})
	if err != nil {
		return "", err
	}
	_, err = s.emails.QueueTx(ctx, tx, email.Request{To: transaction.Identifier, Template: "otp", ChallengeID: &challenge.ID, Data: email.Data{Code: code, ExpiresAt: expiresAt}})
	if err != nil {
		return "", err
	}
	if _, err = queries.SetAuthTransactionState(ctx, sqlc.SetAuthTransactionStateParams{ID: transactionID, State: "otp_sent"}); err != nil {
		return "", err
	}
	return "", tx.Commit(ctx)
}

// VerifyOTP verifies the active OTP challenge and authenticates the
// corresponding transaction.
func (s *Service) VerifyOTP(ctx context.Context, transactionID uuid.UUID, code string) (sqlc.User, error) {
	if s.pool == nil {
		return s.verifyOTP(ctx, transactionID, code)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return sqlc.User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, `SELECT id FROM auth_transactions WHERE id=$1 FOR UPDATE`, transactionID); err != nil {
		return sqlc.User{}, err
	}
	var identifier string
	if err := tx.QueryRow(ctx, `SELECT identifier FROM auth_transactions WHERE id=$1`, transactionID).Scan(&identifier); err != nil {
		return sqlc.User{}, apperror.NewUnauthorized("invalid authentication transaction")
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, identifier); err != nil {
		return sqlc.User{}, err
	}
	local := NewService(NewRepository(sqlc.New(tx)))
	user, verifyErr := local.verifyOTP(ctx, transactionID, code)
	// Preserve failed-code counters, but roll back unexpected verification failures.
	if verifyErr != nil {
		var app *apperror.AppError
		if !errors.As(verifyErr, &app) || app.Code != "UNAUTHORIZED" {
			return sqlc.User{}, verifyErr
		}
	}
	// Commit failed-code attempt counters too; database errors abort and roll back.
	if err := tx.Commit(ctx); err != nil {
		return sqlc.User{}, err
	}
	return user, verifyErr
}

func (s *Service) verifyOTP(ctx context.Context, transactionID uuid.UUID, code string) (sqlc.User, error) {
	transaction, err := s.getValidTransaction(ctx, transactionID)
	if err != nil {
		return sqlc.User{}, err
	}

	challenge, err := s.repo.GetActiveAuthChallenge(ctx, sqlc.GetActiveAuthChallengeParams{
		AuthTransactionID: &transactionID,
		Purpose:           "email_otp",
	})
	if err != nil {
		return sqlc.User{}, apperror.NewUnauthorized("invalid or expired authentication code")
	}

	if challenge.ConsumedAt.Valid {
		return sqlc.User{}, apperror.NewUnauthorized("authentication code has already been used")
	}

	if pgconv.TimestamptzToTime(challenge.ExpiresAt).Before(time.Now()) {
		return sqlc.User{}, apperror.NewUnauthorized("authentication code has expired")
	}

	if challenge.Attempts >= challenge.MaxAttempts {
		return sqlc.User{}, apperror.NewUnauthorized("too many authentication attempts")
	}

	if !token.Verify(code, challenge.SecretHash) {
		if _, err := s.repo.IncrementAuthChallengeAttempts(ctx, challenge.ID); err != nil {
			return sqlc.User{}, err
		}
		return sqlc.User{}, apperror.NewUnauthorized("invalid authentication code")
	}

	if _, err := s.repo.ConsumeAuthChallenge(ctx, challenge.ID); err != nil {
		return sqlc.User{}, err
	}

	user, err := s.getOrCreateOTPUser(ctx, transaction)
	if err != nil {
		return sqlc.User{}, err
	}

	if user.DisabledAt.Valid {
		return sqlc.User{}, apperror.NewUnauthorized("account is disabled")
	}

	if _, err := s.repo.MarkUserEmailVerified(ctx, user.ID); err != nil {
		return sqlc.User{}, err
	}

	if _, err := s.repo.MarkAuthTransactionAuthenticated(ctx, transactionID); err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

func (s *Service) getValidTransaction(ctx context.Context, transactionID uuid.UUID) (sqlc.AuthTransaction, error) {
	if transactionID == uuid.Nil {
		return sqlc.AuthTransaction{}, apperror.NewBadRequest("invalid transaction_id")
	}

	transaction, err := s.repo.GetAuthTransactionByID(ctx, transactionID)
	if err != nil {
		return sqlc.AuthTransaction{}, apperror.NewUnauthorized("invalid authentication transaction")
	}

	if transaction.ExpiresAt.Valid && pgconv.TimestamptzToTime(transaction.ExpiresAt).Before(time.Now()) {
		_ = s.repo.ExpireAuthTransaction(ctx, transactionID)
		return sqlc.AuthTransaction{}, apperror.NewUnauthorized("authentication transaction has expired")
	}

	return transaction, nil
}

func (s *Service) getOrCreateOTPUser(ctx context.Context, transaction sqlc.AuthTransaction) (sqlc.User, error) {
	if transaction.UserID != nil {
		return s.repo.GetUserByID(ctx, *transaction.UserID)
	}

	email := normalizeEmail(transaction.Identifier)
	if strings.TrimSpace(email) == "" {
		return sqlc.User{}, apperror.NewBadRequest("authentication identifier is missing")
	}

	// The recipient lock also serializes first-time signups across transactions.
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.User{}, err
	}
	return s.repo.CreateUser(ctx, sqlc.CreateUserParams{Email: email})
}
