package calls

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

func parseFilter(values url.Values) (Filter, error) {
	filter := Filter{
		Search:         strings.TrimSpace(values.Get("q")),
		Page:           1,
		State:          strings.TrimSpace(values.Get("state")),
		Direction:      strings.TrimSpace(values.Get("direction")),
		OrganizationID: strings.TrimSpace(values.Get("organization_id")),
		AgentID:        strings.TrimSpace(values.Get("agent_id")),
		From:           strings.TrimSpace(values.Get("from")),
		To:             strings.TrimSpace(values.Get("to")),
	}
	if utf8.RuneCountInString(filter.Search) > 200 {
		return Filter{}, fmt.Errorf("search must be at most 200 characters")
	}
	switch filter.State {
	case "", "initiating", "ringing", "answered", "active", "completed", "failed", "cancelled":
	default:
		return Filter{}, fmt.Errorf("invalid state")
	}
	switch filter.Direction {
	case "", "inbound", "outbound":
	default:
		return Filter{}, fmt.Errorf("invalid direction")
	}
	if filter.OrganizationID != "" {
		id, err := uuid.Parse(filter.OrganizationID)
		if err != nil {
			return Filter{}, fmt.Errorf("invalid organization ID")
		}
		filter.OrganizationID = id.String()
	}
	if filter.AgentID != "" {
		id, err := uuid.Parse(filter.AgentID)
		if err != nil {
			return Filter{}, fmt.Errorf("invalid agent ID")
		}
		filter.AgentID = id.String()
	}
	for _, value := range []string{filter.From, filter.To} {
		if value == "" {
			continue
		}
		date, err := time.Parse(time.DateOnly, value)
		if err != nil || date.Year() < 1 || date.Year() > 9998 {
			return Filter{}, fmt.Errorf("dates must use YYYY-MM-DD between years 0001 and 9998")
		}
	}
	if filter.From != "" && filter.To != "" && filter.From > filter.To {
		return Filter{}, fmt.Errorf("from date must not follow to date")
	}
	if value := values.Get("page"); value != "" {
		page, err := strconv.ParseInt(value, 10, 32)
		if err != nil || page < 1 || page > 1000000 {
			return Filter{}, fmt.Errorf("page must be between 1 and 1000000")
		}
		filter.Page = int32(page)
	}
	return filter, nil
}

func pageURL(filter Filter, page int32) string {
	values := url.Values{}
	values.Set("q", filter.Search)
	values.Set("state", filter.State)
	values.Set("direction", filter.Direction)
	values.Set("organization_id", filter.OrganizationID)
	values.Set("agent_id", filter.AgentID)
	values.Set("from", filter.From)
	values.Set("to", filter.To)
	values.Set("page", strconv.FormatInt(int64(page), 10))
	return "/calls?" + values.Encode()
}
