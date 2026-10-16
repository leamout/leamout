package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCallbackRejectsMissingOrMismatchedBrowserState(t *testing.T) {
	for _, test := range []struct {
		name   string
		state  string
		cookie string
	}{
		{name: "missing state", cookie: "browser-state"},
		{name: "missing cookie", state: "returned-state"},
		{name: "wrong browser", state: "returned-state", cookie: "another-state"},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := &Handler{development: true}
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/auth/callback?code=code&state="+test.state, nil)
			if test.cookie != "" {
				request.AddCookie(&http.Cookie{Name: stateCookie, Value: test.cookie})
			}
			recorder := httptest.NewRecorder()
			handler.Callback(recorder, request)
			response := recorder.Result()
			defer func() { _ = response.Body.Close() }()
			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("status = %d", response.StatusCode)
			}
			for _, cookie := range response.Cookies() {
				if cookie.Name == "leamout-session" {
					t.Fatal("invalid callback issued a session")
				}
			}
		})
	}
}
