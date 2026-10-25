package numbers

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
		Country:        strings.TrimSpace(values.Get("country")),
		Voice:          strings.TrimSpace(values.Get("voice")),
		OrganizationID: strings.TrimSpace(values.Get("organization_id")),
		TrunkID:        strings.TrimSpace(values.Get("trunk_id")),
	}
	if utf8.RuneCountInString(filter.Search) > 200 {
		return Filter{}, fmt.Errorf("search must be at most 200 characters")
	}
	switch filter.Status {
	case "", "active", "disabled", "porting", "released":
	default:
		return Filter{}, fmt.Errorf("invalid status filter")
	}
	switch filter.Voice {
	case "", "yes", "no":
	default:
		return Filter{}, fmt.Errorf("invalid voice filter")
	}
	if filter.OrganizationID != "" {
		id, err := uuid.Parse(filter.OrganizationID)
		if err != nil {
			return Filter{}, fmt.Errorf("invalid organization ID")
		}
		filter.OrganizationID = id.String()
	}
	if filter.TrunkID != "" {
		id, err := uuid.Parse(filter.TrunkID)
		if err != nil {
			return Filter{}, fmt.Errorf("invalid trunk ID")
		}
		filter.TrunkID = id.String()
	}
	filter.Country = strings.ToUpper(filter.Country)
	if filter.Country != "" && (len(filter.Country) != 2 || filter.Country[0] < 'A' || filter.Country[0] > 'Z' || filter.Country[1] < 'A' || filter.Country[1] > 'Z') {
		return Filter{}, fmt.Errorf("country must be a two-letter country code")
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
	values.Set("country", filter.Country)
	values.Set("voice", filter.Voice)
	values.Set("organization_id", filter.OrganizationID)
	values.Set("trunk_id", filter.TrunkID)
	values.Set("page", strconv.FormatInt(int64(page), 10))
	return "/numbers?" + values.Encode()
}
