package calls

import "github.com/google/uuid"

const pageSize = 25

type Filter struct {
	Search         string
	State          string
	Direction      string
	OrganizationID string
	AgentID        string
	From           string
	To             string
	Page           int32
}

type Call struct {
	ID                 uuid.UUID
	OrganizationID     uuid.UUID
	OrganizationName   string
	OrganizationStatus string
	VoiceAgentID       *uuid.UUID
	AgentName          string
	Direction          string
	State              string
	FromURI            string
	ToURI              string
	CreatedAt          string
}

type List struct {
	Calls   []Call
	Filter  Filter
	HasNext bool
}

type Detail struct {
	Call
	MediaState      string
	SIPCallID       string
	HangupReason    string
	TrunkID         string
	TrunkEndpointID string
	StartedAt       string
	AnsweredAt      string
	EndedAt         string
	UpdatedAt       string
	Duration        string
	Session         *Session
}

type Session struct {
	ID                    uuid.UUID
	VoiceAgentID          uuid.UUID
	State                 string
	ConfigurationRevision int32
	TurnCount             int32
	InterruptionCount     int32
	FirstResponseLatency  string
	AverageTurnLatency    string
	StartedAt             string
	EndedAt               string
	MediaStartedAt        string
	MediaEndedAt          string
}
