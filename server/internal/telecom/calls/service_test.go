package calls

import "testing"

func TestCanAnswer(t *testing.T) {
	tests := []struct {
		state string
		want  bool
	}{
		{state: "initiating", want: true},
		{state: "ringing", want: true},
		{state: "answered", want: false},
		{state: "active", want: false},
		{state: "completed", want: false},
		{state: "failed", want: false},
		{state: "cancelled", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if got := canAnswer(tt.state); got != tt.want {
				t.Fatalf("canAnswer(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestIsAnswerIdempotent(t *testing.T) {
	tests := []struct {
		state string
		want  bool
	}{
		{state: "initiating", want: false},
		{state: "ringing", want: false},
		{state: "answered", want: true},
		{state: "active", want: true},
		{state: "completed", want: false},
		{state: "failed", want: false},
		{state: "cancelled", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if got := isAnswerIdempotent(tt.state); got != tt.want {
				t.Fatalf("isAnswerIdempotent(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestIsTerminal(t *testing.T) {
	tests := []struct {
		state string
		want  bool
	}{
		{state: "initiating", want: false},
		{state: "ringing", want: false},
		{state: "answered", want: false},
		{state: "active", want: false},
		{state: "completed", want: true},
		{state: "failed", want: true},
		{state: "cancelled", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if got := isTerminal(tt.state); got != tt.want {
				t.Fatalf("isTerminal(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}
