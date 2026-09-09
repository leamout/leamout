package calls

import "fmt"

type Call struct {
	ID           string
	Organization string
	From         string
	To           string
	Direction    string
	Duration     string
	Status       string
	CreatedAt    string
}

func formatDuration(seconds int64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remaining := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, remaining)
}
