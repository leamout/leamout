package agents

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestDetailEscapesCustomerContent(t *testing.T) {
	var body bytes.Buffer
	err := DetailPage(Detail{
		Agent: Agent{
			ID:               uuid.New(),
			Name:             "<script>alert(1)</script>",
			OrganizationID:   uuid.New(),
			OrganizationName: "<img src=x onerror=alert(1)>",
		},
		Instructions: "<script>prompt(1)</script>",
	}).Render(t.Context(), &body)
	if err != nil {
		t.Fatal(err)
	}
	html := body.String()
	for _, unsafe := range []string{"<script>alert", "<script>prompt", "<img src=x"} {
		if strings.Contains(html, unsafe) {
			t.Fatalf("unescaped customer content: %s", unsafe)
		}
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Fatal("instructions should render as escaped text")
	}
}
