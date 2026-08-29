package freeswitch

import "testing"

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := freeSWITCHStatusUp(tt.reply); got != tt.want {
				t.Fatalf("freeSWITCHStatusUp() = %v, want %v", got, tt.want)
			}
		})
	}
}
