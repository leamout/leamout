package carriers

import (
	"context"
	"net/netip"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(queries *sqlc.Queries) *Repository { return &Repository{queries: queries} }

func (r *Repository) Create(ctx context.Context, arg sqlc.CreateCarrierConnectionParams) (sqlc.CarrierConnection, error) {
	return r.queries.CreateCarrierConnection(ctx, arg)
}
func (r *Repository) List(ctx context.Context, organizationID uuid.UUID) ([]sqlc.ListCarrierConnectionsByOrganizationIDRow, error) {
	return r.queries.ListCarrierConnectionsByOrganizationID(ctx, organizationID)
}
func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (sqlc.GetCarrierConnectionByIDRow, error) {
	return r.queries.GetCarrierConnectionByID(ctx, sqlc.GetCarrierConnectionByIDParams{ID: id, OrganizationID: organizationID})
}
func (r *Repository) Update(ctx context.Context, arg sqlc.UpdateCarrierConnectionParams) (sqlc.CarrierConnection, error) {
	return r.queries.UpdateCarrierConnection(ctx, arg)
}
func (r *Repository) Disable(ctx context.Context, organizationID, id uuid.UUID) error {
	return r.queries.DisableCarrierConnection(ctx, sqlc.DisableCarrierConnectionParams{ID: id, OrganizationID: organizationID})
}
func (r *Repository) CreateSourceIP(ctx context.Context, organizationID, connectionID uuid.UUID, cidr netip.Prefix) (sqlc.CarrierConnectionSourceIp, error) {
	return r.queries.CreateCarrierConnectionSourceIP(ctx, sqlc.CreateCarrierConnectionSourceIPParams{OrganizationID: organizationID, CarrierConnectionID: connectionID, Cidr: cidr})
}
func (r *Repository) ListSourceIPs(ctx context.Context, organizationID, connectionID uuid.UUID) ([]sqlc.CarrierConnectionSourceIp, error) {
	return r.queries.ListCarrierConnectionSourceIPs(ctx, sqlc.ListCarrierConnectionSourceIPsParams{CarrierConnectionID: connectionID, OrganizationID: organizationID})
}
func (r *Repository) DeleteSourceIP(ctx context.Context, organizationID, connectionID, id uuid.UUID) error {
	return r.queries.DeleteCarrierConnectionSourceIP(ctx, sqlc.DeleteCarrierConnectionSourceIPParams{ID: id, CarrierConnectionID: connectionID, OrganizationID: organizationID})
}
