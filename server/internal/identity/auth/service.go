package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/platform/email"
	"github.com/leamout/leamout/server/internal/security/otp"
	"github.com/leamout/leamout/server/internal/security/password"
	"github.com/leamout/leamout/server/internal/security/token"
	"github.com/leamout/leamout/server/pkg/apperror"
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
	if s.pool == nil {
		return s.loginWithPassword(ctx, transactionID, value)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return sqlc.User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := sqlc.New(tx).LockAuthTransaction(ctx, transactionID); err != nil {
		return sqlc.User{}, err
	}
	local := NewService(NewRepository(sqlc.New(tx)))
	user, err := local.loginWithPassword(ctx, transactionID, value)
	if err != nil {
		return sqlc.User{}, err
	}
	if s.emails != nil {
		if err := s.emails.CancelTx(ctx, tx, otpEmailCancellationKey(transactionID)); err != nil {
			return sqlc.User{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return sqlc.User{}, err
	}
	return user, nil
}

func (s *Service) loginWithPassword(ctx context.Context, transactionID uuid.UUID, value string) (sqlc.User, error) {
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
	if _, err = sqlc.New(tx).LockAuthTransaction(ctx, transactionID); err != nil {
		return "", err
	}
	queries := sqlc.New(tx)
	local := NewService(NewRepository(queries))
	transaction, err := local.getValidTransaction(ctx, transactionID)
	if err != nil {
		return "", err
	}
	// Serialize recipient quotas across different transactions.
	if err = queries.LockAuthRecipient(ctx, transaction.Identifier); err != nil {
		return "", err
	}
	now := time.Now()
	minute, err := queries.CountEmailDeliveriesSince(ctx, sqlc.CountEmailDeliveriesSinceParams{Recipient: transaction.Identifier, Template: "otp", Since: pgconv.TimeToTimestamptz(now.Add(-time.Minute))})
	if err != nil {
		return "", err
	}
	hour, err := queries.CountEmailDeliveriesSince(ctx, sqlc.CountEmailDeliveriesSinceParams{Recipient: transaction.Identifier, Template: "otp", Since: pgconv.TimeToTimestamptz(now.Add(-time.Hour))})
	if err != nil {
		return "", err
	}
	if minute > 0 || hour >= 5 {
		return "", apperror.NewTooManyRequests("please wait before requesting another code")
	}
	if err := queries.InvalidateAuthOTPChallenges(ctx, &transactionID); err != nil {
		return "", err
	}
	key := otpEmailCancellationKey(transactionID)
	if err := s.emails.CancelTx(ctx, tx, key); err != nil {
		return "", err
	}
	code, err := otp.GenerateNumeric(6)
	if err != nil {
		return "", err
	}
	expiresAt := pgconv.TimestamptzToTime(transaction.ExpiresAt)
	_, err = queries.CreateAuthChallenge(ctx, sqlc.CreateAuthChallengeParams{
		Identifier: transaction.Identifier, SecretHash: token.Hash(code), ExpiresAt: pgconv.NullableTimestamptz(&expiresAt), Purpose: "email_otp", State: []byte(`{}`), AuthTransactionID: &transactionID, MaxAttempts: 5,
	})
	if err != nil {
		return "", err
	}
	_, err = s.emails.QueueTx(ctx, tx, email.Request{To: transaction.Identifier, Template: "otp", CancellationKey: &key, Data: email.Data{Code: code, ExpiresAt: expiresAt}})
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
	transaction, err := sqlc.New(tx).LockAuthTransaction(ctx, transactionID)
	if err != nil {
		return sqlc.User{}, apperror.NewUnauthorized("invalid authentication transaction")
	}
	if err := sqlc.New(tx).LockAuthRecipient(ctx, transaction.Identifier); err != nil {
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
	if s.emails != nil {
		cancel := verifyErr == nil
		if verifyErr != nil {
			_, activeErr := local.repo.GetActiveAuthChallenge(ctx, sqlc.GetActiveAuthChallengeParams{AuthTransactionID: &transactionID, Purpose: "email_otp"})
			if activeErr != nil && !errors.Is(activeErr, pgx.ErrNoRows) {
				return sqlc.User{}, activeErr
			}
			cancel = errors.Is(activeErr, pgx.ErrNoRows)
		}
		if cancel {
			if err := s.emails.CancelTx(ctx, tx, otpEmailCancellationKey(transactionID)); err != nil {
				return sqlc.User{}, err
			}
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

func otpEmailCancellationKey(id uuid.UUID) string { return "auth:otp:" + id.String() }
