package numbers

const pageSize = 25

type Filter struct {
	Search         string
	Status         string
	Country        string
	Voice          string
	OrganizationID string
	TrunkID        string
	Page           int32
}

type Number struct {
	ID                 string
	OrganizationID     string
	OrganizationName   string
	OrganizationStatus string
	Number             string
	CountryCode        string
	TrunkID            string
	TrunkName          string
	TrunkStatus        string
	AgentID            string
	AgentName          string
	AgentStatus        string
	VoiceEnabled       bool
	Status             string
	CreatedAt          string
}

type List struct {
	Numbers []Number
	Filter  Filter
	HasNext bool
}

type Detail struct {
	Number
	UpdatedAt string
}
