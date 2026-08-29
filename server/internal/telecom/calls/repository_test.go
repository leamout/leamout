package calls

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

func TestCallDomainEventUsesLeamoutAndMediaIDs(t *testing.T) {
	organizationID := uuid.New()
	applicationID := uuid.New()
	callID := uuid.New()
	sipCallID := "freeswitch-channel-123"
	occurredAt := time.Date(2026, time.August, 29, 5, 0, 0, 0, time.UTC)

	event := callDomainEvent(sqlc.Call{
		ID:             callID,
		OrganizationID: organizationID,
		ApplicationID:  &applicationID,
		Direction:      string(DirectionInbound),
		State:          "answered",
		FromUri:        "+233201234567",
		ToUri:          "+233301234567",
		SipCallID:      &sipCallID,
	}, EventCallAnswered, StatusAnswered, occurredAt)

	if event.CallID != callID.String() {
		t.Fatalf("call id = %q, want %q", event.CallID, callID.String())
	}
	if event.SIPCallID != sipCallID {
		t.Fatalf("sip call id = %q, want %q", event.SIPCallID, sipCallID)
	}
	if event.OrganizationID != organizationID.String() {
		t.Fatalf("organization id = %q, want %q", event.OrganizationID, organizationID.String())
	}
	if event.ApplicationID != applicationID.String() {
		t.Fatalf("application id = %q, want %q", event.ApplicationID, applicationID.String())
	}
	if event.EventType != EventCallAnswered || event.Status != StatusAnswered {
		t.Fatalf("event = %q/%q, want %q/%q", event.EventType, event.Status, EventCallAnswered, StatusAnswered)
	}
	if !event.OccurredAt.Equal(occurredAt) {
		t.Fatalf("occurred at = %s, want %s", event.OccurredAt, occurredAt)
	}
}
