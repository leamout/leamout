package providers

import "strings"

type PlatformCredentials struct {
	OpenAI     string
	Gemini     string
	Deepgram   string
	AssemblyAI string
	Groq       string
	Cartesia   string
	ElevenLabs string
}

func (p PlatformCredentials) Get(provider string) (string, bool) {
	var secret string
	switch strings.TrimSpace(provider) {
	case ProviderOpenAI:
		secret = p.OpenAI
	case ProviderGemini:
		secret = p.Gemini
	case ProviderDeepgram:
		secret = p.Deepgram
	case ProviderAssemblyAI:
		secret = p.AssemblyAI
	case ProviderGroq:
		secret = p.Groq
	case ProviderCartesia:
		secret = p.Cartesia
	case ProviderElevenLabs:
		secret = p.ElevenLabs
	default:
		return "", false
	}
	secret = strings.TrimSpace(secret)
	return secret, secret != ""
}

func (p PlatformCredentials) Has(provider string) bool {
	_, ok := p.Get(provider)
	return ok
}
