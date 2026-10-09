// Package listquery parses optional organization list query parameters.
package listquery

import (
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/apperror"
)

// Parser records the first invalid parameter while parsing a typed request.
type Parser struct {
	Values url.Values
	Err    error
}

func (p *Parser) invalid(key, reason string) {
	if p.Err == nil {
		p.Err = apperror.NewBadRequest(key + " " + reason)
	}
}

func (p *Parser) Text(key string) *string {
	values, ok := p.Values[key]
	if !ok {
		return nil
	}
	if len(values) != 1 {
		p.invalid(key, "must be supplied once")
		return nil
	}
	value := strings.TrimSpace(values[0])
	if value == "" {
		p.invalid(key, "must not be empty")
		return nil
	}
	return &value
}

func (p *Parser) UUID(key string) *uuid.UUID {
	value := p.Text(key)
	if value == nil {
		return nil
	}
	id, err := uuid.Parse(*value)
	if err != nil || id == uuid.Nil {
		p.invalid(key, "must be a non-zero UUID")
		return nil
	}
	return &id
}

func (p *Parser) Bool(key string) *bool {
	value := p.Text(key)
	if value == nil {
		return nil
	}
	if *value != "true" && *value != "false" {
		p.invalid(key, "must be true or false")
		return nil
	}
	result := *value == "true"
	return &result
}

func (p *Parser) Time(key string) *time.Time {
	value := p.Text(key)
	if value == nil {
		return nil
	}
	result, err := time.Parse(time.RFC3339Nano, *value)
	if err != nil || result.IsZero() {
		p.invalid(key, "must be an RFC3339 timestamp")
		return nil
	}
	result = result.UTC()
	return &result
}

func Enum(key string, value *string, allowed ...string) error {
	if value == nil {
		return nil
	}
	for _, candidate := range allowed {
		if *value == candidate {
			return nil
		}
	}
	return apperror.NewBadRequest("invalid " + key)
}

func ID(key string, value *uuid.UUID) error {
	if value != nil && *value == uuid.Nil {
		return apperror.NewBadRequest("invalid " + key)
	}
	return nil
}

// Range validates a half-open interval [from, before).
func Range(from, before *time.Time) error {
	if from != nil && from.IsZero() || before != nil && before.IsZero() {
		return apperror.NewBadRequest("date bounds must not be zero")
	}
	if from != nil && before != nil && !from.Before(*before) {
		return apperror.NewBadRequest("date range start must precede end")
	}
	return nil
}
