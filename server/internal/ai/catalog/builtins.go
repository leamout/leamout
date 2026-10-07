package catalog

import (
	"github.com/leamout/ai-providers/assemblyai"
	"github.com/leamout/ai-providers/cartesia"
	"github.com/leamout/ai-providers/deepgram"
	"github.com/leamout/ai-providers/elevenlabs"
	"github.com/leamout/ai-providers/gemini"
	"github.com/leamout/ai-providers/groq"
	"github.com/leamout/ai-providers/openai"
)

// Builtins returns the provider catalog shipped with the Leamout Agent Runtime.
func Builtins() (*Catalog, error) {
	return New(
		deepgram.Provider{},
		assemblyai.Provider{},
		groq.Provider{},
		openai.Provider{},
		openai.RealtimeProvider{},
		cartesia.Provider{},
		elevenlabs.Provider{},
		gemini.RealtimeProvider{},
	)
}
