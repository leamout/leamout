package organizations

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for organizations.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]Organization, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	rows, err := r.queries.ListBackofficeOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	organizations := make([]Organization, 0, len(rows))
	for _, row := range rows {
		organizations = append(organizations, Organization{
			ID:         row.ID,
			Name:       row.Name,
			Members:    row.MemberCount,
			Wallets:    row.WalletCount,
			Currencies: row.Currencies,
			Status:     row.Status,
			CreatedAt:  row.CreatedAt,
		})
	}
	return organizations, nil
}

func (r *Repository) Get(ctx context.Context, organizationID uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeOrganization(ctx, organizationID)
	if err != nil {
		return Detail{}, err
	}

	memberRows, err := r.queries.ListBackofficeOrganizationMembers(ctx, organizationID)
	if err != nil {
		return Detail{}, err
	}
	members := make([]Member, 0, len(memberRows))
	for _, member := range memberRows {
		members = append(members, Member{
			UserID:           member.UserID,
			Name:             member.Name,
			Email:            member.Email,
			Role:             member.Role,
			MembershipStatus: member.Status,
			UserStatus:       member.UserStatus,
			EmailVerified:    member.EmailVerified,
			PlatformAdmin:    member.IsPlatformAdmin,
			JoinedAt:         member.JoinedAt,
		})
	}

	return Detail{
		ID:           row.ID,
		Name:         row.Name,
		Status:       row.Status,
		Members:      row.MemberCount,
		Wallets:      row.WalletCount,
		Currencies:   row.Currencies,
		BillingModel: row.BillingModel,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		MembersList:  members,
	}, nil
}
