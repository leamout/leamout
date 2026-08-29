package calls

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (s *Service) EnsureInbound(ctx context.Context, event InboundCallEvent) error {
	if _, err := s.repo.GetBySIPCallID(ctx, event.OrganizationID, event.ChannelID); err == nil {
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("find inbound call: %w", err)
	}

	if _, err := s.repo.CreateInbound(ctx, event); err == nil {
		return nil
	} else if _, getErr := s.repo.GetBySIPCallID(ctx, event.OrganizationID, event.ChannelID); getErr == nil {
		// A concurrent or replayed event may have won the unique-key race.
		return nil
	} else {
		return fmt.Errorf("create inbound call: %w", err)
	}
}

func (s *Service) MarkInboundAnswered(ctx context.Context, event InboundCallEvent) error {
	if err := s.EnsureInbound(ctx, event); err != nil {
		return err
	}

	call, err := s.repo.GetBySIPCallID(ctx, event.OrganizationID, event.ChannelID)
	if err != nil {
		return fmt.Errorf("get inbound call for answer: %w", err)
	}
	if call.State == "answered" || call.State == "active" || isTerminal(call.State) {
		return nil
	}

	if _, err := s.repo.MarkAnswered(ctx, event.OrganizationID, call.ID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("mark inbound call answered: %w", err)
	}
	return nil
}

func (s *Service) FinishInbound(ctx context.Context, event InboundCallEvent) error {
	if err := s.EnsureInbound(ctx, event); err != nil {
		return err
	}

	call, err := s.repo.GetBySIPCallID(ctx, event.OrganizationID, event.ChannelID)
	if err != nil {
		return fmt.Errorf("get inbound call for hangup: %w", err)
	}
	if isTerminal(call.State) {
		return nil
	}

	if event.WasAnswered && (call.State == "initiating" || call.State == "ringing") {
		if updated, answerErr := s.repo.MarkAnswered(ctx, event.OrganizationID, call.ID); answerErr == nil {
			call = updated
		} else if !errors.Is(answerErr, pgx.ErrNoRows) {
			return fmt.Errorf("reconcile inbound answer before hangup: %w", answerErr)
		}
	}

	var reason *string
	if event.HangupCause != "" {
		reason = &event.HangupCause
	}

	var finishErr error
	switch {
	case event.WasAnswered || call.State == "answered" || call.State == "active":
		_, finishErr = s.repo.MarkCompleted(ctx, event.OrganizationID, call.ID, reason)
	case cancelledHangupCause(event.HangupCause):
		_, finishErr = s.repo.MarkCancelled(ctx, event.OrganizationID, call.ID, reason)
	default:
		_, finishErr = s.repo.MarkFailed(ctx, event.OrganizationID, call.ID, reason)
	}
	if finishErr != nil && !errors.Is(finishErr, pgx.ErrNoRows) {
		return fmt.Errorf("finish inbound call: %w", finishErr)
	}
	return nil
}

func cancelledHangupCause(cause string) bool {
	switch strings.ToUpper(strings.TrimSpace(cause)) {
	case "NORMAL_CLEARING", "ORIGINATOR_CANCEL":
		return true
	default:
		return false
	}
}
