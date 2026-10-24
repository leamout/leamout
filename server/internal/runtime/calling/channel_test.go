package calling

import (
	"testing"

	"github.com/google/uuid"
)

func TestChannelKey(t *testing.T) {
	callID := uuid.New()
	want := "telecom:calls:channel:" + callID.String()
	if got := channelKey(callID); got != want {
		t.Fatalf("channel key = %q, want %q", got, want)
	}
}
