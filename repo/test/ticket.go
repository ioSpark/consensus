package test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"consensus/app"
)

func createUser(t *testing.T, repo app.Repository, name string) *app.UserID {
	t.Helper()

	u := app.NewUser(name)
	err := repo.CreateUser(u)
	if err != nil {
		t.Fatalf("create user '%s' failed: %v", name, err)
	}

	return &u
}

func createTicket(
	t *testing.T,
	repo app.Repository,
	name string,
	user app.UserID,
) *app.Ticket {
	t.Helper()

	ticket, err := repo.CreateTicket(
		app.NewTicket(name, "http://whatever"+name, user),
	)
	if err != nil {
		t.Errorf("CreateTicket failed: %v", err)
	}

	return &ticket
}

func init() {
	registerRepoTest("TicketCreateUserNotExist", testTicketCreateUserNotExist)
	registerRepoTest("TicketCreateGenerateUniqueIDs", testTicketCreateGenerateUniqueIDs)
	registerRepoTest("TicketNoLeakage", testTicketNoLeakage)
	registerRepoTest("TicketUpdateNonExistent", testTicketUpdateNonExistent)
	registerRepoTest("TicketUpdateUserNotExist", testTicketUpdateUserNotExist)
	registerRepoTest("TicketDeleteNonExistent", testTicketDeleteNonExistent)
	registerRepoTest("TicketCRUD", testTicketCRUD)
	registerRepoTest("TicketCreateDuplicate", testTicketCreateDuplicate)
}

func testTicketCRUD(t *testing.T, repo app.Repository) {
	if len(repo.Tickets()) != 0 {
		t.Fatalf("expected 0 tickets: got %d", len(repo.Tickets()))
	}

	user := createUser(t, repo, "test")

	t1 := createTicket(t, repo, "i am the first ticket", *user)
	_ = createTicket(t, repo, "i am the second ticket", *user)

	if len(repo.Tickets()) != 2 {
		t.Fatalf("expected 2 tickets: got %d", len(repo.Tickets()))
	}

	fetched, err := repo.Ticket(t1.ID)
	if err != nil {
		t.Fatalf("could not fetch created ticket %d", t1.ID)
	}
	if fetched.Name != "i am the first ticket" {
		t.Errorf("expected name %s, got %s", "i am the first ticket", fetched.Name)
	}

	if !slices.ContainsFunc(repo.Tickets(), func(t app.Ticket) bool {
		return t.ID == t1.ID
	}) {
		t.Errorf("Tickets() missing created ticket: %d", t1.ID)
	}

	newName := "i am renaming myself to Ticket Prime"
	t1.Name = newName
	err = repo.UpdateTicket(*t1)
	if err != nil {
		t.Fatalf("error updating ticket with new name: %v", err)
	}

	fetched, err = repo.Ticket(t1.ID)
	if err != nil {
		t.Fatalf("could not get updated ticket: %v", err)
	}
	if fetched.Name != newName {
		t.Errorf("expected name of ticket to be %s, got %s", newName, fetched.Name)
	}

	err = repo.DeleteTicket(t1.ID)
	if err != nil {
		t.Fatalf("deleting ticket failed: %v", err)
	}

	_, err = repo.Ticket(t1.ID)
	if err == nil {
		t.Fatal("expected non-existent ticket to fail")
	} else if !errors.Is(err, app.ErrTicketNotExist) {
		t.Fatalf("expected ErrTicketNotExist, got %v", err)
	}
}

func testTicketCreateUserNotExist(t *testing.T, repo app.Repository) {
	t1 := app.NewTicket("test", "test", app.NewUser("1"))
	_, err := repo.CreateTicket(t1)
	if err == nil {
		t.Errorf("expected create ticket to fail")
	} else if !errors.Is(err, app.ErrUserNotExist) {
		t.Errorf("expected ErrUserNotExist, got %v", err)
	}
}

// TODO: Characterisation test - memory implementation is limited to 8192 tickets
func testTicketCreateGenerateUniqueIDs(t *testing.T, repo app.Repository) {
	seen := make(map[int]struct{})
	user := createUser(t, repo, "test")

	const n = 8192
	for i := range n {
		num := fmt.Sprintf("%d", i)
		tk, err := repo.CreateTicket(app.NewTicket(num, "whatever-"+num, *user))
		if err != nil {
			t.Fatalf("create ticket failed: %v", err)
		}
		if _, exists := seen[tk.ID]; exists {
			t.Fatalf("duplicate ID generated: %d", tk.ID)
		}
		seen[tk.ID] = struct{}{}
	}

	if len(repo.Tickets()) != n {
		t.Fatalf("expected %d tickets, got %d", n, len(repo.Tickets()))
	}
}

func testTicketCreateDuplicate(t *testing.T, repo app.Repository) {
	user := createUser(t, repo, "1")

	newTicket1 := app.NewTicket("1", "whatever", *user)
	_, err := repo.CreateTicket(newTicket1)
	if err != nil {
		t.Errorf("CreateTicket failed: %v", err)
	}

	_, err = repo.CreateTicket(newTicket1)
	if err == nil {
		t.Errorf("expected duplicate ticket creation to fail")
	} else if !errors.Is(err, app.ErrTicketAlreadyExists) {
		t.Fatalf("expected ErrTicketAlreadyExists, got %v", err)
	}
}

func testTicketUpdateUserNotExist(t *testing.T, repo app.Repository) {
	u1 := createUser(t, repo, "1")
	t1 := app.NewTicket("test", "test", *u1)
	_, err := repo.CreateTicket(t1)
	if err != nil {
		t.Errorf("create ticket failed %v", err)
	}

	u2 := app.NewUser("2")
	t1.RaisedBy = u2

	err = repo.UpdateTicket(t1)
	if err == nil {
		t.Errorf("expected update ticket to fail")
	} else if !errors.Is(err, app.ErrUserNotExist) {
		t.Errorf("expected ErrUserNotExist, got %v", err)
	}
}

func testTicketUpdateNonExistent(t *testing.T, repo app.Repository) {
	// Create valid user, as we already test for non-existent user
	u1 := createUser(t, repo, "test")

	err := repo.UpdateTicket(app.Ticket{ID: 9999, RaisedBy: *u1})
	if err == nil {
		t.Fatal("expected non-existent ticket update to fail")
	} else if !errors.Is(err, app.ErrTicketNotExist) {
		t.Fatalf("expected ErrTicketNotExist, got %v", err)
	}
}

func testTicketDeleteNonExistent(t *testing.T, repo app.Repository) {
	err := repo.DeleteTicket(9999)
	if err == nil {
		t.Fatal("expected non-existent ticket update to fail")
	} else if !errors.Is(err, app.ErrTicketNotExist) {
		t.Fatalf("expected ErrTicketNotExist, got %v", err)
	}
}

func testTicketNoLeakage(t *testing.T, repo app.Repository) {
	user := createUser(t, repo, "test")

	t1 := createTicket(t, repo, "1", *user)
	_ = createTicket(t, repo, "2", *user)

	tickets := repo.Tickets()
	if len(tickets) != 2 {
		t.Fatalf("expected 2 tickets, got %d", len(tickets))
	}

	tickets[0].Name = "1-MODIFIED"

	refetch, err := repo.Ticket(t1.ID)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if refetch.Name != "1" {
		t.Errorf("ticket mutation leak %s, got %s", "1", refetch.Name)
	}
}
