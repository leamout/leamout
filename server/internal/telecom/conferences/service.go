package conferences

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	org uuid.UUID,
	req CreateRequest,
) (sqlc.Conference, error) {
	if err := validateOrganizationID(org); err != nil {
		return sqlc.Conference{}, err
	}
	if err := validateCreateRequest(&req); err != nil {
		return sqlc.Conference{}, err
	}

	value, err := s.repo.Create(ctx, org, req)
	return value, conferenceWriteError(err, "create conference")
}

func (s *Service) Get(
	ctx context.Context,
	org uuid.UUID,
	id uuid.UUID,
) (sqlc.Conference, error) {
	if err := validateOrganizationID(org); err != nil {
		return sqlc.Conference{}, err
	}

	value, err := s.repo.Get(ctx, org, id)
	return value, conferenceReadError(err, "conference not found")
}

func (s *Service) List(
	ctx context.Context,
	org uuid.UUID,
	offset int32,
	limit int32,
) ([]sqlc.Conference, error) {
	if err := validateOrganizationID(org); err != nil {
		return nil, err
	}
	if offset < 0 {
		return nil, apperror.NewBadRequest("offset cannot be negative")
	}
	if limit < 1 || limit > 100 {
		return nil, apperror.NewBadRequest("limit must be between 1 and 100")
	}

	values, err := s.repo.List(ctx, org, offset, limit)
	if err != nil {
		return nil, apperror.NewInternal("list conferences", err)
	}

	return values, nil
}

func (s *Service) End(
	ctx context.Context,
	org uuid.UUID,
	id uuid.UUID,
) (sqlc.Conference, error) {
	if _, err := s.Get(ctx, org, id); err != nil {
		return sqlc.Conference{}, err
	}

	value, err := s.repo.SetState(ctx, org, id, "ended")
	return value, conferenceWriteError(err, "end conference")
}

func (s *Service) AddParticipant(
	ctx context.Context,
	org uuid.UUID,
	conferenceID uuid.UUID,
	req AddParticipantRequest,
) (sqlc.ConferenceParticipant, error) {
	conference, err := s.Get(ctx, org, conferenceID)
	if err != nil {
		return sqlc.ConferenceParticipant{}, err
	}
	if conference.State != "active" {
		return sqlc.ConferenceParticipant{}, apperror.NewConflict("conference has ended")
	}
	if err := validateParticipantRequest(&req); err != nil {
		return sqlc.ConferenceParticipant{}, err
	}

	value, err := s.repo.CreateParticipant(ctx, org, conferenceID, req)
	return value, conferenceWriteError(err, "add conference participant")
}

func (s *Service) ListParticipants(
	ctx context.Context,
	org uuid.UUID,
	conferenceID uuid.UUID,
) ([]sqlc.ConferenceParticipant, error) {
	if _, err := s.Get(ctx, org, conferenceID); err != nil {
		return nil, err
	}

	values, err := s.repo.ListParticipants(ctx, org, conferenceID)
	if err != nil {
		return nil, apperror.NewInternal("list conference participants", err)
	}

	return values, nil
}

func (s *Service) participant(
	ctx context.Context,
	org uuid.UUID,
	conferenceID uuid.UUID,
	id uuid.UUID,
) (sqlc.ConferenceParticipant, error) {
	if _, err := s.Get(ctx, org, conferenceID); err != nil {
		return sqlc.ConferenceParticipant{}, err
	}

	value, err := s.repo.GetParticipant(ctx, org, id)
	if err = conferenceReadError(err, "conference participant not found"); err != nil {
		return sqlc.ConferenceParticipant{}, err
	}
	if value.ConferenceID != conferenceID {
		return sqlc.ConferenceParticipant{}, apperror.NewNotFound("conference participant not found")
	}

	return value, nil
}

func (s *Service) SetParticipant(
	ctx context.Context,
	org uuid.UUID,
	conferenceID uuid.UUID,
	id uuid.UUID,
	state string,
	muted *bool,
	deaf *bool,
) (sqlc.ConferenceParticipant, error) {
	if _, err := s.participant(ctx, org, conferenceID, id); err != nil {
		return sqlc.ConferenceParticipant{}, err
	}

	value, err := s.repo.SetParticipant(ctx, org, id, state, muted, deaf)
	return value, conferenceWriteError(err, "update conference participant")
}

func conferenceReadError(err error, message string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NewNotFound(message)
	}
	if err != nil {
		return apperror.NewInternal("get conference", err)
	}

	return nil
}

func conferenceWriteError(err error, message string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NewNotFound("conference not found")
	}
	if err != nil {
		return apperror.NewInternal(message, err)
	}

	return nil
}
