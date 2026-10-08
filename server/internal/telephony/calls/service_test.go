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
