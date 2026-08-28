package numbers

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

type Service struct{ repo *Repository }

func NewService(r *Repository) *Service { return &Service{r} }
func (s *Service) Create(c context.Context, org uuid.UUID, req CreateRequest) (sqlc.PhoneNumber, error) {
	if e := validateOrganizationID(org); e != nil {
		return sqlc.PhoneNumber{}, e
	}
	var e error
	req.Number, e = normalizeNumber(req.Number)
	if e != nil {
		return sqlc.PhoneNumber{}, e
	}
	req.CountryCode, e = normalizeCountryCode(req.CountryCode)
	if e != nil {
		return sqlc.PhoneNumber{}, e
	}
	v, e := s.repo.Create(c, org, req)
	return v, writeError(e, "create phone number")
}
func (s *Service) List(c context.Context, org uuid.UUID) ([]sqlc.PhoneNumber, error) {
	if e := validateOrganizationID(org); e != nil {
		return nil, e
	}
	v, e := s.repo.List(c, org)
	if e != nil {
		return nil, apperror.NewInternal("list phone numbers", e)
	}
	return v, nil
}
func (s *Service) Get(c context.Context, org, id uuid.UUID) (sqlc.PhoneNumber, error) {
	if e := validIDs(org, id); e != nil {
		return sqlc.PhoneNumber{}, e
	}
	v, e := s.repo.Get(c, org, id)
	return v, readError(e, "phone number not found")
}
func (s *Service) Update(c context.Context, org, id uuid.UUID, req UpdateRequest) (sqlc.PhoneNumber, error) {
	if e := validIDs(org, id); e != nil {
		return sqlc.PhoneNumber{}, e
	}
	if req.CountryCode == nil && req.VoiceEnabled == nil && req.SMSEnabled == nil {
		return sqlc.PhoneNumber{}, apperror.NewBadRequest("at least one field is required")
	}
	if req.CountryCode != nil {
		v, e := normalizeCountryCode(*req.CountryCode)
		if e != nil {
			return sqlc.PhoneNumber{}, e
		}
		req.CountryCode = &v
	}
	v, e := s.repo.Update(c, org, id, req)
	return v, writeError(e, "phone number not found")
}
func (s *Service) Delete(c context.Context, org, id uuid.UUID) error {
	if _, e := s.Get(c, org, id); e != nil {
		return e
	}
	return writeError(s.repo.Disable(c, org, id), "disable phone number")
}
func validIDs(org, id uuid.UUID) error {
	if e := validateOrganizationID(org); e != nil {
		return e
	}
	return validateNumberID(id)
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
		return apperror.NewConflict("phone number already exists")
	}
	return readError(e, msg)
}
