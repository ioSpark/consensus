package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) setupRoutes() {
	s.mux.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(middleware.RequestID)
		r.Use(middleware.Recoverer)

		// TODO: Add cache busting middleware
		Static(r)

		Health(r)

		Index(r)
		Ticket(r)
	})
}
