// Package orchestration coordinates durable AI state with realtime media execution.
package orchestration

import (
	"github.com/leamout/leamout/server/internal/ai/agents"
	"github.com/leamout/leamout/server/internal/ai/conversations"
	"github.com/leamout/leamout/server/internal/ai/providers"
	"github.com/leamout/leamout/server/internal/ai/tools"
)

type Service struct {
	agents        *agents.Service
	conversations *conversations.Service
	tools         *tools.Executor
	providers     *providers.Service
}

func NewService(
	agentService *agents.Service,
	conversationService *conversations.Service,
	executor *tools.Executor,
	providerServices ...*providers.Service,
) *Service {
	var providerService *providers.Service
	if len(providerServices) > 0 {
		providerService = providerServices[0]
	}
	return &Service{
		agents:        agentService,
		conversations: conversationService,
		tools:         executor,
		providers:     providerService,
	}
}
