package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireOperatorKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, configured, authorization string
		wantStatus                      int
	}{
		{name: "valid", configured: "operator-secret", authorization: "Bearer operator-secret", wantStatus: http.StatusNoContent},
		{name: "missing", configured: "operator-secret", wantStatus: http.StatusUnauthorized},
		{name: "incorrect", configured: "operator-secret", authorization: "Bearer other", wantStatus: http.StatusUnauthorized},
		{name: "requires bearer scheme", configured: "operator-secret", authorization: "operator-secret", wantStatus: http.StatusUnauthorized},
		{name: "disabled when unconfigured", authorization: "Bearer operator-secret", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequireOperatorKey(tt.configured)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			req := httptest.NewRequest(http.MethodGet, "/internal/v1/commercial/products", nil)
			req.Header.Set("Authorization", tt.authorization)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tt.wantStatus)
			}
		})
	}
}
