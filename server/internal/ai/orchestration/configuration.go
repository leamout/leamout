package orchestration

import (
	"encoding/json"

	"github.com/coffeyvidzro/monogo/internal/database/sqlc"
	"github.com/coffeyvidzro/monogo/internal/media/session"
)

type configurationSnapshot struct {
	Engine             session.Engine  `json:"engine"`
	Instructions       string          `json:"instructions"`
	Voice              *string         `json:"voice"`
	Language           string          `json:"language"`
	EngineConfig       json.RawMessage `json:"engine_config"`
	InterruptionPolicy string          `json:"interruption_policy"`
	RecordingPolicy    string          `json:"recording_policy"`
	Providers          json.RawMessage `json:"providers"`
	Tools              json.RawMessage `json:"tools"`
}

func MediaConfig(agent sqlc.VoiceAgent, config session.Config) session.Config {
	config.Engine = session.Engine(agent.Engine)
	config.Instructions = agent.Instructions
	config.EngineConfig = json.RawMessage(append([]byte(nil), agent.EngineConfig...))
	if agent.Voice != nil {
		config.Voice = *agent.Voice
	}
	config.Language = agent.Language
	return config
}

// MediaConfigFromSession builds live media configuration only from the durable
// session snapshot. Editing an agent after answer must not change an active call.
func MediaConfigFromSession(record sqlc.VoiceAgentSession, config session.Config) session.Config {
	var snapshot configurationSnapshot
	if err := json.Unmarshal(record.ConfigurationSnapshot, &snapshot); err != nil {
		return config
	}
	config.Engine = snapshot.Engine
	config.Instructions = snapshot.Instructions
	config.EngineConfig = append(json.RawMessage(nil), snapshot.EngineConfig...)
	if snapshot.Voice != nil {
		config.Voice = *snapshot.Voice
	}
	config.Language = snapshot.Language
	return config
}

func decodeConfigurationSnapshot(value []byte) (configurationSnapshot, error) {
	var snapshot configurationSnapshot
	err := json.Unmarshal(value, &snapshot)
	return snapshot, err
}
