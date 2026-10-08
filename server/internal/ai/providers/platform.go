package providers

import "strings"

type PlatformCredentials map[string]string

func NewPlatformCredentials(values map[string]string) PlatformCredentials {
	result := make(PlatformCredentials, len(values))
	for provider, secret := range values {
		provider = strings.TrimSpace(provider)
		secret = strings.TrimSpace(secret)
		if provider == "" || secret == "" {
			continue
		}
		result[provider] = secret
	}
	return result
}

func (p PlatformCredentials) Get(provider string) (string, bool) {
	if len(p) == 0 {
		return "", false
	}
	secret, ok := p[strings.TrimSpace(provider)]
	return secret, ok && strings.TrimSpace(secret) != ""
}

func (p PlatformCredentials) Has(provider string) bool {
	_, ok := p.Get(provider)
	return ok
}
