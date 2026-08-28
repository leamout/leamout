package subscribers

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/hasher"
)

type Service struct{ repo *Repository }

func NewService(r *Repository) *Service { return &Service{r} }
func (s *Service) Create(c context.Context, org uuid.UUID, req CreateRequest) (sqlc.Subscriber, error) {
	if e := validOrg(org); e != nil {
		return sqlc.Subscriber{}, e
	}
	if req.SIPDomainID == uuid.Nil {
		return sqlc.Subscriber{}, apperror.NewBadRequest("sip_domain_id is required")
	}
	var e error
	req.Username, e = normalizeUsername(req.Username)
	if e != nil {
		return sqlc.Subscriber{}, e
	}
	if e = validatePassword(req.Password); e != nil {
		return sqlc.Subscriber{}, e
	}
	req.DisplayName, e = normalizeDisplayName(req.DisplayName)
	if e != nil {
		return sqlc.Subscriber{}, e
	}
	domain, e := s.repo.Domain(c, org, req.SIPDomainID)
	if e != nil {
		return sqlc.Subscriber{}, readError(e, "SIP domain not found")
	}
	return s.repo.Create(c, org, req, hasher.ComputeHA1(req.Username, domain.Domain, req.Password))
}
func (s *Service) List(c context.Context, org uuid.UUID) ([]sqlc.Subscriber, error) {
	if e := validOrg(org); e != nil {
		return nil, e
	}
	v, e := s.repo.List(c, org)
	if e != nil {
		return nil, apperror.NewInternal("list subscribers", e)
	}
	return v, nil
}
func (s *Service) Get(c context.Context, org, id uuid.UUID) (sqlc.Subscriber, error) {
	if e := validIDs(org, id); e != nil {
		return sqlc.Subscriber{}, e
	}
	v, e := s.repo.Get(c, org, id)
	return v, readError(e, "subscriber not found")
}
func (s *Service) Update(c context.Context, org, id uuid.UUID, req UpdateRequest) (sqlc.Subscriber, error) {
	if e := validIDs(org, id); e != nil {
		return sqlc.Subscriber{}, e
	}
	if req.DisplayName == nil {
		return sqlc.Subscriber{}, apperror.NewBadRequest("at least one field is required")
	}
	name, e := normalizeDisplayName(req.DisplayName)
	if e != nil {
		return sqlc.Subscriber{}, e
	}
	v, e := s.repo.Update(c, org, id, name)
	return v, writeError(e, "subscriber not found")
}
func (s *Service) Rotate(c context.Context, org, id uuid.UUID, req RotateCredentialsRequest) (sqlc.Subscriber, error) {
	v, e := s.Get(c, org, id)
	if e != nil {
		return sqlc.Subscriber{}, e
	}
	if e = validatePassword(req.Password); e != nil {
		return sqlc.Subscriber{}, e
	}
	result, e := s.repo.Rotate(c, org, id, hasher.ComputeHA1(v.Username, v.Domain, req.Password))
	return result, writeError(e, "subscriber not found")
}
func (s *Service) Delete(c context.Context, org, id uuid.UUID) error {
	if _, e := s.Get(c, org, id); e != nil {
		return e
	}
	return writeError(s.repo.Disable(c, org, id), "disable subscriber")
}
func validIDs(org, id uuid.UUID) error {
	if e := validOrg(org); e != nil {
		return e
	}
	return validID(id)
}
func readError(e error, msg string) error {
	if e == nil {
		return nil
	}
	if errors.Is(e, pgx.ErrNoRows) {
		return apperror.NewNotFound(msg)
	}
	return apperror.NewInternal(msg, e)
}
func writeError(e error, msg string) error {
	if e == nil {
		return nil
	}
	var pg *pgconn.PgError
	if errors.As(e, &pg) && pg.Code == "23505" {
		return apperror.NewConflict("subscriber already exists")
	}
	return readError(e, msg)
}
