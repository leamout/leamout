package webhooks

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
	"net/url"
	"strings"
)

type Service struct{ repo *Repository }

func NewService(r *Repository) *Service { return &Service{r} }
func (s *Service) Create(c context.Context, org uuid.UUID, req CreateRequest) (sqlc.WebhookEndpoint, []byte, error) {
	if err := validOrg(org); err != nil {
		return sqlc.WebhookEndpoint{}, nil, err
	}
	if err := normalizeCreate(&req); err != nil {
		return sqlc.WebhookEndpoint{}, nil, err
	}
	secret, err := newSigningSecret()
	if err != nil {
		return sqlc.WebhookEndpoint{}, nil, apperror.NewInternal("generate webhook secret", err)
	}
	v, err := s.repo.Create(c, org, req, secret)
	return v, secret, writeErr(err, "create webhook")
}
func (s *Service) List(c context.Context, org uuid.UUID) ([]sqlc.WebhookEndpoint, error) {
	if err := validOrg(org); err != nil {
		return nil, err
	}
	return s.repo.List(c, org)
}
func (s *Service) Get(c context.Context, org, id uuid.UUID) (sqlc.WebhookEndpoint, error) {
	if err := validIDs(org, id); err != nil {
		return sqlc.WebhookEndpoint{}, err
	}
	v, e := s.repo.Get(c, org, id)
	return v, readErr(e, "webhook not found")
}
func (s *Service) Update(c context.Context, org, id uuid.UUID, req UpdateRequest) (sqlc.WebhookEndpoint, error) {
	if err := validIDs(org, id); err != nil {
		return sqlc.WebhookEndpoint{}, err
	}
	if err := normalizeUpdate(&req); err != nil {
		return sqlc.WebhookEndpoint{}, err
	}
	v, e := s.repo.Update(c, org, id, req)
	return v, readErr(e, "webhook not found")
}
func (s *Service) Delete(c context.Context, org, id uuid.UUID) error {
	if err := validIDs(org, id); err != nil {
		return err
	}
	if _, err := s.Get(c, org, id); err != nil {
		return err
	}
	if err := s.repo.Disable(c, org, id); err != nil {
		return writeErr(err, "disable webhook")
	}
	return writeErr(s.repo.Cancel(c, org, id), "cancel webhook deliveries")
}
func (s *Service) RotateSecret(c context.Context, org, id uuid.UUID) (sqlc.WebhookEndpoint, []byte, error) {
	if err := validIDs(org, id); err != nil {
		return sqlc.WebhookEndpoint{}, nil, err
	}
	secret, err := newSigningSecret()
	if err != nil {
		return sqlc.WebhookEndpoint{}, nil, apperror.NewInternal("generate webhook secret", err)
	}
	v, e := s.repo.RotateSecret(c, org, id, secret)
	return v, secret, readErr(e, "webhook not found")
}
func (s *Service) ListDeliveries(c context.Context, org, id uuid.UUID, limit, offset int32) ([]sqlc.WebhookDelivery, error) {
	if err := validIDs(org, id); err != nil {
		return nil, err
	}
	if limit < 1 || limit > 100 {
		return nil, apperror.NewBadRequest("limit must be between 1 and 100")
	}
	if offset < 0 {
		return nil, apperror.NewBadRequest("offset must not be negative")
	}
	return s.repo.ListDeliveries(c, org, id, limit, offset)
}
func (s *Service) GetDelivery(c context.Context, org, endpoint, id uuid.UUID) (sqlc.WebhookDelivery, error) {
	if err := validIDs(org, endpoint); err != nil {
		return sqlc.WebhookDelivery{}, err
	}
	v, e := s.repo.GetDelivery(c, org, id)
	if e = readErr(e, "webhook delivery not found"); e != nil {
		return sqlc.WebhookDelivery{}, e
	}
	if v.EndpointID != endpoint {
		return sqlc.WebhookDelivery{}, apperror.NewNotFound("webhook delivery not found")
	}
	return v, nil
}
func (s *Service) Retry(c context.Context, org, endpoint, id uuid.UUID) (sqlc.WebhookDelivery, error) {
	if _, err := s.GetDelivery(c, org, endpoint, id); err != nil {
		return sqlc.WebhookDelivery{}, err
	}
	v, e := s.repo.Retry(c, org, id)
	return v, readErr(e, "webhook delivery is not retryable")
}
func validOrg(v uuid.UUID) error {
	if v == uuid.Nil {
		return apperror.NewBadRequest("organization_id is required")
	}
	return nil
}
func validIDs(org, id uuid.UUID) error {
	if e := validOrg(org); e != nil {
		return e
	}
	if id == uuid.Nil {
		return apperror.NewBadRequest("webhook id is required")
	}
	return nil
}
func normalizeURL(v string) (string, error) {
	v = strings.TrimSpace(v)
	u, e := url.ParseRequestURI(v)
	if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return "", apperror.NewBadRequest("url must be an https URL")
	}
	return v, nil
}
func normalizeEvents(v []string) ([]string, error) {
	if len(v) == 0 {
		return nil, apperror.NewBadRequest("subscribed_events is required")
	}
	out := make([]string, 0, len(v))
	seen := map[string]struct{}{}
	for _, e := range v {
		e = strings.TrimSpace(e)
		if e == "" || strings.ContainsAny(e, " \t\r\n") {
			return nil, apperror.NewBadRequest("subscribed_events must contain non-empty event names")
		}
		if _, ok := seen[e]; !ok {
			seen[e] = struct{}{}
			out = append(out, e)
		}
	}
	return out, nil
}
func normalizeCreate(r *CreateRequest) error {
	var e error
	r.URL, e = normalizeURL(r.URL)
	if e != nil {
		return e
	}
	r.SubscribedEvents, e = normalizeEvents(r.SubscribedEvents)
	return e
}
func normalizeUpdate(r *UpdateRequest) error {
	var e error
	if r.URL != nil {
		v := *r.URL
		v, e = normalizeURL(v)
		if e != nil {
			return e
		}
		r.URL = &v
	}
	if r.SubscribedEvents != nil {
		v, e := normalizeEvents(*r.SubscribedEvents)
		if e != nil {
			return e
		}
		r.SubscribedEvents = &v
	}
	return nil
}
func readErr(e error, msg string) error {
	if e == nil {
		return nil
	}
	if errors.Is(e, pgx.ErrNoRows) {
		return apperror.NewNotFound(msg)
	}
	return apperror.NewInternal(msg, e)
}
func writeErr(e error, msg string) error { return readErr(e, msg) }
func (s *Service) Test(c context.Context, org, id uuid.UUID) (int, error) {
	v, e := s.Get(c, org, id)
	if e != nil {
		return 0, e
	}
	if !v.Enabled {
		return 0, apperror.NewBadRequest("webhook is disabled")
	}
	status, e := sendTest(c, v)
	if e != nil {
		return 0, apperror.NewInternal("send webhook test", e)
	}
	return status, nil
}
