// Package packages defines the portable agent-package contract consumed by Leamout.
package packages

import "encoding/json"

const (
	SchemaVersionV1 = "1"

	EngineComposable = "composable"
	EngineRealtime   = "realtime"

	RoleRealtime = "realtime"
	RoleSTT      = "stt"
	RoleLLM      = "llm"
	RoleTTS      = "tts"
)

// Manifest is a portable, credential-free description of one deployable agent package.
type Manifest struct {
	SchemaVersion string             `json:"schema_version"`
	Name          string             `json:"name"`
	Agents        []AgentDefinition  `json:"agents"`
	Tools         []ToolDefinition   `json:"tools,omitempty"`
	Routing       *RoutingDefinition `json:"routing,omitempty"`
}

// AgentDefinition describes one runtime voice agent without tenant-specific IDs.
type AgentDefinition struct {
	Alias              string            `json:"alias"`
	Name               string            `json:"name"`
	Engine             string            `json:"engine"`
	Instructions       string            `json:"instructions"`
	Voice              string            `json:"voice,omitempty"`
	Language           string            `json:"language,omitempty"`
	EngineConfig       json.RawMessage   `json:"engine_config,omitempty"`
	Providers          []ProviderBinding `json:"providers"`
	Tools              []string          `json:"tools,omitempty"`
	InterruptionPolicy string            `json:"interruption_policy,omitempty"`
	RecordingPolicy    string            `json:"recording_policy,omitempty"`
}

// ProviderBinding identifies a provider capability and public provider config.
// Credential or integration IDs are intentionally resolved at deployment time.
type ProviderBinding struct {
	Role     string          `json:"role"`
	Provider string          `json:"provider"`
	Config   json.RawMessage `json:"config,omitempty"`
}

// ToolDefinition is a reusable tool definition referenced by agent aliases.
type ToolDefinition struct {
	Alias       string          `json:"alias"`
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

// RoutingDefinition selects an agent alias for an inbound package deployment.
type RoutingDefinition struct {
	DefaultAgent string        `json:"default_agent"`
	Rules        []RoutingRule `json:"rules,omitempty"`
}

// RoutingRule keeps the condition payload opaque until the telephony routing
// policy language is formalized. The package layer still validates the target.
type RoutingRule struct {
	Agent string          `json:"agent"`
	When  json.RawMessage `json:"when"`
}
