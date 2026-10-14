package freeswitchexporter

import "testing"

func TestParseCount(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  uint64
		valid bool
	}{
		{"0 total.", 0, true}, {"42 total.", 42, true}, {"garbage", 0, false}, {"-1 total.", 0, false},
	} {
		got, err := parseCount(tc.input)
		if (err == nil) != tc.valid || (tc.valid && got != tc.want) {
			t.Errorf("parseCount(%q): %d %v", tc.input, got, err)
		}
	}
}
