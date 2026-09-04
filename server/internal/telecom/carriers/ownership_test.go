package carriers

import "testing"

func TestConnectionScopeValid(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name  string
		scope ConnectionScope
		valid bool
	}{
		{name: "organization", scope: ConnectionScopeOrganization, valid: true},
		{name: "platform", scope: ConnectionScopePlatform, valid: true},
		{name: "empty", scope: "", valid: false},
		{name: "unknown", scope: "unknown", valid: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.scope.Valid(); got != tt.valid {
				t.Fatalf("Valid() = %v, want %v", got, tt.valid)
			}
		})
	}
}
