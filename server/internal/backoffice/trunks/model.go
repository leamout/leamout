package trunks

type Trunk struct {
	ID             string
	OrganizationID string
	Organization   string
	Name           string
	Mode           string
	Provider       string
	Direction      string
	Status         string
	Endpoints      int64
}

type Endpoint struct {
	ID, Host, Transport, Direction, Health, LastResponse, LastLatency, LastError, LastChecked string
	Port, Priority, Weight, Failures                                                          int32
	Enabled                                                                                   bool
}
type Detail struct {
	Trunk
	ManagedDefault                                               bool
	CarrierConnectionID, CarrierConnection, CreatedAt, UpdatedAt string
	EnabledEndpoints                                             int64
	EndpointList                                                 []Endpoint
}
