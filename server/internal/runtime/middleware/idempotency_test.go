package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/security/authn"
)

func TestIdempotencyMiddlewarePassesThroughWithoutKey(t *testing.T) {
	handler := NewIdempotencyMiddleware(nil).Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/resource", nil))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", response.Code)
	}
}

func TestIdempotencyScopePrefersOrganization(t *testing.T) {
	organizationID := uuid.New()
	principal := authn.Principal{
		Subject:    authn.Subject{ID: uuid.New(), Type: authn.SubjectUser},
		Credential: authn.Credential{ID: uuid.New(), Type: authn.CredentialSession},
	}
	ctx := authn.WithPrincipal(context.Background(), principal)
	ctx = withOrganizationContext(ctx, organizationContext{ID: organizationID})
	request := httptest.NewRequestWithContext(ctx, http.MethodPost, "/resource", nil)

	scope, ok := idempotencyScope(request)
	if !ok || scope != "organization:"+organizationID.String() {
		t.Fatalf("unexpected scope %q", scope)
	}
}

func TestBufferedResponseWriterPreservesFirstStatus(t *testing.T) {
	response := newBufferedResponseWriter()
	response.WriteHeader(http.StatusCreated)
	response.WriteHeader(http.StatusNoContent)
	_, _ = response.Write([]byte("created"))
	if response.status != http.StatusCreated || response.body.String() != "created" {
		t.Fatalf("unexpected buffered response: status=%d body=%q", response.status, response.body.String())
	}
}

func TestReplayHeadersExcludeUnsafeHeaders(t *testing.T) {
	header := make(http.Header)
	header.Set("Location", "/resource/1")
	header.Set("Set-Cookie", "secret=value")
	header.Set("Content-Type", "application/json")

	replayed := replayHeaders(header)
	if http.Header(replayed).Get("Location") != "/resource/1" {
		t.Fatalf("expected Location header, got %#v", replayed)
	}
	if http.Header(replayed).Get("Set-Cookie") != "" || http.Header(replayed).Get("Content-Type") != "" {
		t.Fatalf("unsafe replay headers were retained: %#v", replayed)
	}
}
