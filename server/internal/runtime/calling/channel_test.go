package calling

import (
	"github.com/google/uuid"
	"testing"
)

func TestChannelKey(t *testing.T) {
	callID := uuid.New()
	want := "telecom:calls:channel:" + callID.String()
	if got := channelKey(callID); got != want {
		t.Fatalf("channel key = %q, want %q", got, want)
	}
}
