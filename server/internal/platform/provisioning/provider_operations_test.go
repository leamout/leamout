package provisioning

import "testing"

func TestNormalizeProviderOperationState(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		allowEmpty bool
		want       string
		wantErr    bool
	}{
		{name: "empty filter", input: "", allowEmpty: true, want: ""},
		{name: "normalizes", input: " Provider_Accepted ", allowEmpty: true, want: "provider_accepted"},
		{name: "failed", input: "FAILED", allowEmpty: true, want: "failed"},
		{name: "rejects invalid", input: "retrying", allowEmpty: true, wantErr: true},
		{name: "rejects empty when required", input: "", allowEmpty: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeProviderOperationState(tt.input, tt.allowEmpty)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeProviderOperationState(%q) succeeded; want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("state = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeProviderOperationLimit(t *testing.T) {
	if got, err := normalizeProviderOperationLimit(0); err != nil || got != defaultProviderOperationDiagnosticLimit {
		t.Fatalf("default limit = %d, %v; want %d", got, err, defaultProviderOperationDiagnosticLimit)
	}
	if got, err := normalizeProviderOperationLimit(25); err != nil || got != 25 {
		t.Fatalf("explicit limit = %d, %v; want 25", got, err)
	}
	if _, err := normalizeProviderOperationLimit(-1); err == nil {
		t.Fatal("negative limit accepted")
	}
	if _, err := normalizeProviderOperationLimit(maxProviderOperationDiagnosticLimit + 1); err == nil {
		t.Fatal("oversized limit accepted")
	}
}
