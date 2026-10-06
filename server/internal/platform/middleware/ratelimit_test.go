package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func TestAuthRateLimitUsesSocketPeer(t *testing.T) {
	m, err := NewRateLimitMiddleware(memory.NewStore())
	if err != nil {
		t.Fatal(err)
	}
	handler := m.HandleAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for i := range 31 {
		r := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/otp/send", nil)
		r.RemoteAddr = "192.0.2.1:1234"
		r.Header.Set("X-Forwarded-For", "203.0.113.1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		expected := http.StatusNoContent
		if i == 30 {
			expected = http.StatusTooManyRequests
		}
		if w.Code != expected {
			t.Fatalf("request %d: status %d", i, w.Code)
		}
	}
	r := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/auth/otp/send", nil)
	r.RemoteAddr = "192.0.2.1:5678"
	r.Header.Set("X-Forwarded-For", "203.0.113.2")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") == "" {
		t.Fatal("peer budget bypassed")
	}
	r.RemoteAddr = "192.0.2.2:1234"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatal("independent peer rejected")
	}
	r = httptest.NewRequestWithContext(context.Background(), "GET", "/v1/users/me", nil)
	r.RemoteAddr = "192.0.2.1:1234"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatal("non-auth endpoint limited by auth budget")
	}
}
