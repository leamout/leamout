package commpeak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
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

// PollCDRs exposes CommPeak's paginated CDR endpoints through the provider-neutral
// polling contract. CommPeak documents date-range filters at day granularity, so
// each cursor window is one UTC calendar day and current-day windows are replayed
// safely by the durable polling inbox.
func (c *Client) PollCDRs(ctx context.Context, direction string, windowDate time.Time, page, perPage int) (json.RawMessage, int, error) {
	date := windowDate.UTC().Format("2006-01-02")
	request := CDRRequest{
		TimeRange: date + " - " + date,
		Page:      page,
		PerPage:   perPage,
	}

	var (
		result CDRPage
		err    error
	)
	switch CDRDirection(strings.ToLower(strings.TrimSpace(direction))) {
	case CDRDirectionTermination:
		result, err = c.ListTerminationCDRs(ctx, request)
	case CDRDirectionOrigination:
		result, err = c.ListOriginationCDRs(ctx, request)
	default:
		return nil, 0, fmt.Errorf("commpeak: unsupported CDR direction %q", direction)
	}
	if err != nil {
		return nil, 0, err
	}
	count, err := cdrRecordCount(result.Raw)
	if err != nil {
		return nil, 0, err
	}
	return result.Raw, count, nil
}

func cdrRecordCount(payload json.RawMessage) (int, error) {
	var records []json.RawMessage
	if err := json.Unmarshal(payload, &records); err == nil {
		return len(records), nil
	}
	var envelope struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return 0, fmt.Errorf("commpeak: unsupported CDR response shape: %w", err)
	}
	if envelope.Data == nil {
		return 0, fmt.Errorf("commpeak: CDR response does not contain a data array")
	}
	return len(envelope.Data), nil
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
