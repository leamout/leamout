package providers

import "testing"

func TestPlatformCredentials(t *testing.T) {
	t.Parallel()

	values := NewPlatformCredentials(map[string]string{
		" openai ": " platform-secret ",
		"empty":    " ",
		"":         "ignored",
	})

	secret, ok := values.Get("openai")
	if !ok {
		t.Fatal("expected OpenAI platform credential")
	}
	if secret != "platform-secret" {
		t.Fatalf("Get(openai) = %q, want platform-secret", secret)
	}
	if values.Has("empty") {
		t.Fatal("empty platform credential should not be configured")
	}
	if values.Has("missing") {
		t.Fatal("missing platform credential should not be configured")
	}
}
