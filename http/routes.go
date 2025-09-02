package http

import (
	"net/http"

	"consensus/app"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func provideStorage(
	s app.Storage,
	fn func(http.ResponseWriter, *http.Request, app.Storage),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, s)
	}
}

func (s *Server) setupRoutes() {
	s.mux.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(middleware.RequestID)
		r.Use(middleware.Recoverer)

		// TODO: Add cache busting middleware
		Static(r)

		Health(r)

		Index(r, s.storage)
		Ticket(r, s.storage)
	})
}
