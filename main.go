package main

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"consensus/app"
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
// TODO: Review and use more idiomatic error handling
// TODO: Show errors to users, instead of failing silently
// TODO: Change some variable/struct arg/field names (e.g. ticket name -> title,
//	link -> URL)
// TODO: Generally remove all of these TODO's
// TODO: URL validation

// TODO: Unsafe, temporary in-memory "storage" layer. Could be coded better
type storage struct {
	tickets []*app.Ticket
	users   []*app.User
}

func (s *storage) Tickets() []*app.Ticket {
	return s.tickets
}

func (s *storage) Ticket(ID int) (*app.Ticket, error) {
	for _, t := range s.Tickets() {
		if t.ID == ID {
			return t, nil
		}
	}
	return nil, app.ErrTicketNotExist
}

func (s *storage) TicketByName(name string) (*app.Ticket, error) {
	for _, t := range s.Tickets() {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, app.ErrTicketNotExist
}

func (s *storage) CreateTicket(t app.Ticket) (*app.Ticket, error) {
	// Generate un-used ID
	var newID int
	// Can be improved
	for {
		newID = rand.IntN(8192) // Enough space without IDs being unweidly
		_, err := s.Ticket(newID)
		if errors.Is(err, app.ErrTicketNotExist) {
			break
		} else if err != nil {
			panic(err)
		}
	}

	t.ID = newID

	s.tickets = append(s.tickets, &t)
	return &t, nil
}

func (s *storage) DeleteTicket(ID int) error {
	_, err := s.Ticket(ID)
	if err != nil && err == app.ErrTicketNotExist {
		panic(err)
	}

	// Not the best way to do this
	s.tickets = slices.DeleteFunc(s.tickets, func(t *app.Ticket) bool {
		return t.ID == ID
	})

	return nil
}

func (s *storage) UpdateTicket(t app.Ticket) error {
	_, err := s.Ticket(t.ID)
	if err != nil && err == app.ErrTicketNotExist {
		panic(err)
	}

	err = s.DeleteTicket(t.ID)
	if err != nil {
		panic(err)
	}

	newTicket, err := s.CreateTicket(t)
	if err != nil {
		panic(err)
	}

	newTicket.ID = t.ID

	return nil
}

func (s *storage) Users() []*app.User {
	return s.users
}

func (s *storage) User(name string) (*app.User, error) {
	for _, u := range s.Users() {
		if u.Name == name {
			return u, nil
		}
	}
	return nil, app.ErrUserNotExist
}

func (s *storage) CreateUser(u app.User) error {
	_, err := s.User(u.Name)
	if err != app.ErrUserNotExist && err != nil {
		panic(err)
	} else if err == nil {
		return app.ErrUserAlreadyExists
	}

	s.users = append(s.users, &u)
	return nil
}

func (s *storage) DeleteUser(name string) error {
	_, err := s.User(name)
	if err != nil && err == app.ErrUserNotExist {
		panic(err)
	}

	// Not the best way to do this
	s.users = slices.DeleteFunc(s.users, func(u *app.User) bool {
		return u.Name == name
	})

	return nil
}

func newStorage() *storage {
	return &storage{}
}

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

	s := newStorage()

	server := http.NewServer(http.NewServerOptions{
		Log:     log,
		Storage: s,
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
