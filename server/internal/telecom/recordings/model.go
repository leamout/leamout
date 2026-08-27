package recordings

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type RecordingResponse struct {
	ID              uuid.UUID  `json:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id"`
	CallID          uuid.UUID  `json:"call_id"`
	Status          string     `json:"status"`
	FileSizeBytes   *int64     `json:"file_size_bytes,omitempty"`
	Format          *string    `json:"format,omitempty"`
	DurationSeconds *int32     `json:"duration_seconds,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type PlaybackResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

func recordingResponse(recording sqlc.Recording) RecordingResponse {
	return RecordingResponse{
		ID:              recording.ID,
		OrganizationID:  recording.OrganizationID,
		CallID:          recording.CallID,
		Status:          recording.Status,
		FileSizeBytes:   recording.FileSizeBytes,
		Format:          recording.Format,
		DurationSeconds: recording.DurationSeconds,
		StartedAt:       pgconv.TimestamptzToTimePtr(recording.StartedAt),
		CompletedAt:     pgconv.TimestamptzToTimePtr(recording.CompletedAt),
		CreatedAt:       pgconv.TimestamptzToTime(recording.CreatedAt),
		UpdatedAt:       pgconv.TimestamptzToTime(recording.UpdatedAt),
	}
}
