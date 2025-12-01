package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"consensus/app"
	"consensus/http"
	"consensus/repo/bbolt"
	"consensus/repo/memory"

	"golang.org/x/sync/errgroup"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	repo, err := cli(log)
	if err != nil {
		log.Error("Error parsing CLI", "error", err)
		os.Exit(1)
	}

	err = start(log, repo)
	if err != nil {
		log.Error("Error starting app", "error", err)
		os.Exit(1)
	}
}

func cli(log *slog.Logger) (app.Repository, error) {
	var repoType string
	flag.StringVar(
		&repoType,
		"repo",
		"",
		"Storage repo to use (memory, bbolt). If not specified, uses $REPO_TYPE. Default: memory",
	)

	var storageDir string
	flag.StringVar(
		&storageDir,
		"storage",
		"",
		"Directory to store bbolt database. If not specified, uses $STORAGE_DIR",
	)

	flag.Parse()

	if repoType == "" {
		repoType = os.Getenv("REPO_TYPE")
		if repoType == "" {
			repoType = "memory"
		}
	}
	repoType = strings.ToLower(repoType)

	if repoType == "" {
		return nil, fmt.Errorf("--repo or REPO_TYPE not specified")
	} else if repoType != "memory" && repoType != "bbolt" {
		return nil, fmt.Errorf("invalid repo type, expected 'bbolt' or 'memory' got: %s", repoType)
	}

	if storageDir == "" {
		storageDir = os.Getenv("STORAGE_DIR")
	}
	if storageDir == "" && repoType == "bbolt" {
		return nil, fmt.Errorf("bbolt repo type specified, but no storage dir given")
	}

	var repo app.Repository
	switch repoType {
	case "memory":
		log.Info("Using in-memory repository")
		repo = memory.NewRepository()
	case "bbolt":
		dbPath := filepath.Join(storageDir, "consensus.db")
		log.Info("Using bbolt DB", "path", dbPath)

		bboltRepo, err := bbolt.NewRepository(dbPath, bbolt.RepositoryOptions{Log: log})
		if err != nil {
			return nil, err
		}

		log.Debug("Initialising bbolt DB")
		err = bboltRepo.Initialise()
		if err != nil {
			return nil, fmt.Errorf("intialising bbolt repo: %w", err)
		}

		repo = bboltRepo
	}

	return repo, nil
}

func start(log *slog.Logger, repo app.Repository) error {
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
		Repository: repo,
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
