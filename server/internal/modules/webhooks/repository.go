package webhooks

import (
	"context"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
	"time"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(q *sqlc.Queries) *Repository { return &Repository{q} }
func (r *Repository) Create(c context.Context, org uuid.UUID, req CreateRequest, secret []byte) (sqlc.WebhookEndpoint, error) {
	return r.queries.CreateWebhookEndpoint(c, sqlc.CreateWebhookEndpointParams{OrganizationID: org, URL: req.URL, SigningSecret: secret, SubscribedEvents: req.SubscribedEvents, Enabled: req.Enabled})
}
func (r *Repository) List(c context.Context, org uuid.UUID) ([]sqlc.WebhookEndpoint, error) {
	return r.queries.ListWebhookEndpointsByOrganizationID(c, org)
}
func (r *Repository) Get(c context.Context, org, id uuid.UUID) (sqlc.WebhookEndpoint, error) {
	return r.queries.GetWebhookEndpointByID(c, sqlc.GetWebhookEndpointByIDParams{ID: id, OrganizationID: org})
}
func (r *Repository) Update(c context.Context, org, id uuid.UUID, req UpdateRequest) (sqlc.WebhookEndpoint, error) {
	return r.queries.UpdateWebhookEndpoint(c, sqlc.UpdateWebhookEndpointParams{ID: id, OrganizationID: org, URL: req.URL, SubscribedEvents: req.SubscribedEvents, Enabled: req.Enabled})
}
func (r *Repository) Disable(c context.Context, org, id uuid.UUID) error {
	return r.queries.DisableWebhookEndpoint(c, sqlc.DisableWebhookEndpointParams{ID: id, OrganizationID: org})
}
func (r *Repository) RotateSecret(c context.Context, org, id uuid.UUID, secret []byte) (sqlc.WebhookEndpoint, error) {
	return r.queries.RotateWebhookEndpointSecret(c, sqlc.RotateWebhookEndpointSecretParams{ID: id, OrganizationID: org, SigningSecret: secret})
}
func (r *Repository) ListDeliveries(c context.Context, org, id uuid.UUID, limit, offset int32) ([]sqlc.WebhookDelivery, error) {
	return r.queries.ListWebhookDeliveriesForEndpoint(c, sqlc.ListWebhookDeliveriesForEndpointParams{OrganizationID: org, EndpointID: id, LimitCount: limit, OffsetCount: offset})
}
func (r *Repository) GetDelivery(c context.Context, org, id uuid.UUID) (sqlc.WebhookDelivery, error) {
	return r.queries.GetWebhookDelivery(c, sqlc.GetWebhookDeliveryParams{ID: id, OrganizationID: org})
}
func (r *Repository) Retry(c context.Context, org, id uuid.UUID) (sqlc.WebhookDelivery, error) {
	return r.queries.RetryWebhookDelivery(c, sqlc.RetryWebhookDeliveryParams{ID: id, OrganizationID: org})
}
func (r *Repository) Cancel(c context.Context, org, id uuid.UUID) error {
	_, err := r.queries.CancelWebhookDeliveriesForEndpoint(c, sqlc.CancelWebhookDeliveriesForEndpointParams{EndpointID: id, OrganizationID: org})
	return err
}
func (r *Repository) CreateTestDelivery(c context.Context, org, id uuid.UUID, eventID uuid.UUID, at time.Time) (sqlc.WebhookDelivery, error) {
	return r.queries.CreateWebhookDelivery(c, sqlc.CreateWebhookDeliveryParams{EndpointID: id, EventID: eventID, NextAttemptAt: at})
}
