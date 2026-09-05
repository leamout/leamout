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

type SearchNumbersRequest struct {
	NumberContains string
	DIDGroupID     string
	DIDGroupTypeID string
	CountryID      string
	RegionID       string
	CityID         string
	Feature        string
	NANPAPrefixID  string
}

type OrderNumberRequest struct {
	AvailableDIDID      string
	SKUID               string
	BillingCyclesCount  *int
	ExternalReferenceID string
	CallbackURL         string
	CallbackMethod      string
}

func (c *Client) SearchNumbers(ctx context.Context, request SearchNumbersRequest) ([]AvailableNumber, error) {
	query := url.Values{}
	setFilter(query, "number_contains", request.NumberContains)
	setFilter(query, "did_group.id", request.DIDGroupID)
	setFilter(query, "did_group_type.id", request.DIDGroupTypeID)
	setFilter(query, "country.id", request.CountryID)
	setFilter(query, "region.id", request.RegionID)
	setFilter(query, "city.id", request.CityID)
	setFilter(query, "did_group.features", request.Feature)
	setFilter(query, "nanpa_prefix.id", request.NANPAPrefixID)

	var response collectionResponse[availableDIDAttributes]
	if err := c.do(ctx, http.MethodGet, "/v3/available_dids", query, nil, &response); err != nil {
		return nil, err
	}

	result := make([]AvailableNumber, 0, len(response.Data))
	for _, item := range response.Data {
		result = append(result, AvailableNumber{
			ID:     item.ID,
			Number: item.Attributes.Number,
		})
	}
	return result, nil
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

func (c *Client) GetNumber(ctx context.Context, didID string) (DID, error) {
	didID = strings.TrimSpace(didID)
	if didID == "" {
		return DID{}, fmt.Errorf("didww: DID id is required")
	}

	query := url.Values{}
	query.Set("include", "voice_in_trunk")
	var response singleResponse[didAttributes]
	if err := c.do(ctx, http.MethodGet, "/v3/dids/"+url.PathEscape(didID), query, nil, &response); err != nil {
		return DID{}, err
	}
	return didFromResource(response.Data), nil
}

func (c *Client) FindNumber(ctx context.Context, number string) (DID, error) {
	number = strings.TrimPrefix(strings.TrimSpace(number), "+")
	if number == "" {
		return DID{}, fmt.Errorf("didww: number is required")
	}

	query := url.Values{}
	query.Set("filter[number]", number)
	query.Set("include", "voice_in_trunk")
	var response collectionResponse[didAttributes]
	if err := c.do(ctx, http.MethodGet, "/v3/dids", query, nil, &response); err != nil {
		return DID{}, err
	}
	if len(response.Data) == 0 {
		return DID{}, fmt.Errorf("didww: DID %s not found", number)
	}
	if len(response.Data) > 1 {
		return DID{}, fmt.Errorf("didww: DID %s resolved ambiguously", number)
	}
	return didFromResource(response.Data[0]), nil
}

func (c *Client) OrderNumber(ctx context.Context, request OrderNumberRequest) (Order, error) {
	availableDIDID := strings.TrimSpace(request.AvailableDIDID)
	skuID := strings.TrimSpace(request.SKUID)
	if availableDIDID == "" || skuID == "" {
		return Order{}, fmt.Errorf("didww: available DID id and SKU id are required")
	}
	callbackMethod := strings.ToLower(strings.TrimSpace(request.CallbackMethod))
	if callbackMethod != "" && callbackMethod != "get" && callbackMethod != "post" {
		return Order{}, fmt.Errorf("didww: callback method must be get or post")
	}

	payload := orderRequest{
		Data: orderRequestData{
			Type: "orders",
			Attributes: orderRequestAttributes{
				AllowBackOrdering:   false,
				ExternalReferenceID: strings.TrimSpace(request.ExternalReferenceID),
				CallbackURL:         strings.TrimSpace(request.CallbackURL),
				CallbackMethod:      callbackMethod,
				Items: []orderRequestItem{{
					Type: "did_order_items",
					Attributes: orderRequestItemAttributes{
						SKUID:              skuID,
						AvailableDIDID:     availableDIDID,
						BillingCyclesCount: request.BillingCyclesCount,
					},
				}},
			},
		},
	}

	var response singleResponse[orderAttributes]
	if err := c.do(ctx, http.MethodPost, "/v3/orders", nil, payload, &response); err != nil {
		return Order{}, err
	}
	return Order{
		ID:                  response.Data.ID,
		Reference:           response.Data.Attributes.Reference,
		ExternalReferenceID: response.Data.Attributes.ExternalReferenceID,
		Amount:              response.Data.Attributes.Amount,
		Status:              response.Data.Attributes.Status,
		Description:         response.Data.Attributes.Description,
		CreatedAt:           response.Data.Attributes.CreatedAt,
	}, nil
}

// ReleaseNumber relinquishes an owned DID. Callers must persist release intent
// before invoking this irreversible provider operation.
func (c *Client) ReleaseNumber(ctx context.Context, didID string) error {
	didID = strings.TrimSpace(didID)
	if didID == "" {
		return fmt.Errorf("didww: DID id is required")
	}
	return c.do(ctx, http.MethodDelete, "/v3/dids/"+url.PathEscape(didID), nil, nil, nil)
}

func setFilter(query url.Values, name, value string) {
	if value = strings.TrimSpace(value); value != "" {
		query.Set("filter["+name+"]", value)
	}
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

func didFromResource(item resource[didAttributes]) DID {
	voiceInTrunkID := ""
	if rel, ok := item.Relationships["voice_in_trunk"]; ok && rel.Data != nil {
		voiceInTrunkID = rel.Data.ID
	}
	return DID{
		ID:                     item.ID,
		Number:                 item.Attributes.Number,
		Blocked:                item.Attributes.Blocked,
		AwaitingRegistration:   item.Attributes.AwaitingRegistration,
		PendingRemoval:         item.Attributes.PendingRemoval,
		Terminated:             item.Attributes.Terminated,
		CapacityLimit:          item.Attributes.CapacityLimit,
		ChannelsIncludedCount:  item.Attributes.ChannelsIncludedCount,
		DedicatedChannelsCount: item.Attributes.DedicatedChannelsCount,
		BillingCyclesCount:     item.Attributes.BillingCyclesCount,
		ExpiresAt:              item.Attributes.ExpiresAt,
		CreatedAt:              item.Attributes.CreatedAt,
		VoiceInTrunkID:         voiceInTrunkID,
	}
}
