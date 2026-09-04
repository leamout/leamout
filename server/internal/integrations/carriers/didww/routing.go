package didww

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ConfigureRouting assigns a DIDWW DID to an existing DIDWW Voice IN trunk.
// The trunk itself represents provider-specific ingress configuration; Leamout
// routing and tenant ownership remain outside this adapter.
func (c *Client) ConfigureRouting(ctx context.Context, didID, voiceInTrunkID string) (DID, error) {
	didID = strings.TrimSpace(didID)
	voiceInTrunkID = strings.TrimSpace(voiceInTrunkID)
	if didID == "" || voiceInTrunkID == "" {
		return DID{}, fmt.Errorf("didww: DID id and Voice IN trunk id are required")
	}

	payload := relationshipPatch{
		Data: relationshipPatchData{
			Type: "dids",
			ID:   didID,
			Relationships: map[string]relationshipWrite{
				"voice_in_trunk": {
					Data: resourceIdentifier{
						Type: "voice_in_trunks",
						ID:   voiceInTrunkID,
					},
				},
			},
		},
	}

	query := url.Values{}
	query.Set("include", "voice_in_trunk")
	var response singleResponse[didAttributes]
	if err := c.do(ctx, http.MethodPatch, "/v3/dids/"+url.PathEscape(didID), query, payload, &response); err != nil {
		return DID{}, err
	}
	return didFromResource(response.Data), nil
}
