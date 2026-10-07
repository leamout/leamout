package media

import aicatalog "github.com/leamout/leamout/server/internal/ai/catalog"

func builtInProviderCatalog() (*aicatalog.Catalog, error) {
	return aicatalog.Builtins()
}
