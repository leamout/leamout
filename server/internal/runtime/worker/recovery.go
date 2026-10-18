package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/ai/conversations"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/integrations/freeswitch"
	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/runtime/agent"
)

const recoveryInterval = 15 * time.Second

type sessionRecoveryJob struct {
	queries    *sqlc.Queries
	freeSwitch *freeswitch.Client
	runtime    *agent.Runtime
	logger     *logging.Logger
}

// Reconcile attempts to reconnect only sessions which still have an original
// FreeSWITCH channel and media attachment. Missing infrastructure is retried;
// it never creates a duplicate fork or terminates a healthy call.
func (j *sessionRecoveryJob) Reconcile(ctx context.Context) error {
	if j == nil || j.queries == nil ||
		j.freeSwitch == nil || j.runtime == nil {
		return fmt.Errorf("active Voice Agent recovery is not initialized")
	}
	candidates, err := j.queries.ListActiveVoiceAgentSessionsForRecovery(ctx)
	if err != nil {
		return fmt.Errorf("list active Voice Agent sessions: %w", err)
	}
	if len(candidates) == 0 {
		return nil
	}
	channels, err := j.freeSwitch.Channels(ctx)
	if err != nil {
		return fmt.Errorf("list live FreeSWITCH channels: %w", err)
	}
	live := make(map[string]bool, len(channels))
	for _, channel := range channels {
		live[channel.UUID] = true
	}
	// Scan each channel once, not once per durable session. The FreeSWITCH
	// session variable is authoritative for the existing audio fork.
	attached := make(map[uuid.UUID]string, len(live))
	for channelID := range live {
		marked, err := j.freeSwitch.GetVariable(ctx, channelID, "leamout_voice_agent_session_id")
		if err != nil {
			continue
		}
		if sessionID, err := uuid.Parse(strings.TrimSpace(marked)); err == nil {
			attached[sessionID] = channelID
		}
	}
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return err
		}
		channelID := attached[candidate.SessionID]
		if channelID == "" {
			continue
		}
		if err := j.recoverCandidate(ctx, candidate, channelID); err != nil {
			if j.logger != nil {
				j.logger.Warn(ctx, "retry active Voice Agent control recovery",
					"session_id", candidate.SessionID,
					"call_id", candidate.CallID,
					"error", err)
			} else {
				log.Printf("retry Voice Agent control recovery: %v", err)
			}
		}
	}
	return nil
}

func (j *sessionRecoveryJob) recoverCandidate(
	ctx context.Context,
	candidate sqlc.ListActiveVoiceAgentSessionsForRecoveryRow,
	channelID string,
) error {
	record, err := j.queries.GetVoiceAgentSessionByID(ctx, sqlc.GetVoiceAgentSessionByIDParams{
		ID:             candidate.SessionID,
		OrganizationID: candidate.OrganizationID,
	})
	if err != nil {
		return err
	}
	persisted, err := j.queries.ListVoiceAgentTurnsBySessionID(ctx, sqlc.ListVoiceAgentTurnsBySessionIDParams{
		OrganizationID: candidate.OrganizationID,
		SessionID:      candidate.SessionID,
	})
	if err != nil {
		return err
	}
	turns := make([]conversations.TurnRecovery, 0, len(persisted))
	for _, turn := range persisted {
		turns = append(turns, conversations.TurnRecovery{Sequence: turn.Sequence})
	}
	agentID := candidate.VoiceAgentID
	call := sqlc.Call{
		ID:             candidate.CallID,
		OrganizationID: candidate.OrganizationID,
		VoiceAgentID:   &agentID,
		State:          candidate.State,
	}
	return j.runtime.RecoverSession(ctx, record, call, channelID, turns)
}

func (j *sessionRecoveryJob) Run(ctx context.Context) error {
	ticker := time.NewTicker(recoveryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := j.Reconcile(ctx); err != nil && !errors.Is(err, context.Canceled) {
				if j.logger != nil {
					j.logger.Warn(ctx, "Voice Agent recovery pass failed", "error", err)
				}
			}
		}
	}
}
