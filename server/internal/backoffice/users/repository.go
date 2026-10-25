package users

import (
	"context"

	"github.com/google/uuid"

	"github.com/leamout/leamout/server/internal/database/sqlc"
)

// Repository owns explicit cross-organization operator read projections.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context, filter Filter) (List, error) {
	organizationID, err := optionalID(filter.OrganizationID)
	if err != nil {
		return List{}, err
	}
	rows, err := r.queries.ListBackofficeUsers(ctx, sqlc.ListBackofficeUsersParams{
		Search:         filter.Search,
		Status:         filter.Status,
		OrganizationID: organizationID,
		PageLimit:      pageSize + 1,
		PageOffset:     (filter.Page - 1) * pageSize,
		Admin:          filter.Admin,
		Verified:       filter.Verified,
	})
	if err != nil {
		return List{}, err
	}
	result := List{
		Filter:  filter,
		HasNext: len(rows) > pageSize,
	}
	if result.HasNext {
		rows = rows[:pageSize]
	}
	for _, row := range rows {
		result.Users = append(result.Users, User{
			ID:                row.ID,
			Name:              row.Name,
			Email:             row.Email,
			EmailVerified:     row.EmailVerified,
			IsPlatformAdmin:   row.IsPlatformAdmin,
			Status:            row.Status,
			OrganizationCount: row.OrganizationCount,
			CreatedAt:         row.CreatedAt,
		})
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID, memberPage int32) (Detail, error) {
	row, err := r.queries.GetBackofficeUser(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	result := Detail{
		User: User{
			ID:                row.ID,
			Name:              row.Name,
			Email:             row.Email,
			EmailVerified:     row.EmailVerified,
			IsPlatformAdmin:   row.IsPlatformAdmin,
			Status:            row.Status,
			OrganizationCount: row.OrganizationCount,
			CreatedAt:         row.CreatedAt,
		},
		UpdatedAt:          row.UpdatedAt,
		ActiveSessionCount: row.ActiveSessionCount,
		LastSeenAt:         row.LastSeenAt,
		DisabledAt:         row.DisabledAt,
		MemberPage:         memberPage,
	}
	members, err := r.queries.ListBackofficeUserOrganizations(ctx, sqlc.ListBackofficeUserOrganizationsParams{
		UserID:     id,
		PageLimit:  pageSize + 1,
		PageOffset: (memberPage - 1) * pageSize,
	})
	if err != nil {
		return Detail{}, err
	}
	result.HasNextMembers = len(members) > pageSize
	if result.HasNextMembers {
		members = members[:pageSize]
	}
	for _, member := range members {
		result.Memberships = append(result.Memberships, Membership{
			OrganizationID:     member.OrganizationID,
			Name:               member.Name,
			OrganizationStatus: member.OrganizationStatus,
			Role:               member.Role,
			Status:             member.MembershipStatus,
			JoinedAt:           member.JoinedAt,
		})
	}
	sessions, err := r.queries.ListBackofficeUserSessions(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	for _, session := range sessions {
		result.Sessions = append(result.Sessions, Session{
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
	return result, nil
}

func optionalID(value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
