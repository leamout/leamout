package subscribers

import (
	"context"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/hasher"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(q *sqlc.Queries) *Repository { return &Repository{q} }
func (r *Repository) Domain(c context.Context, org, id uuid.UUID) (sqlc.SipDomain, error) {
	return r.queries.GetSipDomainByID(c, sqlc.GetSipDomainByIDParams{ID: id, OrganizationID: org})
}
func (r *Repository) Create(c context.Context, org uuid.UUID, req CreateRequest, hashes hasher.HA1Hashes) (sqlc.Subscriber, error) {
	return r.queries.CreateSubscriber(c, sqlc.CreateSubscriberParams{OrganizationID: org, SipDomainID: req.SIPDomainID, Username: req.Username, Ha1Md5: &hashes.MD5, Ha1Sha256: &hashes.SHA256, Ha1Sha512256: &hashes.SHA512_256, DisplayName: req.DisplayName})
}
func (r *Repository) List(c context.Context, org uuid.UUID) ([]sqlc.Subscriber, error) {
	return r.queries.ListSubscribersByOrganizationID(c, org)
}
func (r *Repository) Get(c context.Context, org, id uuid.UUID) (sqlc.Subscriber, error) {
	return r.queries.GetSubscriberByID(c, sqlc.GetSubscriberByIDParams{ID: id, OrganizationID: org})
}
func (r *Repository) Update(c context.Context, org, id uuid.UUID, name *string) (sqlc.Subscriber, error) {
	return r.queries.UpdateSubscriber(c, sqlc.UpdateSubscriberParams{ID: id, OrganizationID: org, DisplayName: name})
}
func (r *Repository) Rotate(c context.Context, org, id uuid.UUID, hashes hasher.HA1Hashes) (sqlc.Subscriber, error) {
	return r.queries.SetSubscriberPassword(c, sqlc.SetSubscriberPasswordParams{ID: id, OrganizationID: org, Ha1Md5: &hashes.MD5, Ha1Sha256: &hashes.SHA256, Ha1Sha512256: &hashes.SHA512_256})
}
func (r *Repository) Disable(c context.Context, org, id uuid.UUID) error {
	return r.queries.DisableSubscriber(c, sqlc.DisableSubscriberParams{ID: id, OrganizationID: org})
}
