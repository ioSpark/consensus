package test

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

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

	ticket, err := repo.CreateTicket(name, "http://whatever"+name, user)
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
	registerRepoTest("TicketDeleteNonExistent", testTicketDeleteNonExistent)
	registerRepoTest("TicketCRUD", testTicketCRUD)
	registerRepoTest("TicketCreateDuplicate", testTicketCreateDuplicate)
	registerRepoTest("TicketTicketTime", testTicketTime)
	registerRepoTest("TicketReveal", testTicketReveal)
	registerRepoTest("TicketRevealNoVote", testTicketRevealNoVote)
	registerRepoTest("TicketRevealBadUser", testTicketRevealBadUser)
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
	updated, err := repo.UpdateTicket(t1.ID, newName, t1.Link)
	if err != nil {
		t.Fatalf("error updating ticket with new name: %v", err)
	}
	if updated.Name != newName {
		t.Fatalf(
			"expected updated ticket to have newName %s, got %s",
			newName,
			updated.Name,
		)
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
	_, err := repo.CreateTicket("test", "test", app.NewUser("1"))
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
		tk, err := repo.CreateTicket(num, "whatever-"+num, *user)
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

	name := "initial"
	link := "whatever"

	_, err := repo.CreateTicket(name, link, *user)
	if err != nil {
		t.Errorf("CreateTicket failed: %v", err)
	}

	_, err = repo.CreateTicket(name, link, *user)
	if err == nil {
		t.Errorf("expected duplicate ticket creation to fail")
	} else if !errors.Is(err, app.ErrTicketAlreadyExists) {
		t.Fatalf("expected ErrTicketAlreadyExists, got %v", err)
	}

	_, err = repo.CreateTicket("never before seen name", link, *user)
	if err == nil {
		t.Errorf("expected duplicate ticket creation to fail")
	} else if !errors.Is(err, app.ErrTicketAlreadyExists) {
		t.Fatalf("expected ErrTicketAlreadyExists, got %v", err)
	}

	_, err = repo.CreateTicket(name, "never before seen link", *user)
	if err == nil {
		t.Errorf("expected duplicate ticket creation to fail")
	} else if !errors.Is(err, app.ErrTicketAlreadyExists) {
		t.Fatalf("expected ErrTicketAlreadyExists, got %v", err)
	}
}

func testTicketUpdateNonExistent(t *testing.T, repo app.Repository) {
	_, err := repo.UpdateTicket(9999, "whatever", "whatever")
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

func testTicketTime(t *testing.T, repo app.Repository) {
	user := createUser(t, repo, "clock")

	ticket := createTicket(t, repo, "whatever", *user)
	if ticket.CreatedAt.Before(time.Now().Add(-1 * time.Minute)) {
		t.Errorf("CreatedAt time is more than a minute old: %v", ticket.CreatedAt)
	}

	_, err := repo.Vote(ticket.ID, *user, 1)
	if err != nil {
		t.Fatalf("could not vote: %v", err)
	}

	revealed, err := repo.Reveal(ticket.ID, *user)
	if err != nil {
		t.Fatalf("could not reveal ticket: %v", err)
	}

	if ticket.CreatedAt != revealed.CreatedAt {
		t.Errorf(
			"CreatedAt time differs after revealing, %v, %v",
			ticket.CreatedAt,
			revealed.CreatedAt,
		)
	}

	if revealed.RevealedAt.Before(time.Now().Add(-1 * time.Minute)) {
		t.Errorf("RevealedAt time is more than a minute old: %v", ticket.RevealedAt)
	}
}

func testTicketReveal(t *testing.T, repo app.Repository) {
	user := createUser(t, repo, "reveal")
	ticket := createTicket(t, repo, "reveal", *user)

	_, err := repo.Vote(ticket.ID, *user, 2)
	if err != nil {
		t.Fatalf("could not vote: %v", err)
	}
	revealedTicket, err := repo.Reveal(ticket.ID, *user)
	if err != nil {
		t.Fatalf("reveal ticket: %v", err)
	}
	if !revealedTicket.Revealed {
		t.Errorf("revealed ticket returned as not revealed")
	}

	renewedTicket, err := repo.Ticket(ticket.ID)
	if err != nil {
		t.Errorf("retrieve ticket: %v", err)
	}

	if !renewedTicket.Revealed {
		t.Errorf("ticket was revealed but repo returned: %v", renewedTicket.Revealed)
	}
}

func testTicketRevealNoVote(t *testing.T, repo app.Repository) {
	user := createUser(t, repo, "reveal")
	ticket := createTicket(t, repo, "reveal", *user)

	_, err := repo.Reveal(ticket.ID, *user)
	if err == nil {
		t.Fatalf("reveal with no votes did not fail")
	} else if !errors.Is(err, app.ErrCantRevealNoVotes) {
		t.Errorf("expected ErrCantRevealNoVotes, got %v", err)
	}

	// Verify it works after voting
	_, err = repo.Vote(ticket.ID, *user, 2)
	if err != nil {
		t.Errorf("could not reveal even after voting")
	}
}

func testTicketRevealBadUser(t *testing.T, repo app.Repository) {
	user := createUser(t, repo, "reveal")
	badUser := createUser(t, repo, "bad-user")
	ticket := createTicket(t, repo, "reveal", *user)

	_, err := repo.Reveal(ticket.ID, *badUser)
	if err == nil {
		t.Fatalf("reveal with non-reporter user did not fail")
	} else if !errors.Is(err, app.ErrUserCantReveal) {
		t.Errorf("expected ErrUserCantReveal, got %v", err)
	}

	// Verify that reveal works as expected
	_, err = repo.Vote(ticket.ID, *user, 2)
	if err != nil {
		t.Fatalf("vote failed: %v", err)
	}
	_, err = repo.Reveal(ticket.ID, *user)
	if err != nil {
		t.Errorf("could not reveal even after voting")
	}
}
