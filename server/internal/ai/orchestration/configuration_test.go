package orchestration

import (
	"testing"

	"github.com/coffeyvidzro/monogo/internal/database/sqlc"
	"github.com/coffeyvidzro/monogo/internal/media/session"
)

func TestMediaConfigFromSessionNormalizesLegacyEngineSnapshot(t *testing.T) {
	record := sqlc.VoiceAgentSession{
		ConfigurationRevision: 7,
		ConfigurationSnapshot: []byte(`{
			"engine":"integrated",
			"instructions":"snapshot instructions",
			"voice":"voice-a",
			"language":"en",
			"engine_config":{"model":"realtime-test"},
			"interruption_policy":"allow",
			"recording_policy":"none",
			"providers":[{"role":"realtime","provider":"openai"}],
			"tools":[]
		}`),
	}
	got := MediaConfigFromSession(record, session.Config{
		Engine:       session.EngineEcho,
		Instructions: "mutable current value",
		Voice:        "voice-b",
		Language:     "fr",
	})

	if got.Engine != session.EngineRealtime {
		t.Fatalf("engine = %q", got.Engine)
	}
	if string(got.EngineConfig) != `{"model":"realtime-test"}` {
		t.Fatalf("engine config = %s", got.EngineConfig)
	}
	if got.Instructions != "snapshot instructions" {
		t.Fatalf("instructions = %q", got.Instructions)
	}
	if got.Voice != "voice-a" {
		t.Fatalf("voice = %q", got.Voice)
	}
	if got.Language != "en" {
		t.Fatalf("language = %q", got.Language)
	}
}
