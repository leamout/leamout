package didww

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

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

	numbers := make([]AvailableNumber, 0, len(response.Data))
	for _, item := range response.Data {
		numbers = append(numbers, AvailableNumber{
			ID:     item.ID,
			Number: item.Attributes.Number,
		})
	}
	return numbers, nil
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
