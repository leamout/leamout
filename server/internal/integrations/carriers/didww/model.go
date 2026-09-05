package didww

import "time"

type resource[T any] struct {
	ID            string                  `json:"id"`
	Type          string                  `json:"type"`
	Attributes    T                       `json:"attributes"`
	Relationships map[string]relationship `json:"relationships,omitempty"`
}

type relationship struct {
	Data *resourceIdentifier `json:"data,omitempty"`
}

type resourceIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type collectionResponse[T any] struct {
	Data []resource[T] `json:"data"`
	Meta responseMeta  `json:"meta,omitempty"`
}

type singleResponse[T any] struct {
	Data resource[T]  `json:"data"`
	Meta responseMeta `json:"meta,omitempty"`
}

type responseMeta struct {
	APIVersion   string `json:"api_version,omitempty"`
	TotalRecords int    `json:"total_records,omitempty"`
}

type AvailableNumber struct {
	ID     string
	Number string
}

type availableDIDAttributes struct {
	Number string `json:"number"`
}

type countryAttributes struct {
	ISO string `json:"iso"`
}

type availabilityRelationship struct {
	Data any `json:"data,omitempty"`
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
	Data     []availableDIDSearchResource   `json:"data"`
	Included []includedAvailabilityResource `json:"included,omitempty"`
}

type DID struct {
	ID                     string
	Number                 string
	Blocked                bool
	AwaitingRegistration   bool
	PendingRemoval         bool
	Terminated             bool
	CapacityLimit          int
	ChannelsIncludedCount  int
	DedicatedChannelsCount int
	BillingCyclesCount     *int
	ExpiresAt              *time.Time
	CreatedAt              time.Time
	VoiceInTrunkID         string
}

type didAttributes struct {
	Blocked                bool       `json:"blocked"`
	AwaitingRegistration   bool       `json:"awaiting_registration"`
	PendingRemoval         bool       `json:"pending_removal"`
	Terminated             bool       `json:"terminated"`
	Number                 string     `json:"number"`
	CapacityLimit          int        `json:"capacity_limit"`
	ChannelsIncludedCount  int        `json:"channels_included_count"`
	DedicatedChannelsCount int        `json:"dedicated_channels_count"`
	BillingCyclesCount     *int       `json:"billing_cycles_count"`
	ExpiresAt              *time.Time `json:"expires_at"`
	CreatedAt              time.Time  `json:"created_at"`
}

type InboundTrunk struct {
	ID                  string
	Name                string
	ExternalReferenceID string
	Host                string
	Port                int
	TransportProtocolID int
}

type inboundTrunkAttributes struct {
	Name                string                    `json:"name"`
	ExternalReferenceID string                    `json:"external_reference_id"`
	Priority            int                       `json:"priority"`
	Weight              int                       `json:"weight"`
	CapacityLimit       int                       `json:"capacity_limit"`
	RingingTimeout      int                       `json:"ringing_timeout"`
	CLIFormat           string                    `json:"cli_format"`
	CLIPrefix           string                    `json:"cli_prefix"`
	Configuration       inboundTrunkConfiguration `json:"configuration"`
}

type inboundTrunkConfiguration struct {
	Type       string                     `json:"type"`
	Attributes inboundSIPConfiguration    `json:"attributes"`
}

type inboundSIPConfiguration struct {
	Username            string `json:"username"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	TransportProtocolID int    `json:"transport_protocol_id"`
	RXDTMFFormatID      int    `json:"rx_dtmf_format_id"`
	TXDTMFFormatID      int    `json:"tx_dtmf_format_id"`
	AuthEnabled         bool   `json:"auth_enabled"`
	MediaEncryptionMode string `json:"media_encryption_mode"`
}

type inboundTrunkRequest struct {
	Data inboundTrunkRequestData `json:"data"`
}

type inboundTrunkRequestData struct {
	ID         string                 `json:"id,omitempty"`
	Type       string                 `json:"type"`
	Attributes inboundTrunkAttributes `json:"attributes"`
}

type Order struct {
	ID                  string
	Reference           string
	ExternalReferenceID *string
	Amount              string
	Status              string
	Description         string
	CreatedAt           time.Time
}

type orderAttributes struct {
	Reference           string    `json:"reference"`
	ExternalReferenceID *string   `json:"external_reference_id"`
	Amount              string    `json:"amount"`
	Status              string    `json:"status"`
	Description         string    `json:"description"`
	CreatedAt           time.Time `json:"created_at"`
}

type orderRequest struct {
	Data orderRequestData `json:"data"`
}

type orderRequestData struct {
	Type       string                 `json:"type"`
	Attributes orderRequestAttributes `json:"attributes"`
}

type orderRequestAttributes struct {
	AllowBackOrdering   bool               `json:"allow_back_ordering"`
	ExternalReferenceID string             `json:"external_reference_id,omitempty"`
	CallbackURL         string             `json:"callback_url,omitempty"`
	CallbackMethod      string             `json:"callback_method,omitempty"`
	Items               []orderRequestItem `json:"items"`
}

type orderRequestItem struct {
	Type       string                     `json:"type"`
	Attributes orderRequestItemAttributes `json:"attributes"`
}

type orderRequestItemAttributes struct {
	SKUID              string `json:"sku_id"`
	AvailableDIDID     string `json:"available_did_id,omitempty"`
	BillingCyclesCount *int   `json:"billing_cycles_count,omitempty"`
}

type relationshipPatch struct {
	Data relationshipPatchData `json:"data"`
}

type relationshipPatchData struct {
	Type          string                       `json:"type"`
	ID            string                       `json:"id"`
	Relationships map[string]relationshipWrite `json:"relationships"`
}

type relationshipWrite struct {
	Data resourceIdentifier `json:"data"`
}
