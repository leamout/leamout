package agents

import (
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestFilterValidation(t *testing.T) {
	for _, query := range []string{"page=0", "page=-1", "page=1000001", "page=2147483648", "status=deleted", "engine=unknown", "organization_id=invalid", "q=" + strings.Repeat("a", 201)} {
		values, err := url.ParseQuery(query)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parseFilter(values); err == nil {
			t.Errorf("accepted invalid filter %q", query)
		}
	}
	organizationID := uuid.NewString()
	filter, err := parseFilter(url.Values{"q": {"  A&B  "}, "status": {"disabled"}, "engine": {"realtime"}, "organization_id": {organizationID}, "page": {"2"}})
	if err != nil {
		t.Fatal(err)
	}
	next, err := url.Parse(pageURL(filter, 3))
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"q": "A&B", "status": "disabled", "engine": "realtime", "organization_id": organizationID, "page": "3"} {
		if got := next.Query().Get(key); got != want {
			t.Errorf("pagination %s = %q, want %q", key, got, want)
		}
	}
}
