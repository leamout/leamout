package freeswitch

import (
	"net"
	"testing"
)

func TestFreeSWITCHStatusUp(t *testing.T) {
	tests := []struct {
		name  string
		reply Reply
		want  bool
	}{
		{
			name: "status body reports uptime",
			reply: Reply{
				Body: "UP 0 years, 0 days, 1 hour, 2 minutes, 3 seconds",
			},
			want: true,
		},
		{
			name: "reply text reports uptime",
			reply: Reply{
				Text: "+OK UP",
			},
			want: true,
		},
		{
			name: "missing uptime marker",
			reply: Reply{
				Body: "not running",
			},
			want: false,
		},
		{
			name: "incidental up substring is not healthy",
			reply: Reply{
				Body: "BACKUP",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := freeSWITCHStatusUp(tt.reply); got != tt.want {
				t.Fatalf("freeSWITCHStatusUp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandReplyError(t *testing.T) {
	if err := commandReplyError(Reply{Body: "+OK"}); err != nil {
		t.Fatalf("commandReplyError(+OK) = %v", err)
	}
	err := commandReplyError(Reply{Body: "-ERR Conference not found"})
	if err == nil || err.Error() != "FreeSWITCH command failed: -ERR Conference not found" {
		t.Fatalf("commandReplyError(-ERR) = %v", err)
	}
}

func TestWriteCommandRejectsLineBreaks(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	if _, err := writeCommand(client, "api status\napi reloadxml"); err == nil {
		t.Fatal("writeCommand() error = nil for a multi-line command")
	}
}
