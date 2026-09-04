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
