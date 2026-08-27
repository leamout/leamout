package voice

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/leamout/leamout/internal/integrations/freeswitch"
)

func TranslateEvent(fsEvent freeswitch.Event, appID, organizationID string) (*CallEvent, error) {
	callID := fsEvent.Header("Unique-ID")
	occurredAt := parseTimestamp(fsEvent.Header("Event-Date-Timestamp"))

	switch fsEvent.Name {
	case "CHANNEL_CREATE":
		return &CallEvent{
			EventType:      EventCallInitiated,
			CallID:         callID,
			ApplicationID:  appID,
			OrganizationID: organizationID,
			From:           fsEvent.Header("Caller-Caller-ID-Number"),
			To:             fsEvent.Header("Caller-Destination-Number"),
			Direction:      determineDirection(fsEvent),
			Status:         StatusInitiated,
			OccurredAt:     occurredAt,
		}, nil

	case "CHANNEL_ANSWER":
		return &CallEvent{
			EventType:      EventCallAnswered,
			CallID:         callID,
			ApplicationID:  appID,
			OrganizationID: organizationID,
			From:           fsEvent.Header("Caller-Caller-ID-Number"),
			To:             fsEvent.Header("Caller-Destination-Number"),
			Direction:      determineDirection(fsEvent),
			Status:         StatusAnswered,
			OccurredAt:     occurredAt,
		}, nil

	case "CHANNEL_HANGUP_COMPLETE":
		return &CallEvent{
			EventType:      EventCallCompleted,
			CallID:         callID,
			ApplicationID:  appID,
			OrganizationID: organizationID,
			From:           fsEvent.Header("Caller-Caller-ID-Number"),
			To:             fsEvent.Header("Caller-Destination-Number"),
			Direction:      determineDirection(fsEvent),
			Status:         mapHangupCause(fsEvent.Header("Hangup-Cause")),
			DurationSec:    int(parseInt(fsEvent.Header("billmsec")) / 1000),
			OccurredAt:     occurredAt,
		}, nil
	}

	return nil, fmt.Errorf("unhandled FreeSWITCH event: %s", fsEvent.Name)
}

func determineDirection(event freeswitch.Event) CallDirection {
	if strings.EqualFold(event.Header("Call-Direction"), "inbound") {
		return DirectionInbound
	}
	return DirectionOutbound
}

func mapHangupCause(cause string) CallStatus {
	switch strings.ToUpper(strings.TrimSpace(cause)) {
	case "NORMAL_CLEARING", "NORMAL_UNSPECIFIED":
		return StatusCompleted
	case "USER_BUSY", "CALL_REJECTED", "NORMAL_CIRCUIT_CONGESTION":
		return StatusBusy
	case "NO_ANSWER", "ALLOTTED_TIMEOUT", "RECOVERY_ON_TIMER_EXPIRE":
		return StatusNoAnswer
	default:
		return StatusFailed
	}
}

func parseTimestamp(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Now().UTC()
	}

	micros, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return time.UnixMicro(micros).UTC()
	}

	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.000000",
	} {
		if timestamp, err := time.Parse(layout, value); err == nil {
			return timestamp.UTC()
		}
	}

	return time.Now().UTC()
}

func parseInt(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
