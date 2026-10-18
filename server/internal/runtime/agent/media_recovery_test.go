package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestControlURLForSessionUsesOriginalNode(t *testing.T) {
	id := uuid.New()
	for _, tc := range []struct{ base, expected string }{
		{"http://node.internal:8080", "ws://node.internal:8080/internal/v1/sessions/" + id.String() + "/control"},
		{"https://media.example", "wss://media.example/internal/v1/sessions/" + id.String() + "/control"},
	} {
		got, err := controlURLForSession(tc.base, id)
		if err != nil || got != tc.expected {
			t.Fatalf("url for %q = %q, %v", tc.base, got, err)
		}
	}
	if _, err := controlURLForSession("ftp://node.internal", id); err == nil {
		t.Fatal("unsupported media control transport accepted")
	}
}

func TestMediaSessionStatusRejectsMissingMedia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := &mediaClient{
		token:  "secret",
		client: server.Client(),
	}
	_, err := client.GetSessionAt(t.Context(), server.URL, uuid.New())
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("missing media session = %v, want HTTP 404", err)
	}
}
