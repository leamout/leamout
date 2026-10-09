package providers

import "testing"

func TestPlatformCredentials(t *testing.T) {
	t.Parallel()

	values := PlatformCredentials{
		OpenAI:   " platform-secret ",
		Deepgram: "deepgram-secret",
	}

	secret, ok := values.Get(ProviderOpenAI)
	if !ok {
		t.Fatal("expected OpenAI platform credential")
	}
	if secret != "platform-secret" {
		t.Fatalf("Get(openai) = %q, want platform-secret", secret)
	}

	secret, ok = values.Get(ProviderDeepgram)
	if !ok || secret != "deepgram-secret" {
		t.Fatalf("Get(deepgram) = %q, %v", secret, ok)
	}

	if values.Has(ProviderCartesia) {
		t.Fatal("unset Cartesia credential should not be configured")
	}
}
