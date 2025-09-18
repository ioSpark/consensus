package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"consensus/http"
	"consensus/repo/memory"

	"golang.org/x/sync/errgroup"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	err := start(log)
	if err != nil {
		log.Error("Error starting app", "error", err)
		os.Exit(1)
	}
}

func start(log *slog.Logger) error {
	log.Info("Starting")

	// Graceful shutfown
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	server := http.NewServer(http.NewServerOptions{
		Log:        log,
		Repository: memory.NewRepository(),
	})

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return server.Start()
	})

	<-ctx.Done()
	log.Info("Shutting down")

	// Stop Gracefully
	eg.Go(func() error {
		return server.Stop()
	})

	err := eg.Wait()
	if err != nil {
		return err
	}

	return nil
}
