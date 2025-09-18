package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// TODO: Try to do something. Perhaps use chi health probe feature? #2
func Health(r chi.Router) {
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}
