package providers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestUpsertBindingRejectsUnapprovedProviderEndpoints(t *testing.T) {
	service := NewService(nil, nil)

	tests := []struct {
		name     string
		role     string
		provider string
		endpoint string
	}{
		{name: "deepgram", role: RoleSTT, provider: ProviderDeepgram, endpoint: "wss://tenant.example.test/ws"},
		{name: "assemblyai", role: RoleSTT, provider: ProviderAssemblyAI, endpoint: "wss://tenant.example.test/ws"},
		{name: "groq", role: RoleLLM, provider: ProviderGroq, endpoint: "https://tenant.example.test/v1/chat"},
		{name: "openai-llm", role: RoleLLM, provider: ProviderOpenAI, endpoint: "https://tenant.example.test/v1/chat"},
		{name: "cartesia", role: RoleTTS, provider: ProviderCartesia, endpoint: "wss://tenant.example.test/ws"},
		{name: "elevenlabs", role: RoleTTS, provider: ProviderElevenLabs, endpoint: "wss://tenant.example.test/ws"},
		{name: "openai-realtime", role: RoleRealtime, provider: ProviderOpenAI, endpoint: "wss://tenant.example.test/ws"},
		{name: "gemini-realtime", role: RoleRealtime, provider: ProviderGemini, endpoint: "wss://tenant.example.test/ws"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			credentialID := uuid.New()
			config, err := json.Marshal(map[string]any{
				"endpoint":          test.endpoint,
				"allowed_endpoints": []string{test.endpoint},
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = service.UpsertBinding(
				context.Background(),
				uuid.New(),
				uuid.New(),
				test.role,
				UpsertBindingRequest{
					Provider:     test.provider,
					CredentialID: &credentialID,
					Config:       config,
				},
			)
			if err == nil || !strings.Contains(err.Error(), "not approved") {
				t.Fatalf("UpsertBinding() error = %v, want endpoint rejection", err)
			}
		})
	}
}
