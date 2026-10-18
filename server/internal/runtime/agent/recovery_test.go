package agent

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/ai/conversations"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

func TestRestoreConversationStateContinuesDurableSequence(t *testing.T) {
	turns := []conversations.TurnRecovery{
		{Sequence: 1}, {Sequence: 2}, {Sequence: 5},
	}
	first := int32(180)
	avg := int32(420)
	state := restoreConversationState(turns, 3, 2, &first, &avg)
	turn := state.toolTurn("lookup", "tool-call", `{"ok":true}`, false)
	if turn.Sequence != 6 {
		t.Fatalf("recovered turn sequence = %d, want 6", turn.Sequence)
	}
	summary := state.summary()
	if summary.TurnCount != 3 || summary.InterruptionCount != 2 {
		t.Fatalf("recovered summary = %+v", summary)
	}
	if summary.FirstResponseLatencyMS == nil || *summary.FirstResponseLatencyMS != first {
		t.Fatalf("first response latency = %v", summary.FirstResponseLatencyMS)
	}
	if summary.AverageTurnLatencyMS == nil || *summary.AverageTurnLatencyMS != avg {
		t.Fatalf("average turn latency = %v", summary.AverageTurnLatencyMS)
	}
}

func TestRecoverSessionRejectsMismatchedDurableIdentityBeforeConnecting(t *testing.T) {
	organizationID := uuid.New()
	callID := uuid.New()
	agentID := uuid.New()
	record := sqlc.VoiceAgentSession{
		ID: uuid.New(), OrganizationID: organizationID,
		CallID: callID, VoiceAgentID: agentID, State: "active",
	}
	call := sqlc.Call{
		ID: callID, OrganizationID: organizationID,
		VoiceAgentID: &agentID, State: "answered",
	}
	runtime := &Runtime{}
	for name, changed := range map[string]sqlc.Call{
		"other_call":  {ID: uuid.New(), OrganizationID: organizationID, VoiceAgentID: &agentID, State: "answered"},
		"other_agent": {ID: callID, OrganizationID: organizationID, VoiceAgentID: func() *uuid.UUID { id := uuid.New(); return &id }(), State: "answered"},
		"terminal":    {ID: callID, OrganizationID: organizationID, VoiceAgentID: &agentID, State: "completed"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := runtime.RecoverSession(t.Context(), record, changed, uuid.NewString(), nil); err == nil {
				t.Fatal("mismatched durable recovery identity accepted")
			}
		})
	}
}
