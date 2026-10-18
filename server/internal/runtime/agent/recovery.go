package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/ai/conversations"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

// RecoverSession reconnects control to an already-running media stream.
// It does not issue a new media session or restart the FreeSWITCH audio fork.
func (r *Runtime) RecoverSession(
	ctx context.Context,
	record sqlc.VoiceAgentSession,
	call sqlc.Call,
	channelID string,
	turns []conversations.TurnRecovery,
) (bool, error) {
	if record.ID == uuid.Nil || record.State != "active" ||
		call.ID != record.CallID ||
		call.OrganizationID != record.OrganizationID ||
		call.VoiceAgentID == nil || *call.VoiceAgentID != record.VoiceAgentID ||
		(call.State != "answered" && call.State != "active") {
		return false, fmt.Errorf("active session and call identity are required for recovery")
	}
	channel, err := uuid.Parse(strings.TrimSpace(channelID))
	if err != nil {
		return false, fmt.Errorf("invalid recovery channel ID: %w", err)
	}
	r.mu.Lock()
	if r.controls[record.ID] != nil {
		r.mu.Unlock()
		return false, nil
	}
	r.mu.Unlock()

	node, ok := r.mediaNode(ctx, record.ID)
	if !ok {
		return false, fmt.Errorf("original media node unavailable for session %s", record.ID)
	}
	status, err := r.media.GetSessionAt(ctx, node.ControlURL, record.ID)
	if err != nil {
		return false, fmt.Errorf("inspect existing media session: %w", err)
	}
	if status.ID != record.ID || status.OrganizationID != record.OrganizationID ||
		status.CallID != record.CallID || status.ChannelID != channel || !status.Attached {
		return false, fmt.Errorf("existing media attachment does not match durable session")
	}
	controlURL, err := controlURLForSession(node.ControlURL, record.ID)
	if err != nil {
		return false, err
	}
	state := restoreConversationState(
		turns,
		record.TurnCount,
		record.InterruptionCount,
		record.FirstResponseLatencyMs,
		record.AvgTurnLatencyMs,
	)
	if err := r.registerControlWithState(ctx, call, record.ID, controlURL, state); err != nil {
		return false, fmt.Errorf("recover media control attachment: %w", err)
	}
	return true, nil
}
