package http

import (
	"net/http"

	"consensus/app"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	g "maragu.dev/gomponents"
	ghttp "maragu.dev/gomponents/http"
)

// provideRepo passes the repository interface, and wraps the returned output into
// a http.Handler with gomponents http.Adapt. This provides the calling functions
// to only respond with Gomponents node, and a fairly easy way to return HTTP errors.
func provideRepo(
	s app.Repository,
	fn func(http.ResponseWriter, *http.Request, app.Repository) (g.Node, error),
) http.HandlerFunc {
	return ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (g.Node, error) {
		return fn(w, r, s)
	})
}

func (s *Server) setupRoutes() {
	s.mux.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(middleware.RequestID)
		r.Use(middleware.Recoverer)

		// TODO: Add cache busting middleware
		Static(r)

		Health(r)

		Index(r, s.repository)
		Ticket(r, s.repository)
	})
}
