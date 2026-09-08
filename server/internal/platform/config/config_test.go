package config

import "testing"

func TestNormalizeCommPeakConfig(t *testing.T) {
	cfg := Config{
		CommPeak: CommPeakConfig{
			Authorization: "  Bearer test-token  ",
			APIBaseURL:    "  https://api.commpeak.com/  ",
		},
	}

	cfg.normalize()

	if got, want := cfg.CommPeak.Authorization, "Bearer test-token"; got != want {
		t.Fatalf("CommPeak API authorization = %q, want %q", got, want)
	}
	if got, want := cfg.CommPeak.APIBaseURL, "https://api.commpeak.com"; got != want {
		t.Fatalf("CommPeak API base URL = %q, want %q", got, want)
	}
}
