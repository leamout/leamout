package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for users.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]User, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}

	rows, err := r.queries.ListBackofficeUsers(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, User{
			ID:                row.ID,
			Name:              row.Name,
			Email:             row.Email,
			EmailVerified:     row.EmailVerified,
			PlatformAdmin:     row.IsPlatformAdmin,
			Status:            row.Status,
			OrganizationCount: row.OrganizationCount,
			CreatedAt:         row.CreatedAt,
		})
	}

	return users, nil
}

func (r *Repository) Get(ctx context.Context, userID uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeUser(ctx, userID)
	if err != nil {
		return Detail{}, err
	}

	organizationRows, err := r.queries.ListBackofficeUserOrganizations(ctx, userID)
	if err != nil {
		return Detail{}, err
	}
	organizations := make([]OrganizationMembership, 0, len(organizationRows))
	for _, organization := range organizationRows {
		organizations = append(organizations, OrganizationMembership{
			OrganizationID:     organization.OrganizationID,
			OrganizationName:   organization.Name,
			OrganizationStatus: organization.OrganizationStatus,
			DetailAvailable:    organization.OrganizationStatus != "deleted",
			Role:               organization.Role,
			MembershipStatus:   organization.MembershipStatus,
			JoinedAt:           organization.JoinedAt,
		})
	}

	sessionRows, err := r.queries.ListBackofficeUserSessions(ctx, userID)
	if err != nil {
		return Detail{}, err
	}
	sessions := make([]Session, 0, len(sessionRows))
	for _, session := range sessionRows {
		sessions = append(sessions, Session{
			ID:         session.SessionID,
			Assurance:  session.Assurance,
			IPAddress:  session.IpAddress,
			UserAgent:  session.UserAgent,
			Status:     session.Status,
			CreatedAt:  session.CreatedAt,
			LastSeenAt: session.LastSeenAt,
			ExpiresAt:  session.ExpiresAt,
			RevokedAt:  session.RevokedAt,
		})
	}

	return Detail{
		User: User{
			ID:                row.ID,
			Name:              row.Name,
			Email:             row.Email,
			EmailVerified:     row.EmailVerified,
			PlatformAdmin:     row.IsPlatformAdmin,
			Status:            row.Status,
			OrganizationCount: row.OrganizationCount,
			CreatedAt:         row.CreatedAt,
		},
		ActiveSessions: row.ActiveSessionCount,
		LastSeenAt:     row.LastSeenAt,
		UpdatedAt:      row.UpdatedAt,
		DisabledAt:     row.DisabledAt,
		Organizations:  organizations,
		Sessions:       sessions,
	}, nil
}
