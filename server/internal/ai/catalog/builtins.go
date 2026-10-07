package catalog

import providercatalog "github.com/leamout/ai-providers/catalog"

// Builtins returns the provider catalog shipped with the Leamout Agent Runtime.
func Builtins() (*Catalog, error) {
	return New(providercatalog.Providers()...)
}
