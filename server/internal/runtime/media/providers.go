package media

import aicatalog "github.com/coffeyvidzro/monogo/internal/ai/catalog"

func builtInProviderCatalog() (*aicatalog.Catalog, error) {
	return aicatalog.Builtins()
}
