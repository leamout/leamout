package commpeak

import (
	"fmt"
	"strings"
)

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	if body := strings.TrimSpace(e.Body); body != "" {
		return fmt.Sprintf("commpeak: API returned HTTP %d: %s", e.StatusCode, body)
	}
	return fmt.Sprintf("commpeak: API returned HTTP %d", e.StatusCode)
}

func newAPIError(statusCode int, payload []byte) error {
	return &APIError{StatusCode: statusCode, Body: string(payload)}
}
