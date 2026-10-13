package polar

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) GetCustomerStateByExternalID(ctx context.Context, externalID string) (CustomerState, error) {
	if strings.TrimSpace(externalID) == "" {
		return CustomerState{}, fmt.Errorf("polar external customer ID is required")
	}

	var result CustomerState
	path := "/v1/customers/external/" + url.PathEscape(externalID) + "/state"
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return CustomerState{}, err
	}
	if err := validateUUID(result.ID); err != nil {
		return CustomerState{}, fmt.Errorf("polar customer state ID is invalid: %w", err)
	}
	if result.ExternalID != externalID {
		return CustomerState{}, fmt.Errorf("polar customer state external ID does not match")
	}
	return result, nil
}
