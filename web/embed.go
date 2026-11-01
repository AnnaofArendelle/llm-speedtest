// Package web embeds the compiled single-page application.
package web

import (
	"embed"
	"io/fs"
)

// dist holds the Vite build output. The directory always exists in the
// repository (web/dist/.gitkeep) so that the Go toolchain can compile even
// when the frontend has not been built; the HTTP server falls back to a
// placeholder page in that case.
//
//go:embed all:dist
var dist embed.FS

// Dist returns the embedded frontend file system rooted at the build output.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}

// Built reports whether a real frontend build is embedded.
func Built() bool {
	_, err := fs.Stat(Dist(), "index.html")
	return err == nil
}
