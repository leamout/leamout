package worker

import "testing"

func TestUsageNotificationThreshold(t *testing.T) {
	tests := []struct {
		name      string
		used      int64
		allowance int64
		want      int32
	}{
		{name: "below warning", used: 799, allowance: 1000, want: 0},
		{name: "at warning", used: 800, allowance: 1000, want: 80},
		{name: "above warning", used: 999, allowance: 1000, want: 80},
		{name: "at exhaustion", used: 1000, allowance: 1000, want: 100},
		{name: "over exhaustion", used: 1200, allowance: 1000, want: 100},
		{name: "unlimited or invalid allowance", used: 900, allowance: 0, want: 0},
		{name: "invalid usage", used: -1, allowance: 1000, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := usageNotificationThreshold(test.used, test.allowance); got != test.want {
				t.Fatalf("usageNotificationThreshold(%d, %d) = %d, want %d", test.used, test.allowance, got, test.want)
			}
		})
	}
}
