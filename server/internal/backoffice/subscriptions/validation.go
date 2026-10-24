package subscriptions

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

func parseFilter(values url.Values) (Filter, error) {
	filter := Filter{
		Search:         strings.TrimSpace(values.Get("q")),
		Page:           1,
		Status:         strings.TrimSpace(values.Get("status")),
		Plan:           strings.TrimSpace(values.Get("plan")),
		Cancellation:   strings.TrimSpace(values.Get("cancellation")),
		OrganizationID: strings.TrimSpace(values.Get("organization_id")),
	}
	if utf8.RuneCountInString(filter.Search) > 200 {
		return Filter{}, fmt.Errorf("search must be at most 200 characters")
	}
	switch filter.Status {
	case "", "trialing", "active", "past_due", "canceled", "incomplete":
	default:
		return Filter{}, fmt.Errorf("invalid status")
	}
	switch filter.Cancellation {
	case "", "scheduled", "not_scheduled":
	default:
		return Filter{}, fmt.Errorf("invalid cancellation")
	}
	if filter.OrganizationID != "" {
		id, err := uuid.Parse(filter.OrganizationID)
		if err != nil {
			return Filter{}, fmt.Errorf("invalid organization ID")
		}
		filter.OrganizationID = id.String()
	}
	if utf8.RuneCountInString(filter.Plan) > 64 {
		return Filter{}, fmt.Errorf("plan code must be at most 64 characters")
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
	values.Set("status", filter.Status)
	values.Set("plan", filter.Plan)
	values.Set("cancellation", filter.Cancellation)
	values.Set("organization_id", filter.OrganizationID)
	values.Set("page", strconv.FormatInt(int64(page), 10))
	return "/subscriptions?" + values.Encode()
}
