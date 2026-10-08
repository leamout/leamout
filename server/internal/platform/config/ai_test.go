package config

import "testing"

func TestPlatformAICredentials(t *testing.T) {
	t.Parallel()

	cfg := Config{
		AIPlatformCredentials: `{"openai":" secret ","deepgram":"token"}`,
	}
	values, err := cfg.PlatformAICredentials()
	if err != nil {
		t.Fatalf("PlatformAICredentials() error = %v", err)
	}
	if values["openai"] != "secret" {
		t.Fatalf("OpenAI credential = %q, want secret", values["openai"])
	}
	if values["deepgram"] != "token" {
		t.Fatalf("Deepgram credential = %q, want token", values["deepgram"])
	}
}

func TestPlatformAICredentialsOptional(t *testing.T) {
	t.Parallel()

	values, err := (Config{}).PlatformAICredentials()
	if err != nil {
		t.Fatalf("PlatformAICredentials() error = %v", err)
	}
	if len(values) != 0 {
		t.Fatalf("PlatformAICredentials() = %v, want empty", values)
	}
}

func TestPlatformAICredentialsRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := (Config{
		AIPlatformCredentials: `{"openai":`,
	}).PlatformAICredentials()
	if err == nil {
		t.Fatal("expected invalid AI_PLATFORM_CREDENTIALS error")
	}
}

func TestPlatformAICredentialsRejectsEmptySecret(t *testing.T) {
	t.Parallel()

	_, err := (Config{
		AIPlatformCredentials: `{"openai":" "}`,
	}).PlatformAICredentials()
	if err == nil {
		t.Fatal("expected empty platform credential error")
	}
}
