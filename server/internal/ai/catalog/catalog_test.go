package catalog

import (
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
	assertProviderIDs(t, catalog.List(ai.KindRealtime), nil)

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

	if !catalog.Has("assemblyai") || !catalog.Has("elevenlabs") {
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
