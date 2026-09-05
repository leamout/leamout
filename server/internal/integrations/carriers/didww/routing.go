package didww

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type EnsureInboundTrunkRequest struct {
	Name                string
	ExternalReferenceID string
	Host                string
	Port                int
	Transport           string
}

func (c *Client) EnsureInboundTrunk(ctx context.Context, request EnsureInboundTrunkRequest) (InboundTrunk, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.ExternalReferenceID = strings.TrimSpace(request.ExternalReferenceID)
	request.Host = strings.TrimSpace(request.Host)
	request.Transport = strings.ToLower(strings.TrimSpace(request.Transport))
	if request.Name == "" || request.ExternalReferenceID == "" || request.Host == "" {
		return InboundTrunk{}, fmt.Errorf("didww: inbound trunk name, external reference id and host are required")
	}
	if request.Port <= 0 || request.Port > 65535 {
		return InboundTrunk{}, fmt.Errorf("didww: inbound trunk port must be between 1 and 65535")
	}
	transportProtocolID, err := inboundTransportProtocolID(request.Transport)
	if err != nil {
		return InboundTrunk{}, err
	}

	query := url.Values{}
	query.Set("filter[external_reference_id]", request.ExternalReferenceID)
	var existing collectionResponse[inboundTrunkAttributes]
	if err := c.do(ctx, http.MethodGet, "/v3/voice_in_trunks", query, nil, &existing); err != nil {
		return InboundTrunk{}, err
	}
	if len(existing.Data) > 1 {
		return InboundTrunk{}, fmt.Errorf("didww: inbound trunk external reference resolved ambiguously")
	}

	attributes := inboundTrunkAttributes{
		Name:                request.Name,
		ExternalReferenceID: request.ExternalReferenceID,
		Priority:            1,
		Weight:              100,
		CapacityLimit:       100,
		RingingTimeout:      30,
		CLIFormat:           "e164",
		CLIPrefix:           "+",
		Configuration: inboundTrunkConfiguration{
			Type: "sip_configurations",
			Attributes: inboundSIPConfiguration{
				Username:            "+{DID}",
				Host:                request.Host,
				Port:                request.Port,
				TransportProtocolID: transportProtocolID,
				RXDTMFFormatID:      1,
				TXDTMFFormatID:      1,
				AuthEnabled:         false,
				MediaEncryptionMode: "disabled",
			},
		},
	}

	if len(existing.Data) == 0 {
		payload := inboundTrunkRequest{Data: inboundTrunkRequestData{Type: "voice_in_trunks", Attributes: attributes}}
		var response singleResponse[inboundTrunkAttributes]
		if err := c.do(ctx, http.MethodPost, "/v3/voice_in_trunks", nil, payload, &response); err != nil {
			return InboundTrunk{}, err
		}
		return inboundTrunkFromResource(response.Data), nil
	}

	current := inboundTrunkFromResource(existing.Data[0])
	if current.Name == request.Name && current.Host == request.Host && current.Port == request.Port && current.TransportProtocolID == transportProtocolID {
		return current, nil
	}

	payload := inboundTrunkRequest{Data: inboundTrunkRequestData{
		ID:         existing.Data[0].ID,
		Type:       "voice_in_trunks",
		Attributes: attributes,
	}}
	var response singleResponse[inboundTrunkAttributes]
	if err := c.do(ctx, http.MethodPatch, "/v3/voice_in_trunks/"+url.PathEscape(existing.Data[0].ID), nil, payload, &response); err != nil {
		return InboundTrunk{}, err
	}
	return inboundTrunkFromResource(response.Data), nil
}

func inboundTransportProtocolID(transport string) (int, error) {
	switch transport {
	case "udp":
		return 1, nil
	case "tcp":
		return 2, nil
	case "tls":
		return 3, nil
	default:
		return 0, fmt.Errorf("didww: unsupported inbound SIP transport %q", transport)
	}
}

func inboundTrunkFromResource(item resource[inboundTrunkAttributes]) InboundTrunk {
	return InboundTrunk{
		ID:                  item.ID,
		Name:                item.Attributes.Name,
		ExternalReferenceID: item.Attributes.ExternalReferenceID,
		Host:                item.Attributes.Configuration.Attributes.Host,
		Port:                item.Attributes.Configuration.Attributes.Port,
		TransportProtocolID: item.Attributes.Configuration.Attributes.TransportProtocolID,
	}
}

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
