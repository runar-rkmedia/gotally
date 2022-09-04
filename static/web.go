package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:static
var webDir embed.FS

func StaticWebHandler() http.Handler {

	fsys := fs.FS(webDir)
	html, _ := fs.Sub(fsys, "static")
	return http.FileServer(http.FS(html))
}
