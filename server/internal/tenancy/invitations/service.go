package invitations

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/coffeyvidzro/monogo/internal/database/pgconv"
	"github.com/coffeyvidzro/monogo/internal/database/sqlc"
	"github.com/coffeyvidzro/monogo/internal/platform/email"
	"github.com/coffeyvidzro/monogo/internal/security/token"
	"github.com/coffeyvidzro/monogo/pkg/apperror"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	repo   *Repository
	emails *email.Service
	appURL string
}

func NewService(repo *Repository, emails *email.Service, appURL string) *Service {
	return &Service{repo: repo, emails: emails, appURL: strings.TrimRight(appURL, "/")}
}

func cancellationKey(id uuid.UUID) string { return "invitation:" + id.String() }

func (s *Service) Create(ctx context.Context, actorID, organizationID uuid.UUID, req CreateRequest) (Response, error) {
	req, err := normalizeRequest(req)
	if err != nil {
		return Response{}, err
	}
	if actorID == uuid.Nil || organizationID == uuid.Nil {
		return Response{}, apperror.NewBadRequest("invalid invitation identity")
	}
	base, err := url.Parse(s.appURL)
	if err != nil || base.Scheme != "https" || base.Host == "" || base.User != nil {
		return Response{}, apperror.NewInternal("invitation application URL is invalid", nil)
	}
	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return Response{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := sqlc.New(tx)
	member, err := q.GetOrganizationMember(ctx, sqlc.GetOrganizationMemberParams{OrganizationID: organizationID, UserID: actorID})
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		return Response{}, apperror.NewForbidden("organization administrator required")
	}
	organization, err := q.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		return Response{}, err
	}
	inviter, err := q.GetUserByID(ctx, actorID)
	if err != nil {
		return Response{}, err
	}
	if err := q.ExpireOrganizationInvitations(ctx, organizationID); err != nil {
		return Response{}, err
	}
	secret, err := token.Generate(32)
	if err != nil {
		return Response{}, err
	}
	expiry := time.Now().Add(7 * 24 * time.Hour)
	invitation, err := q.CreateInvitation(ctx, sqlc.CreateInvitationParams{
		OrganizationID: organizationID, InvitedBy: actorID, Email: req.Email,
		Role: req.Role, TokenHash: token.Hash(secret), ExpiresAt: pgconv.TimeToTimestamptz(expiry),
	})
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23505" {
			return Response{}, apperror.NewConflict("a pending invitation already exists")
		}
		return Response{}, err
	}
	key := cancellationKey(invitation.ID)
	name := inviter.Email
	if inviter.Name != nil && strings.TrimSpace(*inviter.Name) != "" {
		name = *inviter.Name
	}
	_, err = s.emails.QueueTx(ctx, tx, email.Request{
		To: req.Email, Template: "invitation", ExpiresAt: expiry, CancellationKey: &key,
		Data: email.Data{Organization: organization.Name, Inviter: name, Role: req.Role,
			AcceptURL: s.appURL + "/invitations/accept?token=" + url.QueryEscape(secret), ExpiresAt: expiry},
	})
	if err != nil {
		return Response{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Response{}, err
	}
	return toResponse(invitation), nil
}

func (s *Service) Accept(ctx context.Context, userID uuid.UUID, secret string) (Response, error) {
	if userID == uuid.Nil || len(secret) != 64 {
		return Response{}, apperror.NewBadRequest("invalid invitation token")
	}
	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return Response{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := sqlc.New(tx)
	i, err := q.LockInvitationByTokenHash(ctx, token.Hash(secret))
	if err != nil {
		return Response{}, apperror.NewNotFound("invitation is invalid or expired")
	}
	user, err := q.GetUserByID(ctx, userID)
	if err != nil || !user.EmailVerified || !strings.EqualFold(user.Email, i.Email) {
		return Response{}, apperror.NewForbidden("invitation requires its verified email recipient")
	}
	i, err = q.AcceptInvitation(ctx, i.ID)
	if err != nil {
		return Response{}, err
	}
	if _, err := q.AddInvitedOrganizationMember(ctx, sqlc.AddInvitedOrganizationMemberParams{InvitationID: i.ID, UserID: userID}); err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23505" {
			return Response{}, apperror.NewConflict("recipient is already an organization member")
		}
		return Response{}, err
	}
	if err := s.emails.CancelTx(ctx, tx, cancellationKey(i.ID)); err != nil {
		return Response{}, err
	}
	organization, err := q.GetOrganizationByID(ctx, i.OrganizationID)
	if err != nil {
		return Response{}, err
	}
	inviter, err := q.GetUserByID(ctx, i.InvitedBy)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Response{}, err
	}
	if err == nil {
		name := user.Email
		if user.Name != nil && strings.TrimSpace(*user.Name) != "" {
			name = *user.Name
		}
		_, err = s.emails.QueueTx(ctx, tx, email.Request{
			To: inviter.Email, Template: "invitation-accepted", ExpiresAt: time.Now().Add(24 * time.Hour),
			Data: email.Data{Organization: organization.Name, MemberName: name, Role: i.Role},
		})
		if err != nil {
			return Response{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Response{}, err
	}
	return toResponse(i), nil
}

func (s *Service) Revoke(ctx context.Context, actorID, organizationID, invitationID uuid.UUID) error {
	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := sqlc.New(tx)
	member, err := q.GetOrganizationMember(ctx, sqlc.GetOrganizationMemberParams{OrganizationID: organizationID, UserID: actorID})
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		return apperror.NewForbidden("organization administrator required")
	}
	if _, err := q.RevokeInvitation(ctx, sqlc.RevokeInvitationParams{ID: invitationID, OrganizationID: organizationID}); err != nil {
		return apperror.NewNotFound("pending invitation not found")
	}
	if err := s.emails.CancelTx(ctx, tx, cancellationKey(invitationID)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
