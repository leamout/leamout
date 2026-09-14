package leamout

import "testing"

func TestParseDrainCount(t *testing.T) {
	for _, test := range []struct {
		output    string
		want      int
		wantError bool
	}{
		{output: "0\n", want: 0},
		{output: " 12 \n", want: 12},
		{output: "unknown\n", wantError: true},
	} {
		got, err := parseDrainCount([]byte(test.output))
		if (err != nil) != test.wantError || (!test.wantError && got != test.want) {
			t.Errorf("parseDrainCount(%q) = %d, %v; want %d, error=%v", test.output, got, err, test.want, test.wantError)
		}
	}
}
