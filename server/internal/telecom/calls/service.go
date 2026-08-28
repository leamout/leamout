package calls

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

// Controller is the media-server contract used by the call API.
type Controller interface {
	Originate(context.Context, CreateCallRequest) (string, error)
	Answer(context.Context, string) error
	Hangup(context.Context, string) error
	Transfer(context.Context, string, TransferRequest) error
	Hold(context.Context, string) error
	Unhold(context.Context, string) error
	Play(context.Context, string, string) error
	Stop(context.Context, string) error
	Record(context.Context, string, RecordRequest) error
	DTMF(context.Context, string, string) error
}

type Service struct {
	repo       *Repository
	controller Controller
}

func NewService(repo *Repository, controller Controller) *Service {
	if repo == nil {
		panic("calls: repository is required")
	}
	if controller == nil {
		panic("calls: controller is required")
	}

	return &Service{repo: repo, controller: controller}
}

func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, req CreateCallRequest) (sqlc.Call, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return sqlc.Call{}, err
	}
	if err := validateCreateRequest(&req); err != nil {
		return sqlc.Call{}, err
	}

	sipCallID, err := s.controller.Originate(ctx, req)
	if err != nil {
		return sqlc.Call{}, mediaError("originate call", err)
	}
	if sipCallID == "" {
		return sqlc.Call{}, apperror.NewServiceUnavailable("media server returned an empty call id", nil)
	}

	call, err := s.repo.Create(ctx, organizationID, req, sipCallID)
	if err == nil {
		return call, nil
	}

	if hangupErr := s.controller.Hangup(ctx, sipCallID); hangupErr != nil {
		return sqlc.Call{}, apperror.NewInternal(
			"create call and clean up media call",
			errors.Join(err, hangupErr),
		)
	}

	return sqlc.Call{}, writeError(err, "create call")
}

func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Call, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return sqlc.Call{}, err
	}
	return readCall(s.repo.Get(ctx, organizationID, id))
}

func (s *Service) List(ctx context.Context, organizationID uuid.UUID, state *string, offset, limit int32) ([]sqlc.Call, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return nil, err
	}
	if offset < 0 {
		return nil, apperror.NewBadRequest("offset cannot be negative")
	}
	if limit < 1 || limit > 100 {
		return nil, apperror.NewBadRequest("limit must be between 1 and 100")
	}
	return s.repo.List(ctx, organizationID, state, offset, limit)
}

func (s *Service) Answer(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Call, error) {
	call, externalID, err := s.controlCall(ctx, organizationID, id)
	if err != nil {
		return sqlc.Call{}, err
	}

	if isAnswerIdempotent(call.State) {
		return call, nil
	}
	if !canAnswer(call.State) {
		return sqlc.Call{}, apperror.NewConflict("call cannot be answered from state: " + call.State)
	}

	if err := s.controller.Answer(ctx, externalID); err != nil {
		return sqlc.Call{}, mediaError("answer call", err)
	}

	updated, err := s.repo.MarkAnswered(ctx, organizationID, call.ID)
	if err == nil {
		return updated, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return s.Get(ctx, organizationID, call.ID)
	}
	return sqlc.Call{}, writeError(err, "answer call")
}

func (s *Service) Hangup(ctx context.Context, organizationID, id uuid.UUID) (sqlc.Call, error) {
	call, externalID, err := s.controlCall(ctx, organizationID, id)
	if err != nil {
		return sqlc.Call{}, err
	}

	if isTerminal(call.State) {
		return call, nil
	}

	if err := s.controller.Hangup(ctx, externalID); err != nil {
		return sqlc.Call{}, mediaError("hang up call", err)
	}

	updated, err := s.repo.MarkCompleted(ctx, organizationID, call.ID, nil)
	if err == nil {
		return updated, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return s.Get(ctx, organizationID, call.ID)
	}
	return sqlc.Call{}, writeError(err, "hang up call")
}

func canAnswer(state string) bool {
	return state == "initiating" || state == "ringing"
}

func isAnswerIdempotent(state string) bool {
	return state == "answered" || state == "active"
}

func isTerminal(state string) bool {
	return state == "completed" || state == "failed" || state == "cancelled"
}

func (s *Service) Transfer(ctx context.Context, org, id uuid.UUID, req TransferRequest) error {
	_, externalID, err := s.controlCall(ctx, org, id)
	if err != nil {
		return err
	}
	if req.Destination, err = require(req.Destination, "destination"); err != nil {
		return err
	}
	return mediaError("transfer call", s.controller.Transfer(ctx, externalID, req))
}

func (s *Service) Hold(ctx context.Context, org, id uuid.UUID) error {
	_, externalID, err := s.controlCall(ctx, org, id)
	if err != nil {
		return err
	}
	return mediaError("hold call", s.controller.Hold(ctx, externalID))
}

func (s *Service) Unhold(ctx context.Context, org, id uuid.UUID) error {
	_, externalID, err := s.controlCall(ctx, org, id)
	if err != nil {
		return err
	}
	return mediaError("resume call", s.controller.Unhold(ctx, externalID))
}

func (s *Service) Play(ctx context.Context, org, id uuid.UUID, req PlayRequest) error {
	_, externalID, err := s.controlCall(ctx, org, id)
	if err != nil {
		return err
	}
	if req.Path, err = require(req.Path, "path"); err != nil {
		return err
	}
	return mediaError("play audio", s.controller.Play(ctx, externalID, req.Path))
}

func (s *Service) Stop(ctx context.Context, org, id uuid.UUID) error {
	_, externalID, err := s.controlCall(ctx, org, id)
	if err != nil {
		return err
	}
	return mediaError("stop audio", s.controller.Stop(ctx, externalID))
}

func (s *Service) Record(ctx context.Context, org, id uuid.UUID, req RecordRequest) error {
	_, externalID, err := s.controlCall(ctx, org, id)
	if err != nil {
		return err
	}
	if err := validateRecordRequest(&req); err != nil {
		return err
	}
	return mediaError("record call", s.controller.Record(ctx, externalID, req))
}

func (s *Service) DTMF(ctx context.Context, org, id uuid.UUID, req DTMFRequest) error {
	_, externalID, err := s.controlCall(ctx, org, id)
	if err != nil {
		return err
	}
	if req.Digits, err = require(req.Digits, "digits"); err != nil {
		return err
	}
	return mediaError("send DTMF", s.controller.DTMF(ctx, externalID, req.Digits))
}

func (s *Service) controlCall(ctx context.Context, org, id uuid.UUID) (sqlc.Call, string, error) {
	call, err := s.Get(ctx, org, id)
	if err != nil {
		return sqlc.Call{}, "", err
	}
	if call.SipCallID == nil || *call.SipCallID == "" {
		return sqlc.Call{}, "", apperror.NewConflict("call has no media-server id")
	}
	return call, *call.SipCallID, nil
}

func readCall(call sqlc.Call, err error) (sqlc.Call, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Call{}, apperror.NewNotFound("call not found")
	}
	if err != nil {
		return sqlc.Call{}, apperror.NewInternal("get call", err)
	}
	return call, nil
}

func writeError(err error, message string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NewNotFound("call not found")
	}
	if err != nil {
		return apperror.NewInternal(message, err)
	}
	return nil
}

func mediaError(message string, err error) error {
	if err != nil {
		return apperror.NewServiceUnavailable(message, err)
	}
	return nil
}
