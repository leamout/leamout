package calls

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/leamout/leamout/server/internal/database/sqlc"
)

// Repository owns explicit cross-organization operator read projections.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context, filter Filter) (List, error) {
	organizationID, err := optionalID(filter.OrganizationID)
	if err != nil {
		return List{}, err
	}
	agentID, err := optionalID(filter.AgentID)
	if err != nil {
		return List{}, err
	}
	from, err := dateBoundary(filter.From, false)
	if err != nil {
		return List{}, err
	}
	before, err := dateBoundary(filter.To, true)
	if err != nil {
		return List{}, err
	}
	rows, err := r.queries.ListBackofficeCalls(ctx, sqlc.ListBackofficeCallsParams{
		Search:         filter.Search,
		OrganizationID: organizationID,
		PageLimit:      pageSize + 1,
		PageOffset:     (filter.Page - 1) * pageSize,
		State:          filter.State,
		Direction:      filter.Direction,
		AgentID:        agentID,
		CreatedFrom:    from,
		CreatedBefore:  before,
	})
	if err != nil {
		return List{}, err
	}
	result := List{
		Filter:  filter,
		HasNext: len(rows) > pageSize,
	}
	if result.HasNext {
		rows = rows[:pageSize]
	}
	for _, row := range rows {
		result.Calls = append(result.Calls, Call{
			ID:                 row.ID,
			OrganizationID:     row.OrganizationID,
			OrganizationName:   row.OrganizationName,
			OrganizationStatus: row.OrganizationStatus,
			VoiceAgentID:       row.VoiceAgentID,
			AgentName:          row.AgentName,
			Direction:          row.Direction,
			State:              row.State,
			FromURI:            row.FromUri,
			ToURI:              row.ToUri,
			CreatedAt:          formatTime(row.CreatedAt),
		})
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeCall(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	result := Detail{
		Call: Call{
			ID:                 row.ID,
			OrganizationID:     row.OrganizationID,
			OrganizationName:   row.OrganizationName,
			OrganizationStatus: row.OrganizationStatus,
			VoiceAgentID:       row.VoiceAgentID,
			AgentName:          row.AgentName,
			Direction:          row.Direction,
			State:              row.State,
			FromURI:            row.FromUri,
			ToURI:              row.ToUri,
			CreatedAt:          formatTime(row.CreatedAt),
		},
		StartedAt:       formatTime(row.StartedAt),
		AnsweredAt:      formatTime(row.AnsweredAt),
		EndedAt:         formatTime(row.EndedAt),
		UpdatedAt:       formatTime(row.UpdatedAt),
		MediaState:      row.MediaState,
		SIPCallID:       row.SipCallID,
		HangupReason:    row.HangupReason,
		TrunkID:         displayID(row.TrunkID),
		TrunkEndpointID: displayID(row.TrunkEndpointID),
		Duration:        callDuration(row.AnsweredAt, row.EndedAt, row.State),
	}
	session, err := r.queries.GetLatestBackofficeCallSession(ctx, sqlc.GetLatestBackofficeCallSessionParams{
		CallID:         row.ID,
		OrganizationID: row.OrganizationID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return Detail{}, err
	}
	result.Session = &Session{
		ID:                    session.ID,
		VoiceAgentID:          session.VoiceAgentID,
		State:                 session.State,
		ConfigurationRevision: session.ConfigurationRevision,
		TurnCount:             session.TurnCount,
		InterruptionCount:     session.InterruptionCount,
		FirstResponseLatency:  milliseconds(session.FirstResponseLatencyMs),
		AverageTurnLatency:    milliseconds(session.AvgTurnLatencyMs),
		StartedAt:             formatTime(session.StartedAt),
		EndedAt:               formatTime(session.EndedAt),
		MediaStartedAt:        formatTime(session.MediaStartedAt),
		MediaEndedAt:          formatTime(session.MediaEndedAt),
	}
	return result, nil
}

func displayID(value *uuid.UUID) string {
	if value == nil {
		return "—"
	}
	return value.String()
}

func milliseconds(value *int32) string {
	if value == nil {
		return "—"
	}
	return fmt.Sprintf("%d ms", *value)
}

// Duration measures answered time, never ringing time. Live durations are snapshots.
func callDuration(answered, ended pgtype.Timestamptz, state string) string {
	if !answered.Valid {
		return "Not answered"
	}
	end := ended.Time
	suffix := ""
	if !ended.Valid {
		if state == "completed" || state == "failed" || state == "cancelled" {
			return "Unknown"
		}
		end = time.Now()
		suffix = " (ongoing)"
	}
	seconds := max(int64(end.Sub(answered.Time).Seconds()), 0)
	return fmt.Sprintf("%d:%02d:%02d%s", seconds/3600, seconds%3600/60, seconds%60, suffix)
}

// The to-date is inclusive in the UI and exclusive at the next UTC midnight in SQL.
func dateBoundary(value string, nextDay bool) (pgtype.Timestamptz, error) {
	if value == "" {
		return pgtype.Timestamptz{}, nil
	}
	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return pgtype.Timestamptz{}, err
	}
	if nextDay {
		date = date.AddDate(0, 0, 1)
	}
	return pgtype.Timestamptz{Time: date, Valid: true}, nil
}

func formatTime(value pgtype.Timestamptz) string {
	if !value.Valid {
		return "—"
	}
	return value.Time.UTC().Format("2006-01-02 15:04:05 UTC")
}

func optionalID(value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
