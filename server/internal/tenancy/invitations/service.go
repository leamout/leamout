package invitations

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/platform/email"
	"github.com/leamout/leamout/server/internal/security/token"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Service struct {
	repo      *Repository
	emails    *email.Service
	acceptURL string
}

func NewService(repo *Repository, emails *email.Service, acceptURL string) *Service {
	return &Service{
		repo:      repo,
		emails:    emails,
		acceptURL: acceptURL,
	}
}

func (s *Service) requireAdmin(ctx context.Context, queries *sqlc.Queries, userID, organizationID uuid.UUID) error {
	member, err := queries.GetOrganizationMember(ctx, sqlc.GetOrganizationMemberParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		return apperror.NewForbidden("organization admin access required")
	}

	if member.Role != "admin" && member.Role != "owner" {
		return apperror.NewForbidden("organization admin access required")
	}

	return nil
}

func (s *Service) Create(ctx context.Context, userID, organizationID uuid.UUID, req CreateRequest) (Response, error) {
	req, err := validateCreate(req)
	if err != nil {
		return Response{}, err
	}

	u, err := url.Parse(s.acceptURL)
	if err != nil || !email.ValidURL(s.acceptURL) || s.emails == nil {
		return Response{}, apperror.NewServiceUnavailable("invitation email is not configured", nil)
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return Response{}, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queries := sqlc.New(tx)

	if err := s.requireAdmin(ctx, queries, userID, organizationID); err != nil {
		return Response{}, err
	}

	if err := queries.ExpireStaleInvitations(ctx); err != nil {
		return Response{}, err
	}

	rawToken, err := token.Generate(32)
	if err != nil {
		return Response{}, err
	}

	expiry := time.Now().Add(7 * 24 * time.Hour)
	invitation, err := queries.CreateInvitation(ctx, sqlc.CreateInvitationParams{
		OrganizationID: organizationID,
		InvitedBy:      userID,
		Email:          req.Email,
		Role:           req.Role,
		TokenHash:      token.Hash(rawToken),
		ExpiresAt:      pgconv.TimeToTimestamptz(expiry),
	})
	if err != nil {
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) && databaseError.Code == "23505" {
			return Response{}, apperror.NewConflict("pending invitation already exists")
		}
		return Response{}, err
	}

	organization, err := queries.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		return Response{}, err
	}

	inviter, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		return Response{}, err
	}

	inviterName := inviter.Email
	if inviter.Name != nil && *inviter.Name != "" {
		inviterName = *inviter.Name
	}

	values := u.Query()
	values.Set("token", rawToken)
	u.RawQuery = values.Encode()
	key := cancellationKey(invitation.ID)
	if _, err := s.emails.QueueTx(ctx, tx, email.Request{
		To:              req.Email,
		Template:        "organization-invitation",
		CancellationKey: &key,
		Data: email.Data{
			Organization: organization.Name,
			Inviter:      inviterName,
			Role:         req.Role,
			AcceptURL:    u.String(),
			ExpiresAt:    expiry,
		},
	}); err != nil {
		return Response{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Response{}, err
	}

	return response(invitation), nil
}

func (s *Service) List(ctx context.Context, userID, organizationID uuid.UUID) ([]Response, error) {
	queries := sqlc.New(s.repo.pool)
	if err := s.requireAdmin(ctx, queries, userID, organizationID); err != nil {
		return nil, err
	}

	rows, err := queries.ListPendingInvitationsByOrganizationID(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	result := make([]Response, 0, len(rows))
	for _, row := range rows {
		result = append(result, response(row))
	}

	return result, nil
}

func (s *Service) Revoke(ctx context.Context, userID, organizationID, id uuid.UUID) error {
	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queries := sqlc.New(tx)

	if err := s.requireAdmin(ctx, queries, userID, organizationID); err != nil {
		return err
	}

	if _, err := queries.RevokeInvitation(ctx, sqlc.RevokeInvitationParams{
		ID:             id,
		OrganizationID: organizationID,
	}); err != nil {
		return apperror.NewNotFound("pending invitation not found")
	}

	if err := s.emails.CancelTx(ctx, tx, cancellationKey(id)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) Accept(ctx context.Context, userID uuid.UUID, rawToken string) error {
	if len(rawToken) != 64 {
		return apperror.NewBadRequest("invalid invitation token")
	}

	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queries := sqlc.New(tx)
	invitation, err := queries.GetInvitationByTokenHash(ctx, token.Hash(rawToken))
	if err != nil {
		return apperror.NewNotFound("invitation is invalid or expired")
	}

	user, err := queries.GetUserByID(ctx, userID)
	if err != nil || !user.EmailVerified || user.DisabledAt.Valid || !strings.EqualFold(user.Email, invitation.Email) {
		return apperror.NewForbidden("invited verified email required")
	}

	// The conditional update serializes concurrent acceptance and revocation.
	if _, err := queries.AcceptInvitation(ctx, invitation.ID); err != nil {
		return apperror.NewConflict("invitation is no longer pending")
	}

	if _, err := queries.AddInvitedOrganizationMember(ctx, sqlc.AddInvitedOrganizationMemberParams{
		InvitationID: invitation.ID,
		UserID:       userID,
	}); err != nil {
		return err
	}

	if err := s.emails.CancelTx(ctx, tx, cancellationKey(invitation.ID)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func cancellationKey(id uuid.UUID) string {
	return "invitation:" + id.String()
}

func response(value sqlc.OrganizationInvitation) Response {
	return Response{
		ID:        value.ID,
		Email:     value.Email,
		Role:      value.Role,
		Status:    value.Status,
		ExpiresAt: pgconv.TimestamptzToTime(value.ExpiresAt),
	}
}
