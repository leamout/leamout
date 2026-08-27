package recordings

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

// Storage manages recording objects without exposing storage URLs in metadata.
type Storage interface {
	PlaybackURL(context.Context, sqlc.Recording) (string, time.Time, error)
	Delete(context.Context, sqlc.Recording) error
}

type Service struct {
	repo    *Repository
	storage Storage
}

func NewService(repo *Repository, storage Storage) *Service {
	return &Service{repo: repo, storage: storage}
}
func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Recording, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return sqlc.Recording{}, err
	}
	recording, err := s.repo.Get(ctx, organizationID, id)
	return recording, readError(err)
}
func (s *Service) List(ctx context.Context, organizationID uuid.UUID, offset, limit int32) ([]sqlc.Recording, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return nil, err
	}
	if err := validatePagination(offset, limit); err != nil {
		return nil, err
	}
	recordings, err := s.repo.List(ctx, organizationID, offset, limit)
	if err != nil {
		return nil, apperror.NewInternal("list recordings", err)
	}
	return recordings, nil
}
func (s *Service) Playback(ctx context.Context, organizationID, id uuid.UUID) (PlaybackResponse, error) {
	recording, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return PlaybackResponse{}, err
	}
	if recording.Status != "completed" {
		return PlaybackResponse{}, apperror.NewConflict("recording is not available for playback")
	}
	if s.storage == nil {
		return PlaybackResponse{}, apperror.NewServiceUnavailable("recording storage is unavailable", nil)
	}
	url, expiresAt, err := s.storage.PlaybackURL(ctx, recording)
	if err != nil {
		return PlaybackResponse{}, apperror.NewServiceUnavailable("create recording playback URL", err)
	}
	if url == "" || expiresAt.IsZero() {
		return PlaybackResponse{}, apperror.NewServiceUnavailable("recording storage returned an invalid playback URL", nil)
	}
	return PlaybackResponse{URL: url, ExpiresAt: expiresAt.UTC()}, nil
}
func (s *Service) Delete(ctx context.Context, organizationID, id uuid.UUID) error {
	recording, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return err
	}
	if s.storage == nil {
		return apperror.NewServiceUnavailable("recording storage is unavailable", nil)
	}
	if err := s.storage.Delete(ctx, recording); err != nil {
		return apperror.NewServiceUnavailable("delete recording object", err)
	}
	_, err = s.repo.Delete(ctx, organizationID, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NewNotFound("recording not found")
	}
	if err != nil {
		return apperror.NewInternal("delete recording", err)
	}
	return nil
}
func readError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NewNotFound("recording not found")
	}
	if err != nil {
		return apperror.NewInternal("get recording", err)
	}
	return nil
}
