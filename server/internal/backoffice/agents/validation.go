package agents

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
		Status:         values.Get("status"),
		Engine:         values.Get("engine"),
		OrganizationID: values.Get("organization_id"),
		Page:           1,
	}
	if utf8.RuneCountInString(filter.Search) > 200 {
		return Filter{}, fmt.Errorf("search must be at most 200 characters")
	}
	switch filter.Status {
	case "", "active", "disabled":
	default:
		return Filter{}, fmt.Errorf("invalid agent status")
	}
	switch filter.Engine {
	case "", "composable", "realtime":
	default:
		return Filter{}, fmt.Errorf("invalid agent engine")
	}
	if filter.OrganizationID != "" {
		id, err := uuid.Parse(filter.OrganizationID)
		if err != nil {
			return Filter{}, fmt.Errorf("invalid organization ID")
		}
		filter.OrganizationID = id.String()
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
	values.Set("engine", filter.Engine)
	values.Set("organization_id", filter.OrganizationID)
	values.Set("page", strconv.FormatInt(int64(page), 10))
	return "/agents?" + values.Encode()
}
