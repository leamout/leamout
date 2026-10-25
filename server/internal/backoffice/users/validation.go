package users

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
		Status:         strings.TrimSpace(values.Get("status")),
		Admin:          strings.TrimSpace(values.Get("admin")),
		Verified:       strings.TrimSpace(values.Get("verified")),
		OrganizationID: strings.TrimSpace(values.Get("organization_id")),
	}
	if utf8.RuneCountInString(filter.Search) > 200 {
		return Filter{}, fmt.Errorf("search must be at most 200 characters")
	}
	switch filter.Status {
	case "", "active", "disabled":
	default:
		return Filter{}, fmt.Errorf("invalid status filter")
	}
	switch filter.Admin {
	case "", "yes", "no":
	default:
		return Filter{}, fmt.Errorf("invalid admin filter")
	}
	switch filter.Verified {
	case "", "yes", "no":
	default:
		return Filter{}, fmt.Errorf("invalid verified filter")
	}
	if filter.OrganizationID != "" {
		id, err := uuid.Parse(filter.OrganizationID)
		if err != nil {
			return Filter{}, fmt.Errorf("invalid organization ID")
		}
		filter.OrganizationID = id.String()
	}
	page, err := parsePage(values.Get("page"))
	if err != nil {
		return Filter{}, err
	}
	filter.Page = page
	return filter, nil
}

func parsePage(value string) (int32, error) {
	if value == "" {
		return 1, nil
	}
	page, err := strconv.ParseInt(value, 10, 32)
	if err != nil || page < 1 || page > 1000000 {
		return 0, fmt.Errorf("page must be between 1 and 1000000")
	}
	return int32(page), nil
}

func pageURL(filter Filter, page int32) string {
	values := url.Values{}
	values.Set("q", filter.Search)
	values.Set("status", filter.Status)
	values.Set("admin", filter.Admin)
	values.Set("verified", filter.Verified)
	values.Set("organization_id", filter.OrganizationID)
	values.Set("page", strconv.FormatInt(int64(page), 10))
	return "/users?" + values.Encode()
}

func memberPageURL(id string, page int32) string {
	values := url.Values{}
	values.Set("members_page", strconv.FormatInt(int64(page), 10))
	return "/users/" + id + "?" + values.Encode()
}
