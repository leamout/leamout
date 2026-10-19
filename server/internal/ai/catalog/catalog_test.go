package catalog

import (
	"encoding/json"
	"testing"

	"github.com/leamout/contracts/ai"
)

func TestBuiltins(t *testing.T) {
	catalog, err := Builtins()
	if err != nil {
		t.Fatal(err)
	}

	assertProviderIDs(t, catalog.List(ai.KindSTT), []string{"assemblyai", "deepgram"})
	assertProviderIDs(t, catalog.List(ai.KindLLM), []string{"groq", "openai"})
	assertProviderIDs(t, catalog.List(ai.KindTTS), []string{"cartesia", "elevenlabs"})
	assertProviderIDs(t, catalog.List(ai.KindRealtime), []string{"gemini", "openai"})

	if _, ok := catalog.STT("deepgram"); !ok {
		t.Fatal("deepgram STT provider is not registered")
	}
	if _, ok := catalog.STT("assemblyai"); !ok {
		t.Fatal("assemblyai STT provider is not registered")
	}
	if _, ok := catalog.LLM("groq"); !ok {
		t.Fatal("groq LLM provider is not registered")
	}
	if _, ok := catalog.LLM("openai"); !ok {
		t.Fatal("openai LLM provider is not registered")
	}
	if _, ok := catalog.TTS("cartesia"); !ok {
		t.Fatal("cartesia TTS provider is not registered")
	}
	if _, ok := catalog.TTS("elevenlabs"); !ok {
		t.Fatal("elevenlabs TTS provider is not registered")
	}
	if _, ok := catalog.Realtime("openai"); !ok {
		t.Fatal("openai realtime provider is not registered")
	}
	if _, ok := catalog.Realtime("gemini"); !ok {
		t.Fatal("gemini realtime provider is not registered")
	}

	if !catalog.Has("assemblyai") || !catalog.Has("elevenlabs") || !catalog.Has("gemini") {
		t.Fatal("catalog Has() does not include all built-in providers")
	}
	if catalog.Has("missing") {
		t.Fatal("catalog Has() accepted an unknown provider")
	}
	if _, ok := catalog.ConfigValidator(ai.KindLLM, "openai"); !ok {
		t.Fatal("OpenAI config validator is unavailable")
	}
	if _, ok := catalog.CredentialVerifier("openai"); !ok {
		t.Fatal("OpenAI credential verifier is unavailable")
	}
	capabilities := catalog.Capabilities("openai")
	if len(capabilities) == 0 {
		t.Fatal("OpenAI capabilities are empty")
	}
	capabilities[0] = "changed"
	if catalog.Capabilities("openai")[0] == "changed" {
		t.Fatal("Capabilities() reused mutable storage")
	}
}

func assertProviderIDs(t *testing.T, descriptors []ai.Descriptor, want []string) {
	t.Helper()
	if len(descriptors) != len(want) {
		t.Fatalf("provider count = %d, want %d", len(descriptors), len(want))
	}
	for i := range want {
		if descriptors[i].ID != want[i] {
			t.Fatalf("provider %d = %q, want %q", i, descriptors[i].ID, want[i])
		}
	}
}

func TestBuiltinsRejectTenantControlledProviderEndpoints(t *testing.T) {
	catalog, err := Builtins()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		kind     ai.Kind
		provider string
		endpoint string
	}{
		{name: "deepgram", kind: ai.KindSTT, provider: "deepgram", endpoint: "wss://tenant.example.test/ws"},
		{name: "assemblyai", kind: ai.KindSTT, provider: "assemblyai", endpoint: "wss://tenant.example.test/ws"},
		{name: "groq", kind: ai.KindLLM, provider: "groq", endpoint: "https://tenant.example.test/v1/chat"},
		{name: "openai-llm", kind: ai.KindLLM, provider: "openai", endpoint: "https://tenant.example.test/v1/chat"},
		{name: "cartesia", kind: ai.KindTTS, provider: "cartesia", endpoint: "wss://tenant.example.test/ws"},
		{name: "elevenlabs", kind: ai.KindTTS, provider: "elevenlabs", endpoint: "wss://tenant.example.test/ws"},
		{name: "openai-realtime", kind: ai.KindRealtime, provider: "openai", endpoint: "wss://tenant.example.test/ws"},
		{name: "gemini-realtime", kind: ai.KindRealtime, provider: "gemini", endpoint: "wss://tenant.example.test/ws"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validator, ok := catalog.ConfigValidator(test.kind, test.provider)
			if !ok {
				t.Fatalf("%s provider validator is unavailable", test.provider)
			}
			if err := validator.ValidateConfig(json.RawMessage(`{}`)); err != nil {
				t.Fatalf("default provider config was rejected: %v", err)
			}

			config, err := json.Marshal(map[string]any{
				"endpoint":          test.endpoint,
				"allowed_endpoints": []string{test.endpoint},
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := validator.ValidateConfig(config); err == nil {
				t.Fatal("tenant-controlled endpoint was accepted by built-in provider")
			}
		})
	}
}
