package tools

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/pkg/apperror"
)

// sessionTool freezes the public tool definition and its non-secret execution
// settings when the Voice Agent session is created. Credentials stay outside
// the snapshot and are fetched from the secret store at execution time.
type sessionTool struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	Type        string          `json:"type"`
	EndpointURL *string         `json:"endpoint_url"`
	TimeoutMS   int32           `json:"timeout_ms"`
}

func validateExecutionSession(stored sqlc.VoiceAgentSession, req ExecuteRequest) error {
	if stored.ID != req.SessionID ||
		stored.OrganizationID != req.OrganizationID ||
		stored.VoiceAgentID != req.VoiceAgentID ||
		stored.CallID != req.CallID ||
		stored.State != "active" {
		return apperror.NewForbidden("voice agent session does not authorize tool execution")
	}
	return nil
}

// resolveSessionTool fails closed on legacy/incomplete snapshots and refuses
// names/IDs that were not advertised to the model at session creation.
func resolveSessionTool(snapshot []byte, requestedID uuid.UUID, requestedName string) (sessionTool, error) {
	deny := func() error {
		return apperror.NewForbidden("tool is not authorized by the voice agent session snapshot")
	}
	var configuration struct {
		Tools json.RawMessage `json:"tools"`
	}
	if err := json.Unmarshal(snapshot, &configuration); err != nil || len(configuration.Tools) == 0 {
		return sessionTool{}, deny()
	}
	var definitions []sessionTool
	if err := json.Unmarshal(configuration.Tools, &definitions); err != nil || len(definitions) == 0 {
		return sessionTool{}, deny()
	}

	ids := make(map[uuid.UUID]bool, len(definitions))
	names := make(map[string]bool, len(definitions))
	var matched sessionTool
	found := false
	for _, tool := range definitions {
		var parameters map[string]json.RawMessage
		if err := json.Unmarshal(tool.Parameters, &parameters); err != nil ||
			parameters == nil || tool.ID == uuid.Nil || strings.TrimSpace(tool.Name) == "" ||
			tool.TimeoutMS < 100 || tool.TimeoutMS > 30000 ||
			ids[tool.ID] || names[tool.Name] {
			return sessionTool{}, deny()
		}
		switch tool.Type {
		case TypeBuiltin:
			if tool.EndpointURL != nil {
				return sessionTool{}, deny()
			}
		case TypeWebhook:
			if tool.EndpointURL == nil || strings.TrimSpace(*tool.EndpointURL) == "" {
				return sessionTool{}, deny()
			}
		default:
			return sessionTool{}, deny()
		}
		ids[tool.ID] = true
		names[tool.Name] = true
		if requestedName != "" && tool.Name != requestedName {
			continue
		}
		if requestedID != uuid.Nil && tool.ID != requestedID {
			continue
		}
		if requestedID == uuid.Nil && requestedName == "" {
			continue
		}
		matched = tool
		found = true
	}
	if !found {
		return sessionTool{}, deny()
	}
	return matched, nil
}

// bindExecutionTool uses the live row only for emergency revocation, never for
// executable settings. Disabled/deleted tools are denied for new executions;
// edits take effect only on subsequent sessions.
func bindExecutionTool(current sqlc.VoiceAgentTool, frozen sessionTool) (sqlc.VoiceAgentTool, error) {
	if current.ID != frozen.ID || !current.Enabled {
		return sqlc.VoiceAgentTool{}, apperror.NewForbidden("voice agent tool has been revoked")
	}
	current.Name = frozen.Name
	current.Type = frozen.Type
	current.Description = frozen.Description
	current.Parameters = append([]byte(nil), frozen.Parameters...)
	current.EndpointUrl = frozen.EndpointURL
	current.TimeoutMs = frozen.TimeoutMS
	return current, nil
}
