package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetCookieUsesDedicatedSecureHostOnlyCookie(t *testing.T) {
	recorder := httptest.NewRecorder()
	SetCookie(recorder, "secret", time.Now().Add(time.Hour))

	response := recorder.Result()
	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != CookieName {
		t.Fatalf("cookie name = %q, want %q", cookie.Name, CookieName)
	}
	if cookie.Domain != "" {
		t.Fatalf("cookie domain = %q, want host-only cookie", cookie.Domain)
	}
	if cookie.Path != "/" {
		t.Fatalf("cookie path = %q, want /", cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Fatal("cookie must be HttpOnly")
	}
	if !cookie.Secure {
		t.Fatal("cookie must be Secure")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie SameSite = %v, want Strict", cookie.SameSite)
	}
	if cookie.MaxAge <= 0 {
		t.Fatalf("cookie MaxAge = %d, want positive", cookie.MaxAge)
	}
}

func TestClearCookieExpiresDedicatedCookie(t *testing.T) {
	recorder := httptest.NewRecorder()
	ClearCookie(recorder)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != CookieName {
		t.Fatalf("cookie name = %q, want %q", cookie.Name, CookieName)
	}
	if cookie.Value != "" {
		t.Fatalf("cookie value = %q, want empty", cookie.Value)
	}
	if cookie.MaxAge >= 0 {
		t.Fatalf("cookie MaxAge = %d, want negative", cookie.MaxAge)
	}
	if cookie.Domain != "" {
		t.Fatalf("cookie domain = %q, want host-only cookie", cookie.Domain)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("cleared cookie must preserve Backoffice security attributes")
	}
}
