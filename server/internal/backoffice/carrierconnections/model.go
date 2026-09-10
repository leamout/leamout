package carrierconnections

type CarrierConnection struct {
	ID                 string
	OrganizationID     string
	Organization       string
	Name               string
	Provider           string
	Scope              string
	Status             string
	InboundEnabled     bool
	MaxCPS             int32
	MaxConcurrentCalls int32
	Trunks             int64
}

type SourceIP struct{ ID, CIDR, CreatedAt string }
type ProviderResource struct{ Type, ProviderResourceID, CreatedAt, UpdatedAt string }
type Detail struct {
	CarrierConnection
	ProviderID, OutboundAuth, InboundAuth, MaxDailyMinutes, Codecs, CreatedAt, UpdatedAt string
	SupportsVideo, SupportsFax                                                           bool
	SourceIPs                                                                            []SourceIP
	Resources                                                                            []ProviderResource
}
