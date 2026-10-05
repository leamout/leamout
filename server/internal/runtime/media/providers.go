package media

import (
	aicatalog "github.com/coffeyvidzro/monogo/internal/ai/catalog"
	"github.com/coffeyvidzro/monogo/internal/integrations/openai"
	providersdk "github.com/coffeyvidzro/monogo/internal/providers"
)

func builtInProviderCatalog() (*aicatalog.Catalog, error) {
	return aicatalog.Builtins()
}

// builtInRealtimeProviderRegistry keeps the existing integrated OpenAI path
// isolated while the composable engine moves to the external provider modules.
// It can be removed when realtime adapters move to github.com/leamout/ai-providers.
func builtInRealtimeProviderRegistry(cfg Config) (*providersdk.Registry, error) {
	return providersdk.NewRegistry(
		openai.Provider{
			Config: openai.Config{
				Endpoint: cfg.OpenAIEndpoint,
			},
		},
	)
}
