package manifest

import (
	"encoding/json"
	"testing"

	agentcontract "github.com/leamout/contracts/agent"
	"github.com/leamout/contracts/ai"
)

func TestValidatorAcceptsOfficialGeminiRealtime(t *testing.T) {
	t.Parallel()

	validator := NewValidator()
	err := validator.Validate(Package{
		Agent: agentcontract.Manifest{
			SchemaVersion: agentcontract.SchemaVersion,
			Name:          "Gemini",
			Engine:        agentcontract.EngineRealtime,
			Instructions:  "Answer the caller.",
			Providers: []agentcontract.ProviderBinding{
				{
					Role:     ai.KindRealtime,
					Provider: "gemini",
					Config:   json.RawMessage(`{"model":"gemini-3.8-live"}`),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidatorRejectsUnknownToolReference(t *testing.T) {
	t.Parallel()

	validator := NewValidator()
	err := validator.Validate(Package{
		Agent: agentcontract.Manifest{
			SchemaVersion: agentcontract.SchemaVersion,
			Name:          "Composable",
			Engine:        agentcontract.EngineComposable,
			Instructions:  "Answer the caller.",
			Providers: []agentcontract.ProviderBinding{
				{Role: ai.KindSTT, Provider: "deepgram"},
				{Role: ai.KindLLM, Provider: "groq"},
				{Role: ai.KindTTS, Provider: "cartesia"},
			},
			Tools: []string{"missing_tool"},
		},
	})
	if err == nil {
		t.Fatal("Validate() error = nil, want missing tool error")
	}
}

func TestResolveValue(t *testing.T) {
	t.Parallel()

	got, err := resolveValue("${DIRECTORY_WEBHOOK_URL}", map[string]string{
		"DIRECTORY_WEBHOOK_URL": "https://example.com/directory",
	})
	if err != nil {
		t.Fatalf("resolveValue() error = %v", err)
	}
	if got != "https://example.com/directory" {
		t.Fatalf("resolveValue() = %q", got)
	}
}
