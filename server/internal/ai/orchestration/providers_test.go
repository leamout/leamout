package orchestration

import (
	"strings"
	"testing"

	"github.com/leamout/leamout/server/internal/media/session"
)

func TestValidateProviderTopologyReportsMissingBinding(t *testing.T) {
	tests := []struct {
		name      string
		engine    session.Engine
		providers []session.ProviderRuntime
		want      string
	}{
		{
			name:   "realtime",
			engine: session.EngineRealtime,
			want:   "Voice Agent requires a realtime provider binding",
		},
		{
			name:   "stt",
			engine: session.EngineComposable,
			want:   "Voice Agent requires an stt provider binding",
		},
		{
			name:   "llm",
			engine: session.EngineComposable,
			providers: []session.ProviderRuntime{
				{
					Role:     "stt",
					Provider: "deepgram",
				},
			},
			want: "Voice Agent requires an llm provider binding",
		},
		{
			name:   "tts",
			engine: session.EngineComposable,
			providers: []session.ProviderRuntime{
				{
					Role:     "stt",
					Provider: "assemblyai",
				},
				{
					Role:     "llm",
					Provider: "openai",
				},
			},
			want: "Voice Agent requires an tts provider binding",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateProviderTopology(test.providers, test.engine)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validateProviderTopology() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidateProviderTopologyAcceptsCompleteBindings(t *testing.T) {
	tests := []struct {
		name      string
		engine    session.Engine
		providers []session.ProviderRuntime
	}{
		{
			name:   "realtime",
			engine: session.EngineRealtime,
			providers: []session.ProviderRuntime{
				{
					Role:     "realtime",
					Provider: "openai",
				},
			},
		},
		{
			name:   "composable",
			engine: session.EngineComposable,
			providers: []session.ProviderRuntime{
				{
					Role:     "stt",
					Provider: "assemblyai",
				},
				{
					Role:     "llm",
					Provider: "openai",
				},
				{
					Role:     "tts",
					Provider: "elevenlabs",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateProviderTopology(test.providers, test.engine); err != nil {
				t.Fatalf("validateProviderTopology() error = %v", err)
			}
		})
	}
}
