package ai

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/server/internal/ai/agents"
	"github.com/leamout/leamout/server/internal/ai/providers"
	"github.com/leamout/leamout/server/internal/ai/tools"
)

func RegisterRoutes(router chi.Router, module *Module, organizationAccess func(string) func(http.Handler) http.Handler) {
	agents.RegisterRoutes(router, module.Agents.Handler, organizationAccess("voice-agents"))
	tools.RegisterRoutes(router, module.Tools.Handler, organizationAccess("voice-agents"))
	providers.RegisterRoutes(router, module.Providers.Handler, organizationAccess("voice-agents"))
}
