package organizations

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

func parseFilter(values url.Values) (Filter, error) {
	filter := Filter{
		Search: strings.TrimSpace(values.Get("q")),
		Status: values.Get("status"),
		Page:   1,
	}
	if utf8.RuneCountInString(filter.Search) > 200 {
		return Filter{}, fmt.Errorf("search must be at most 200 characters")
	}
	switch filter.Status {
	case "", "active", "disabled", "deleted":
	default:
		return Filter{}, fmt.Errorf("invalid organization status")
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
	values.Set("page", strconv.FormatInt(int64(page), 10))
	return "/backoffice/organizations?" + values.Encode()
}
