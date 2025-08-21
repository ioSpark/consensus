package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type contextValues int

const (
	contextUser contextValues = iota
	contextTicket
)

type Server struct {
	log    *slog.Logger
	mux    chi.Router
	server *http.Server
}

type NewServerOptions struct {
	Log *slog.Logger
}

func NewServer(opts NewServerOptions) *Server {
	if opts.Log == nil {
		opts.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	mux := chi.NewRouter()

	return &Server{
		log: opts.Log,
		mux: mux,
		server: &http.Server{
			Addr:              ":8088",
			Handler:           mux,
			IdleTimeout:       5 * time.Second,
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	s.log.Info("Starting http server", "address", "http://localhost:8088")

	s.setupRoutes()

	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	s.log.Info("Stopping http server")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		return err
	}
	s.log.Info("Stopped http server")
	return nil
}
