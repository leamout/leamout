package numbers

type Number struct {
	ID             string
	OrganizationID string
	Organization   string
	Number         string
	CountryCode    string
	Source         string
	Provider       string
	Voice          bool
	SMS            bool
	Status         string
	CreatedAt      string
}

type PhoneNumber = Number

type Detail struct {
	PhoneNumber
	CarrierConnectionID  string
	CarrierConnection    string
	ProviderConnectionID string
	ProviderConnection   string
	ProviderID           string
	ProviderResourceID   string
	ErrorCode            string
	ErrorMessage         string
	UpdatedAt            string
}
