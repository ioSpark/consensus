package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"consensus/http"

	"golang.org/x/sync/errgroup"
)

// TODO: If user goes "back" the page does not refresh
// TODO: The page is not live, when people vote, when new people login, when ticket is revealed
// TODO: (perhaps) we cache the github avatars
// TODO: New ticket section does not expand when table gets wider
// TODO: There is no "view only" page. i.e. so that the scrum master can vote, but not reveal their own votes
// TODO: Question type.
// TODO: voters is not padded. i.e. if we live-refresh and a user logs in, it could move buttons
// TODO: tooltip over avg/mean to explain how they work.
// TODO: Add tests for all kinds of characters.
// TODO: A way to clear/remove tickets
// TODO: A way to edit tickets
// TODO: Simple way to test without oauth2 proxy

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
		Log: log,
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
