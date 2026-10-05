package providers

import (
	"encoding/json"
	"testing"

	aicatalog "github.com/coffeyvidzro/monogo/internal/ai/catalog"
	"github.com/leamout/contracts/ai"
)

func TestValidateProviderConfig(t *testing.T) {
	catalog, err := aicatalog.Builtins()
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(nil, nil, catalog)

	tests := []struct {
		name     string
		kind     ai.Kind
		provider string
		config   json.RawMessage
		wantErr  bool
	}{
		{
			name:     "Deepgram",
			kind:     ai.KindSTT,
			provider: ProviderDeepgram,
			config:   json.RawMessage(`{"model":"flux-general-en"}`),
		},
		{
			name:     "AssemblyAI",
			kind:     ai.KindSTT,
			provider: ProviderAssemblyAI,
			config:   json.RawMessage(`{"speech_model":"universal-streaming-english"}`),
		},
		{
			name:     "Groq",
			kind:     ai.KindLLM,
			provider: ProviderGroq,
			config:   json.RawMessage(`{"model":"llama","temperature":0.3}`),
		},
		{
			name:     "OpenAI",
			kind:     ai.KindLLM,
			provider: ProviderOpenAI,
			config:   json.RawMessage(`{"model":"gpt-4.1-mini"}`),
		},
		{
			name:     "Cartesia",
			kind:     ai.KindTTS,
			provider: ProviderCartesia,
			config:   json.RawMessage(`{"model":"sonic-3","voice_id":"voice"}`),
		},
		{
			name:     "ElevenLabs",
			kind:     ai.KindTTS,
			provider: ProviderElevenLabs,
			config:   json.RawMessage(`{"model":"eleven_flash_v2_5","voice_id":"voice"}`),
		},
		{
			name:     "unknown field",
			kind:     ai.KindLLM,
			provider: ProviderOpenAI,
			config:   json.RawMessage(`{"secret":"not-allowed"}`),
			wantErr:  true,
		},
		{
			name:     "wrong type",
			kind:     ai.KindLLM,
			provider: ProviderGroq,
			config:   json.RawMessage(`{"temperature":"hot"}`),
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.validateProviderConfig(test.kind, test.provider, test.config)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateProviderConfig() error = %v, wantErr = %v", err, test.wantErr)
			}
		})
	}
}
