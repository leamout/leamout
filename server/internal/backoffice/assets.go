package backoffice

import (
	"embed"
	"io/fs"
)

// publicAssets contains the generated frontend bundle so the backoffice binary
// can serve its UI independently of its working directory.
//
//go:embed assets/public
var publicAssets embed.FS

func PublicAssets() fs.FS {
	public, err := fs.Sub(publicAssets, "assets/public")
	if err != nil {
		panic(err)
	}
	return public
}
