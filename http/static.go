package http

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed static
var staticFS embed.FS

func Static(r chi.Router) {
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServerFS(static)))
	r.Handle("/favicon.ico", http.FileServerFS(staticFS))
}
