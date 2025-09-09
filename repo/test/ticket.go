package test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"consensus/app"
)

func createTicket(t *testing.T, repo app.Repository, name string) *app.Ticket {
	t.Helper()

	ticket, err := repo.CreateTicket(
		app.NewTicket(name, "http://whatever"+name, app.NewUser("test")),
	)
	if err != nil {
		t.Errorf("CreateTicket failed: %v", err)
	}

	return &ticket
}

func init() {
	registerRepoTest("TicketCreateGenerateUniqueIDs", testTicketCreateGenerateUniqueIDs)
	registerRepoTest("TicketNoLeakage", testTicketNoLeakage)
	registerRepoTest("TicketUpdateNonExistent", testTicketUpdateNonExistent)
	registerRepoTest("TicketDeleteNonExistent", testTicketDeleteNonExistent)
	registerRepoTest("TicketCRUD", testTicketCRUD)
	registerRepoTest("TicketCreateDuplicate", testTicketCreateDuplicate)
}

func testTicketCRUD(t *testing.T, repo app.Repository) {
	if len(repo.Tickets()) != 0 {
		t.Fatalf("expected 0 tickets: got %d", len(repo.Tickets()))
	}

	t1 := createTicket(t, repo, "i am the first ticket")
	_ = createTicket(t, repo, "i am the second ticket")

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

// TODO: Characterisation test - memory implementation is limited to 8192 tickets
func testTicketCreateGenerateUniqueIDs(t *testing.T, repo app.Repository) {
	seen := make(map[int]struct{})
	const n = 8192
	user := app.NewUser("1")
	for i := range n {
		num := fmt.Sprintf("%d", i)
		tk, err := repo.CreateTicket(app.NewTicket(num, "whatever-"+num, user))
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
	newTicket1 := app.NewTicket("1", "whatever", "1")
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

func testTicketUpdateNonExistent(t *testing.T, repo app.Repository) {
	err := repo.UpdateTicket(app.Ticket{ID: 9999})
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
	t1 := createTicket(t, repo, "1")
	_ = createTicket(t, repo, "2")

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
