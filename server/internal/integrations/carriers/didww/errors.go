package didww

import (
	"encoding/json"
	"fmt"
	"strings"
)

type APIErrorObject struct {
	Title  string         `json:"title"`
	Detail string         `json:"detail"`
	Code   string         `json:"code"`
	Status string         `json:"status"`
	Meta   map[string]any `json:"meta,omitempty"`
}

type APIError struct {
	StatusCode int
	Errors     []APIErrorObject
	Body       string
}

func (e *APIError) Error() string {
	for _, item := range e.Errors {
		if detail := strings.TrimSpace(item.Detail); detail != "" {
			return fmt.Sprintf("didww: API returned HTTP %d: %s", e.StatusCode, detail)
		}
		if title := strings.TrimSpace(item.Title); title != "" {
			return fmt.Sprintf("didww: API returned HTTP %d: %s", e.StatusCode, title)
		}
	}
	if body := strings.TrimSpace(e.Body); body != "" {
		return fmt.Sprintf("didww: API returned HTTP %d: %s", e.StatusCode, body)
	}
	return fmt.Sprintf("didww: API returned HTTP %d", e.StatusCode)
}

func newAPIError(statusCode int, payload []byte) error {
	errorResponse := struct {
		Errors []APIErrorObject `json:"errors"`
	}{}
	_ = json.Unmarshal(payload, &errorResponse)
	return &APIError{
		StatusCode: statusCode,
		Errors:     errorResponse.Errors,
		Body:       string(payload),
	}
}
