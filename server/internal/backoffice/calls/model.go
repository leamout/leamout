package calls

import "fmt"

type Call struct {
	ID             string
	OrganizationID string
	Organization   string
	From           string
	To             string
	Direction      string
	Duration       string
	Status         string
	CreatedAt      string
}

// Detail is the deliberately secret-free operator projection for one call.
type Detail struct {
	Call
	MediaState          string
	SIPCallID           string
	ApplicationID       string
	Application         string
	CarrierConnectionID string
	CarrierConnection   string
	ProviderID          string
	Provider            string
	TrunkID             string
	Trunk               string
	TrunkEndpointID     string
	HangupReason        string
	RecordingCount      int64
	RecordingStatus     string
	StartedAt           string
	AnsweredAt          string
	EndedAt             string
	UpdatedAt           string
}

func formatDuration(seconds int64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remaining := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, remaining)
}
