package commpeak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) ListTerminationCDRs(ctx context.Context, request CDRRequest) (CDRPage, error) {
	query, err := cdrQuery(request, CDRDirectionTermination)
	if err != nil {
		return CDRPage{}, err
	}
	payload, err := c.get(ctx, "/call_records/termination", query)
	if err != nil {
		return CDRPage{}, err
	}
	if !json.Valid(payload) {
		return CDRPage{}, fmt.Errorf("commpeak: termination CDR response is not valid JSON")
	}
	return CDRPage{Direction: CDRDirectionTermination, Raw: json.RawMessage(payload)}, nil
}

func (c *Client) ListOriginationCDRs(ctx context.Context, request CDRRequest) (CDRPage, error) {
	query, err := cdrQuery(request, CDRDirectionOrigination)
	if err != nil {
		return CDRPage{}, err
	}
	payload, err := c.get(ctx, "/call_records/origination", query)
	if err != nil {
		return CDRPage{}, err
	}
	if !json.Valid(payload) {
		return CDRPage{}, fmt.Errorf("commpeak: origination CDR response is not valid JSON")
	}
	return CDRPage{Direction: CDRDirectionOrigination, Raw: json.RawMessage(payload)}, nil
}

func cdrQuery(request CDRRequest, direction CDRDirection) (url.Values, error) {
	if request.Page < 0 {
		return nil, fmt.Errorf("commpeak: page must be positive")
	}
	if request.PerPage < 0 || request.PerPage > 1000 {
		return nil, fmt.Errorf("commpeak: per_page must be between 1 and 1000")
	}

	query := url.Values{}
	setQuery(query, "time_range", request.TimeRange)
	if request.Page > 0 {
		query.Set("page", strconv.Itoa(request.Page))
	}
	if request.PerPage > 0 {
		query.Set("per_page", strconv.Itoa(request.PerPage))
	}

	switch direction {
	case CDRDirectionTermination:
		setQuery(query, "tech_prefix", request.TechPrefix)
		setQuery(query, "cli", request.CLI)
		setQuery(query, "destination", request.Destination)
		setQuery(query, "sip_account_id", request.SIPAccountID)
	case CDRDirectionOrigination:
		for _, did := range request.DIDs {
			if did = strings.TrimSpace(did); did != "" {
				query.Add("did[]", did)
			}
		}
	default:
		return nil, fmt.Errorf("commpeak: unsupported CDR direction %q", direction)
	}
	return query, nil
}

func setQuery(query url.Values, key, value string) {
	if value = strings.TrimSpace(value); value != "" {
		query.Set(key, value)
	}
}
