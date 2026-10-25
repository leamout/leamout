package calls

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/runtime/calling"
)

func TestAdmissionFailureReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "cps", err: calling.ErrAdmissionCPS, want: "trunk_cps_limit"},
		{name: "concurrent", err: calling.ErrAdmissionConcurrent, want: "trunk_concurrent_limit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := admissionFailureReason(tt.err); got != tt.want {
				t.Fatalf("admissionFailureReason() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateExistingInbound(t *testing.T) {
	organizationID := uuid.New()
	voiceAgentID := uuid.New()
	trunkID := uuid.New()

	req := InboundAdmissionRequest{
		OrganizationID: organizationID,
		VoiceAgentID:   voiceAgentID,
		TrunkID:        trunkID,
		ToURI:          "+14155550100",
	}

	call := sqlc.Call{
		OrganizationID: organizationID,
		VoiceAgentID:   &voiceAgentID,
		TrunkID:        &trunkID,
		Direction:      string(DirectionInbound),
		State:          string(StateRinging),
		ToUri:          req.ToURI,
	}

	if err := validateExistingInbound(call, req); err != nil {
		t.Fatalf("validateExistingInbound() unexpected error: %v", err)
	}

	call.State = string(StateCompleted)
	if err := validateExistingInbound(call, req); err == nil {
		t.Fatal("expected terminal inbound call conflict")
	}
}

func TestValidateExistingInboundRejectsTrunkMismatch(t *testing.T) {
	organizationID := uuid.New()
	voiceAgentID := uuid.New()
	trunkID := uuid.New()
	otherTrunkID := uuid.New()

	req := InboundAdmissionRequest{
		OrganizationID: organizationID,
		VoiceAgentID:   voiceAgentID,
		TrunkID:        trunkID,
		ToURI:          "+14155550100",
	}

	call := sqlc.Call{
		OrganizationID: organizationID,
		VoiceAgentID:   &voiceAgentID,
		TrunkID:        &otherTrunkID,
		Direction:      string(DirectionInbound),
		State:          string(StateRinging),
		ToUri:          req.ToURI,
	}

	if err := validateExistingInbound(call, req); err == nil {
		t.Fatal("expected trunk attribution conflict")
	}
}

func TestRequireVoiceAgentReadyRejectsUnavailableGuard(t *testing.T) {
	t.Parallel()

	service := &Service{}
	err := service.requireVoiceAgentReady(
		context.Background(),
		uuid.New(),
		uuid.New(),
	)
	if err == nil {
		t.Fatal("requireVoiceAgentReady() error = nil")
	}
}

func TestRequireVoiceAgentReadyPropagatesReadinessFailure(t *testing.T) {
	t.Parallel()

	want := errors.New("voice agent is not ready")
	service := &Service{
		voiceAgentReadiness: func(
			context.Context,
			uuid.UUID,
			uuid.UUID,
		) error {
			return want
		},
	}
	err := service.requireVoiceAgentReady(
		context.Background(),
		uuid.New(),
		uuid.New(),
	)
	if !errors.Is(err, want) {
		t.Fatalf("requireVoiceAgentReady() error = %v, want %v", err, want)
	}
}

func TestRequireVoiceAgentReadyAcceptsReadyAgent(t *testing.T) {
	t.Parallel()

	service := &Service{
		voiceAgentReadiness: func(
			context.Context,
			uuid.UUID,
			uuid.UUID,
		) error {
			return nil
		},
	}
	if err := service.requireVoiceAgentReady(
		context.Background(),
		uuid.New(),
		uuid.New(),
	); err != nil {
		t.Fatalf("requireVoiceAgentReady() error = %v", err)
	}
}

func TestRequireCallVoiceAgentReadySkipsProgrammableCall(t *testing.T) {
	t.Parallel()

	service := &Service{}
	if err := service.requireCallVoiceAgentReady(
		context.Background(),
		sqlc.Call{OrganizationID: uuid.New()},
	); err != nil {
		t.Fatalf("requireCallVoiceAgentReady() error = %v", err)
	}
}

func TestRequireCallVoiceAgentReadyChecksBoundAgent(t *testing.T) {
	t.Parallel()

	organizationID := uuid.New()
	voiceAgentID := uuid.New()
	called := false
	service := &Service{
		voiceAgentReadiness: func(
			ctx context.Context,
			gotOrganizationID uuid.UUID,
			gotVoiceAgentID uuid.UUID,
		) error {
			called = true
			if gotOrganizationID != organizationID {
				t.Fatalf("organization id = %s, want %s", gotOrganizationID, organizationID)
			}
			if gotVoiceAgentID != voiceAgentID {
				t.Fatalf("voice agent id = %s, want %s", gotVoiceAgentID, voiceAgentID)
			}
			return nil
		},
	}
	if err := service.requireCallVoiceAgentReady(
		context.Background(),
		sqlc.Call{
			OrganizationID: organizationID,
			VoiceAgentID:   &voiceAgentID,
		},
	); err != nil {
		t.Fatalf("requireCallVoiceAgentReady() error = %v", err)
	}
	if !called {
		t.Fatal("voice agent readiness check was not called")
	}
}

// Recording cannot begin without an explicit organization-owned BYOS
// destination. The check applies only to start, never to stop.
func TestRequireRecordingStorageRejectsUnavailableGuard(t *testing.T) {
	service := &Service{}
	if err := service.requireRecordingStorage(context.Background(), uuid.New()); err == nil {
		t.Fatal("unconfigured recording admission check accepted")
	}
}

func TestRequireRecordingStoragePropagatesFailure(t *testing.T) {
	want := errors.New("customer BYOS storage is not configured")
	service := &Service{
		recordingStorage: func(context.Context, uuid.UUID) error {
			return want
		},
	}
	if err := service.requireRecordingStorage(context.Background(), uuid.New()); !errors.Is(err, want) {
		t.Fatalf("recording storage check error = %v, want %v", err, want)
	}
}

func TestRequireRecordingStorageAcceptsConfiguredIntegration(t *testing.T) {
	organizationID := uuid.New()
	service := &Service{
		recordingStorage: func(_ context.Context, got uuid.UUID) error {
			if got != organizationID {
				t.Fatalf("organization ID = %s, want %s", got, organizationID)
			}
			return nil
		},
	}
	if err := service.requireRecordingStorage(context.Background(), organizationID); err != nil {
		t.Fatalf("recording storage check error = %v", err)
	}
}

func TestLifecycleAlreadyApplied(t *testing.T) {
	tests := []struct {
		name      string
		snapshot  LifecycleSnapshot
		eventType LifecycleEventType
		want      bool
	}{
		{
			name:      "initiated is idempotent",
			snapshot:  LifecycleSnapshot{State: string(StateInitiating)},
			eventType: LifecycleInitiated,
			want:      true,
		},
		{
			name:      "answered already applied",
			snapshot:  LifecycleSnapshot{State: string(StateAnswered)},
			eventType: LifecycleAnswered,
			want:      true,
		},
		{
			name:      "answered not yet applied",
			snapshot:  LifecycleSnapshot{State: string(StateRinging)},
			eventType: LifecycleAnswered,
			want:      false,
		},
		{
			name: "held already applied",
			snapshot: LifecycleSnapshot{
				State:      string(StateActive),
				MediaState: string(MediaStateHeld),
			},
			eventType: LifecycleHeld,
			want:      true,
		},
		{
			name:      "terminal ignores later answer",
			snapshot:  LifecycleSnapshot{State: string(StateCompleted)},
			eventType: LifecycleAnswered,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lifecycleAlreadyApplied(tt.snapshot, tt.eventType); got != tt.want {
				t.Fatalf("lifecycleAlreadyApplied() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerminalLifecycleAndState(t *testing.T) {
	for _, eventType := range []LifecycleEventType{
		LifecycleCompleted,
		LifecycleFailed,
		LifecycleCancelled,
	} {
		if !isTerminalLifecycle(eventType) {
			t.Fatalf("%q should be terminal lifecycle", eventType)
		}
	}

	for _, state := range []State{
		StateCompleted,
		StateFailed,
		StateCancelled,
	} {
		if !isTerminalState(string(state)) {
			t.Fatalf("%q should be terminal state", state)
		}
	}

	if isTerminalLifecycle(LifecycleAnswered) {
		t.Fatal("answered must not be terminal lifecycle")
	}
	if isTerminalState(string(StateActive)) {
		t.Fatal("active must not be terminal state")
	}
}
