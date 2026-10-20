package organizations

import (
	"net/url"
	"testing"
)

func TestFilterValidation(t *testing.T) {
	for _, query := range []string{"page=0", "page=-1", "page=2147483648", "page=1000001", "status=unknown"} {
		values, err := url.ParseQuery(query)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parseFilter(values); err == nil {
			t.Errorf("accepted invalid filter %q", query)
		}
	}
	filter, err := parseFilter(url.Values{"q": {"  A&B  "}, "status": {"deleted"}, "page": {"2"}})
	if err != nil {
		t.Fatal(err)
	}
	next, err := url.Parse(pageURL(filter, 3))
	if err != nil {
		t.Fatal(err)
	}
	if next.Query().Get("q") != "A&B" || next.Query().Get("status") != "deleted" || next.Query().Get("page") != "3" {
		t.Fatalf("pagination lost filters: %s", next)
	}
}
