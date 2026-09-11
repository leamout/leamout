package payments

import "sync"

// ProviderRegistry contains runtime-supplied collection adapters. Commercial
// depends on the Provider port, never on Stripe or Paystack implementations.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewProviderRegistry(initial ...map[string]Provider) *ProviderRegistry {
	registry := &ProviderRegistry{providers: make(map[string]Provider)}
	if len(initial) > 0 {
		for name, provider := range initial[0] {
			registry.Set(name, provider)
		}
	}
	return registry
}

func (r *ProviderRegistry) Set(name string, provider Provider) {
	if provider == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

func (r *ProviderRegistry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, ok := r.providers[name]
	return provider, ok
}
