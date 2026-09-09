// Package assets embeds the Backoffice public asset tree.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed public
var publicAssets embed.FS

func Public() fs.FS {
	public, err := fs.Sub(publicAssets, "public")
	if err != nil {
		panic(err)
	}
	return public
}
