package agents

import "github.com/google/uuid"

const pageSize = 25

type Filter struct {
	Search         string
	Status         string
	Engine         string
	OrganizationID string
	Page           int32
}

type Agent struct {
	ID                 uuid.UUID
	Name               string
	OrganizationID     uuid.UUID
	OrganizationName   string
	OrganizationStatus string
	Engine             string
	Status             string
	Language           string
	CreatedAt          string
}

type List struct {
	Agents  []Agent
	Filter  Filter
	HasNext bool
}

type Detail struct {
	Agent
	Instructions          string
	Voice                 string
	InterruptionPolicy    string
	RecordingPolicy       string
	ConfigurationRevision int32
	UpdatedAt             string
}
