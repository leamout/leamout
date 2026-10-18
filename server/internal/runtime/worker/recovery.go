package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/ai/conversations"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/integrations/freeswitch"
	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/runtime/agent"
)

const recoveryInterval = 15 * time.Second

type sessionRecoveryJob struct {
	pool       *pgxpool.Pool
	queries    *sqlc.Queries
	freeSwitch *freeswitch.Client
	runtime    *agent.Runtime
	logger     *logging.Logger
}

type activeRecoveryCandidate struct {
	sessionID      uuid.UUID
	organizationID uuid.UUID
	callID         uuid.UUID
	voiceAgentID   uuid.UUID
	state          string
}

// Reconcile attempts to reconnect only sessions which still have an original
// FreeSWITCH channel and media attachment. Missing infrastructure is retried;
// it never creates a duplicate fork or terminates a healthy call.
func (j *sessionRecoveryJob) Reconcile(ctx context.Context) error {
	if j == nil || j.pool == nil || j.queries == nil ||
		j.freeSwitch == nil || j.runtime == nil {
		return fmt.Errorf("active Voice Agent recovery is not initialized")
	}
	rows, err := j.pool.Query(ctx, `
SELECT session.id, session.organization_id, session.call_id,
       session.voice_agent_id, calls.state
FROM voice_agent_sessions AS session
JOIN calls ON calls.id = session.call_id
          AND calls.organization_id = session.organization_id
JOIN organizations AS org ON org.id = session.organization_id
WHERE session.state = 'active'
  AND calls.state IN ('answered', 'active')
  AND calls.ended_at IS NULL
  AND session.media_started_at IS NOT NULL
  AND org.status = 'active'
  AND org.deleted_at IS NULL
ORDER BY session.started_at ASC
LIMIT 500
`)
	if err != nil {
		return fmt.Errorf("list active Voice Agent sessions: %w", err)
	}
	var candidates []activeRecoveryCandidate
	for rows.Next() {
		var candidate activeRecoveryCandidate
		if err := rows.Scan(
			&candidate.sessionID,
			&candidate.organizationID,
			&candidate.callID,
			&candidate.voiceAgentID,
			&candidate.state,
		); err != nil {
			rows.Close()
			return fmt.Errorf("scan active Voice Agent recovery session: %w", err)
		}
		candidates = append(candidates, candidate)
	}
	readErr := rows.Err()
	rows.Close()
	if readErr != nil {
		return fmt.Errorf("read active Voice Agent recovery sessions: %w", readErr)
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
		channelID := attached[candidate.sessionID]
		if channelID == "" {
			continue
		}
		if err := j.recoverCandidate(ctx, candidate, channelID); err != nil {
			if j.logger != nil {
				j.logger.Warn(ctx, "retry active Voice Agent control recovery",
					"session_id", candidate.sessionID,
					"call_id", candidate.callID,
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
	candidate activeRecoveryCandidate,
	channelID string,
) error {
		record, err := j.queries.GetVoiceAgentSessionByID(ctx, sqlc.GetVoiceAgentSessionByIDParams{
			ID: candidate.sessionID, OrganizationID: candidate.organizationID,
		})
		if err != nil {
			return err
		}
		persisted, err := j.queries.ListVoiceAgentTurnsBySessionID(ctx, sqlc.ListVoiceAgentTurnsBySessionIDParams{
			OrganizationID: candidate.organizationID, SessionID: candidate.sessionID,
		})
		if err != nil {
			return err
		}
		turns := make([]conversations.TurnRecovery, 0, len(persisted))
		for _, turn := range persisted {
			turns = append(turns, conversations.TurnRecovery{Sequence: turn.Sequence})
		}
		agentID := candidate.voiceAgentID
		call := sqlc.Call{
			ID: candidate.callID, OrganizationID: candidate.organizationID,
			VoiceAgentID: &agentID, State: candidate.state,
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
