package didww

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/leamout/leamout/internal/telecom/numbers"
)

const providerSlug = "didww"

type countryAttributes struct {
	ISO string `json:"iso"`
}

type availabilityRelationship struct {
	Data any `json:"data,omitempty"`
}

type availabilityIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type availableDIDSearchResource struct {
	ID            string                              `json:"id"`
	Type          string                              `json:"type"`
	Attributes    availableDIDAttributes              `json:"attributes"`
	Relationships map[string]availabilityRelationship `json:"relationships,omitempty"`
}

type includedAvailabilityResource struct {
	ID            string                              `json:"id"`
	Type          string                              `json:"type"`
	Attributes    stockKeepingUnitAttributes          `json:"attributes"`
	Relationships map[string]availabilityRelationship `json:"relationships,omitempty"`
}

type stockKeepingUnitAttributes struct {
	ChannelsIncludedCount int `json:"channels_included_count"`
}

type availableDIDSearchResponse struct {
	Data     []availableDIDSearchResource  `json:"data"`
	Included []includedAvailabilityResource `json:"included,omitempty"`
}

func (c *Client) SearchAvailable(ctx context.Context, request numbers.AvailableSearchRequest) ([]numbers.ManagedNumberCandidate, error) {
	countryID, err := c.countryIDByISO(ctx, request.CountryCode)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Add("include", "did_group")
	query.Add("include", "did_group.stock_keeping_units")
	setFilter(query, "country.id", countryID)
	setFilter(query, "number_contains", request.Contains)
	setFilter(query, "did_group.features", "voice_in")

	var response availableDIDSearchResponse
	if err := c.do(ctx, http.MethodGet, "/v3/available_dids", query, nil, &response); err != nil {
		return nil, err
	}

	groupSKUs := make(map[string][]string)
	skus := make(map[string]stockKeepingUnitAttributes)
	for _, item := range response.Included {
		switch item.Type {
		case "did_groups":
			groupSKUs[item.ID] = relationshipIDs(item.Relationships["stock_keeping_units"])
		case "stock_keeping_units":
			skus[item.ID] = item.Attributes
		}
	}

	result := make([]numbers.ManagedNumberCandidate, 0, len(response.Data))
	for _, item := range response.Data {
		groupID := relationshipID(item.Relationships["did_group"])
		skuID, channels := selectVoiceSKU(groupSKUs[groupID], skus)
		if groupID == "" || skuID == "" || channels <= 0 {
			continue
		}
		number := strings.TrimSpace(item.Attributes.Number)
		if number == "" {
			continue
		}
		result = append(result, numbers.ManagedNumberCandidate{
			Provider:              providerSlug,
			ProviderInventoryID:   item.ID,
			ProviderProductID:     skuID,
			Number:                "+" + strings.TrimPrefix(number, "+"),
			CountryCode:           request.CountryCode,
			ChannelsIncludedCount: channels,
		})
	}
	return result, nil
}

func (c *Client) countryIDByISO(ctx context.Context, iso string) (string, error) {
	query := url.Values{}
	setFilter(query, "iso", strings.ToUpper(strings.TrimSpace(iso)))

	var response collectionResponse[countryAttributes]
	if err := c.do(ctx, http.MethodGet, "/v3/countries", query, nil, &response); err != nil {
		return "", err
	}
	if len(response.Data) == 0 {
		return "", fmt.Errorf("didww: country %s is not available", iso)
	}
	if len(response.Data) != 1 {
		return "", fmt.Errorf("didww: country %s resolved ambiguously", iso)
	}
	return response.Data[0].ID, nil
}

func selectVoiceSKU(ids []string, skus map[string]stockKeepingUnitAttributes) (string, int) {
	type option struct {
		id       string
		channels int
	}
	options := make([]option, 0, len(ids))
	for _, id := range ids {
		attributes, ok := skus[id]
		if !ok || attributes.ChannelsIncludedCount <= 0 {
			continue
		}
		options = append(options, option{id: id, channels: attributes.ChannelsIncludedCount})
	}
	if len(options) == 0 {
		return "", 0
	}
	sort.Slice(options, func(i, j int) bool {
		if options[i].channels == options[j].channels {
			return options[i].id < options[j].id
		}
		return options[i].channels < options[j].channels
	})
	return options[0].id, options[0].channels
}

func relationshipID(relationship availabilityRelationship) string {
	data, ok := relationship.Data.(map[string]any)
	if !ok {
		return ""
	}
	id, _ := data["id"].(string)
	return id
}

func relationshipIDs(relationship availabilityRelationship) []string {
	data, ok := relationship.Data.([]any)
	if !ok {
		return nil
	}
	ids := make([]string, 0, len(data))
	for _, raw := range data {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id, _ := item["id"].(string)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}
