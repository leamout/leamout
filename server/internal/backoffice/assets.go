package backoffice

import (
	"embed"
	"io/fs"
)

// staticAssets contains the generated frontend bundle so the backoffice binary
// can serve its UI independently of its working directory.
//
//go:embed assets
var staticAssets embed.FS

func StaticAssets() fs.FS {
	static, err := fs.Sub(staticAssets, "assets")
	if err != nil {
		panic(err)
	}
	return static
}
