package providers

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/contracts/ai"
)

const (
	ProviderOpenAI     = "openai"
	ProviderDeepgram   = "deepgram"
	ProviderAssemblyAI = "assemblyai"
	ProviderGroq       = "groq"
	ProviderCartesia   = "cartesia"
	ProviderElevenLabs = "elevenlabs"

	RoleRealtime = string(ai.KindRealtime)
	RoleSTT      = string(ai.KindSTT)
	RoleLLM      = string(ai.KindLLM)
	RoleTTS      = string(ai.KindTTS)

	ConnectionUnchecked   = "unchecked"
	ConnectionReady       = "ready"
	ConnectionInvalid     = "invalid"
	ConnectionUnavailable = "unavailable"
)

type Credential struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	Provider        string
	Name            string
	ConnectionState string
	VerifiedAt      *time.Time
	FailureCode     *string
	CreatedAt       time.Time
	RotatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateCredentialRequest struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	Secret   string `json:"secret"`
}

type RotateCredentialRequest struct {
	Secret string `json:"secret"`
}

type CredentialResponse struct {
	ID              uuid.UUID   `json:"id"`
	OrganizationID  uuid.UUID   `json:"organization_id"`
	Provider        string      `json:"provider"`
	Name            string      `json:"name"`
	ConnectionState string      `json:"connection_state"`
	VerifiedAt      *time.Time  `json:"verified_at,omitempty"`
	FailureCode     *string     `json:"failure_code,omitempty"`
	Capabilities    []string    `json:"capabilities"`
	VoiceAgentIDs   []uuid.UUID `json:"voice_agent_ids"`
	CreatedAt       time.Time   `json:"created_at"`
	RotatedAt       time.Time   `json:"rotated_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type VerificationResponse struct {
	ID              uuid.UUID  `json:"id"`
	ConnectionState string     `json:"connection_state"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	FailureCode     *string    `json:"failure_code,omitempty"`
}

type Binding struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	VoiceAgentID   uuid.UUID
	Role           string
	Provider       string
	CredentialID   uuid.UUID
	Config         json.RawMessage
}

type UpsertBindingRequest struct {
	Provider     string          `json:"provider"`
	CredentialID uuid.UUID       `json:"credential_id"`
	Config       json.RawMessage `json:"config,omitempty"`
}

type BindingResponse struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	VoiceAgentID   uuid.UUID       `json:"voice_agent_id"`
	Role           string          `json:"role"`
	Provider       string          `json:"provider"`
	CredentialID   uuid.UUID       `json:"credential_id"`
	Config         json.RawMessage `json:"config"`
}

type ResolvedBinding struct {
	Role     string
	Provider string
	APIKey   string
	Config   json.RawMessage
}

type Integration struct {
	Credential
	VoiceAgentIDs []uuid.UUID
}

type BindingStatus struct {
	Role            string
	Provider        string
	IntegrationID   uuid.UUID
	ConnectionState string
	FailureCode     *string
	Config          json.RawMessage
}

func bindingResponse(value Binding) BindingResponse {
	return BindingResponse(value)
}
